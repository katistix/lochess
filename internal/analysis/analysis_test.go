package analysis

import (
	"context"
	"errors"
	"testing"
	"time"

	"lochess/internal/engine"
	"lochess/internal/game"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		loss float64
		best bool
		want string
	}{
		{0, true, Best}, {1, false, Excellent}, {3, false, Good},
		{7, false, Inaccuracy}, {15, false, Mistake}, {40, false, Blunder},
	}
	for _, c := range cases {
		if got := classify(c.loss, c.best); got != c.want {
			t.Errorf("classify(%v, %v) = %s, want %s", c.loss, c.best, got, c.want)
		}
	}
}

func TestWinChance(t *testing.T) {
	if w := winChance(0); w != 50 {
		t.Fatalf("equal position: %v", w)
	}
	if w := winChance(300); w < 70 || w > 80 {
		t.Fatalf("+3 should be ~75%%: %v", w)
	}
}

// Fool's mate: White's g4 lets mate happen; it must be graded a blunder.
func TestRunFoolsMate(t *testing.T) {
	path, err := engine.Find()
	if errors.Is(err, engine.ErrNotFound) {
		t.Skip(err)
	}
	eng, err := engine.Start(path)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()

	g := game.New()
	for _, m := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		if _, err := g.Move(m); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	a, err := Run(ctx, eng, g.Plies(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Evals) != 5 || len(a.Moves) != 4 {
		t.Fatalf("sizes: %d evals, %d moves", len(a.Evals), len(a.Moves))
	}
	if c := a.Moves[2].Class; c != Blunder {
		t.Errorf("g4 graded %s, want blunder", c)
	}
	if c := a.Moves[3].Class; c != Best {
		t.Errorf("Qh4# graded %s, want best", c)
	}
	if txt := a.Evals[4].Text; txt != "0-1" {
		t.Errorf("final eval %q, want 0-1", txt)
	}
	if km := a.FindKeyMoments([]string{"white"}, 3); len(km) == 0 || km[len(km)-1] != 2 {
		t.Errorf("key moments %v should include g4 (ply 2)", km)
	}
}
