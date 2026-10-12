package notify

import (
	"testing"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

// TestToastDelivery pushes a clearly labeled self-test toast. It passes when
// Windows accepts the notification. Watch for the popup when running tests.
func TestToastDelivery(t *testing.T) {
	n := toast.Notification{
		AppID: "NEST File Organizer",
		Title: "NEST self-test",
		Body:  "Notifications work. Nothing was organized.",
	}
	if err := n.Push(); err != nil {
		t.Fatalf("toast push failed: %v", err)
	}
}
