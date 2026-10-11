package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"lochess/internal/history"
	"lochess/internal/match"
	"lochess/internal/review"
	"lochess/internal/settings"
)

// App is the binding layer between the frontend and the Go backend.
// Keep it thin: game logic lives in internal packages.
type App struct {
	ctx      context.Context
	match    *match.Match
	history  *history.Store
	review   *review.Service
	settings *settings.Store
}

func NewApp() (*App, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	root := filepath.Join(dir, "lochess")
	store, err := history.Open(filepath.Join(root, "games"))
	if err != nil {
		return nil, fmt.Errorf("open game history: %w", err)
	}
	a := &App{history: store, review: review.New(store), settings: settings.Open(filepath.Join(root, "settings.json"))}
	a.match = match.New(a.saveGame, filepath.Join(root, "current.json"))
	return a, nil
}

// saveGame stores a finished game and starts analysing it in the background,
// so the review is usually ready by the time it is opened.
func (a *App) saveGame(g history.Game) error {
	if err := a.history.Save(g); err != nil {
		return err
	}
	go func() {
		if _, err := a.AnalyzeGame(g.ID); err != nil {
			log.Printf("background analysis of %s: %v", g.ID, err)
		}
	}()
	return nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.match.Close()
	a.review.Close()
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

// Takeback undoes the human's last move (and the bot's reply).
func (a *App) Takeback() (match.State, error) {
	return a.match.Takeback()
}

// AgreeDraw ends a local 1v1 game as a draw.
func (a *App) AgreeDraw() (match.State, error) {
	return a.match.AgreeDraw()
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

// --- Settings ---

func (a *App) GetSettings() settings.Settings {
	return a.settings.Load()
}

func (a *App) SaveSettings(s settings.Settings) error {
	return a.settings.Save(s)
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
