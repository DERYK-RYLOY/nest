package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRunFiresOnNewFile watches a temp dir and expects one debounced burst
// after a file lands. Temp-only, no system changes.
func TestRunFiresOnNewFile(t *testing.T) {
	dir := t.TempDir()
	stop := make(chan struct{})

	burst := make(chan struct{}, 4)
	go Run([]string{dir}, 2*time.Second, stop, func() { burst <- struct{}{} })

	time.Sleep(500 * time.Millisecond) // let the watcher arm
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-burst:
	case <-time.After(15 * time.Second):
		t.Fatal("no burst within 15s of file creation")
	}

	// Stop the watcher and let its directory handles drain before TempDir
	// cleanup runs, otherwise RemoveAll races the open handles on Windows.
	close(stop)
	time.Sleep(2 * time.Second)
}
