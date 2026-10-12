// Package settings stores user preferences as a small JSON file.
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"lochess/internal/bot"
)

// Settings are user preferences. Board and Pieces are theme IDs the
// frontend defines; unknown IDs fall back to defaults there. Mode, Level
// and Color remember the last game setup.
type Settings struct {
	Board  string `json:"board"`
	Pieces string `json:"pieces"`
	Sound  bool   `json:"sound"`
	Volume int    `json:"volume"` // 0-100
	Mode   string `json:"mode"`   // "bot" | "local"
	Level  int    `json:"level"`  // bot level
	Color  string `json:"color"`  // "white" | "random" | "black"
}

func Default() Settings {
	return Settings{Board: "brown", Pieces: "cburnett", Sound: true, Volume: 70, Mode: "bot", Level: bot.MinLevel, Color: "white"}
}

// clamp replaces out-of-range values with defaults.
func clamp(v Settings) Settings {
	d := Default()
	v.Volume = max(0, min(100, v.Volume))
	if v.Mode != "bot" && v.Mode != "local" {
		v.Mode = d.Mode
	}
	if v.Level < bot.MinLevel || v.Level > bot.MaxLevel {
		v.Level = d.Level
	}
	if v.Color != "white" && v.Color != "random" && v.Color != "black" {
		v.Color = d.Color
	}
	return v
}

// Store is safe for concurrent use.
type Store struct {
	mu   sync.Mutex
	path string
}

func Open(path string) *Store {
	return &Store{path: path}
}

// Load returns the saved settings, with defaults for anything missing.
// A missing or unreadable file gives the defaults.
func (s *Store) Load() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return Default()
	}
	// Decoding over the defaults keeps them for fields the file lacks.
	out := Default()
	if json.Unmarshal(data, &out) != nil {
		return Default()
	}
	if out.Board == "" {
		out.Board = Default().Board
	}
	if out.Pieces == "" {
		out.Pieces = Default().Pieces
	}
	return clamp(out)
}

func (s *Store) Save(v Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.Board == "" || v.Pieces == "" {
		return errors.New("settings: board and pieces are required")
	}
	v = clamp(v)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
