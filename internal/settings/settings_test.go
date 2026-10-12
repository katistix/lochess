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
	want := Settings{Board: "blue", Pieces: "merida", Sound: false, Volume: 30, Mode: "bot", Level: 4, Color: "random"}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	if got := Open(path).Load(); got != want {
		t.Fatalf("round trip: %+v", got)
	}
	os.WriteFile(path, []byte(`{"board":"green"}`), 0o644)
	// Fields missing from an older file keep their defaults.
	if got := s.Load(); got != (Settings{Board: "green", Pieces: "cburnett", Sound: true, Volume: 70, Mode: "bot", Level: 1, Color: "white"}) {
		t.Fatalf("partial file: %+v", got)
	}
	os.WriteFile(path, []byte(`{"board":"green","pieces":"merida","mode":"chess960","level":99,"color":"purple"}`), 0o644)
	// Invalid game setup values fall back to defaults.
	if got := s.Load(); got.Mode != "bot" || got.Level != 1 || got.Color != "white" {
		t.Fatalf("invalid setup: %+v", got)
	}
	os.WriteFile(path, []byte(`not json`), 0o644)
	if got := s.Load(); got != Default() {
		t.Fatalf("corrupt file: %+v", got)
	}
}
