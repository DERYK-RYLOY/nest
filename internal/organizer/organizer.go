// Package organizer scans a source folder, classifies every file and moves it
// to its home: a real Windows known folder, or a self-created cluster folder.
// It never touches folders, hidden/system files, in-progress downloads or files
// that were modified within the last few seconds.
package organizer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"nest/internal/classifier"
	"nest/internal/knownfolders"
)

// Options controls a single organize pass.
type Options struct {
	Source        string
	Apply         bool // false => dry run, nothing is moved
	MinAgeSeconds int
	MinCluster    int
	MiscName      string
	Rules         map[string]string // lowercase ext (".blend") -> target
}

// PlannedMove is a single file's intended relocation.
type PlannedMove struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
	KnownKey string `json:"knownKey"`
	Kind     string `json:"kind"`
	Err      string `json:"err"`
}

// GroupStat summarizes one destination group.
type GroupStat struct {
	Category string `json:"category"`
	Icon     string `json:"icon"`
	Folder   string `json:"folder"`
	Count    int    `json:"count"`
	Bytes    int64  `json:"bytes"`
}

// Report is the outcome of a pass.
type Report struct {
	Source       string        `json:"source"`
	Apply        bool          `json:"apply"`
	Scanned      int           `json:"scanned"`
	Skipped      int           `json:"skipped"`
	SkippedYoung int           `json:"skippedYoung"`
	Moved        int           `json:"moved"`
	Moves        []PlannedMove `json:"moves"`
	Bytes        int64         `json:"bytes"`
	Groups       []GroupStat   `json:"groups"`
	Errors       []string      `json:"errors"`
}

type candidate struct {
	path    string
	size    int64
	modTime int64
	dest    classifier.Destination
}

var ignoreNames = map[string]bool{
	"desktop.ini": true,
	"thumbs.db":   true,
	".ds_store":   true,
}

var ignoreSuffixes = []string{
	".crdownload", ".part", ".partial", ".download", ".tmp", ".!ut", ".opdownload",
}

// Run performs one organize pass over opts.Source.
func Run(opts Options) Report {
	if opts.MiscName == "" {
		opts.MiscName = "_Misc"
	}
	if opts.MinCluster < 1 {
		opts.MinCluster = 3
	}
	rep := Report{Source: opts.Source, Apply: opts.Apply, Moves: []PlannedMove{}, Errors: []string{}}

	entries, err := os.ReadDir(opts.Source)
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("cannot read %s: %v", opts.Source, err))
		return rep
	}

	var cands []candidate
	clusterCount := map[string]int{}

	for _, e := range entries {
		if e.IsDir() {
			rep.Skipped++
			continue
		}
		name := e.Name()
		if ignoreNames[strings.ToLower(name)] || hasIgnoredSuffix(name) {
			rep.Skipped++
			continue
		}
		full := filepath.Join(opts.Source, name)
		if isHiddenOrSystem(full) {
			rep.Skipped++
			continue
		}
		info, err := e.Info()
		if err != nil {
			rep.Skipped++
			continue
		}
		if opts.MinAgeSeconds > 0 && time.Since(info.ModTime()) < time.Duration(opts.MinAgeSeconds)*time.Second {
			rep.Skipped++
			rep.SkippedYoung++
			continue
		}
		rep.Scanned++

		dest := classifier.ResolvePath(name)
		if tgt, ok := opts.Rules[strings.ToLower(filepath.Ext(name))]; ok && tgt != "" {
			dest = classifier.ResolveCustom(filepath.Ext(name), tgt)
		}
		if dest.Kind == "cluster" {
			clusterCount[dest.Cluster]++
		}
		cands = append(cands, candidate{path: full, size: info.Size(), modTime: info.ModTime().Unix(), dest: dest})
	}

	for _, c := range cands {
		dest := c.dest
		// Named family clusters (Archives, Code, ...) always apply. Only the
		// unknown-extension fallback clusters need a minimum group size;
		// stragglers drop into _Misc so nothing is left loose.
		if dest.Kind == "cluster" && dest.Custom && clusterCount[dest.Cluster] < opts.MinCluster {
			dest = classifier.Destination{Kind: "cluster", Category: "Misc", Icon: "folder", Cluster: opts.MiscName}
		}

		targetDir := resolveTarget(opts.Source, &dest)

		pm := PlannedMove{
			From:     c.path,
			Name:     filepath.Base(c.path),
			Size:     c.size,
			Modified: c.modTime,
			Category: dest.Category,
			Icon:     dest.Icon,
			KnownKey: dest.KnownKey,
			Kind:     dest.Kind,
		}

		if strings.EqualFold(filepath.Dir(c.path), targetDir) {
			rep.Skipped++
			continue
		}

		finalDest := filepath.Join(targetDir, pm.Name)
		if opts.Apply {
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				pm.Err = err.Error()
				rep.Errors = append(rep.Errors, pm.Err)
				rep.Moves = append(rep.Moves, pm)
				continue
			}
			finalDest = uniqueDest(targetDir, pm.Name)
			if err := moveFile(c.path, finalDest); err != nil {
				pm.Err = err.Error()
				rep.Errors = append(rep.Errors, fmt.Sprintf("%s: %v", pm.Name, err))
				rep.Moves = append(rep.Moves, pm)
				continue
			}
			rep.Moved++
		}
		pm.To = finalDest
		rep.Bytes += c.size
		rep.Moves = append(rep.Moves, pm)
	}

	rep.Groups = summarize(rep.Moves)
	return rep
}

// resolveTarget turns a Destination into an absolute folder path. Known folders
// fall back to a cluster inside the source when Windows cannot resolve them.
func resolveTarget(source string, dest *classifier.Destination) string {
	if dest.Kind == "known" {
		if p := knownfolders.Lookup(dest.KnownKey); p != "" {
			return p
		}
		dest.Kind = "cluster"
		dest.Cluster = "_" + dest.Category
	}
	return filepath.Join(source, dest.Cluster)
}

func summarize(moves []PlannedMove) []GroupStat {
	type acc struct {
		icon   string
		folder string
		count  int
		bytes  int64
	}
	order := []string{}
	byCat := map[string]*acc{}
	for _, m := range moves {
		if m.Err != "" {
			continue
		}
		a, ok := byCat[m.Category]
		if !ok {
			a = &acc{icon: m.Icon, folder: filepath.Dir(m.To)}
			byCat[m.Category] = a
			order = append(order, m.Category)
		}
		a.count++
		a.bytes += m.Size
	}
	out := make([]GroupStat, 0, len(order))
	for _, cat := range order {
		a := byCat[cat]
		out = append(out, GroupStat{Category: cat, Icon: a.icon, Folder: a.folder, Count: a.count, Bytes: a.bytes})
	}
	return out
}

func hasIgnoredSuffix(name string) bool {
	lower := strings.ToLower(name)
	for _, s := range ignoreSuffixes {
		if strings.HasSuffix(lower, s) {
			return true
		}
	}
	return false
}

func isHiddenOrSystem(path string) bool {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return false
	}
	const hidden = windows.FILE_ATTRIBUTE_HIDDEN
	const system = windows.FILE_ATTRIBUTE_SYSTEM
	return attrs&(hidden|system) != 0
}

func uniqueDest(dir, name string) string {
	dest := filepath.Join(dir, name)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; i < 100000; i++ {
		cand := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return cand
		}
	}
	return dest
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
