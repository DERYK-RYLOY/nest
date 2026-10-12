package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

type tokenElevation struct {
	Elevated uint32
}

// IsElevated reports whether this process runs with administrator rights.
// Daily tasks work as a standard user; the logon watcher task needs elevation.
func IsElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()

	var elevation tokenElevation
	var outLen uint32
	if err := windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&outLen,
	); err != nil {
		return false
	}
	return elevation.Elevated != 0
}

// RestartAsAdmin relaunches this exact exe elevated (UAC prompt) and exits
// the current instance. The app itself stays asInvoker so scheduled runs
// never trigger prompts.
func RestartAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		fmt.Sprintf(`Start-Process -FilePath "%s" -Verb RunAs`, exe))
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
