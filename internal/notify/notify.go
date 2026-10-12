// Package notify shows Windows toast notifications for headless runs, so the
// user sees a single "NEST arranged your files" note and nothing else.
package notify

import (
	"fmt"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

// Organized tells the user a background pass finished. Best effort: errors are
// returned but callers usually ignore them.
func Organized(moved int, sourceCount int) error {
	body := fmt.Sprintf("Arranged %d file(s) into their folders.", moved)
	if sourceCount > 1 {
		body = fmt.Sprintf("Arranged %d file(s) across %d folders.", moved, sourceCount)
	}
	n := toast.Notification{
		AppID: "NEST File Organizer",
		Title: "NEST organized your files",
		Body:  body,
	}
	return n.Push()
}
