//go:build !bundled_engine

package engine

// Development builds use an installed engine (e.g. brew install fairy-stockfish).
var bundledBinary []byte
