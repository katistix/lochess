// Package bot plays moves at a chosen strength using a UCI engine.
package bot

import (
	"context"
	"fmt"
	"log"
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

// moveTimeout bounds one attempt at a move; the slowest level searches for 1s.
const moveTimeout = 10 * time.Second

func New() (*Bot, error) {
	eng, err := start()
	if err != nil {
		return nil, err
	}
	return &Bot{eng: eng}, nil
}

func start() (*engine.Engine, error) {
	path, err := engine.Find()
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
	return eng, nil
}

// NewGame clears the engine's state from any previous game.
func (b *Bot) NewGame() error {
	return b.retry(func(ctx context.Context) error {
		return b.eng.NewGame(ctx)
	})
}

// Move returns the bot's move in UCI notation for the position at the given level.
func (b *Bot) Move(fen string, level int) (string, error) {
	if level < MinLevel || level > MaxLevel {
		return "", fmt.Errorf("level %d out of range %d-%d", level, MinLevel, MaxLevel)
	}
	l := Levels[level]
	var uci string
	err := b.retry(func(ctx context.Context) error {
		if err := b.eng.SetOption("Skill Level", l.Skill); err != nil {
			return err
		}
		var err error
		uci, err = b.eng.BestMove(ctx, fen, fmt.Sprintf("movetime %d depth %d", l.MoveTime.Milliseconds(), l.Depth))
		return err
	})
	return uci, err
}

// retry runs f; if it fails (engine crashed or hung), it restarts the engine
// and tries once more.
func (b *Bot) retry(f func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), moveTimeout)
	err := f(ctx)
	cancel()
	if err == nil {
		return nil
	}
	log.Printf("bot engine failed, restarting: %v", err)
	b.eng.Close()
	eng, startErr := start()
	if startErr != nil {
		return fmt.Errorf("%w (restart failed: %v)", err, startErr)
	}
	b.eng = eng
	ctx, cancel = context.WithTimeout(context.Background(), moveTimeout)
	defer cancel()
	return f(ctx)
}

func (b *Bot) Close() error {
	return b.eng.Close()
}
