package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"nest/internal/classifier"
	"nest/internal/config"
	"nest/internal/engine"
	"nest/internal/knownfolders"
	"nest/internal/organizer"
	"nest/internal/scheduler"
)

const (
	// AppVersion is shown in the About card.
	AppVersion = "1.2.0"
	// AppAuthor must stay credited wherever NEST is presented.
	AppAuthor = "DERYK RYLOY"
)

// App is the Wails backend exposed to the React frontend.
type App struct {
	ctx context.Context
}

// NewApp creates the App struct.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Dashboard is the single payload the home screen needs.
type Dashboard struct {
	Sources        []string              `json:"sources"`
	KnownFolders   []knownfolders.Folder `json:"knownFolders"`
	Categories     []classifier.Category `json:"categories"`
	Schedule       scheduler.Status      `json:"schedule"`
	LastRun        *organizer.RunRecord  `json:"lastRun"`
	TotalBytes     int64                 `json:"totalBytes"`
	RunCount       int                   `json:"runCount"`
	ConfigPath     string                `json:"configPath"`
	DownloadsCount int                   `json:"downloadsCount"`
}

// About carries product and attribution info for the UI.
type About struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Author     string `json:"author"`
	ConfigPath string `json:"configPath"`
	ExePath    string `json:"exePath"`
}

// GetDashboard returns everything the dashboard renders.
func (a *App) GetDashboard() Dashboard {
	cfg := config.Load()

	history := organizer.LoadHistory()
	d := Dashboard{
		Sources:      cfg.Sources,
		KnownFolders: knownfolders.All(),
		Categories:   classifier.Categories(),
		Schedule:     currentSchedule(cfg),
		LastRun:      organizer.LatestRun(history),
		TotalBytes:   organizer.TotalBytes(),
		RunCount:     len(history),
		ConfigPath:   config.Path(),
	}
	// "In source" counts loose files across every watched folder, not the
	// cluster folders NEST created.
	for _, src := range cfg.Sources {
		entries, err := os.ReadDir(src)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				d.DownloadsCount++
			}
		}
	}
	return d
}

func currentSchedule(cfg config.Config) scheduler.Status {
	registered := scheduler.RegisteredTimes()
	return scheduler.Status{
		Enabled:     len(cfg.Times) > 0,
		Times:       cfg.Times,
		Registered:  registered,
		Instant:     cfg.Instant,
		WatchActive: scheduler.WatchActive(),
		IsAdmin:     scheduler.IsElevated(),
		Detail:      "Managed through Windows Task Scheduler",
	}
}

// RestartAsAdmin relaunches NEST elevated and quits this instance.
func (a *App) RestartAsAdmin() error {
	return scheduler.RestartAsAdmin()
}

// RevealFile selects a file in File Explorer.
func (a *App) RevealFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return exec.Command("explorer", "/select,"+path).Start()
}

// GetAbout returns product info and credits.
func (a *App) GetAbout() About {
	exe, _ := os.Executable()
	return About{
		Name:       "NEST File Organizer",
		Version:    AppVersion,
		Author:     AppAuthor,
		ConfigPath: config.Path(),
		ExePath:    exe,
	}
}

// GetConfig returns the saved settings.
func (a *App) GetConfig() config.Config {
	return config.Load()
}

// SaveConfig persists sources, rules and theme. Automation fields are owned by
// SaveAutomation and are copied from the stored config, so this method can
// never desync Windows Task Scheduler.
func (a *App) SaveConfig(cfg config.Config) error {
	current := config.Load()
	cfg.Times = current.Times
	cfg.Enabled = current.Enabled
	cfg.ScheduleTime = current.ScheduleTime
	cfg.Instant = current.Instant
	cfg.InstantDelay = current.InstantDelay
	return cfg.Save()
}

// GetSchedule returns the current automation state.
func (a *App) GetSchedule() scheduler.Status {
	return currentSchedule(config.Load())
}

// SaveAutomation persists the whole automation page (times, instant mode,
// delay, folders and smart rules) and then syncs Windows Task Scheduler. If
// Windows refuses — typically instant mode without elevation — the previous
// config and tasks are restored so the UI never claims automation that is not
// actually registered.
func (a *App) SaveAutomation(cfg config.Config) (config.Config, error) {
	prev := config.Load()
	next := cfg.Normalized()

	if err := next.Save(); err != nil {
		return prev, err
	}
	if err := scheduler.Sync(next.Times); err != nil {
		rollback(prev)
		return prev, err
	}
	// Recreating an already-correct logon task would need elevation for no
	// reason, so only touch it when the state changed or it went missing.
	if next.Instant != prev.Instant || (next.Instant && !scheduler.WatchActive()) {
		if err := scheduler.SetWatch(next.Instant); err != nil {
			rollback(prev)
			if next.Instant {
				return prev, fmt.Errorf("instant mode could not be registered (%v) — restart NEST as administrator once, then save again", err)
			}
			return prev, err
		}
	}
	if next.Instant {
		// A logon task does not pick up new folders or delays by itself, so
		// restart the watcher to make the saved settings live immediately.
		scheduler.RestartWatch()
	}
	return next, nil
}

// rollback restores the on-disk config and the scheduled tasks after a failed
// save.
func rollback(prev config.Config) {
	_ = prev.Save()
	_ = scheduler.Sync(prev.Times)
	_ = scheduler.SetWatch(prev.Instant)
}

// GetCategories returns the classifier registry.
func (a *App) GetCategories() []classifier.Category {
	return classifier.Categories()
}

// GetKnownFolders returns the real Windows known folders.
func (a *App) GetKnownFolders() []knownfolders.Folder {
	return knownfolders.All()
}

// PreviewOrganize runs a dry-run across every configured source.
func (a *App) PreviewOrganize() []organizer.Report {
	return engine.RunAll(config.Load(), false)
}

// RunOrganize applies the moves and records undoable history entries. The UI
// does a single pass: the user is present and sees which downloads are still
// in progress, so there is no reason to freeze the window for a minute.
func (a *App) RunOrganize() []organizer.Report {
	return engine.RunAll(config.Load(), true)
}

// GetHistory returns previous runs, newest first.
func (a *App) GetHistory() []organizer.RunRecord {
	return organizer.LoadHistory()
}

// UndoLast reverts the most recent run across every source it touched.
func (a *App) UndoLast() (organizer.UndoResult, error) {
	return organizer.UndoLast()
}

// ClearHistory deletes the run log, resetting the stats.
func (a *App) ClearHistory() error {
	return organizer.ClearHistory()
}

// PickFolder opens a native folder picker.
func (a *App) PickFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose a folder for NEST to organize",
	})
}

// OpenFolder reveals a path in File Explorer.
func (a *App) OpenFolder(path string) error {
	if path == "" {
		return fmt.Errorf("no folder selected")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return exec.Command("explorer", path).Start()
}
