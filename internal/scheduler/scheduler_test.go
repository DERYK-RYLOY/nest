package scheduler

import (
	"strings"
	"testing"
)

func accessDenied(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "denied") ||
		strings.Contains(msg, "0x80070005") ||
		strings.Contains(msg, "zugriff")
}

// TestSyncRoundTrip registers a task, verifies it, then removes it again.
// Skips when Task Scheduler needs elevation.
func TestSyncRoundTrip(t *testing.T) {
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
	if err := Sync(nil); err != nil {
		t.Fatalf("cleanup sync failed: %v", err)
	}
	if got := RegisteredTimes(); len(got) != 0 {
		t.Errorf("expected no tasks after cleanup, got %v", got)
	}
}

// TestWatchRoundTrip toggles the logon watcher task on and off.
func TestWatchRoundTrip(t *testing.T) {
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
