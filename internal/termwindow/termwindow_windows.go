//go:build windows

package termwindow

const (
	fallbackRows = 62
	fallbackCols = 210
)

// platformMaximize grows the console grid via ANSI. Windows Terminal,
// ConHost (Windows 10+) and many third-party terminals accept the resize
// escape sequence from an application.
func platformMaximize() (*Rect, error) {
	if err := ansiResize(fallbackRows, fallbackCols); err != nil {
		return nil, err
	}
	return nil, nil
}

func platformRestore(prev *Rect) error {
	return ansiResize(24, 100)
}
