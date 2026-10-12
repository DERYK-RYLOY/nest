package organizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"nest/internal/config"
)

// Move records a single file relocation.
type Move struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Size     int64  `json:"size"`
	Category string `json:"category"`
	KnownKey string `json:"knownKey"`
}

// RunRecord is one complete organize pass, kept so it can be undone.
type RunRecord struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Moves     []Move    `json:"moves"`
	Bytes     int64     `json:"bytes"`
	Undone    bool      `json:"undone"`
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
	return os.WriteFile(historyPath(), data, 0o644)
}

// AppendHistory prepends a run record, keeping the list bounded.
func AppendHistory(rec RunRecord) error {
	recs := LoadHistory()
	recs = append([]RunRecord{rec}, recs...)
	return saveHistory(recs)
}

// UndoLast restores the files moved by the most recent, not-yet-undone run.
// It returns the record that was reverted.
func UndoLast() (RunRecord, error) {
	recs := LoadHistory()
	for i := range recs {
		if recs[i].Undone {
			continue
		}
		for _, m := range recs[i].Moves {
			if _, err := os.Stat(m.To); err != nil {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(m.From), 0o755); err != nil {
				continue
			}
			// Never overwrite: if something now sits at the original path,
			// restore alongside it with a numbered name instead.
			dest := m.From
			if _, err := os.Stat(dest); err == nil {
				dest = uniqueDest(filepath.Dir(m.From), filepath.Base(m.From))
			}
			if err := moveFile(m.To, dest); err != nil {
				continue
			}
		}
		recs[i].Undone = true
		_ = saveHistory(recs)
		return recs[i], nil
	}
	return RunRecord{}, os.ErrNotExist
}

// ClearHistory deletes the entire run log, resetting all stats.
func ClearHistory() error {
	err := os.Remove(historyPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// TotalBytes returns the total number of bytes organized across all runs.
func TotalBytes() int64 {
	var total int64
	for _, r := range LoadHistory() {
		if !r.Undone {
			total += r.Bytes
		}
	}
	return total
}
