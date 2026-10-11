package main

import (
	"context"

	"lochess/internal/match"
)

// App is the binding layer between the frontend and the Go backend.
// Keep it thin: game logic lives in internal packages.
type App struct {
	ctx   context.Context
	match *match.Match
}

func NewApp() *App {
	return &App{match: match.New()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.match.Close()
}

// Ping lets the frontend verify the Go bridge is wired up.
func (a *App) Ping() string {
	return "ok"
}

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
