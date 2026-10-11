package main

import (
	"context"

	"lochess/internal/game"
)

// App is the binding layer between the frontend and the Go backend.
// Keep it thin: game logic lives in internal packages.
type App struct {
	ctx  context.Context
	game *game.Game
}

func NewApp() *App {
	return &App{game: game.New()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Ping lets the frontend verify the Go bridge is wired up.
func (a *App) Ping() string {
	return "ok"
}

func (a *App) GameState() game.State {
	return a.game.State()
}

func (a *App) NewGame() game.State {
	return a.game.Reset()
}

// Move plays a UCI move. On error the promise rejects and the frontend
// should re-render the returned state from GameState.
func (a *App) Move(uci string) (game.State, error) {
	return a.game.Move(uci)
}
