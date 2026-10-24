// Package scheduler manages NEST automation through Windows Task Scheduler:
// one daily task per chosen time ("NEST Auto 1230") plus an optional logon
// task ("NEST Instant Watch") for instant mode. NEST itself never runs in the
// background; these tasks are what wake it up.
package scheduler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
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

var timeRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// ValidTime reports whether s is a HH:MM time.
func ValidTime(s string) bool { return timeRe.MatchString(s) }

func taskNameFor(hhmm string) string {
	return TaskPrefix + strings.ReplaceAll(hhmm, ":", "")
}

// Sync registers one daily task per time in times and removes NEST daily tasks
// that are no longer wanted. Wanted tasks are created before stale ones are
// deleted, so a failure can never leave the user with no schedule at all.
func Sync(times []string) error {
	seen := map[string]bool{}
	wanted := make([]string, 0, len(times))
	for _, t := range times {
		if !ValidTime(t) {
			return fmt.Errorf("invalid schedule time %q (want HH:MM)", t)
		}
		if !seen[t] {
			seen[t] = true
			wanted = append(wanted, t)
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	keep := make(map[string]bool, len(wanted))
	for _, t := range wanted {
		name := taskNameFor(t)
		keep[name] = true
		out, err := exec.Command("schtasks",
			"/Create", "/TN", name,
			"/TR", fmt.Sprintf(`\"%s\" --auto`, exe),
			"/SC", "DAILY", "/ST", t, "/F",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
		}
	}

	for _, suffix := range registeredSuffixes() {
		name := TaskPrefix + suffix
		if !keep[name] {
			_ = deleteTask(name)
		}
	}
	return nil
}

// registeredSuffixes returns the HHMM suffixes of existing daily tasks. The
// output is parsed as CSV because task names may contain commas.
func registeredSuffixes() []string {
	out, err := exec.Command("schtasks", "/Query", "/FO", "CSV", "/NH").CombinedOutput()
	if err != nil {
		return nil
	}
	r := csv.NewReader(bytes.NewReader(out))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	var found []string
	for {
		rec, err := r.Read()
		if err != nil {
			break
		}
		if len(rec) == 0 {
			continue
		}
		name := strings.TrimPrefix(strings.TrimSpace(rec[0]), `\`)
		if strings.HasPrefix(name, TaskPrefix) {
			found = append(found, strings.TrimPrefix(name, TaskPrefix))
		}
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
	sort.Strings(out)
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

// StopWatch asks Task Scheduler to end a running watcher. Best effort: the
// task may not be running at all.
func StopWatch() {
	_ = exec.Command("schtasks", "/End", "/TN", WatchTask).Run()
}

// StartWatch runs the logon watcher task immediately. Best effort: the task
// also starts by itself at the next logon.
func StartWatch() {
	_ = exec.Command("schtasks", "/Run", "/TN", WatchTask).Run()
}

// RestartWatch stops the running watcher and starts it again, so changed
// folders and delays take effect without logging off.
func RestartWatch() {
	StopWatch()
	StartWatch()
}

// SetWatch toggles the logon task that runs "nest.exe --watch".
func SetWatch(on bool) error {
	if !on {
		// Stop a watcher that is running right now before removing the task,
		// otherwise it would keep organizing until the session ends.
		StopWatch()
		return deleteTask(WatchTask)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	out, err := exec.Command("schtasks",
		"/Create", "/TN", WatchTask,
		"/TR", fmt.Sprintf(`\"%s\" --watch`, exe),
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
