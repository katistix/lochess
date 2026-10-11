# lochess

**lochess** is a local chess coach based on lichess and stockfish. it is written in go + wails.

## Development

Requires Go, Node, pnpm and the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation).
Playing against the bot needs [Fairy-Stockfish](https://fairy-stockfish.github.io/) (`brew install fairy-stockfish`), the engine lichess uses for its AI levels.

```bash
wails dev     # run with hot reload
wails build   # produce build/bin/lochess.app
```

## Layout

- `main.go`, `app.go` — Wails entry point and the thin binding layer exposed to the frontend
- `internal/game` — rules and authoritative game state (notnil/chess)
- `internal/engine` — UCI subprocess wrapper
- `internal/bot` — lichess AI levels 1–8 on Fairy-Stockfish
- `internal/match` — a game plus its mode (local 1v1 or vs bot), level and colours
- `frontend/` — Vite + React + TypeScript + Tailwind v4 (render only, no game logic)
- `frontend/src/routes/` — TanStack Router file-based routes (`routeTree.gen.ts` is generated)
- `frontend/wailsjs/` — generated Go bindings (do not edit)
- `build/` — platform packaging assets
