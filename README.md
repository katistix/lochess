<div align="center">

# lochess

**A quiet, local chess app for getting better at chess.**<br>
Play Stockfish at lichess's beginner-friendly levels, then see exactly where each game turned.

[![Release](https://img.shields.io/github/v/release/katistix/lochess?style=flat-square&color=81b64c)](https://github.com/katistix/lochess/releases/latest)
[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square)](LICENSE)
![Platform: macOS](https://img.shields.io/badge/platform-macOS-lightgrey?style=flat-square)

<img src="docs/screenshots/review.png" alt="Reviewing a game: a blunder marked on the board, the best move shown as an arrow, and accuracy for both sides" width="860">

</div>

## Features

- **Play Stockfish at levels 1–8**, using the same engine and settings lichess uses for its AI, so even level 1 is beatable.
- **Review every game.** Each move is graded from *Best* to *Blunder*, with the better move drawn on the board, an eval bar, accuracy for both sides and the three moments that mattered most.
- **Your history, kept locally.** Every finished game is saved and analysed in the background, ready when you are.
- **Local 1v1** on one board, with takebacks and draws by agreement.
- **Make it yours.** 6 board colours, 12 open-licensed piece sets from lichess, and move sounds.
- **Private by design.** No account and no network: games, analysis and settings stay on your Mac.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/katistix/lochess/main/install.sh | sh
```

This installs **lochess.app** into `/Applications` (or `~/Applications`). If you use [Homebrew](https://brew.sh), it also installs [Fairy-Stockfish](https://fairy-stockfish.github.io/), the engine behind the bot and the analysis. Local 1v1 works without it.

Run the same command again to update. Works on Apple Silicon and Intel Macs.

<details>
<summary>Uninstall</summary>

```bash
rm -rf /Applications/lochess.app ~/Library/Application\ Support/lochess
```

</details>

## Screenshots

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/play.png" alt="Playing Stockfish level 3"></td>
    <td width="50%"><img src="docs/screenshots/settings.png" alt="Board, piece and sound settings with a live preview"></td>
  </tr>
  <tr>
    <td align="center"><sub>Playing Stockfish level 3</sub></td>
    <td align="center"><sub>Boards, piece sets and sounds, with a live preview</sub></td>
  </tr>
</table>

<img src="docs/screenshots/history.png" alt="Game history with results and accuracy">

## How it works

### Bot levels

Stockfish is far too strong for beginners, even at its lowest setting. lochess copies lichess's approach instead ([fishnet](https://github.com/lichess-org/fishnet)): it plays with **Fairy-Stockfish** using classical evaluation, a limited search, and a *Skill Level* that goes below zero. Official Stockfish can't do that.

| Level | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Skill | −9 | −5 | −1 | 3 | 7 | 11 | 16 | 20 |
| Depth | 5 | 5 | 5 | 5 | 5 | 8 | 13 | 22 |
| Time | 50 ms | 100 ms | 150 ms | 200 ms | 300 ms | 400 ms | 500 ms | 1 s |

### Move grades

After a game, a full-strength engine evaluates every position. Each evaluation becomes a winning chance (lichess's formula), and a move is graded by how much winning chance it gave away, using chess.com's thresholds:

| Best | Excellent | Good | Inaccuracy `?!` | Mistake `?` | Blunder `??` |
|---|---|---|---|---|---|
| engine's choice | < 2% | < 5% | < 10% | < 20% | ≥ 20% |

Accuracy uses lichess's per-move formula, and *key moments* are your three costliest mistakes.

### Your data

Everything lives in `~/Library/Application Support/lochess/`: one JSON file per game in `games/`, your settings, and an unfinished game, which is restored when you reopen the app.

## Development

Requires [Go](https://go.dev), [Node](https://nodejs.org) with [pnpm](https://pnpm.io), the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation) and `brew install fairy-stockfish`.

```bash
wails dev               # run with hot reload
wails build             # build build/bin/lochess.app
go test ./...           # backend tests (engine tests skip without Fairy-Stockfish)
```

Set `LOCHESS_DATA_DIR` to run against a separate data folder, e.g. for testing without touching your own games.

<details>
<summary>Project layout</summary>

Go owns all game state and logic. The frontend only renders it and sends moves back.

| Path | |
|---|---|
| `main.go`, `app.go` | Wails entry point and the thin binding layer exposed to the frontend |
| `internal/game` | Rules and authoritative game state ([notnil/chess](https://github.com/notnil/chess)) |
| `internal/engine` | UCI subprocess wrapper |
| `internal/bot` | lichess AI levels 1–8 on Fairy-Stockfish, with crash recovery |
| `internal/match` | A game plus its mode, level and colours; saves finished games, resumes unfinished ones |
| `internal/history` | Finished games as JSON files |
| `internal/analysis` | Move grades, accuracy and key moments |
| `internal/review` | Runs and caches analysis on a dedicated full-strength engine |
| `internal/settings` | Display and sound preferences |
| `frontend/` | Vite, React, TypeScript, Tailwind v4 and [chessground](https://github.com/lichess-org/chessground) |
| `frontend/src/routes/` | File-based routes ([TanStack Router](https://tanstack.com/router)) |
| `frontend/wailsjs/` | Generated Go bindings (do not edit) |
| `install.sh`, `.github/workflows/` | One-line installer and the release build |

</details>

### Releasing

Push a version tag. GitHub Actions tests, builds a universal app and publishes the release that `install.sh` downloads:

```bash
git tag -a v0.1.1 -m "lochess v0.1.1" && git push origin v0.1.1
```

## Credits

- [lichess](https://lichess.org): the board ([chessground](https://github.com/lichess-org/chessground)), bot levels and piece sets. Authors and licences are in [`frontend/src/assets/pieces/CREDITS.md`](frontend/src/assets/pieces/CREDITS.md).
- [Fairy-Stockfish](https://fairy-stockfish.github.io/), the chess engine.
- [notnil/chess](https://github.com/notnil/chess), the move generation and rules.
- [Kenney](https://kenney.nl), the sounds (CC0), credited in [`frontend/src/assets/sounds/CREDITS.md`](frontend/src/assets/sounds/CREDITS.md).

## License

[GPL-3.0-or-later](LICENSE). Piece sets and sounds keep their own licences, listed in their `CREDITS.md` files.
