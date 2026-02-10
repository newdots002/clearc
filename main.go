package main

import (
	"context"
	"embed"

	"clearc/internal/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var appCtx context.Context

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create system tray
	trayApp := tray.New()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "ClearC",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 13, A: 1},
		OnStartup: func(ctx context.Context) {
			appCtx = ctx
			app.startup(ctx)

			// Setup tray callbacks
			trayApp.SetCallbacks(
				func() { // onShow
					runtime.WindowShow(ctx)
					runtime.WindowSetAlwaysOnTop(ctx, true)
					runtime.WindowSetAlwaysOnTop(ctx, false)
				},
				func() { // onQuickScan
					runtime.WindowShow(ctx)
					runtime.EventsEmit(ctx, "navigate", "scan")
					runtime.EventsEmit(ctx, "startScan")
				},
				func() { // onClean
					runtime.WindowShow(ctx)
					runtime.EventsEmit(ctx, "navigate", "scan")
					runtime.EventsEmit(ctx, "startClean")
				},
				func() { // onSettings
					runtime.WindowShow(ctx)
					runtime.EventsEmit(ctx, "navigate", "settings")
				},
				func() { // onQuit
					runtime.Quit(ctx)
				},
			)

			// Start tray in background
			go trayApp.Run()
		},
		OnShutdown: func(ctx context.Context) {
			app.shutdown(ctx)
			trayApp.Quit()
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// Minimize to tray instead of closing if configured
			if app.config.MinimizeToTray {
				runtime.WindowHide(ctx)
				return true // Prevent close
			}
			return false
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
