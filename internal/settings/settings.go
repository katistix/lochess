// Package settings stores user preferences as a small JSON file.
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// Settings are display preferences. Values are theme IDs the frontend
// defines; unknown IDs fall back to defaults there.
type Settings struct {
	Board  string `json:"board"`
	Pieces string `json:"pieces"`
	Sound  bool   `json:"sound"`
	Volume int    `json:"volume"` // 0-100
}

func Default() Settings {
	return Settings{Board: "brown", Pieces: "cburnett", Sound: true, Volume: 70}
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
	out.Volume = max(0, min(100, out.Volume))
	return out
}

func (s *Store) Save(v Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.Board == "" || v.Pieces == "" {
		return errors.New("settings: board and pieces are required")
	}
	v.Volume = max(0, min(100, v.Volume))
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
