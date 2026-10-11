package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s := Open(path)
	if got := s.Load(); got != Default() {
		t.Fatalf("missing file: %+v", got)
	}
	want := Settings{Board: "blue", Pieces: "merida", Sound: false, Volume: 30}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	if got := Open(path).Load(); got != want {
		t.Fatalf("round trip: %+v", got)
	}
	os.WriteFile(path, []byte(`{"board":"green"}`), 0o644)
	// Fields missing from an older file keep their defaults.
	if got := s.Load(); got != (Settings{Board: "green", Pieces: "cburnett", Sound: true, Volume: 70}) {
		t.Fatalf("partial file: %+v", got)
	}
	os.WriteFile(path, []byte(`not json`), 0o644)
	if got := s.Load(); got != Default() {
		t.Fatalf("corrupt file: %+v", got)
	}
}
