#!/bin/sh
# Installs lochess on macOS:
#   curl -fsSL https://raw.githubusercontent.com/katistix/lochess/main/install.sh | sh
# Set LOCHESS_DIR to install somewhere other than /Applications.
set -eu

REPO="katistix/lochess"
URL="https://github.com/$REPO/releases/latest/download/lochess-macos.zip"

if [ "$(uname -s)" != "Darwin" ]; then
  echo "lochess currently supports macOS only." >&2
  exit 1
fi

# The bot and game analysis need Fairy-Stockfish (the engine lichess uses).
have_engine() {
  command -v fairy-stockfish >/dev/null 2>&1 ||
    [ -x /opt/homebrew/bin/fairy-stockfish ] || [ -x /usr/local/bin/fairy-stockfish ]
}
if ! have_engine; then
  BREW="$(command -v brew || true)"
  [ -z "$BREW" ] && [ -x /opt/homebrew/bin/brew ] && BREW=/opt/homebrew/bin/brew
  [ -z "$BREW" ] && [ -x /usr/local/bin/brew ] && BREW=/usr/local/bin/brew
  if [ -n "$BREW" ]; then
    echo "Installing Fairy-Stockfish with Homebrew..."
    "$BREW" install fairy-stockfish
  else
    echo "Note: Homebrew not found, so Fairy-Stockfish was not installed." >&2
    echo "      1v1 works without it; for the bot and analysis run: brew install fairy-stockfish" >&2
  fi
fi

DEST="${LOCHESS_DIR:-/Applications}"
if [ -z "${LOCHESS_DIR:-}" ] && [ ! -w "$DEST" ]; then
  DEST="$HOME/Applications"
fi
mkdir -p "$DEST"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading lochess..."
curl -fsSL "$URL" -o "$TMP/lochess.zip"
ditto -x -k "$TMP/lochess.zip" "$TMP/app"

rm -rf "$DEST/lochess.app"
mv "$TMP/app/lochess.app" "$DEST/lochess.app"

echo "Installed lochess to $DEST/lochess.app"
echo "Open it from Launchpad, or run: open \"$DEST/lochess.app\""
