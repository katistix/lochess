# lochess

**lochess** is a local chess coach based on lichess and stockfish. it is written in go + wails.

## Development

Requires Go, Node, pnpm and the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation).

```bash
wails dev     # run with hot reload
wails build   # produce build/bin/lochess.app
```

## Layout

- `main.go`, `app.go` — Wails entry point and the thin binding layer exposed to the frontend
- `frontend/` — Vite + React + TypeScript + Tailwind v4 (render only, no game logic)
- `frontend/src/routes/` — TanStack Router file-based routes (`routeTree.gen.ts` is generated)
- `frontend/wailsjs/` — generated Go bindings (do not edit)
- `build/` — platform packaging assets
