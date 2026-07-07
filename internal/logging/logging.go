package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelTrace Level = iota - 2
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

type Logger struct {
	mu       sync.Mutex
	level    Level
	logger   *log.Logger
	file     *os.File
	rotate   bool
	dir      string
	basename string
	uiBuf    []LogEntry
}

func New(level string, path string, rotate bool) (*Logger, error) {
	l := &Logger{
		level:  parseLevel(level),
		rotate: rotate,
		uiBuf:  make([]LogEntry, 0, 100),
	}

	writers := []io.Writer{os.Stderr}

	if path != "" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		l.dir = dir
		l.basename = base

		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		l.file = f
		writers = append(writers, f)
	}

	l.logger = log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime|log.Lshortfile)
	return l, nil
}

func (l *Logger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = parseLevel(level)
}

func (l *Logger) log(level Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	if l.rotate && l.file != nil {
		l.rotateIfNeeded()
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.Output(3, fmt.Sprintf("[%s] %s", levelNames[level], msg))

	entry := LogEntry{
		Time:    time.Now(),
		Level:   levelNames[level],
		Message: msg,
	}
	if len(l.uiBuf) >= 100 {
		l.uiBuf = l.uiBuf[1:]
	}
	l.uiBuf = append(l.uiBuf, entry)
}

func (l *Logger) UIMessages() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LogEntry, len(l.uiBuf))
	copy(out, l.uiBuf)
	return out
}

func (l *Logger) Trace(format string, args ...any) { l.log(LevelTrace, format, args...) }
func (l *Logger) Debug(format string, args ...any) { l.log(LevelDebug, format, args...) }
func (l *Logger) Info(format string, args ...any)  { l.log(LevelInfo, format, args...) }
func (l *Logger) Warn(format string, args ...any)  { l.log(LevelWarn, format, args...) }
func (l *Logger) Error(format string, args ...any) { l.log(LevelError, format, args...) }

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) rotateIfNeeded() {
	info, err := l.file.Stat()
	if err != nil {
		return
	}

	if info.Size() < 10*1024*1024 {
		return
	}

	l.file.Close()

	ts := time.Now().Format("20060102-150405")
	rotated := filepath.Join(l.dir, l.basename+"."+ts)
	os.Rename(filepath.Join(l.dir, l.basename), rotated)

	f, err := os.OpenFile(filepath.Join(l.dir, l.basename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	l.file = f

	writers := []io.Writer{os.Stderr, f}
	l.logger.SetOutput(io.MultiWriter(writers...))
}

func parseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}
