// Package engine runs full organize passes over every configured source.
// It is shared by the UI, the scheduled --auto mode and the --watch mode so
// all three behave identically.
package engine

import (
	"fmt"
	"time"

	"nest/internal/config"
	"nest/internal/organizer"
)

// RunAll organizes every source in cfg. When apply is true the moves are
// executed and recorded as undoable history entries. Every source of one pass
// shares a RunID so the whole pass can be undone as a single action.
func RunAll(cfg config.Config, apply bool) []organizer.Report {
	return runPass(cfg, apply, newRunID())
}

func newRunID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func runPass(cfg config.Config, apply bool, runID string) []organizer.Report {
	rules := make(map[string]string, len(cfg.CustomRules))
	for _, r := range cfg.CustomRules {
		rules[r.Ext] = r.Target
	}
	reports := make([]organizer.Report, 0, len(cfg.Sources))
	for i, src := range cfg.Sources {
		rep := organizer.Run(organizer.Options{
			Source:        src,
			Apply:         apply,
			MinAgeSeconds: cfg.MinAgeSeconds,
			MinCluster:    cfg.MinCluster,
			MiscName:      "_Misc",
			Rules:         rules,
		})
		if apply && rep.Moved > 0 {
			_ = organizer.AppendHistory(recordFor(runID, i, src, &rep))
		}
		reports = append(reports, rep)
	}
	return reports
}

func recordFor(runID string, index int, src string, rep *organizer.Report) organizer.RunRecord {
	rec := organizer.RunRecord{
		ID:        fmt.Sprintf("%s-%d", runID, index),
		RunID:     runID,
		Timestamp: time.Now(),
		Source:    src,
		Bytes:     rep.Bytes,
		Moves:     make([]organizer.Move, 0, len(rep.Moves)),
	}
	for _, m := range rep.Moves {
		if m.Err != "" {
			continue
		}
		rec.Moves = append(rec.Moves, organizer.Move{
			From: m.From, To: m.To, Size: m.Size,
			Category: m.Category, KnownKey: m.KnownKey,
		})
	}
	return rec
}

// Totals summarizes a batch of reports.
func Totals(reports []organizer.Report) (moved int, bytes int64) {
	for _, r := range reports {
		moved += r.Moved
		bytes += r.Bytes
	}
	return moved, bytes
}

// Errors collects every failure from a batch of reports.
func Errors(reports []organizer.Report) []string {
	var out []string
	for _, r := range reports {
		out = append(out, r.Errors...)
	}
	return out
}

// RunSettled behaves like RunAll, but when applying it repeats the pass while
// young (still-downloading) files were skipped, giving them time to finish.
// At most two extra rounds, 30 seconds apart. All rounds share one RunID so the
// settled pass is still a single undoable action.
func RunSettled(cfg config.Config, apply bool) []organizer.Report {
	runID := newRunID()
	reports := runPass(cfg, apply, runID)
	if !apply {
		return reports
	}
	all := [][]organizer.Report{reports}
	for round := 0; round < 2; round++ {
		young := 0
		for _, r := range reports {
			young += r.SkippedYoung
		}
		if young == 0 {
			break
		}
		time.Sleep(30 * time.Second)
		reports = runPass(cfg, apply, runID)
		all = append(all, reports)
	}
	return mergeRounds(all)
}

// mergeRounds combines the passes of a settled run so callers (the headless
// toast and log lines) report everything that moved, not just the last round.
// Groups are left from the first round; only the UI renders them and the UI
// never settles.
func mergeRounds(rounds [][]organizer.Report) []organizer.Report {
	if len(rounds) == 0 {
		return nil
	}
	out := make([]organizer.Report, len(rounds[0]))
	copy(out, rounds[0])
	for _, r := range rounds[1:] {
		for i := range r {
			if i >= len(out) {
				continue
			}
			out[i].Moved += r[i].Moved
			out[i].Bytes += r[i].Bytes
			out[i].Moves = append(out[i].Moves, r[i].Moves...)
			out[i].Errors = append(out[i].Errors, r[i].Errors...)
			out[i].SkippedYoung = r[i].SkippedYoung
		}
	}
	return out
}
