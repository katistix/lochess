// Package match ties a game to its mode (local or vs bot) and settings.
package match

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"lochess/internal/bot"
	"lochess/internal/game"
)

const (
	ModeLocal = "local" // two humans on one board
	ModeBot   = "bot"   // human vs engine
)

// Phases of the play screen.
const (
	PhaseSetup   = "setup"   // choosing options; board at the start position
	PhasePlaying = "playing" // game in progress
	PhaseOver    = "over"    // game finished; final position shown
)

// Options configure a new match.
type Options struct {
	Mode  string `json:"mode"`  // ModeLocal | ModeBot
	Level int    `json:"level"` // bot level 1-8, ignored for local
	Color string `json:"color"` // human side vs bot: "white" | "black" | "random"
}

// State is what the frontend renders.
type State struct {
	Game       game.State `json:"game"`
	Phase      string     `json:"phase"` // PhaseSetup | PhasePlaying | PhaseOver
	Mode       string     `json:"mode"`
	Level      int        `json:"level"`
	HumanColor string     `json:"humanColor"` // "white" | "black"; "white" in local mode
	BotToMove  bool       `json:"botToMove"`  // frontend should call BotMove
}

// Match is safe for concurrent use. The bot engine is started lazily on the
// first bot game so local play works without it installed.
type Match struct {
	mu         sync.Mutex
	game       *game.Game
	started    bool // false while in setup
	mode       string
	level      int
	humanColor string
	bot        *bot.Bot
}

func New() *Match {
	return &Match{game: game.New(), mode: ModeLocal, level: bot.MinLevel, humanColor: "white"}
}

func (m *Match) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state(m.game.State())
}

// Start begins a new game with the given options.
func (m *Match) Start(opts Options) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch opts.Mode {
	case ModeLocal:
		m.humanColor = "white"
	case ModeBot:
		if opts.Level < bot.MinLevel || opts.Level > bot.MaxLevel {
			return m.state(m.game.State()), fmt.Errorf("level %d out of range", opts.Level)
		}
		if m.bot == nil {
			b, err := bot.New()
			if err != nil {
				return m.state(m.game.State()), err
			}
			m.bot = b
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.bot.NewGame(ctx); err != nil {
			return m.state(m.game.State()), err
		}
		m.level = opts.Level
		m.humanColor = resolveColor(opts.Color)
	default:
		return m.state(m.game.State()), fmt.Errorf("unknown mode %q", opts.Mode)
	}
	m.mode = opts.Mode
	m.started = true
	return m.state(m.game.Reset()), nil
}

// Setup abandons any game and returns to the options screen.
func (m *Match) Setup() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = false
	return m.state(m.game.Reset())
}

// Resign ends the game. Against the bot the human resigns; in local play
// the side to move does.
func (m *Match) Resign() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.game.State()
	if m.phase(s) != PhasePlaying {
		return m.state(s), errors.New("no game in progress")
	}
	loser := s.Turn
	if m.mode == ModeBot {
		loser = m.humanColor
	}
	return m.state(m.game.Resign(loser)), nil
}

// Move plays a human move.
func (m *Match) Move(uci string) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur := m.game.State()
	if m.phase(cur) != PhasePlaying {
		return m.state(cur), errors.New("no game in progress")
	}
	if m.botToMove(cur) {
		return m.state(cur), errors.New("not your turn")
	}
	s, err := m.game.Move(uci)
	return m.state(s), err
}

// BotMove lets the bot play if it is its turn; otherwise it returns the
// current state unchanged, so duplicate calls are harmless.
func (m *Match) BotMove() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.game.State()
	if !m.botToMove(s) {
		return m.state(s), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uci, err := m.bot.Move(ctx, s.FEN, m.level)
	if err != nil {
		return m.state(s), fmt.Errorf("bot: %w", err)
	}
	s, err = m.game.Move(uci)
	return m.state(s), err
}

// Close stops the engine, if running.
func (m *Match) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bot != nil {
		m.bot.Close()
		m.bot = nil
	}
}

func (m *Match) phase(s game.State) string {
	switch {
	case !m.started:
		return PhaseSetup
	case s.Outcome != "*":
		return PhaseOver
	default:
		return PhasePlaying
	}
}

func (m *Match) botToMove(s game.State) bool {
	return m.phase(s) == PhasePlaying && m.mode == ModeBot && s.Turn != m.humanColor
}

func (m *Match) state(s game.State) State {
	return State{
		Game:       s,
		Phase:      m.phase(s),
		Mode:       m.mode,
		Level:      m.level,
		HumanColor: m.humanColor,
		BotToMove:  m.botToMove(s),
	}
}

func resolveColor(c string) string {
	switch c {
	case "white", "black":
		return c
	}
	if rand.IntN(2) == 0 {
		return "white"
	}
	return "black"
}
