#!/usr/bin/env bash
# Places Fairy-Stockfish at internal/engine/bundled/fairy-stockfish.bin so a
# build with `-tags bundled_engine` ships its own engine.
#
#   scripts/fetch-engine.sh linux|windows|darwin
#
# Linux and Windows use the official Fairy-Stockfish 14 release binaries,
# verified against pinned SHA-256 hashes. There is no official macOS build, so
# on macOS it compiles the same release from source as a universal binary.
set -euo pipefail

cd "$(dirname "$0")/.."
OUT=internal/engine/bundled/fairy-stockfish.bin
RELEASE=https://github.com/fairy-stockfish/Fairy-Stockfish/releases/download/fairy_sf_14
COMMIT=f3e6969d11d1bec17eba26e7ae0e629ad4af71dd # tag fairy_sf_14

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | cut -d' ' -f1
  else shasum -a 256 "$1" | cut -d' ' -f1; fi
}

fetch() { # asset, expected sha256
  curl -fsSL "$RELEASE/$1" -o "$OUT"
  local got
  got="$(sha256 "$OUT")"
  if [ "$got" != "$2" ]; then
    rm -f "$OUT"
    echo "checksum mismatch for $1: got $got, want $2" >&2
    exit 1
  fi
}

mkdir -p "$(dirname "$OUT")"
case "${1:-}" in
  linux)
    fetch fairy-stockfish_x86-64 ab6b85823152e78654092dc2fbb154956a559c6ef0455d728268544390ee150f ;;
  windows)
    fetch fairy-stockfish_x86-64.exe 28d5c18fd7352d66800d13e1fdd57ba7d64d95c38d93642fd12cbd89e9d0ed73 ;;
  darwin)
    src="$(mktemp -d)"
    trap 'rm -rf "$src"' EXIT
    git -C "$src" init --quiet
    git -C "$src" fetch --quiet --depth 1 https://github.com/fairy-stockfish/Fairy-Stockfish "$COMMIT"
    git -C "$src" checkout --quiet FETCH_HEAD
    jobs="$(sysctl -n hw.ncpu)"
    for arch in apple-silicon x86-64; do
      make -C "$src/src" clean >/dev/null
      make -C "$src/src" -j"$jobs" build ARCH="$arch" COMP=clang >/dev/null
      mv "$src/src/stockfish" "$src/$arch"
    done
    lipo -create -output "$OUT" "$src/apple-silicon" "$src/x86-64"
    codesign --force --sign - "$OUT" ;;
  *)
    echo "usage: $0 linux|windows|darwin" >&2
    exit 2 ;;
esac
echo "bundled engine: $OUT ($(wc -c <"$OUT" | tr -d ' ') bytes)"
