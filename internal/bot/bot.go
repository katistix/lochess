// Package bot plays moves at a chosen strength using a UCI engine.
package bot

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"lochess/internal/engine"
)

// Level mirrors one lichess AI level.
type Level struct {
	Skill    int           // UCI "Skill Level"; Fairy-Stockfish accepts -20..20
	Depth    int           // max search depth
	MoveTime time.Duration // max search time
}

// Levels 1-8 copied from lichess fishnet (src/api.rs, SkillLevel).
// Index 0 is unused so Levels[n] is level n.
var Levels = [...]Level{
	{},
	{Skill: -9, Depth: 5, MoveTime: 50 * time.Millisecond},
	{Skill: -5, Depth: 5, MoveTime: 100 * time.Millisecond},
	{Skill: -1, Depth: 5, MoveTime: 150 * time.Millisecond},
	{Skill: 3, Depth: 5, MoveTime: 200 * time.Millisecond},
	{Skill: 7, Depth: 5, MoveTime: 300 * time.Millisecond},
	{Skill: 11, Depth: 8, MoveTime: 400 * time.Millisecond},
	{Skill: 16, Depth: 13, MoveTime: 500 * time.Millisecond},
	{Skill: 20, Depth: 22, MoveTime: 1000 * time.Millisecond},
}

const MinLevel, MaxLevel = 1, len(Levels) - 1

// Bot plays like lichess's "Stockfish level N". Like fishnet, it uses
// Fairy-Stockfish with classical eval: official Stockfish clamps Skill Level
// to 0..20, which would make levels 1-3 far too strong for beginners.
type Bot struct {
	eng *engine.Engine
}

// enginePaths are tried in order. GUI apps launched from Finder don't inherit
// the shell PATH, so the Homebrew locations are listed explicitly.
var enginePaths = []string{
	"fairy-stockfish",
	"/opt/homebrew/bin/fairy-stockfish",
	"/usr/local/bin/fairy-stockfish",
}

// ErrEngineNotFound means no Fairy-Stockfish binary could be located.
var ErrEngineNotFound = errors.New("fairy-stockfish not found (install with: brew install fairy-stockfish)")

func New() (*Bot, error) {
	path, err := findEngine()
	if err != nil {
		return nil, err
	}
	eng, err := engine.Start(path)
	if err != nil {
		return nil, err
	}
	// Same options fishnet sends for move work on its MultiVariant engine.
	for name, value := range map[string]any{
		"UCI_Variant":     "chess",
		"Use NNUE":        false,
		"UCI_AnalyseMode": false,
	} {
		if err := eng.SetOption(name, value); err != nil {
			eng.Close()
			return nil, err
		}
	}
	return &Bot{eng: eng}, nil
}

func findEngine() (string, error) {
	for _, p := range enginePaths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}
	return "", ErrEngineNotFound
}

// NewGame clears the engine's state from any previous game.
func (b *Bot) NewGame(ctx context.Context) error {
	return b.eng.NewGame(ctx)
}

// Move returns the bot's move in UCI notation for the position at the given level.
func (b *Bot) Move(ctx context.Context, fen string, level int) (string, error) {
	if level < MinLevel || level > MaxLevel {
		return "", fmt.Errorf("level %d out of range %d-%d", level, MinLevel, MaxLevel)
	}
	l := Levels[level]
	if err := b.eng.SetOption("Skill Level", l.Skill); err != nil {
		return "", err
	}
	return b.eng.BestMove(ctx, fen, fmt.Sprintf("movetime %d depth %d", l.MoveTime.Milliseconds(), l.Depth))
}

func (b *Bot) Close() error {
	return b.eng.Close()
}
