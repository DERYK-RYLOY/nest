// Package watch waits for changes inside watched folders using
// ReadDirectoryChangesW, which blocks with genuinely zero CPU until something
// actually happens. It reports bursts through a debounced callback.
package watch

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// Wait blocks until dir changes and returns true, or returns false when stop
// is closed.
func Wait(dir string, stop <-chan struct{}) bool {
	pathPtr, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return false
	}
	h, err := windows.CreateFile(
		pathPtr,
		windows.FILE_LIST_DIRECTORY,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		select {
		case <-stop:
			return false
		case <-time.After(30 * time.Second):
			return false
		}
	}
	defer windows.CloseHandle(h)

	fired := make(chan struct{}, 1)
	go func() {
		select {
		case <-stop:
			windows.CloseHandle(h) // unblocks ReadDirectoryChanges
		case <-fired:
		}
	}()

	buf := make([]byte, 64*1024)
	var n uint32
	err = windows.ReadDirectoryChanges(
		h,
		&buf[0],
		uint32(len(buf)),
		false,
		windows.FILE_NOTIFY_CHANGE_FILE_NAME|
			windows.FILE_NOTIFY_CHANGE_DIR_NAME|
			windows.FILE_NOTIFY_CHANGE_SIZE|
			windows.FILE_NOTIFY_CHANGE_LAST_WRITE,
		&n,
		nil,
		0,
	)
	fired <- struct{}{}
	if err != nil {
		select {
		case <-stop:
			return false
		default:
			time.Sleep(5 * time.Second)
			return false
		}
	}
	return true
}

// Run watches every dir and calls onBurst once things have been quiet for the
// debounce period. It returns when stop is closed.
func Run(dirs []string, debounce time.Duration, stop <-chan struct{}, onBurst func()) {
	trigger := make(chan struct{}, 1)
	notify := func() {
		select {
		case trigger <- struct{}{}:
		default:
		}
	}

	for _, d := range dirs {
		go func(dir string) {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := os.Stat(dir); err != nil {
					select {
					case <-stop:
						return
					case <-time.After(time.Minute):
						continue
					}
				}
				select {
				case <-stop:
					return
				default:
				}
				if Wait(dir, stop) {
					notify()
				}
			}
		}(d)
	}

	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	for {
		select {
		case <-stop:
			return
		case <-trigger:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(debounce)
		case <-timer.C:
			onBurst()
		}
	}
}
