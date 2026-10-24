package notify

import (
	"os"
	"testing"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

// TestToastDelivery pushes a clearly labeled self-test toast. It passes when
// Windows accepts the notification. Watch for the popup when running tests.
// Gated so the default suite never pops notifications at the user.
func TestToastDelivery(t *testing.T) {
	if os.Getenv("NEST_LIVE_TESTS") == "" {
		t.Skip("set NEST_LIVE_TESTS=1 to push a real notification")
	}
	n := toast.Notification{
		AppID: "NEST File Organizer",
		Title: "NEST self-test",
		Body:  "Notifications work. Nothing was organized.",
	}
	if err := n.Push(); err != nil {
		t.Fatalf("toast push failed: %v", err)
	}
}
