package scheduler

import (
	"os"
	"strings"
	"testing"
)

// requireLive keeps the suite from touching the user's real Task Scheduler.
// Set NEST_LIVE_TESTS=1 to exercise the real tasks.
func requireLive(t *testing.T) {
	t.Helper()
	if os.Getenv("NEST_LIVE_TESTS") == "" {
		t.Skip("set NEST_LIVE_TESTS=1 to exercise the real Task Scheduler")
	}
}

func accessDenied(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "denied") ||
		strings.Contains(msg, "0x80070005") ||
		strings.Contains(msg, "zugriff")
}

func sameTimes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]bool, len(a))
	for _, t := range a {
		seen[t] = true
	}
	for _, t := range b {
		if !seen[t] {
			return false
		}
	}
	return true
}

// TestSyncRoundTrip registers a task, verifies it, then restores whatever
// schedule existed before the test. Skips when Task Scheduler needs elevation.
func TestSyncRoundTrip(t *testing.T) {
	requireLive(t)
	original := RegisteredTimes()
	t.Cleanup(func() {
		if err := Sync(original); err != nil {
			t.Logf("could not restore the original schedule %v: %v", original, err)
		}
	})

	if err := Sync([]string{"23:59"}); err != nil {
		if accessDenied(err) {
			t.Skipf("task scheduler needs admin: %v", err)
		}
		t.Fatalf("sync failed: %v", err)
	}
	found := false
	for _, tm := range RegisteredTimes() {
		if tm == "23:59" {
			found = true
		}
	}
	if !found {
		t.Errorf("registered task not found, got %v", RegisteredTimes())
	}

	if err := Sync(original); err != nil {
		t.Fatalf("restore sync failed: %v", err)
	}
	if got := RegisteredTimes(); !sameTimes(got, original) {
		t.Errorf("schedule not restored: want %v, got %v", original, got)
	}
}

// TestWatchRoundTrip toggles the logon watcher task on and off, then restores
// its previous state.
func TestWatchRoundTrip(t *testing.T) {
	requireLive(t)
	original := WatchActive()
	t.Cleanup(func() {
		if err := SetWatch(original); err != nil {
			t.Logf("could not restore watcher state (%v): %v", original, err)
		}
	})

	if err := SetWatch(true); err != nil {
		if accessDenied(err) {
			t.Skipf("task scheduler needs admin: %v", err)
		}
		t.Fatalf("enabling watch failed: %v", err)
	}
	if !WatchActive() {
		t.Errorf("watch task not active after enabling")
	}
	if err := SetWatch(false); err != nil {
		t.Fatalf("disabling watch failed: %v", err)
	}
	if WatchActive() {
		t.Errorf("watch task still active after disabling")
	}
}

func TestValidTime(t *testing.T) {
	for _, ok := range []string{"00:00", "08:30", "23:59"} {
		if !ValidTime(ok) {
			t.Errorf("ValidTime(%q) = false", ok)
		}
	}
	for _, bad := range []string{"", "9:00", "24:00", "12:60", "noon", "12:00:00", "08:30 "} {
		if ValidTime(bad) {
			t.Errorf("ValidTime(%q) = true", bad)
		}
	}
}

// TestSyncRejectsInvalidTimes verifies validation happens before any task is
// touched, so a bad value can never disturb an existing schedule.
func TestSyncRejectsInvalidTimes(t *testing.T) {
	before := RegisteredTimes()
	for _, bad := range []string{"", "9:00", "24:00", "12:60", "noon"} {
		if err := Sync([]string{bad}); err == nil {
			t.Errorf("Sync(%q) accepted an invalid time", bad)
		}
	}
	after := RegisteredTimes()
	if len(before) != len(after) {
		t.Errorf("invalid times changed the schedule: %v -> %v", before, after)
	}
}
