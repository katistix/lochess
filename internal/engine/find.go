package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Find locates a Fairy-Stockfish binary: the one bundled into release builds
// if present, otherwise one installed on the system.
func Find() (string, error) {
	if len(bundledBinary) > 0 {
		cache, err := os.UserCacheDir()
		if err == nil {
			var path string
			if path, err = extract(bundledBinary, cache); err == nil {
				return path, nil
			}
		}
		log.Printf("bundled engine unavailable, looking for an installed one: %v", err)
	}
	for _, p := range searchPaths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}
	return "", ErrNotFound
}

// searchPaths are tried in order. GUI apps launched from Finder don't inherit
// the shell PATH, so the Homebrew locations are listed explicitly.
var searchPaths = []string{
	"fairy-stockfish",
	"/opt/homebrew/bin/fairy-stockfish",
	"/usr/local/bin/fairy-stockfish",
}

// extract writes the engine to <cacheDir>/lochess/engine/<hash>/ once and
// returns its path. The content hash in the path means an app update with a
// new engine never reuses a stale file.
func extract(data []byte, cacheDir string) (string, error) {
	sum := sha256.Sum256(data)
	name := "fairy-stockfish"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dir := filepath.Join(cacheDir, "lochess", "engine", hex.EncodeToString(sum[:8]))
	path := filepath.Join(dir, name)
	if st, err := os.Stat(path); err == nil && st.Size() == int64(len(data)) {
		return path, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("extract engine: %w", err)
	}
	tmp, err := os.CreateTemp(dir, name+".*.tmp")
	if err != nil {
		return "", fmt.Errorf("extract engine: %w", err)
	}
	_, err = tmp.Write(data)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0o755)
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("extract engine: %w", err)
	}
	return path, nil
}
