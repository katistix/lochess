package bot

import (
	"context"
	"testing"
	"time"

	"lochess/internal/engine"
)

func TestBotPlaysEveryLevel(t *testing.T) {
	b, err := New()
	if err == engine.ErrNotFound {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := b.NewGame(ctx); err != nil {
		t.Fatal(err)
	}
	const start = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	for level := MinLevel; level <= MaxLevel; level++ {
		m, err := b.Move(ctx, start, level)
		if err != nil {
			t.Fatalf("level %d: %v", level, err)
		}
		if len(m) < 4 {
			t.Fatalf("level %d: bad move %q", level, m)
		}
	}
}
