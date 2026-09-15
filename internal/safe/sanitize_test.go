package safe

import "testing"

func TestTextKeepsPlainContent(t *testing.T) {
	in := "hello, world! 你好\nsecond line"
	if got := Text(in); got != in {
		t.Fatalf("plain content changed: %q", got)
	}
}

func TestTextStripsC0ControlsExcludingNewlines(t *testing.T) {
	in := "a\x00b\x01c\n\t\r"
	want := "abc\n\t\r"
	if got := Text(in); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTextStripsDel(t *testing.T) {
	in := "a\x7fb"
	if got := Text(in); got != "ab" {
		t.Fatalf("got %q, want %q", got, "ab")
	}
}

func TestTextStripsCSI(t *testing.T) {
	in := "before\x1b[2Jafter"
	if got := Text(in); got != "beforeafter" {
		t.Fatalf("got %q, want %q", got, "beforeafter")
	}
}

func TestTextStripsCSIWithParams(t *testing.T) {
	in := "a\x1b[38;5;196mred"
	if got := Text(in); got != "ared" {
		t.Fatalf("got %q, want %q", got, "ared")
	}
}

func TestTextStripsOSCWithBEL(t *testing.T) {
	in := "a\x1b]0;fake title\x07b"
	if got := Text(in); got != "ab" {
		t.Fatalf("got %q, want %q", got, "ab")
	}
}

func TestTextStripsOSCWithST(t *testing.T) {
	in := "a\x1b]0;fake title\x1b\\b"
	if got := Text(in); got != "ab" {
		t.Fatalf("got %q, want %q", got, "ab")
	}
}

func TestTextStripsLoneEsc(t *testing.T) {
	// ESC followed by a single byte is treated as a two-byte sequence (for
	// example cursor-save `ESC 7` or charset-shift `ESC ( D`).
	in := "a\x1bb"
	if got := Text(in); got != "a" {
		t.Fatalf("got %q, want %q", got, "a")
	}
}

func TestTextPreservesUnicodeAroundEscapes(t *testing.T) {
	in := "héllo \x1b[m" + " мир"
	if got := Text(in); got != "héllo  мир" {
		t.Fatalf("got %q, want %q", got, "héllo  мир")
	}
}

func TestTextHandlesOscAtEnd(t *testing.T) {
	in := "a\x1b]7;file:///tmp"
	if got := Text(in); got != "a" {
		t.Fatalf("got %q, want %q", got, "a")
	}
}
