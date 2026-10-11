// Package review analyses saved games on demand and caches the result.
package review

import (
	"context"
	"fmt"
	"sync"

	"lochess/internal/analysis"
	"lochess/internal/engine"
	"lochess/internal/history"
)

// Service is safe for concurrent use. Analyses run one at a time on a
// dedicated full-strength engine, separate from the bot's.
type Service struct {
	mu    sync.Mutex
	store *history.Store
	eng   *engine.Engine
}

func New(store *history.Store) *Service {
	return &Service{store: store}
}

// Analyze returns the game with its analysis, running it first if needed.
// Concurrent calls for the same game wait and then get the cached result.
func (s *Service) Analyze(ctx context.Context, id string, progress func(done, total int)) (history.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, err := s.store.Get(id)
	if err != nil || g.Analysis != nil {
		return g, err
	}
	if err := s.startEngine(); err != nil {
		return g, err
	}
	a, err := analysis.Run(ctx, s.eng, g.Plies, progress)
	if err != nil {
		return g, fmt.Errorf("analyse: %w", err)
	}
	colors := []string{g.HumanColor}
	if g.Mode == "local" {
		colors = []string{"white", "black"}
	}
	a.KeyMoments = a.FindKeyMoments(colors, 3)
	g.Analysis = a
	return g, s.store.Save(g)
}

func (s *Service) startEngine() error {
	if s.eng != nil {
		return nil
	}
	path, err := engine.Find()
	if err != nil {
		return err
	}
	eng, err := engine.Start(path)
	if err != nil {
		return err
	}
	for name, value := range map[string]any{"Use NNUE": false, "Skill Level": 20} {
		if err := eng.SetOption(name, value); err != nil {
			eng.Close()
			return err
		}
	}
	s.eng = eng
	return nil
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.eng != nil {
		s.eng.Close()
		s.eng = nil
	}
}
