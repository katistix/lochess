// Package match ties a game to its mode (local or vs bot) and settings.
package match

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"lochess/internal/bot"
	"lochess/internal/game"
	"lochess/internal/history"
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
	GameID     string     `json:"gameId"`     // history ID once a finished game is saved
}

// snapshot is an in-progress game, persisted so it survives an app restart.
type snapshot struct {
	Mode       string   `json:"mode"`
	Level      int      `json:"level"`
	HumanColor string   `json:"humanColor"`
	Moves      []string `json:"moves"` // UCI
}

// Match is safe for concurrent use. The bot engine is started lazily on the
// first bot move so local play works without it installed.
type Match struct {
	mu           sync.Mutex
	game         *game.Game
	started      bool // false while in setup
	mode         string
	level        int
	humanColor   string
	bot          *bot.Bot
	save         func(history.Game) error // called once when a game ends; may be nil
	savedID      string
	snapshotPath string // where the in-progress game is kept; "" disables
}

// botDelay is the least time a bot move takes, so its replies don't land
// instantly and are easy to follow. A random extra is added on top.
const botDelay = 400 * time.Millisecond

// New creates a match. save, if non-nil, stores each finished game.
// snapshotPath, if set, persists the game in progress and restores it.
func New(save func(history.Game) error, snapshotPath string) *Match {
	m := &Match{game: game.New(), mode: ModeLocal, level: bot.MinLevel, humanColor: "white", save: save, snapshotPath: snapshotPath}
	m.restore()
	return m
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
		if err := m.ensureBot(); err != nil {
			return m.state(m.game.State()), err
		}
		if err := m.bot.NewGame(); err != nil {
			return m.state(m.game.State()), err
		}
		m.level = opts.Level
		m.humanColor = resolveColor(opts.Color)
	default:
		return m.state(m.game.State()), fmt.Errorf("unknown mode %q", opts.Mode)
	}
	m.mode = opts.Mode
	m.started = true
	m.savedID = ""
	return m.update(m.game.Reset()), nil
}

// Setup abandons any game and returns to the options screen.
func (m *Match) Setup() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = false
	m.savedID = ""
	return m.update(m.game.Reset())
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
	return m.update(m.game.Resign(loser)), nil
}

// AgreeDraw ends a local game as a draw. The bot does not take draw offers.
func (m *Match) AgreeDraw() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.game.State()
	if m.phase(s) != PhasePlaying || m.mode != ModeLocal {
		return m.state(s), errors.New("draws can only be agreed in a local game")
	}
	s, err := m.game.AgreeDraw()
	return m.update(s), err
}

// Takeback undoes the human's last move. Against the bot that also undoes
// the bot's reply; in local play it undoes one move.
func (m *Match) Takeback() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.game.State()
	if m.phase(s) != PhasePlaying {
		return m.state(s), errors.New("no game in progress")
	}
	if m.botToMove(s) {
		return m.state(s), errors.New("wait for the bot to move")
	}
	n := 1
	if m.mode == ModeBot {
		n = 2
	}
	if len(s.Moves) < n {
		return m.state(s), errors.New("nothing to take back")
	}
	s, err := m.game.Takeback(n)
	return m.update(s), err
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
	return m.update(s), err
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
	if err := m.ensureBot(); err != nil {
		return m.state(s), err
	}
	begin := time.Now()
	uci, err := m.bot.Move(s.FEN, m.level)
	if err != nil {
		return m.state(s), fmt.Errorf("bot: %w", err)
	}
	time.Sleep(botDelay + rand.N(400*time.Millisecond) - time.Since(begin))
	s, err = m.game.Move(uci)
	return m.update(s), err
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

func (m *Match) ensureBot() error {
	if m.bot != nil {
		return nil
	}
	b, err := bot.New()
	if err != nil {
		return err
	}
	m.bot = b
	return nil
}

// update records a state change: it saves the game to history the first
// time it is seen over and keeps the in-progress snapshot current.
func (m *Match) update(s game.State) State {
	phase := m.phase(s)
	// Games with no moves are aborted, not saved.
	if phase == PhaseOver && m.savedID == "" && len(s.Moves) > 0 && m.save != nil {
		now := time.Now()
		g := history.Game{
			ID:         history.NewID(now),
			PlayedAt:   now,
			Mode:       m.mode,
			HumanColor: m.humanColor,
			Result:     s.Outcome,
			Method:     s.Method,
			Plies:      m.game.Plies(),
		}
		if m.mode == ModeBot {
			g.Level = m.level
		}
		if err := m.save(g); err != nil {
			log.Printf("save game: %v", err)
		} else {
			m.savedID = g.ID
		}
	}
	m.persist(phase)
	return m.state(s)
}

// persist writes the game in progress, or removes the snapshot when there is none.
func (m *Match) persist(phase string) {
	if m.snapshotPath == "" {
		return
	}
	if phase != PhasePlaying {
		if err := os.Remove(m.snapshotPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("clear game snapshot: %v", err)
		}
		return
	}
	snap := snapshot{Mode: m.mode, Level: m.level, HumanColor: m.humanColor, Moves: []string{}}
	for _, p := range m.game.Plies() {
		snap.Moves = append(snap.Moves, p.UCI)
	}
	data, err := json.Marshal(snap)
	if err == nil {
		tmp := m.snapshotPath + ".tmp"
		if err = os.WriteFile(tmp, data, 0o644); err == nil {
			err = os.Rename(tmp, m.snapshotPath)
		}
	}
	if err != nil {
		log.Printf("save game snapshot: %v", err)
	}
}

// restore resumes a game left in progress when the app last closed.
func (m *Match) restore() {
	if m.snapshotPath == "" {
		return
	}
	data, err := os.ReadFile(m.snapshotPath)
	if err != nil {
		return
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		log.Printf("discard game snapshot: %v", err)
		return
	}
	if snap.Mode != ModeLocal && snap.Mode != ModeBot {
		return
	}
	if _, err := m.game.Load(snap.Moves); err != nil {
		log.Printf("discard game snapshot: %v", err)
		return
	}
	m.mode, m.level, m.humanColor, m.started = snap.Mode, snap.Level, snap.HumanColor, true
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
		GameID:     m.savedID,
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
