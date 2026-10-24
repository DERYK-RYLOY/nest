package organizer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// isolateHistory points %APPDATA% at a temp folder so history tests never
// touch the real NEST history file.
func isolateHistory(t *testing.T) {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())
}

func moveOut(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(from), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(from, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveFile(from, to); err != nil {
		t.Fatal(err)
	}
}

func TestUndoRestoresEverySourceOfOneRun(t *testing.T) {
	isolateHistory(t)
	base := t.TempDir()
	srcA := filepath.Join(base, "A")
	srcB := filepath.Join(base, "B")
	dst := filepath.Join(base, "dest")

	aFrom := filepath.Join(srcA, "a.txt")
	bFrom := filepath.Join(srcB, "b.txt")
	aTo := filepath.Join(dst, "a.txt")
	bTo := filepath.Join(dst, "b.txt")
	moveOut(t, aFrom, aTo)
	moveOut(t, bFrom, bTo)

	// Records are appended one per source, newest first on disk.
	if err := AppendHistory(RunRecord{
		ID: "r-0", RunID: "r", Source: srcA,
		Moves: []Move{{From: aFrom, To: aTo, Size: 4}}, Bytes: 4,
	}); err != nil {
		t.Fatal(err)
	}
	if err := AppendHistory(RunRecord{
		ID: "r-1", RunID: "r", Source: srcB,
		Moves: []Move{{From: bFrom, To: bTo, Size: 4}}, Bytes: 4,
	}); err != nil {
		t.Fatal(err)
	}

	res, err := UndoLast()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if res.Records != 2 || res.Moves != 2 {
		t.Errorf("want 2 records and 2 restored files, got %+v", res)
	}
	for _, p := range []string{aFrom, bFrom} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file was not restored to %s: %v", p, err)
		}
	}
	for _, r := range LoadHistory() {
		if !r.Undone {
			t.Errorf("record %s is still marked undoable", r.ID)
		}
	}
	if _, err := UndoLast(); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("second undo: want ErrNotExist, got %v", err)
	}
}

func TestUndoFailureKeepsRunUndoable(t *testing.T) {
	isolateHistory(t)
	base := t.TempDir()
	to := filepath.Join(base, "dest", "x.txt")
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A file occupies the directory the move must be restored into, so
	// restoring cannot succeed.
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(blocked, "nested", "x.txt")

	if err := AppendHistory(RunRecord{
		ID: "r-0", RunID: "r", Source: base,
		Moves: []Move{{From: from, To: to, Size: 4}}, Bytes: 4,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := UndoLast(); err == nil {
		t.Fatal("undo should report the failed restore")
	}
	if recs := LoadHistory(); len(recs) != 1 || recs[0].Undone {
		t.Errorf("failed undo must stay retryable, got %+v", recs)
	}
}

func TestUndoSkipsFilesThatAreAlreadyGone(t *testing.T) {
	isolateHistory(t)
	base := t.TempDir()
	from := filepath.Join(base, "lost.txt")
	to := filepath.Join(base, "dest", "lost.txt") // never created

	if err := AppendHistory(RunRecord{
		ID: "r-0", RunID: "r", Source: base,
		Moves: []Move{{From: from, To: to, Size: 4}}, Bytes: 4,
	}); err != nil {
		t.Fatal(err)
	}

	res, err := UndoLast()
	if err != nil {
		t.Fatalf("missing files should not fail the undo: %v", err)
	}
	if res.Missing != 1 || res.Moves != 0 {
		t.Errorf("want 1 missing and 0 restored, got %+v", res)
	}
	if !LoadHistory()[0].Undone {
		t.Error("run should still be marked undone")
	}
}

func TestLatestRunSkipsUndoneRuns(t *testing.T) {
	recs := []RunRecord{
		{ID: "2-0", RunID: "2", Undone: true, Bytes: 9, Moves: []Move{{}, {}}},
		{ID: "1-0", RunID: "1", Bytes: 4, Moves: []Move{{}}},
	}

	out := LatestRun(recs)
	if out == nil || out.ID != "1-0" || len(out.Moves) != 1 || out.Bytes != 4 {
		t.Errorf("want the newest active run, got %+v", out)
	}
	if LatestRun([]RunRecord{{ID: "x", Undone: true}}) != nil {
		t.Error("a fully undone history should have no latest run")
	}
}

func TestLatestRunAggregatesOnlyTheNewestRun(t *testing.T) {
	recs := []RunRecord{
		{ID: "1-1", RunID: "1", Bytes: 5, Moves: []Move{{}, {}}},
		{ID: "1-0", RunID: "1", Bytes: 10, Moves: []Move{{}}},
		{ID: "0-0", RunID: "0", Bytes: 99, Moves: []Move{{}, {}, {}, {}}},
	}

	out := LatestRun(recs)
	if out == nil {
		t.Fatal("nil result")
	}
	if len(out.Moves) != 3 || out.Bytes != 15 {
		t.Errorf("want 3 moves / 15 bytes from the newest run, got %d / %d", len(out.Moves), out.Bytes)
	}
	if len(recs[0].Moves) != 2 {
		t.Error("LatestRun mutated the stored records")
	}
	if LatestRun(nil) != nil {
		t.Error("LatestRun(nil) should be nil")
	}
}
