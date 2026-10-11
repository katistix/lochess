package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"lochess/internal/history"
	"lochess/internal/match"
	"lochess/internal/review"
)

// App is the binding layer between the frontend and the Go backend.
// Keep it thin: game logic lives in internal packages.
type App struct {
	ctx     context.Context
	match   *match.Match
	history *history.Store
	review  *review.Service
}

func NewApp() (*App, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	store, err := history.Open(filepath.Join(dir, "lochess", "games"))
	if err != nil {
		return nil, fmt.Errorf("open game history: %w", err)
	}
	return &App{
		match:   match.New(store.Save),
		history: store,
		review:  review.New(store),
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.match.Close()
	a.review.Close()
}

// Ping lets the frontend verify the Go bridge is wired up.
func (a *App) Ping() string {
	return "ok"
}

// --- Play ---

func (a *App) GameState() match.State {
	return a.match.State()
}

func (a *App) NewGame(opts match.Options) (match.State, error) {
	return a.match.Start(opts)
}

// Setup leaves the current game and returns to the options screen.
func (a *App) Setup() match.State {
	return a.match.Setup()
}

func (a *App) Resign() (match.State, error) {
	return a.match.Resign()
}

// Move plays the human's UCI move. On error the promise rejects and the
// frontend should re-sync from GameState.
func (a *App) Move(uci string) (match.State, error) {
	return a.match.Move(uci)
}

// BotMove asks the bot to play when State.BotToMove is true.
func (a *App) BotMove() (match.State, error) {
	return a.match.BotMove()
}

// --- History & review ---

func (a *App) ListGames() ([]history.Summary, error) {
	return a.history.List()
}

func (a *App) GetGame(id string) (history.Game, error) {
	return a.history.Get(id)
}

func (a *App) DeleteGame(id string) error {
	return a.history.Delete(id)
}

// AnalyzeGame runs (or returns the cached) engine analysis for a game.
// Progress is emitted as "analysis:progress" events: (id, done, total).
func (a *App) AnalyzeGame(id string) (history.Game, error) {
	return a.review.Analyze(a.ctx, id, func(done, total int) {
		runtime.EventsEmit(a.ctx, "analysis:progress", id, done, total)
	})
}
