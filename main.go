package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"nest/internal/config"
	"nest/internal/engine"
	"nest/internal/notify"
	"nest/internal/watch"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Headless modes: NEST does its job, shows one toast, then dies. No
	// background process, no tray icon, no window.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--auto":
			runAuto()
			return
		case "--watch":
			runWatch()
			return
		}
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "NEST",
		Width:     1220,
		Height:    800,
		MinWidth:  1000,
		MinHeight: 660,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 8, G: 13, B: 26, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

// runAuto is the scheduled, windowless pass.
func runAuto() {
	cfg := config.Load()
	if len(cfg.Sources) == 0 {
		fmt.Println("NEST: no source folders configured")
		return
	}
	reports := engine.RunSettled(cfg, true)
	if moved, _ := engine.Totals(reports); moved > 0 {
		fmt.Printf("NEST: organized %d file(s)\n", moved)
		_ = notify.Organized(moved, len(cfg.Sources))
	} else {
		fmt.Println("NEST: nothing to organize")
	}
}

// runWatch idles with zero CPU until a watched folder changes, then organizes
// the newcomers, toasts once, and goes back to waiting. It exits only when its
// scheduled task is removed or the session ends.
func runWatch() {
	cfg := config.Load()
	if len(cfg.Sources) == 0 {
		fmt.Println("NEST: no source folders configured")
		return
	}
	debounce := time.Duration(cfg.InstantDelay) * time.Second
	if debounce < 10*time.Second {
		debounce = 45 * time.Second
	}
	fmt.Printf("NEST: watching %d folder(s)\n", len(cfg.Sources))
	stop := make(chan struct{})
	watch.Run(cfg.Sources, debounce, stop, func() {
		cfg := config.Load()
		reports := engine.RunSettled(cfg, true)
		if moved, _ := engine.Totals(reports); moved > 0 {
			fmt.Printf("NEST: organized %d new file(s)\n", moved)
			_ = notify.Organized(moved, len(cfg.Sources))
		}
	})
}
