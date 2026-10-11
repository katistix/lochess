package engine

import (
	"os"
	"runtime"
	"testing"
)

func TestExtract(t *testing.T) {
	cache := t.TempDir()
	path, err := extract([]byte("engine v1"), cache)
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Windows has no executable bit; elsewhere the engine must be runnable.
	if runtime.GOOS != "windows" && st.Mode().Perm()&0o100 == 0 {
		t.Fatalf("extracted file not executable: %v", st.Mode())
	}
	// Same content reuses the file; new content goes to a new path.
	if again, _ := extract([]byte("engine v1"), cache); again != path {
		t.Fatalf("re-extracted to %s, want %s", again, path)
	}
	if other, _ := extract([]byte("engine v2"), cache); other == path {
		t.Fatal("different engine reused the same path")
	}
}
