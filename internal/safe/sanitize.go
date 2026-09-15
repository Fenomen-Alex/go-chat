// Package safe provides defensive sanitization helpers for text that crosses
// a trust boundary (for example message or display-name content received from
// another peer) before it is stored or rendered in the terminal.
package safe

import "strings"

const (
	esc = byte(0x1b)
	bel = byte(0x07)
)

// Text strips terminal control sequences and stray C0 control bytes from a
// string while preserving newlines. This prevents ANSI/OSC injection that
// could otherwise clear the screen, move the cursor, spoof prompts or corrupt
// the TUI layout when peer-supplied content is rendered.
//
// Kept characters: \t, \n, \r and everything >= 0x20 except DEL (0x7f).
func Text(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	var i int
	for i < len(s) {
		c := s[i]
		switch {
		case c == esc:
			i = skipEscape(s, i+1)
		case c == '\t' || c == '\n' || c == '\r':
			b.WriteByte(c)
			i++
		case c < 0x20 || c == 0x7f:
			i++
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// skipEscape swallows an escape introduced at position after the ESC byte and
// returns the index of the next byte to process. It understands CSI
// (ESC [ ... final) and OSC (ESC ] ... ST/BEL) sequences, and falls back to
// skipping a single byte otherwise so a lone ESC cannot hide subsequent text.
func skipEscape(s string, i int) int {
	if i >= len(s) {
		return i
	}
	switch s[i] {
	case '[': // CSI: params + final byte in 0x40..0x7e
		i++
		for i < len(s) {
			c := s[i]
			if c >= 0x40 && c <= 0x7e {
				return i + 1
			}
			i++
		}
		return i
	case ']': // OSC: terminated by ST (ESC backslash) or BEL
		i++
		for i < len(s) {
			if s[i] == bel {
				return i + 1
			}
			if s[i] == esc && i+1 < len(s) && s[i+1] == '\\' {
				return i + 2
			}
			i++
		}
		return i
	default:
		return i + 1
	}
}
