//go:build darwin

package termwindow

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	fallbackRows = 62
	fallbackCols = 210
)

// platformMaximize resizes and positions the frontmost window of the
// frontmost application to fill the desktop work area. It returns the
// previous rectangle so it can be restored, or nil if it could not be
// captured (in which case restore falls back to an ANSI grid resize).
func platformMaximize() (*Rect, error) {
	prev, err := frontWindowRect()
	if err != nil {
		return nil, err
	}

	bounds, err := desktopBounds()
	if err != nil {
		// Still managed to capture the old rect; restore is possible.
		_ = ansiResize(fallbackRows, fallbackCols)
		return prev, nil
	}

	target := &Rect{
		X: bounds[0],
		Y: bounds[1],
		W: bounds[2] - bounds[0],
		H: bounds[3] - bounds[1],
	}
	if err := setWindowRect(target); err != nil {
		_ = ansiResize(fallbackRows, fallbackCols)
	}
	return prev, nil
}

func platformRestore(prev *Rect) error {
	if prev != nil {
		if err := setWindowRect(prev); err == nil {
			return nil
		}
	}
	return ansiResize(24, 100)
}

func osascript(script string) (string, error) {
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// frontWindowRect reads the frontmost window's position and size.
func frontWindowRect() (*Rect, error) {
	posOut, err := osascript(`tell application "System Events" to tell (first process whose frontmost is true) to get position of front window`)
	if err != nil {
		return nil, fmt.Errorf("read window position: %w", err)
	}
	sizeOut, err := osascript(`tell application "System Events" to tell (first process whose frontmost is true) to get size of front window`)
	if err != nil {
		return nil, fmt.Errorf("read window size: %w", err)
	}
	pos, err := parsePoint(posOut)
	if err != nil {
		return nil, err
	}
	size, err := parsePoint(sizeOut)
	if err != nil {
		return nil, err
	}
	return &Rect{X: pos[0], Y: pos[1], W: size[0], H: size[1]}, nil
}

// desktopBounds returns the full desktop bounds {left, top, right, bottom}.
func desktopBounds() ([]int, error) {
	out, err := osascript(`tell application "Finder" to get bounds of window of desktop`)
	if err != nil {
		return nil, fmt.Errorf("read desktop bounds: %w", err)
	}
	return parsePoint(out)
}

func setWindowRect(r *Rect) error {
	script := fmt.Sprintf(
		`tell application "System Events" to tell (first process whose frontmost is true) to set {position, size} of front window to {%d, %d, %d, %d}`,
		r.X, r.Y, r.W, r.H,
	)
	if _, err := osascript(script); err != nil {
		return fmt.Errorf("set window rect: %w", err)
	}
	return nil
}

func parsePoint(s string) ([]int, error) {
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected point %q", s)
	}
	out := make([]int, 2)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("parse point %q: %w", s, err)
		}
		out[i] = n
	}
	return out, nil
}
