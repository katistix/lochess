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

// Move plays a move in UCI notation (e.g. "e2e4", "e7e8q").
// A pawn reaching the last rank without a promotion suffix promotes to a queen.
func (g *Game) Move(uci string) (State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(uci) == 4 && g.isPromotion(uci) {
		uci += "q"
	}
	m, err := chess.UCINotation{}.Decode(g.g.Position(), uci)
	if err != nil {
		return g.state(), fmt.Errorf("invalid move %q: %w", uci, err)
	}
	if err := g.g.Move(m); err != nil {
		return g.state(), fmt.Errorf("illegal move %q: %w", uci, err)
	}
	return g.state(), nil
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
		FEN:      pos.String(),
		Turn:     "white",
		Dests:    map[string][]string{},
		LastMove: []string{},
		Outcome:  string(g.g.Outcome()),
	}
	if pos.Turn() == chess.Black {
		s.Turn = "black"
	}
	if s.Outcome == string(chess.NoOutcome) {
		for _, m := range g.g.ValidMoves() {
			from := m.S1().String()
			to := m.S2().String()
			// Promotions appear once per piece type; the board only needs the squares.
			if m.Promo() != chess.NoPieceType && m.Promo() != chess.Queen {
				continue
			}
			s.Dests[from] = append(s.Dests[from], to)
		}
	} else {
		s.Method = g.g.Method().String()
	}
	if moves := g.g.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		s.LastMove = []string{last.S1().String(), last.S2().String()}
		s.Check = last.HasTag(chess.Check)
	}
	return s
}
