// Package termwindow provides best-effort, cross-platform control of the
// terminal window so a chat session can use the full visible screen without
// entering the operating system's fullscreen mode.
//
// The implementation prefers an OS window-manager approach where available
// (macOS System Events, EWMH-based window managers on Linux) and falls back
// to an ANSI terminal resize escape sequence (ESC [ 8 ; rows ; cols t) which
// the common terminal emulators translate into a larger window.
package termwindow

import (
	"fmt"
	"os"
	"sync"
)

// Rect describes a window's position and size on screen.
type Rect struct {
	X, Y, W, H int
}

// Controller toggles the current terminal window between a maximized
// ("full sized", not fullscreen) state and its previous size.
type Controller struct {
	mu     sync.Mutex
	active bool
	saved  *Rect
}

var std = &Controller{}

// Default returns the process-wide window controller.
func Default() *Controller { return std }

// Active reports whether the window is currently in the maximized state.
func (c *Controller) Active() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}

// Toggle switches between maximized and restored and returns the resulting
// state (true means maximized).
func (c *Controller) Toggle() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active {
		c.active = false
		prev := c.saved
		c.saved = nil
		return false, platformRestore(prev)
	}

	prev, pErr := platformMaximize()
	if pErr != nil {
		if err := ansiResize(fallbackRows, fallbackCols); err != nil {
			return false, fmt.Errorf("maximize window: %v (ANSI fallback also failed: %v)", pErr, err)
		}
		c.active = true
		c.saved = nil
		return true, nil
	}

	c.saved = prev
	c.active = true
	return true, nil
}

// Restore returns the window to the size captured before the last maximize.
func (c *Controller) Restore() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return false, nil
	}
	c.active = false
	prev := c.saved
	c.saved = nil
	return false, platformRestore(prev)
}

func ansiResize(rows, cols int) error {
	_, err := fmt.Fprintf(os.Stdout, "\x1b[8;%d;%dt", rows, cols)
	return err
}
