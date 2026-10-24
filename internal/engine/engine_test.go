package engine

import (
	"os"
	"path/filepath"
	"testing"

	"nest/internal/config"
	"nest/internal/organizer"
)

func writeZip(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunAllGroupsSourcesIntoOneRun verifies that a pass over several folders
// is recorded as one run and restored as one undo.
func TestRunAllGroupsSourcesIntoOneRun(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	base := t.TempDir()
	a := filepath.Join(base, "A")
	b := filepath.Join(base, "B")
	writeZip(t, a, "one.zip")
	writeZip(t, b, "two.zip")

	cfg := config.Config{Sources: []string{a, b}, MinAgeSeconds: 0, MinCluster: 3}
	reports := RunAll(cfg, true)
	if len(reports) != 2 {
		t.Fatalf("want 2 reports, got %d", len(reports))
	}
	for _, r := range reports {
		if r.Moved != 1 {
			t.Errorf("%s: want 1 moved file, got %d (errors: %v)", r.Source, r.Moved, r.Errors)
		}
	}

	history := organizer.LoadHistory()
	if len(history) != 2 {
		t.Fatalf("want 2 history records, got %d", len(history))
	}
	if history[0].RunID == "" || history[0].RunID != history[1].RunID {
		t.Errorf("records of one pass should share a RunID: %q vs %q", history[0].RunID, history[1].RunID)
	}
	if last := organizer.LatestRun(history); last == nil || len(last.Moves) != 2 {
		t.Errorf("dashboard aggregate should see both sources, got %+v", last)
	}

	res, err := organizer.UndoLast()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if res.Records != 2 || res.Moves != 2 {
		t.Errorf("want both sources undone, got %+v", res)
	}
	for _, p := range []string{
		filepath.Join(a, "one.zip"),
		filepath.Join(b, "two.zip"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file was not restored to %s: %v", p, err)
		}
	}
}

func TestMergeRoundsAddsUpEveryRound(t *testing.T) {
	rounds := [][]organizer.Report{
		{
			{Source: "a", Moved: 5, Bytes: 50, Errors: []string{"x"}},
			{Source: "b", Moved: 1, Bytes: 10},
		},
		{
			{Source: "a", Moved: 2, Bytes: 20},
			{Source: "b"},
		},
	}
	out := mergeRounds(rounds)
	if len(out) != 2 {
		t.Fatalf("want 2 reports, got %d", len(out))
	}
	if out[0].Moved != 7 || out[0].Bytes != 70 || len(out[0].Errors) != 1 {
		t.Errorf("first source not merged: %+v", out[0])
	}
	if out[1].Moved != 1 || out[1].Bytes != 10 {
		t.Errorf("second source not merged: %+v", out[1])
	}
}

// TestRunAllLeavesDestinationFoldersAlone guards against re-organizing the
// cluster folders NEST itself created.
func TestRunAllIsIdempotent(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	src := t.TempDir()
	writeZip(t, src, "one.zip")

	cfg := config.Config{Sources: []string{src}, MinAgeSeconds: 0}
	if got := RunAll(cfg, true); got[0].Moved != 1 {
		t.Fatalf("first pass moved %d, want 1", got[0].Moved)
	}
	if got := RunAll(cfg, true); got[0].Moved != 0 {
		t.Errorf("second pass moved %d files, want 0", got[0].Moved)
	}
}
