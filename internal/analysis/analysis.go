// Package analysis grades the moves of a finished game with an engine.
package analysis

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/notnil/chess"

	"lochess/internal/engine"
	"lochess/internal/game"
)

// Move classes, from best to worst.
const (
	Best       = "best"
	Excellent  = "excellent"
	Good       = "good"
	Inaccuracy = "inaccuracy"
	Mistake    = "mistake"
	Blunder    = "blunder"
)

// Eval is an engine evaluation of a position, from White's point of view.
type Eval struct {
	Win  float64 `json:"win"`  // White's winning chances, 0-100
	Text string  `json:"text"` // display form: "+1.2", "#3", "#-2", "1-0"
}

// MoveGrade grades one half-move.
type MoveGrade struct {
	Class   string  `json:"class"`
	Loss    float64 `json:"loss"`    // winning chances the mover gave away, 0-100
	BestUCI string  `json:"bestUci"` // engine's choice in the position before the move
	BestSAN string  `json:"bestSan"`
}

type Analysis struct {
	Evals         []Eval      `json:"evals"`         // one per position: start, then after each ply
	Moves         []MoveGrade `json:"moves"`         // one per ply
	WhiteAccuracy float64     `json:"whiteAccuracy"` // 0-100
	BlackAccuracy float64     `json:"blackAccuracy"`
	KeyMoments    []int       `json:"keyMoments"` // plies worth reviewing first, in game order
}

// SearchArgs is the per-position search. Shallow but far stronger than the
// players this app is for; keeps a full game under ~15 seconds.
const SearchArgs = "depth 14 movetime 400"

// searchTimeout bounds one position, so a hung engine fails instead of
// stalling the review forever.
const searchTimeout = 5 * time.Second

// Run analyses every position of the game. progress is called after each one.
func Run(ctx context.Context, eng *engine.Engine, plies []game.Ply, progress func(done, total int)) (*Analysis, error) {
	fens := make([]string, 0, len(plies)+1)
	fens = append(fens, game.StartFEN)
	for _, p := range plies {
		fens = append(fens, p.FEN)
	}

	evals := make([]Eval, len(fens))
	best := make([]string, len(fens))
	for i, fen := range fens {
		sctx, cancel := context.WithTimeout(ctx, searchTimeout)
		info, err := eng.Search(sctx, fen, SearchArgs)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("position %d: %w", i, err)
		}
		evals[i] = toEval(info, whiteToMove(fen))
		best[i] = info.BestMove
		if progress != nil {
			progress(i+1, len(fens))
		}
	}

	a := &Analysis{Evals: evals, Moves: make([]MoveGrade, len(plies))}
	var whiteAcc, blackAcc []float64
	for i, p := range plies {
		white := whiteToMove(fens[i])
		before, after := evals[i].Win, evals[i+1].Win
		if !white {
			before, after = 100-before, 100-after
		}
		loss := math.Max(0, before-after)
		a.Moves[i] = MoveGrade{
			Class:   classify(loss, p.UCI == best[i]),
			Loss:    loss,
			BestUCI: best[i],
			BestSAN: toSAN(fens[i], best[i]),
		}
		if white {
			whiteAcc = append(whiteAcc, moveAccuracy(loss))
		} else {
			blackAcc = append(blackAcc, moveAccuracy(loss))
		}
	}
	a.WhiteAccuracy = mean(whiteAcc)
	a.BlackAccuracy = mean(blackAcc)
	return a, nil
}

// FindKeyMoments returns up to n plies, in game order, where a player of the
// given colors ("white", "black") lost the most. Only real errors qualify.
func (a *Analysis) FindKeyMoments(colors []string, n int) []int {
	cands := []int{}
	for i, m := range a.Moves {
		color := "white"
		if i%2 == 1 {
			color = "black"
		}
		if m.Loss >= 5 && contains(colors, color) {
			cands = append(cands, i)
		}
	}
	sort.SliceStable(cands, func(x, y int) bool { return a.Moves[cands[x]].Loss > a.Moves[cands[y]].Loss })
	if len(cands) > n {
		cands = cands[:n]
	}
	sort.Ints(cands)
	return cands
}

// classify uses chess.com's expected-points thresholds, in percent.
func classify(loss float64, isBest bool) string {
	switch {
	case isBest:
		return Best
	case loss < 2:
		return Excellent
	case loss < 5:
		return Good
	case loss < 10:
		return Inaccuracy
	case loss < 20:
		return Mistake
	default:
		return Blunder
	}
}

// winChance converts centipawns (side to move) to winning chances, 0-100.
// Same curve lichess uses.
func winChance(cp int) float64 {
	return 50 + 50*(2/(1+math.Exp(-0.00368208*float64(cp)))-1)
}

// moveAccuracy is lichess's per-move accuracy from winning chances lost.
func moveAccuracy(loss float64) float64 {
	return math.Max(0, math.Min(100, 103.1668*math.Exp(-0.04354*loss)-3.1669))
}

func toEval(info engine.Info, whiteMoves bool) Eval {
	sign := 1
	if !whiteMoves {
		sign = -1
	}
	var e Eval
	switch {
	case info.HasMate && info.Mate == 0: // side to move is checkmated
		e.Win = 0
		e.Text = map[bool]string{true: "0-1", false: "1-0"}[whiteMoves]
	case info.HasMate:
		e.Win = 0
		if info.Mate > 0 {
			e.Win = 100
		}
		e.Text = fmt.Sprintf("#%d", info.Mate*sign)
	case info.BestMove == "": // stalemate
		e.Win = 50
		e.Text = "½-½"
	default:
		e.Win = winChance(info.CP)
		e.Text = fmt.Sprintf("%+.1f", float64(info.CP*sign)/100)
	}
	if !whiteMoves {
		e.Win = 100 - e.Win
	}
	return e
}

func toSAN(fen, uci string) string {
	if uci == "" {
		return ""
	}
	f, err := chess.FEN(fen)
	if err != nil {
		return uci
	}
	pos := chess.NewGame(f).Position()
	m, err := chess.UCINotation{}.Decode(pos, uci)
	if err != nil {
		return uci
	}
	return chess.AlgebraicNotation{}.Encode(pos, m)
}

func whiteToMove(fen string) bool {
	f := strings.Fields(fen)
	return len(f) < 2 || f[1] == "w"
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
