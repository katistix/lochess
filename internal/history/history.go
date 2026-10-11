// Package history stores finished games on disk, one JSON file per game.
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"lochess/internal/analysis"
	"lochess/internal/game"
)

// Game is a finished game as stored on disk.
type Game struct {
	ID         string             `json:"id"`
	PlayedAt   time.Time          `json:"playedAt"`
	Mode       string             `json:"mode"`       // "local" | "bot"
	Level      int                `json:"level"`      // bot level; 0 for local
	HumanColor string             `json:"humanColor"` // "white" | "black"; "white" for local
	Result     string             `json:"result"`     // "1-0" | "0-1" | "1/2-1/2"
	Method     string             `json:"method"`     // e.g. "Checkmate", "Resignation"
	Plies      []game.Ply         `json:"plies"`
	Analysis   *analysis.Analysis `json:"analysis,omitempty"`
}

// Summary is a Game without its moves, for listing.
type Summary struct {
	ID         string    `json:"id"`
	PlayedAt   time.Time `json:"playedAt"`
	Mode       string    `json:"mode"`
	Level      int       `json:"level"`
	HumanColor string    `json:"humanColor"`
	Result     string    `json:"result"`
	Method     string    `json:"method"`
	Plies      int       `json:"plies"`
	Analyzed   bool      `json:"analyzed"`
	Accuracy   float64   `json:"accuracy"` // human side's accuracy in analysed bot games
}

// Store is safe for concurrent use.
type Store struct {
	mu  sync.Mutex
	dir string
}

var ErrNotFound = errors.New("game not found")

// Open uses dir for storage, creating it if needed.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// NewID returns a time-sortable game ID.
func NewID(t time.Time) string {
	return t.UTC().Format("20060102-150405.000")
}

func (s *Store) Save(g Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.path(g.ID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	// Write then rename so a crash never leaves a half-written game.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Store) Get(id string) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.get(id)
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.path(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return nil
}

// List returns all games, newest first. Unreadable files are skipped.
func (s *Store) List() ([]Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	files, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, err
	}
	out := []Summary{}
	for _, f := range files {
		id := filepath.Base(f)
		id = id[:len(id)-len(".json")]
		g, err := s.get(id)
		if err != nil {
			continue
		}
		sum := Summary{
			ID: g.ID, PlayedAt: g.PlayedAt, Mode: g.Mode, Level: g.Level, HumanColor: g.HumanColor,
			Result: g.Result, Method: g.Method, Plies: len(g.Plies), Analyzed: g.Analysis != nil,
		}
		if g.Analysis != nil && g.Mode == "bot" {
			sum.Accuracy = g.Analysis.WhiteAccuracy
			if g.HumanColor == "black" {
				sum.Accuracy = g.Analysis.BlackAccuracy
			}
		}
		out = append(out, sum)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PlayedAt.After(out[j].PlayedAt) })
	return out, nil
}

func (s *Store) get(id string) (Game, error) {
	path, err := s.path(id)
	if err != nil {
		return Game{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Game{}, ErrNotFound
	} else if err != nil {
		return Game{}, err
	}
	var g Game
	if err := json.Unmarshal(data, &g); err != nil {
		return Game{}, fmt.Errorf("game %s: %w", id, err)
	}
	return g, nil
}

var validID = regexp.MustCompile(`^[0-9A-Za-z.-]+$`)

// path maps an ID to its file, rejecting anything that could escape dir.
func (s *Store) path(id string) (string, error) {
	if !validID.MatchString(id) || id == "." || id == ".." {
		return "", fmt.Errorf("invalid game id %q", id)
	}
	return filepath.Join(s.dir, id+".json"), nil
}
