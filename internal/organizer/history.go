package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nest/internal/config"
	"nest/internal/fsutil"
)

// Move records a single file relocation.
type Move struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Size     int64  `json:"size"`
	Category string `json:"category"`
	KnownKey string `json:"knownKey"`
}

// RunRecord is one source's part of an organize pass, kept so it can be undone.
type RunRecord struct {
	ID string `json:"id"`
	// RunID ties together the records produced by one pass over several
	// sources, so undoing a run restores every folder at once.
	RunID     string    `json:"runId,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Moves     []Move    `json:"moves"`
	Bytes     int64     `json:"bytes"`
	Undone    bool      `json:"undone"`
}

// UndoResult describes what an undo actually restored.
type UndoResult struct {
	RunID   string `json:"runId"`
	Records int    `json:"records"`
	Moves   int    `json:"moves"`
	Bytes   int64  `json:"bytes"`
	Missing int    `json:"missing"`
}

func historyPath() string {
	d, _ := config.Dir()
	return filepath.Join(d, "history.json")
}

// LoadHistory returns all recorded runs, newest first.
func LoadHistory() []RunRecord {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return []RunRecord{}
	}
	var recs []RunRecord
	if err := json.Unmarshal(data, &recs); err != nil {
		return []RunRecord{}
	}
	return recs
}

func saveHistory(recs []RunRecord) error {
	const maxRecords = 100
	if len(recs) > maxRecords {
		recs = recs[:maxRecords]
	}
	data, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(historyPath(), data, 0o644)
}

// AppendHistory prepends a run record, keeping the list bounded.
func AppendHistory(rec RunRecord) error {
	recs := LoadHistory()
	recs = append([]RunRecord{rec}, recs...)
	return saveHistory(recs)
}

// LatestRun folds the newest run that has not been undone into a single
// record, joining every source it touched. Returns nil when nothing is left to
// show (no runs at all, or every run was undone).
func LatestRun(recs []RunRecord) *RunRecord {
	start := -1
	for i := range recs {
		if !recs[i].Undone {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	out := recs[start]
	out.Moves = append([]Move(nil), recs[start].Moves...)
	if out.RunID == "" {
		return &out
	}
	for _, r := range recs[start+1:] {
		if r.RunID != out.RunID {
			continue
		}
		if r.Undone {
			continue
		}
		out.Moves = append(out.Moves, r.Moves...)
		out.Bytes += r.Bytes
	}
	return &out
}

// UndoLast restores the files moved by the most recent run that has not been
// undone yet. Every record that belongs to the same run is restored, so a pass
// over several folders is reverted as one action. Files that no longer exist at
// their destination are skipped; a failed move leaves the run undoable so the
// user can retry.
func UndoLast() (UndoResult, error) {
	recs := LoadHistory()
	newest := -1
	for i := range recs {
		if !recs[i].Undone {
			newest = i
			break
		}
	}
	if newest < 0 {
		return UndoResult{}, os.ErrNotExist
	}

	group := []int{newest}
	if id := recs[newest].RunID; id != "" {
		for i := newest + 1; i < len(recs); i++ {
			if recs[i].RunID == id && !recs[i].Undone {
				group = append(group, i)
			}
		}
	}

	res := UndoResult{RunID: recs[newest].RunID, Records: len(group)}
	var failures []string
	attempted := 0
	for _, i := range group {
		for _, m := range recs[i].Moves {
			attempted++
			missing, err := restoreMove(m)
			if missing {
				res.Missing++
				continue
			}
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", filepath.Base(m.To), err))
				continue
			}
			res.Moves++
			res.Bytes += m.Size
		}
	}

	if len(failures) > 0 {
		sample := failures
		if len(sample) > 3 {
			sample = sample[:3]
		}
		return res, fmt.Errorf("could not restore %d of %d file(s): %s",
			len(failures), attempted, strings.Join(sample, "; "))
	}

	for _, i := range group {
		recs[i].Undone = true
	}
	if err := saveHistory(recs); err != nil {
		return res, err
	}
	return res, nil
}

// restoreMove puts one file back where it came from without overwriting
// anything that has appeared there since.
func restoreMove(m Move) (missing bool, err error) {
	if _, err := os.Stat(m.To); err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(m.From), 0o755); err != nil {
		return false, err
	}
	dest := m.From
	if _, err := os.Stat(dest); err == nil {
		dest = uniqueDest(filepath.Dir(m.From), filepath.Base(m.From))
	}
	return false, moveFile(m.To, dest)
}

// ClearHistory deletes the entire run log, resetting all stats.
func ClearHistory() error {
	err := os.Remove(historyPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// TotalBytes returns the total number of bytes organized across active runs.
func TotalBytes() int64 {
	var total int64
	for _, r := range LoadHistory() {
		if !r.Undone {
			total += r.Bytes
		}
	}
	return total
}
