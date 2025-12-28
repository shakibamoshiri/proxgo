package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
)

func colorForLevel(l slog.Level) string {
	if !ColorFlag {
		return ColorReset
	}

	switch l {
	case slog.LevelDebug:
		return ColorBlue
	case slog.LevelInfo:
		return ColorGreen
	case slog.LevelWarn:
		return ColorYellow
	case slog.LevelError:
		return ColorRed
	default:
		return ColorReset
	}
}

// //////////////////////////////////////////////////////////////////////////////
// Shorten absolute file path → relative
// /home/user/app/pkg/db/file.go → pkg/db/file.go
// //////////////////////////////////////////////////////////////////////////////
func shortenPath(path string) string {
	// Trim to last 2–3 components
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) > 3 {
		return filepath.Join(parts[len(parts)-3], parts[len(parts)-2], parts[len(parts)-1])
	}
	return filepath.Base(path)
}

// ─────────────────────────────────────────────────────────────────────────────
// Handler
// ─────────────────────────────────────────────────────────────────────────────
type SimpleHandler struct {
	level     slog.Level
	writer    *os.File
	addCaller bool
}

func NewSimpleHandler(level slog.Level, w *os.File, caller bool) *SimpleHandler {
	return &SimpleHandler{
		level:     level,
		writer:    w,
		addCaller: caller,
	}
}

func (h *SimpleHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level
}

func (h *SimpleHandler) Handle(_ context.Context, r slog.Record) error {
	location := ""

	attrStr := ""

	color := colorForLevel(r.Level)

	r.Attrs(func(a slog.Attr) bool {
		attrStr += fmt.Sprintf(" %s %+v", a.Key, a.Value)
		return true // keep iterating
	})

	if h.addCaller && r.PC != 0 {
		fn := runtime.FuncForPC(r.PC)
		if fn != nil {
			file, line := fn.FileLine(r.PC)
			file = shortenPath(file)
			funcName := filepath.Base(fn.Name())
			location = fmt.Sprintf("%s +%d %s%s()%s ", file, line, color, funcName, ColorReset)
			// location = fmt.Sprintf("%s:%d ", file, line)
		}
	}

	formatted := time.Now().Format("2006/01/02 15:04:05")
	_, err := fmt.Fprintf(
		h.writer,
		"%s %s%s%s %s%s%s\n",
		formatted,
		color,
		r.Level.String(),
		ColorReset,
		location,
		r.Message,
		attrStr,
	)
	return err
}

func (h *SimpleHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *SimpleHandler) WithGroup(_ string) slog.Handler {
	return h
}

func NewLogger(logLevel string) *slog.Logger {
	level := slog.LevelError
	switch logLevel {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelError + 1
	}
	return slog.New(NewSimpleHandler(level, os.Stdout, true))

}

func SimpleLogger(logLevel string) *slog.Logger {
	var h slog.Handler

	switch logLevel {
	case "error":
		// INFO: only info and above
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError,
		})

	case "warn":
		// INFO: only info and above
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})

	case "info":
		// INFO: only info and above
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	case "debug":
		// DEBUG: everything visible
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

	default:
		// No info/debug flags → discard all logs
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError + 1, // effectively hide everything
		})
	}

	logger := slog.New(h)
	// slog.SetDefault(logger)
	return logger
}
