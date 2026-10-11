package match

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"lochess/internal/engine"
	"lochess/internal/history"
)

func TestBotGame(t *testing.T) {
	m := New(nil, "")
	defer m.Close()
	s, err := m.Start(Options{Mode: ModeBot, Level: 1, Color: "black"})
	if errors.Is(err, engine.ErrNotFound) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if !s.BotToMove || s.HumanColor != "black" {
		t.Fatalf("bot should open as white: %+v", s)
	}
	if _, err := m.Move("e7e5"); err == nil {
		t.Fatal("human moved on bot's turn")
	}
	if s, err = m.BotMove(); err != nil {
		t.Fatal(err)
	}
	if s.BotToMove || s.Game.Turn != "black" {
		t.Fatalf("after bot move: %+v", s)
	}
	// A second call on the human's turn is a no-op.
	if s2, _ := m.BotMove(); s2.Game.FEN != s.Game.FEN {
		t.Fatal("duplicate BotMove changed the position")
	}
}

func TestPhases(t *testing.T) {
	m := New(nil, "")
	if s := m.State(); s.Phase != PhaseSetup {
		t.Fatalf("new match: %s", s.Phase)
	}
	if _, err := m.Move("e2e4"); err == nil {
		t.Fatal("moved during setup")
	}
	if s, _ := m.Start(Options{Mode: ModeLocal}); s.Phase != PhasePlaying {
		t.Fatalf("after start: %s", s.Phase)
	}
	m.Move("e2e4")
	s, err := m.Resign()
	if err != nil || s.Phase != PhaseOver || s.Game.Outcome != "1-0" {
		t.Fatalf("black resigns: %v %+v", err, s)
	}
	if s := m.Setup(); s.Phase != PhaseSetup || len(s.Game.Moves) != 0 {
		t.Fatalf("back to setup: %+v", s)
	}
}

func TestSavesFinishedGame(t *testing.T) {
	var saved []history.Game
	m := New(func(g history.Game) error { saved = append(saved, g); return nil }, "")

	m.Start(Options{Mode: ModeLocal})
	m.Resign() // no moves: aborted, not saved
	if len(saved) != 0 {
		t.Fatal("saved an aborted game")
	}

	m.Start(Options{Mode: ModeLocal})
	var s State
	for _, mv := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		s, _ = m.Move(mv)
	}
	if len(saved) != 1 || s.GameID != saved[0].ID || saved[0].Result != "0-1" || len(saved[0].Plies) != 4 {
		t.Fatalf("saved %+v, state id %q", saved, s.GameID)
	}
	m.State()
	if len(saved) != 1 {
		t.Fatal("saved twice")
	}
}

func TestLocalNeverBot(t *testing.T) {
	m := New(nil, "")
	s, err := m.Start(Options{Mode: ModeLocal})
	if err != nil || s.BotToMove {
		t.Fatalf("%v %+v", err, s)
	}
	if _, err := m.Move("e2e4"); err != nil {
		t.Fatal(err)
	}
	if s := m.State(); s.BotToMove {
		t.Fatal("local mode asked for bot move")
	}
}

func TestResumeAfterRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.json")
	m := New(nil, path)
	m.Start(Options{Mode: ModeLocal})
	m.Move("e2e4")
	m.Move("e7e5")

	resumed := New(nil, path)
	s := resumed.State()
	if s.Phase != PhasePlaying || strings.Join(s.Game.Moves, " ") != "e4 e5" {
		t.Fatalf("resumed: %+v", s)
	}
	resumed.Resign()
	if s := New(nil, path).State(); s.Phase != PhaseSetup {
		t.Fatalf("finished game was resumed: %+v", s)
	}
}

func TestTakebackAndDraw(t *testing.T) {
	m := New(nil, "")
	m.Start(Options{Mode: ModeLocal})
	if _, err := m.Takeback(); err == nil {
		t.Fatal("took back with no moves")
	}
	m.Move("e2e4")
	if s, err := m.Takeback(); err != nil || len(s.Game.Moves) != 0 {
		t.Fatalf("local takeback: %v %+v", err, s)
	}
	s, err := m.AgreeDraw()
	if err != nil || s.Phase != PhaseOver || s.Game.Outcome != "1/2-1/2" {
		t.Fatalf("draw: %v %+v", err, s)
	}
}
