package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFiles creates n files with the given extension inside dir.
func writeFiles(t *testing.T, dir, prefix, ext string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		name := filepath.Join(dir, prefix+string(rune('a'+i))+ext)
		if err := os.WriteFile(name, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunClustersAndMisc(t *testing.T) {
	dir := t.TempDir()

	writeFiles(t, dir, "arc", ".zip", 4)  // -> _Archives
	writeFiles(t, dir, "code", ".go", 5)  // -> _Code
	writeFiles(t, dir, "lone", ".xyz", 1) // < threshold -> _Misc

	// A subdirectory and an in-progress download must be left alone.
	if err := os.Mkdir(filepath.Join(dir, "_Archives"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "partial.crdownload"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep := Run(Options{Source: dir, Apply: true, MinAgeSeconds: 0, MinCluster: 3})

	if rep.Moved != 10 {
		t.Fatalf("expected 10 moves, got %d (errors: %v)", rep.Moved, rep.Errors)
	}
	if len(rep.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", rep.Errors)
	}

	count := func(sub string) int {
		entries, err := os.ReadDir(filepath.Join(dir, sub))
		if err != nil {
			return 0
		}
		return len(entries)
	}
	if got := count("_Archives"); got != 4 {
		t.Errorf("_Archives: want 4, got %d", got)
	}
	if got := count("_Code"); got != 5 {
		t.Errorf("_Code: want 5, got %d", got)
	}
	if got := count("_Misc"); got != 1 {
		t.Errorf("_Misc: want 1, got %d", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "partial.crdownload")); err != nil {
		t.Errorf("in-progress download should not have been moved")
	}
}

func TestFamilyClusterAppliesBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "solo", ".zip", 1) // family cluster, no threshold

	rep := Run(Options{Source: dir, Apply: true, MinAgeSeconds: 0, MinCluster: 3})

	if rep.Moved != 1 {
		t.Fatalf("expected 1 move, got %d (errors: %v)", rep.Moved, rep.Errors)
	}
	if _, err := os.Stat(filepath.Join(dir, "_Archives", "soloa.zip")); err != nil {
		t.Errorf("single archive should land in _Archives: %v", err)
	}
}

func TestCustomRules(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "blend", ".blend", 2) // unknown ext, custom cluster

	rep := Run(Options{
		Source: dir, Apply: true, MinAgeSeconds: 0, MinCluster: 3,
		Rules: map[string]string{".blend": "_Blender"},
	})

	if rep.Moved != 2 {
		t.Fatalf("expected 2 moves, got %d (errors: %v)", rep.Moved, rep.Errors)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "_Blender"))
	if err != nil || len(entries) != 2 {
		t.Errorf("custom cluster _Blender should hold 2 files: %v", err)
	}
}

func TestMoveFileNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("source"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := moveFile(src, dst); err == nil {
		t.Fatal("moveFile replaced an existing destination")
	}
	if b, _ := os.ReadFile(dst); string(b) != "keep" {
		t.Errorf("destination was modified: %q", b)
	}
	if b, _ := os.ReadFile(src); string(b) != "source" {
		t.Errorf("source was removed after a failed move: %q", b)
	}
}

func TestDryRunReportsFinalCollisionName(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "arc", ".zip", 1)
	if err := os.MkdirAll(filepath.Join(dir, "_Archives"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "_Archives", "arca.zip"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep := Run(Options{Source: dir, Apply: false, MinAgeSeconds: 0, MinCluster: 3})

	if len(rep.Moves) != 1 {
		t.Fatalf("want 1 planned move, got %d", len(rep.Moves))
	}
	if got := filepath.Base(rep.Moves[0].To); got != "arca (1).zip" {
		t.Errorf("preview should show the collision-free name, got %q", got)
	}
}

func TestRunDryRunMovesNothing(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "arc", ".zip", 5)

	rep := Run(Options{Source: dir, Apply: false, MinAgeSeconds: 0, MinCluster: 3})

	if rep.Moved != 0 {
		t.Fatalf("dry run reported %d moves", rep.Moved)
	}
	if _, err := os.Stat(filepath.Join(dir, "arca.zip")); err != nil {
		t.Errorf("dry run should leave files in place")
	}
	if len(rep.Moves) != 5 {
		t.Errorf("dry run should still plan 5 moves, got %d", len(rep.Moves))
	}
}
