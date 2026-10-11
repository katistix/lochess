#!/bin/sh
# Installs or updates lochess on macOS and Linux:
#   curl -fsSL https://raw.githubusercontent.com/katistix/lochess/main/install.sh | sh
#
# The chess engine ships inside the app, so nothing else is needed on macOS.
# On Linux the app needs WebKitGTK, which this installs with your package
# manager if it's missing (you may be asked for your password).
#
# LOCHESS_DIR overrides where the app goes.
set -eu

REPO="katistix/lochess"
BASE="https://github.com/$REPO/releases/latest/download"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

install_macos() {
  DEST="${LOCHESS_DIR:-/Applications}"
  if [ -z "${LOCHESS_DIR:-}" ] && [ ! -w "$DEST" ]; then
    DEST="$HOME/Applications"
  fi
  mkdir -p "$DEST"

  say "Downloading lochess for macOS..."
  curl -fsSL "$BASE/lochess-macos.zip" -o "$TMP/lochess.zip"
  ditto -x -k "$TMP/lochess.zip" "$TMP/app"
  rm -rf "$DEST/lochess.app"
  mv "$TMP/app/lochess.app" "$DEST/lochess.app"

  say "Installed lochess to $DEST/lochess.app"
  say "Open it from Launchpad, or run: open \"$DEST/lochess.app\""
}

# Installs WebKitGTK 4.1 (and GTK 3) if the system doesn't have it yet.
ensure_webkit() {
  if ldconfig -p 2>/dev/null | grep -q 'libwebkit2gtk-4.1\.so\.0'; then
    return
  fi
  SUDO=""
  [ "$(id -u)" -ne 0 ] && SUDO="sudo"
  if command -v apt-get >/dev/null 2>&1; then
    PKGS="libwebkit2gtk-4.1-0 libgtk-3-0"
    CMD="$SUDO apt-get install -y $PKGS"
  elif command -v dnf >/dev/null 2>&1; then
    PKGS="webkit2gtk4.1 gtk3"
    CMD="$SUDO dnf install -y $PKGS"
  elif command -v pacman >/dev/null 2>&1; then
    PKGS="webkit2gtk-4.1 gtk3"
    CMD="$SUDO pacman -S --needed --noconfirm $PKGS"
  elif command -v zypper >/dev/null 2>&1; then
    PKGS="libwebkit2gtk-4_1-0 gtk3"
    CMD="$SUDO zypper install -y $PKGS"
  else
    say "Note: couldn't find WebKitGTK 4.1, which lochess needs."
    say "      Install it with your package manager (often named webkit2gtk-4.1), then run lochess."
    return
  fi
  say "lochess needs WebKitGTK. Installing: $PKGS"
  [ -n "$SUDO" ] && say "(sudo may ask for your password)"
  # Read from the terminal: stdin is this script when piped from curl.
  $CMD </dev/tty || say "Note: installing WebKitGTK failed; install it yourself with: $CMD"
}

install_linux() {
  case "$(uname -m)" in
    x86_64 | amd64) ;;
    *) die "lochess for Linux is currently built for x86_64 only (this machine is $(uname -m))." ;;
  esac

  BIN_DIR="${LOCHESS_DIR:-$HOME/.local/bin}"
  APPS_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
  ICON_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/icons/hicolor/256x256/apps"
  mkdir -p "$BIN_DIR" "$APPS_DIR" "$ICON_DIR"

  ensure_webkit

  say "Downloading lochess for Linux..."
  curl -fsSL "$BASE/lochess-linux-amd64.tar.gz" -o "$TMP/lochess.tar.gz"
  tar -xzf "$TMP/lochess.tar.gz" -C "$TMP"

  install -m 755 "$TMP/lochess/lochess" "$BIN_DIR/lochess"
  install -m 644 "$TMP/lochess/lochess.png" "$ICON_DIR/lochess.png"
  # Launchers don't always have ~/.local/bin on PATH, so point at it directly.
  sed "s|^Exec=.*|Exec=$BIN_DIR/lochess|" "$TMP/lochess/lochess.desktop" >"$APPS_DIR/lochess.desktop"
  command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$APPS_DIR" >/dev/null 2>&1 || true

  say "Installed lochess to $BIN_DIR/lochess"
  say "Find it in your applications menu, or run: $BIN_DIR/lochess"
}

case "$(uname -s)" in
  Darwin) install_macos ;;
  Linux) install_linux ;;
  *) die "unsupported system $(uname -s). On Windows, see https://github.com/$REPO#install" ;;
esac
