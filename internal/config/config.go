// Package config loads and persists NEST user settings.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"nest/internal/fsutil"
	"nest/internal/knownfolders"
)

// CustomRule maps one extension onto a destination: either a known-folder key
// (downloads, documents, pictures, videos, music, desktop) or a cluster folder
// name created inside the source (e.g. "_Blender").
type CustomRule struct {
	Ext    string `json:"ext"`
	Target string `json:"target"`
}

// Config holds every user-tunable setting for NEST.
type Config struct {
	Sources []string `json:"sources"`

	CustomRules []CustomRule `json:"customRules"`

	// Daily schedule: run at each of these HH:MM times. Empty => manual only.
	Times []string `json:"times"`
	// Legacy single-time fields, migrated into Times on load.
	Enabled      bool   `json:"enabled"`
	ScheduleTime string `json:"scheduleTime"`

	// Instant mode: watch sources and organize new downloads on arrival.
	Instant      bool `json:"instant"`
	InstantDelay int  `json:"instantDelay"`

	Theme         string `json:"theme"`
	MinCluster    int    `json:"minCluster"`
	MinAgeSeconds int    `json:"minAgeSeconds"`
}

var timeRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// ValidTime reports whether s is a HH:MM time.
func ValidTime(s string) bool { return timeRe.MatchString(s) }

// Dir returns the NEST configuration directory (%APPDATA%\Nest), creating it
// if it does not exist yet.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	d := filepath.Join(base, "Nest")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

// Path returns the absolute path to config.json.
func Path() string {
	d, _ := Dir()
	return filepath.Join(d, "config.json")
}

// Default returns the settings a fresh install starts with.
func Default() Config {
	downloads := knownfolders.Downloads()
	sources := []string{}
	if downloads != "" {
		sources = append(sources, downloads)
	}
	return Config{
		Sources:       sources,
		Times:         []string{},
		Instant:       false,
		InstantDelay:  45,
		Theme:         "dark",
		MinCluster:    3,
		MinAgeSeconds: 60,
	}
}

// Load reads config.json, falling back to defaults when missing or corrupt.
func Load() Config {
	cfg := Default()
	data, err := os.ReadFile(Path())
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default()
	}
	cfg.normalize()
	return cfg
}

// Normalized returns a validated copy of the config with invalid values
// repaired or removed.
func (c Config) Normalized() Config {
	c.normalize()
	return c
}

// Save writes the config to disk atomically.
func (c Config) Save() error {
	c = c.Normalized()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(Path(), data, 0o644)
}

// SanitizeTarget cleans a rule destination. A target is either a known-folder
// key (pictures, documents, ...) or a single folder name NEST creates inside
// the source.
func SanitizeTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" || target == "." || target == ".." || target == "_" {
		return ""
	}
	target = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return -1
		}
		return r
	}, target)
	if target == "" {
		return ""
	}
	if knownfolders.NameOf(strings.ToLower(target)) != "" {
		return strings.ToLower(target)
	}
	return target
}

func (c *Config) normalize() {
	// Migrate the legacy single-time schedule.
	if len(c.Times) == 0 && c.Enabled && ValidTime(c.ScheduleTime) {
		c.Times = []string{c.ScheduleTime}
	}
	// Validate, dedupe and sort times.
	seen := map[string]bool{}
	times := make([]string, 0, len(c.Times))
	for _, t := range c.Times {
		if ValidTime(t) && !seen[t] {
			seen[t] = true
			times = append(times, t)
		}
	}
	sort.Strings(times)
	c.Times = times
	c.Enabled = len(times) > 0
	if len(times) > 0 {
		c.ScheduleTime = times[0]
	} else {
		c.ScheduleTime = "12:00"
	}

	if c.InstantDelay < 10 {
		c.InstantDelay = 10
	}
	if c.InstantDelay > 3600 {
		c.InstantDelay = 3600
	}
	if c.Theme != "light" && c.Theme != "dark" {
		c.Theme = "dark"
	}

	// Sources: trim, absolutize and dedupe (Windows paths are case-insensitive).
	// An empty list stays empty — the user may deliberately watch nothing.
	seenSrc := map[string]bool{}
	sources := make([]string, 0, len(c.Sources))
	for _, s := range c.Sources {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if abs, err := filepath.Abs(s); err == nil {
			s = abs
		}
		s = filepath.Clean(s)
		key := strings.ToLower(s)
		if seenSrc[key] {
			continue
		}
		seenSrc[key] = true
		sources = append(sources, s)
	}
	c.Sources = sources

	clean := make([]CustomRule, 0, len(c.CustomRules))
	for _, r := range c.CustomRules {
		ext := strings.ToLower(strings.TrimSpace(r.Ext))
		if ext == "" || strings.ContainsAny(ext, `\/:`) {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		target := SanitizeTarget(r.Target)
		if len(ext) < 2 || target == "" {
			continue
		}
		dup := false
		for _, c2 := range clean {
			if c2.Ext == ext {
				dup = true
				break
			}
		}
		if !dup {
			clean = append(clean, CustomRule{Ext: ext, Target: target})
		}
	}
	c.CustomRules = clean

	if c.MinCluster < 1 {
		c.MinCluster = 3
	}
	if c.MinAgeSeconds < 0 {
		c.MinAgeSeconds = 60
	}
	if c.MinAgeSeconds > 86400 {
		c.MinAgeSeconds = 86400
	}
}
