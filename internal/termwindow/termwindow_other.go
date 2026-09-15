//go:build !darwin && !windows

package termwindow

import (
	"os/exec"
)

const (
	fallbackRows = 62
	fallbackCols = 210
)

// platformMaximize first tries an EWMH-compatible window manager (wmctrl) to
// maximize the focused window; otherwise it grows the terminal grid via ANSI.
func platformMaximize() (*Rect, error) {
	if _, err := exec.LookPath("wmctrl"); err == nil {
		if err := exec.Command("wmctrl", "-r", ":ACTIVE:", "-b", "add,maximized_vert,maximized_horz").Run(); err == nil {
			return nil, nil
		}
	}
	if err := ansiResize(fallbackRows, fallbackCols); err != nil {
		return nil, err
	}
	return nil, nil
}

func platformRestore(prev *Rect) error {
	if _, err := exec.LookPath("wmctrl"); err == nil {
		if err := exec.Command("wmctrl", "-r", ":ACTIVE:", "-b", "remove,maximized_vert,maximized_horz").Run(); err == nil {
			return nil
		}
	}
	return ansiResize(24, 100)
}
