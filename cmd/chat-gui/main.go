package main

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"go-chat/internal/app"
	"go-chat/internal/gui"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2"
)

func frontendAssetsDir() string {
	// Try relative to executable first (works when launched from anywhere)
	ex, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(ex)
		candidate := filepath.Join(dir, "build", "frontend")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		// Also try one level up (e.g. executable in cmd/build/)
		candidate2 := filepath.Join(dir, "..", "build", "frontend")
		if _, err := os.Stat(candidate2); err == nil {
			return candidate2
		}
	}
	// Fallback to CWD-relative (works when run from cmd/chat-gui/)
	candidate := "../build/frontend"
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return candidate
}

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
		AssetServer: &assetserver.Options{
			Assets: func() fs.FS {
				return os.DirFS(frontendAssetsDir())
			}(),
		},
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