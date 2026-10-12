package main

import (
	"context"
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
	AppVersion = "1.1.0"
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
	var last *organizer.RunRecord
	if len(history) > 0 {
		last = &history[0]
	}

	d := Dashboard{
		Sources:      cfg.Sources,
		KnownFolders: knownfolders.All(),
		Categories:   classifier.Categories(),
		Schedule:     currentSchedule(cfg),
		LastRun:      last,
		TotalBytes:   organizer.TotalBytes(),
		RunCount:     len(history),
		ConfigPath:   config.Path(),
	}
	if len(cfg.Sources) > 0 {
		if entries, err := os.ReadDir(cfg.Sources[0]); err == nil {
			d.DownloadsCount = len(entries)
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

// SaveConfig persists sources, rules and theme. It never touches the
// scheduled tasks; use SaveSchedule for automation.
func (a *App) SaveConfig(cfg config.Config) error {
	return cfg.Save()
}

// GetSchedule returns the current automation state.
func (a *App) GetSchedule() scheduler.Status {
	return currentSchedule(config.Load())
}

// SaveSchedule stores the daily times and instant mode, then syncs the
// Windows Task Scheduler entries to match.
func (a *App) SaveSchedule(times []string, instant bool) error {
	cfg := config.Load()
	cfg.Times = times
	cfg.Instant = instant
	if err := cfg.Save(); err != nil {
		return err
	}
	if err := scheduler.Sync(cfg.Times); err != nil {
		return err
	}
	return scheduler.SetWatch(cfg.Instant)
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

// RunOrganize applies the moves and records undoable history entries.
func (a *App) RunOrganize() []organizer.Report {
	return engine.RunSettled(config.Load(), true)
}

// GetHistory returns previous runs, newest first.
func (a *App) GetHistory() []organizer.RunRecord {
	return organizer.LoadHistory()
}

// UndoLast reverts the most recent run.
func (a *App) UndoLast() (organizer.RunRecord, error) {
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
	return exec.Command("explorer", path).Start()
}
