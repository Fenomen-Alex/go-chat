package network

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

func TestReadLineBasic(t *testing.T) {
	r := bufio.NewReader(bytes.NewBufferString("hello\nworld\n"))
	line, err := readLine(r, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(line) != "hello\n" {
		t.Fatalf("got %q, want %q", line, "hello\n")
	}
}

func TestReadLineLongLineAcrossBuffer(t *testing.T) {
	// Force the high-level reader to use a small internal buffer so the line
	// spans multiple ReadSlice chunks. The original implementation reused the
	// last chunk and silently truncated the message.
	large := bytes.Repeat([]byte("x"), 20000)
	data := append(append([]byte("a"), large...), '\n')
	r := bufio.NewReaderSize(bytes.NewReader(data), 512)

	line, err := readLine(r, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(line) != len(data) {
		t.Fatalf("line length = %d, want %d", len(line), len(data))
	}
	if !bytes.Equal(line, data) {
		t.Fatal("line content does not match input")
	}
}

func TestReadLineTooLong(t *testing.T) {
	r := bufio.NewReaderSize(bytes.NewReader(bytes.Repeat([]byte("y"), 2048)), 256)
	if _, err := readLine(r, 1024); err != errLineTooLong {
		t.Fatalf("got %v, want errLineTooLong", err)
	}
}

func TestReadLineUnderMaxWithNoNewline(t *testing.T) {
	r := bufio.NewReader(bytes.NewBufferString("no newline"))
	line, err := readLine(r, 1024)
	if err != io.EOF {
		t.Fatalf("got %v, want io.EOF", err)
	}
	if string(line) != "no newline" {
		t.Fatalf("got %q", line)
	}
}
