package main

import (
	"context"
	"log"

	"go-chat/internal/app"
	"go-chat/internal/gui"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2"
)

func main() {
	// Create application instance with default config path
	a, err := app.New("")
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}
	defer func() {
		if err := a.Close(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Create GUI service
	guiService := gui.NewService(a, "")

	// Set up Wails application
	err = runWails(a, guiService)
	if err != nil {
		log.Fatalf("Error running Wails: %v", err)
	}
}

func runWails(a *app.App, gs *gui.Service) error {
	// Create Wails app with GUI bindings
	app := &options.App{
		Title:  "go-chat",
		Width:  1024,
		Height: 768,
		MinWidth: 800,
		MinHeight: 600,
		// Enable dark mode by default to match the existing theme
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			// Send initial data to frontend
			runtime.EventsEmit(ctx, "app-ready")
		},
		Bind: []interface{}{
			gs,
		},
		// Windows-specific options
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent: false,
		},
	}

	// Run the application
	err := wails.Run(app)
	if err != nil {
		return err
	}

	return nil
}