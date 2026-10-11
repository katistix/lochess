# lochess

**lochess** is a local chess coach based on lichess and stockfish. it is written in go + wails.

## Install (macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/katistix/lochess/main/install.sh | sh
```

This installs `lochess.app` into `/Applications` (or `~/Applications`), plus
[Fairy-Stockfish](https://fairy-stockfish.github.io/) via Homebrew if you have it — the engine
behind the bot levels and game analysis. Run the same command again to update.

To uninstall: `rm -rf /Applications/lochess.app ~/Library/Application\ Support/lochess`

## Releasing

Push a version tag; GitHub Actions builds a universal (Apple Silicon + Intel) app and publishes it:

```bash
git tag -a v0.1.1 -m "lochess v0.1.1" && git push origin v0.1.1
```

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
- `internal/match` — a game plus its mode (local 1v1 or vs bot), level and colours; saves finished games and resumes an unfinished one (`current.json`)
- `internal/history` — finished games as JSON files in `~/Library/Application Support/lochess/games/`
- `internal/analysis` — engine analysis: move grades (chess.com thresholds), accuracy (lichess formula), key moments
- `internal/review` — runs and caches analysis on a dedicated full-strength engine
- `internal/settings` — display preferences (`settings.json`)
- `frontend/` — Vite + React + TypeScript + Tailwind v4 (render only, no game logic)
- `frontend/src/routes/` — TanStack Router file-based routes (`routeTree.gen.ts` is generated)
- `frontend/src/assets/pieces/` — piece sets from lichess (see `CREDITS.md` for authors and licences)
- `frontend/src/assets/sounds/` — move, capture and notify sounds from Kenney, CC0 (see `CREDITS.md`)
- `frontend/wailsjs/` — generated Go bindings (do not edit)
- `build/` — platform packaging assets
- `install.sh`, `.github/workflows/release.yml` — one-line installer and the tag-triggered release build

## License

[GPL-3.0-or-later](LICENSE). Piece sets and sounds keep their own licences; see the `CREDITS.md`
files under `frontend/src/assets/`.
