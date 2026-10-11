// Package game owns the authoritative chess game state.
// The frontend only renders a State and sends moves back.
package game

import (
	"fmt"
	"sync"

	"github.com/notnil/chess"
)

// State is a snapshot of the game for the frontend to render.
type State struct {
	FEN      string              `json:"fen"`
	Turn     string              `json:"turn"`     // "white" | "black"
	Dests    map[string][]string `json:"dests"`    // legal moves: from square -> to squares
	LastMove []string            `json:"lastMove"` // [from, to], empty at start
	Check    bool                `json:"check"`
	Outcome  string              `json:"outcome"` // "*", "1-0", "0-1", "1/2-1/2"
	Method   string              `json:"method"`  // e.g. "Checkmate", "Stalemate"; empty while ongoing
	Moves    []string            `json:"moves"`   // move history in SAN, e.g. ["e4", "e5", "Nf3"]
	// Promotions lists from+to pairs (e.g. "e7e8") that promote a pawn, so the
	// frontend knows when to ask which piece to promote to.
	Promotions []string `json:"promotions"`
}

// Ply is one half-move of a finished or ongoing game.
type Ply struct {
	UCI   string `json:"uci"`
	SAN   string `json:"san"`
	FEN   string `json:"fen"` // position after the move
	Check bool   `json:"check"`
}

// StartFEN is the standard initial position.
const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// methodNames are the display names for how a game ended.
var methodNames = map[chess.Method]string{
	chess.Checkmate:            "Checkmate",
	chess.Resignation:          "Resignation",
	chess.DrawOffer:            "Agreement",
	chess.Stalemate:            "Stalemate",
	chess.ThreefoldRepetition:  "Threefold repetition",
	chess.FivefoldRepetition:   "Fivefold repetition",
	chess.FiftyMoveRule:        "50-move rule",
	chess.SeventyFiveMoveRule:  "75-move rule",
	chess.InsufficientMaterial: "Insufficient material",
}

// Game is safe for concurrent use; Wails may call bound methods from multiple goroutines.
type Game struct {
	mu sync.Mutex
	g  *chess.Game
}

func New() *Game {
	return &Game{g: chess.NewGame()}
}

// Reset starts a new game from the initial position.
func (g *Game) Reset() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.g = chess.NewGame()
	return g.state()
}

func (g *Game) State() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.state()
}

// Resign ends the game with color ("white" | "black") resigning.
func (g *Game) Resign(color string) State {
	g.mu.Lock()
	defer g.mu.Unlock()
	c := chess.White
	if color == "black" {
		c = chess.Black
	}
	g.g.Resign(c)
	return g.state()
}

// Move plays a move in UCI notation (e.g. "e2e4", "e7e8q").
// A pawn reaching the last rank without a promotion suffix promotes to a queen.
func (g *Game) Move(uci string) (State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.move(uci); err != nil {
		return g.state(), err
	}
	return g.state(), nil
}

// Load replaces the game with a new one built from UCI moves.
func (g *Game) Load(moves []string) (State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.replay(moves)
}

// Takeback undoes the last n half-moves.
func (g *Game) Takeback(n int) (State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	moves := g.g.Moves()
	if n < 1 || n > len(moves) {
		return g.state(), fmt.Errorf("cannot take back %d of %d moves", n, len(moves))
	}
	uci := make([]string, len(moves)-n)
	for i := range uci {
		uci[i] = moves[i].String()
	}
	return g.replay(uci)
}

// AgreeDraw ends an ongoing game as a draw by agreement.
func (g *Game) AgreeDraw() (State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.g.Draw(chess.DrawOffer); err != nil {
		return g.state(), err
	}
	return g.state(), nil
}

// replay rebuilds the game from scratch. On error the previous game is kept.
func (g *Game) replay(moves []string) (State, error) {
	prev := g.g
	g.g = chess.NewGame()
	for _, uci := range moves {
		if err := g.move(uci); err != nil {
			g.g = prev
			return g.state(), err
		}
	}
	return g.state(), nil
}

func (g *Game) move(uci string) error {
	if len(uci) == 4 && g.isPromotion(uci) {
		uci += "q"
	}
	m, err := chess.UCINotation{}.Decode(g.g.Position(), uci)
	if err != nil {
		return fmt.Errorf("invalid move %q: %w", uci, err)
	}
	if err := g.g.Move(m); err != nil {
		return fmt.Errorf("illegal move %q: %w", uci, err)
	}
	// The library only ends games automatically at fivefold repetition and the
	// 75-move rule; like lichess, end them at threefold and 50 moves instead.
	for _, method := range g.g.EligibleDraws() {
		if method == chess.ThreefoldRepetition || method == chess.FiftyMoveRule {
			return g.g.Draw(method)
		}
	}
	return nil
}

// Plies returns every move played so far with the position after it.
func (g *Game) Plies() []Ply {
	g.mu.Lock()
	defer g.mu.Unlock()
	moves := g.g.Moves()
	positions := g.g.Positions()
	san := g.san()
	out := make([]Ply, len(moves))
	for i, m := range moves {
		out[i] = Ply{UCI: m.String(), SAN: san[i], FEN: positions[i+1].String(), Check: m.HasTag(chess.Check)}
	}
	return out
}

// san returns the move history in standard algebraic notation.
func (g *Game) san() []string {
	moves := g.g.Moves()
	positions := g.g.Positions() // positions[i] is the position before moves[i]
	out := make([]string, len(moves))
	for i, m := range moves {
		out[i] = chess.AlgebraicNotation{}.Encode(positions[i], m)
	}
	return out
}

func (g *Game) isPromotion(uci string) bool {
	for _, m := range g.g.ValidMoves() {
		if m.Promo() != chess.NoPieceType && m.S1().String()+m.S2().String() == uci {
			return true
		}
	}
	return false
}

func (g *Game) state() State {
	pos := g.g.Position()
	s := State{
		FEN:        pos.String(),
		Turn:       "white",
		Dests:      map[string][]string{},
		LastMove:   []string{},
		Outcome:    string(g.g.Outcome()),
		Moves:      g.san(),
		Promotions: []string{},
	}
	if pos.Turn() == chess.Black {
		s.Turn = "black"
	}
	if s.Outcome == string(chess.NoOutcome) {
		for _, m := range g.g.ValidMoves() {
			from := m.S1().String()
			to := m.S2().String()
			// Promotions appear once per piece type; the board only needs the squares.
			if m.Promo() != chess.NoPieceType {
				if m.Promo() != chess.Queen {
					continue
				}
				s.Promotions = append(s.Promotions, from+to)
			}
			s.Dests[from] = append(s.Dests[from], to)
		}
	} else {
		s.Method = methodNames[g.g.Method()]
	}
	if moves := g.g.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		s.LastMove = []string{last.S1().String(), last.S2().String()}
		s.Check = last.HasTag(chess.Check)
	}
	return s
}
