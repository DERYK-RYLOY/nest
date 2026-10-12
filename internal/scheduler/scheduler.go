// Package scheduler manages NEST automation through Windows Task Scheduler:
// one daily task per chosen time ("NEST Auto 1230") plus an optional logon
// task ("NEST Instant Watch") for instant mode. NEST itself never runs in the
// background; these tasks are what wake it up.
package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// TaskPrefix prefixes every daily task; the HHMM time is appended.
	TaskPrefix = "NEST Auto "
	// WatchTask is the logon task that runs "nest.exe --watch".
	WatchTask = "NEST Instant Watch"
)

// Status describes the automation state.
type Status struct {
	Enabled     bool     `json:"enabled"`
	Times       []string `json:"times"`
	Registered  []string `json:"registered"`
	Instant     bool     `json:"instant"`
	WatchActive bool     `json:"watchActive"`
	IsAdmin     bool     `json:"isAdmin"`
	Detail      string   `json:"detail"`
}

func taskNameFor(hhmm string) string {
	return TaskPrefix + strings.ReplaceAll(hhmm, ":", "")
}

// Sync replaces all NEST daily tasks with one per time in times.
func Sync(times []string) error {
	for _, suffix := range registeredSuffixes() {
		_ = deleteTask(TaskPrefix + suffix)
	}
	for _, t := range times {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		out, err := exec.Command("schtasks",
			"/Create", "/TN", taskNameFor(t),
			"/TR", fmt.Sprintf(`"%s" --auto`, exe),
			"/SC", "DAILY", "/ST", t, "/F",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

// registeredSuffixes returns the HHMM suffixes of existing daily tasks.
func registeredSuffixes() []string {
	out, err := exec.Command("schtasks", "/Query", "/FO", "CSV", "/NH").CombinedOutput()
	if err != nil {
		return nil
	}
	var found []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, `"`) {
			continue
		}
		parts := strings.Split(line, `","`)
		if len(parts) == 0 {
			continue
		}
		name := strings.TrimPrefix(strings.TrimPrefix(parts[0], `"`), `\`)
		if !strings.HasPrefix(name, TaskPrefix) {
			continue
		}
		found = append(found, strings.TrimPrefix(name, TaskPrefix))
	}
	return found
}

// RegisteredTimes returns the HH:MM times currently in Task Scheduler.
func RegisteredTimes() []string {
	var out []string
	for _, s := range registeredSuffixes() {
		if len(s) == 4 {
			out = append(out, s[:2]+":"+s[2:])
		}
	}
	return out
}

func deleteTask(name string) error {
	out, err := exec.Command("schtasks", "/Delete", "/TN", name, "/F").CombinedOutput()
	if err != nil {
		msg := strings.ToLower(string(out))
		if strings.Contains(msg, "cannot find") || strings.Contains(msg, "does not exist") ||
			strings.Contains(msg, "nicht gefunden") {
			return nil
		}
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// SetWatch toggles the logon task that runs "nest.exe --watch".
func SetWatch(on bool) error {
	if !on {
		return deleteTask(WatchTask)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	out, err := exec.Command("schtasks",
		"/Create", "/TN", WatchTask,
		"/TR", fmt.Sprintf(`"%s" --watch`, exe),
		"/SC", "ONLOGON", "/F",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// WatchActive reports whether the instant-watch task exists.
func WatchActive() bool {
	_, err := exec.Command("schtasks", "/Query", "/TN", WatchTask).CombinedOutput()
	return err == nil
}
