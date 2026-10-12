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
// executed and recorded as one undoable history entry per source.
func RunAll(cfg config.Config, apply bool) []organizer.Report {
	rules := make(map[string]string, len(cfg.CustomRules))
	for _, r := range cfg.CustomRules {
		rules[r.Ext] = r.Target
	}
	reports := make([]organizer.Report, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		rep := organizer.Run(organizer.Options{
			Source:        src,
			Apply:         apply,
			MinAgeSeconds: cfg.MinAgeSeconds,
			MinCluster:    cfg.MinCluster,
			MiscName:      "_Misc",
			Rules:         rules,
		})
		if apply && rep.Moved > 0 {
			_ = organizer.AppendHistory(recordFor(src, &rep))
		}
		reports = append(reports, rep)
	}
	return reports
}

func recordFor(src string, rep *organizer.Report) organizer.RunRecord {
	rec := organizer.RunRecord{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
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

// RunSettled behaves like RunAll, but when applying it repeats the pass while
// young (still-downloading) files were skipped, giving them time to finish.
// At most two extra rounds, 30 seconds apart.
func RunSettled(cfg config.Config, apply bool) []organizer.Report {
	reports := RunAll(cfg, apply)
	if !apply {
		return reports
	}
	for round := 0; round < 2; round++ {
		young := 0
		for _, r := range reports {
			young += r.SkippedYoung
		}
		if young == 0 {
			break
		}
		time.Sleep(30 * time.Second)
		reports = RunAll(cfg, apply)
	}
	return reports
}
