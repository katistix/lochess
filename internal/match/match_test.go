package match

import (
	"errors"
	"testing"

	"lochess/internal/bot"
)

func TestBotGame(t *testing.T) {
	m := New()
	defer m.Close()
	s, err := m.Start(Options{Mode: ModeBot, Level: 1, Color: "black"})
	if errors.Is(err, bot.ErrEngineNotFound) {
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
	m := New()
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

func TestLocalNeverBot(t *testing.T) {
	m := New()
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
