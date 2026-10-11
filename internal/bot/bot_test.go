package bot

import (
	"testing"

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

	if err := b.NewGame(); err != nil {
		t.Fatal(err)
	}
	const start = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	for level := MinLevel; level <= MaxLevel; level++ {
		m, err := b.Move(start, level)
		if err != nil {
			t.Fatalf("level %d: %v", level, err)
		}
		if len(m) < 4 {
			t.Fatalf("level %d: bad move %q", level, m)
		}
	}
}

func TestBotRecoversFromEngineCrash(t *testing.T) {
	b, err := New()
	if err == engine.ErrNotFound {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	b.eng.Close() // simulate the engine process dying
	const start = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if _, err := b.Move(start, 1); err != nil {
		t.Fatalf("bot did not recover: %v", err)
	}
}
