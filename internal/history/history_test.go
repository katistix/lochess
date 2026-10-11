package history

import (
	"errors"
	"testing"
	"time"

	"lochess/internal/game"
)

func TestStore(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	older := Game{ID: NewID(t0), PlayedAt: t0, Mode: "bot", Level: 2, Result: "0-1", Plies: []game.Ply{{UCI: "e2e4"}}}
	newer := Game{ID: NewID(t0.Add(time.Hour)), PlayedAt: t0.Add(time.Hour), Mode: "local", Result: "1-0"}
	for _, g := range []Game{older, newer} {
		if err := s.Save(g); err != nil {
			t.Fatal(err)
		}
	}

	list, err := s.List()
	if err != nil || len(list) != 2 || list[0].ID != newer.ID || list[1].Plies != 1 {
		t.Fatalf("list: %v %+v", err, list)
	}
	if g, err := s.Get(older.ID); err != nil || g.Level != 2 {
		t.Fatalf("get: %v %+v", err, g)
	}
	if err := s.Delete(older.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(older.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
	if _, err := s.Get("../secret"); err == nil {
		t.Fatal("path traversal accepted")
	}
}
