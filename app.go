package main

import "context"

// App is the binding layer between the frontend and the Go backend.
// Keep it thin: game logic will live in internal packages.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Ping lets the frontend verify the Go bridge is wired up.
func (a *App) Ping() string {
	return "ok"
}
