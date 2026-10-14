package gui

import (
	"os"
	"path/filepath"
	"runtime"

	"go-chat/internal/config"
)

// resolvePaths resolves relative paths in the config to be relative to the config directory.
// This ensures that when the GUI is launched from a shortcut, data is stored in the proper
// user data directory rather than the working directory.
func resolvePaths(cfg *config.Config, configPath string) {
	// If configPath is empty, it means the default location was used.
	// In that case, we use the standard config directory.
	var configDir string
	if configPath == "" {
		configDir = getAppConfigDir()
	} else {
		configDir = filepath.Dir(configPath)
	}

	// Resolve database path
	if !filepath.IsAbs(cfg.Database.Path) {
		cfg.Database.Path = filepath.Join(configDir, cfg.Database.Path)
	}

	// Resolve downloads path
	if !filepath.IsAbs(cfg.Downloads.Path) {
		cfg.Downloads.Path = filepath.Join(configDir, cfg.Downloads.Path)
	}

	// If no log file is specified, put it in a logs subdirectory
	if cfg.Logging.File == "" {
		logDir := filepath.Join(configDir, "logs")
		// Ensure the logs directory exists
		_ = os.MkdirAll(logDir, 0700)
		cfg.Logging.File = filepath.Join(logDir, "go-chat.log")
	} else if !filepath.IsAbs(cfg.Logging.File) {
		// If log file is relative, make it relative to config dir
		cfg.Logging.File = filepath.Join(configDir, cfg.Logging.File)
	}
}

// getAppConfigDir returns the platform-specific config directory for the app.
// This is where config.yaml should be stored.
func getAppConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/go-chat
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "go-chat")
	case "windows":
		// Windows: %APPDATA%\go-chat
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = os.Getenv("USERPROFILE") // fallback
		}
		return filepath.Join(appdata, "go-chat")
	default:
		// Linux and others: ~/.config/go-chat
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "go-chat")
	}
}

// getDefaultConfigPath returns the default path for the config file.
func getDefaultConfigPath() string {
	return filepath.Join(getAppConfigDir(), "config.yaml")
}