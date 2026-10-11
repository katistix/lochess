package game

import (
	"strings"
	"testing"
)

func TestOpeningMoves(t *testing.T) {
	g := New()
	s := g.State()
	if s.Turn != "white" || len(s.Dests) != 10 {
		t.Fatalf("start: turn=%s pieces with moves=%d", s.Turn, len(s.Dests))
	}
	s, err := g.Move("e2e4")
	if err != nil {
		t.Fatal(err)
	}
	if s.Turn != "black" || s.LastMove[0] != "e2" || s.LastMove[1] != "e4" {
		t.Fatalf("after e4: %+v", s)
	}
	if _, err := g.Move("e2e4"); err == nil {
		t.Fatal("expected illegal move error")
	}
}

func TestFoolsMate(t *testing.T) {
	g := New()
	var s State
	for _, m := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		var err error
		if s, err = g.Move(m); err != nil {
			t.Fatal(err)
		}
	}
	if s.Outcome != "0-1" || s.Method != "Checkmate" || !s.Check || len(s.Dests) != 0 {
		t.Fatalf("fool's mate: %+v", s)
	}
	if got := strings.Join(s.Moves, " "); got != "f3 e5 g4 Qh4#" {
		t.Fatalf("SAN history: %q", got)
	}
}

func TestAutoQueen(t *testing.T) {
	g := New()
	for _, m := range []string{"a2a4", "b7b5", "a4b5", "a7a6", "b5a6", "c8b7", "a6b7", "h7h6", "b7a8"} {
		if _, err := g.Move(m); err != nil {
			t.Fatalf("%s: %v", m, err)
		}
	}
	if fen := g.State().FEN; fen[:2] != "Qn" {
		t.Fatalf("expected queen on a8, got %s", fen)
	}
}

func TestThreefoldRepetitionDraws(t *testing.T) {
	g := New()
	var s State
	// Knights out and back twice: the start position occurs a third time.
	for _, m := range []string{"g1f3", "g8f6", "f3g1", "f6g8", "g1f3", "g8f6", "f3g1", "f6g8"} {
		var err error
		if s, err = g.Move(m); err != nil {
			t.Fatal(err)
		}
	}
	if s.Outcome != "1/2-1/2" || s.Method != "Threefold repetition" {
		t.Fatalf("expected threefold draw, got %s %s", s.Outcome, s.Method)
	}
}

func TestPromotionChoice(t *testing.T) {
	g := New()
	for _, m := range []string{"a2a4", "b7b5", "a4b5", "a7a6", "b5a6", "c8b7", "a6b7", "h7h6"} {
		if _, err := g.Move(m); err != nil {
			t.Fatal(err)
		}
	}
	s := g.State()
	// b8 is blocked by the knight and c8 is empty, so only bxa8 promotes.
	if strings.Join(s.Promotions, ",") != "b7a8" {
		t.Fatalf("promotions: %v", s.Promotions)
	}
	if s, _ = g.Move("b7a8n"); s.FEN[:1] != "N" {
		t.Fatalf("expected knight on a8: %s", s.FEN)
	}
}

func TestTakebackAndLoad(t *testing.T) {
	g := New()
	g.Move("e2e4")
	g.Move("e7e5")
	s, err := g.Takeback(2)
	if err != nil || len(s.Moves) != 0 || s.Turn != "white" {
		t.Fatalf("takeback: %v %+v", err, s)
	}
	if _, err := g.Takeback(1); err == nil {
		t.Fatal("took back from empty game")
	}
	s, err = g.Load([]string{"d2d4", "d7d5"})
	if err != nil || strings.Join(s.Moves, " ") != "d4 d5" {
		t.Fatalf("load: %v %+v", err, s)
	}
	if _, err := g.Load([]string{"d2d4", "d2d4"}); err == nil {
		t.Fatal("loaded illegal moves")
	}
	if s := g.State(); strings.Join(s.Moves, " ") != "d4 d5" {
		t.Fatalf("failed load changed game: %v", s.Moves)
	}
}
