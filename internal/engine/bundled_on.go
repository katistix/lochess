//go:build bundled_engine

package engine

import _ "embed"

// bundledBinary is Fairy-Stockfish for the target platform, placed by
// scripts/fetch-engine.sh before a release build.
//
//go:embed bundled/fairy-stockfish.bin
var bundledBinary []byte
