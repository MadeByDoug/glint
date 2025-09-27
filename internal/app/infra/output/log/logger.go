// internal/app/infra/output/log/logger.go
package log

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/level"
)

type ctxKey string

func ParseLevel(s string) level.Level {
	trimmed := strings.TrimSpace(s)
	if strings.EqualFold(trimmed, "debug") {
		return level.Info
	}
	lvl, err := level.Parse(trimmed)
	if err != nil {
		return level.Warn
	}
	return lvl
}

// Logger provides structured logging with optional base fields and context.
type Logger struct {
	format     string
	writer     io.Writer
	baseFields map[string]any
	ctx        context.Context
	level      level.Level
}

func New(format string, writer io.Writer) *Logger {
	return &Logger{format: format, writer: writer, level: level.Warn}
}

func (l *Logger) SetLevel(lvl level.Level) *Logger {
	cp := l.clone()
	cp.level = lvl
	return cp
}

// With returns a shallow-cloned logger enriched with a key/value.
func (l *Logger) With(key string, val any) *Logger {
	cp := l.clone()
	if cp.baseFields == nil {
		cp.baseFields = make(map[string]any, 1)
	}
	cp.baseFields[key] = val
	return cp
}

// WithContext attaches a context for correlation (optional).
func (l *Logger) WithContext(ctx context.Context) *Logger {
	cp := l.clone()
	cp.ctx = ctx
	return cp
}

func (l *Logger) Info(message string) {
	if l.level != level.Info {
		return
	}
	l.write("info", "", message)
}

func (l *Logger) Error(message string) {
	if level.Rank(l.level) < level.Rank(level.Warn) {
		return
	}
	l.write("error", "", message)
}

// Fatal no longer exits; caller decides. Returns an error for convenience.
func (l *Logger) Fatal(message string) error {
	l.write("fatal", "", message)
	return errors.New(message)
}

func tryCopyString(ctx context.Context, key string, dst map[string]any) {
	if ctx == nil {
		return
	}

	v := ctx.Value(ctxKey(key))
	s, ok := v.(string)
	if !ok || s == "" {
		return
	}

	dst[key] = s
}

func (l *Logger) write(severity, code, message string) {
	if l.writer == nil || l.writer == io.Discard {
		return
	}

	if err := l.writeFormatted(severity, code, message); err != nil && l.writer != os.Stderr {
		_, _ = fmt.Fprintf(os.Stderr, "glint: logger write error: %v\n", err)
	}
}

func (l *Logger) writeFormatted(severity, code, message string) error {
	switch l.format {
	case "json":
		return l.writeJSON(severity, code, message)
	default:
		return l.writeText(severity, code, message)
	}
}

func (l *Logger) writeJSON(severity, code, message string) error {
	out := map[string]any{
		"event":    "log",
		"severity": severity,
		"code":     code,
		"message":  message,
	}
	for k, v := range l.collectMeta() {
		out[k] = v
	}
	if err := json.NewEncoder(l.writer).Encode(out); err != nil {
		return fmt.Errorf("encode log entry: %w", err)
	}
	return nil
}

func (l *Logger) writeText(severity, code, message string) error {
	header := severity
	if code != "" {
		header = fmt.Sprintf("%s[%s]", severity, code)
	}

	if _, err := fmt.Fprintf(l.writer, "%s: %s", header, message); err != nil {
		return fmt.Errorf("write log header: %w", err)
	}

	if err := l.writeTextMeta(l.collectMeta()); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(l.writer); err != nil {
		return fmt.Errorf("write log newline: %w", err)
	}
	return nil
}

func (l *Logger) writeTextMeta(meta map[string]any) error {
	if len(meta) == 0 {
		return nil
	}

	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := fmt.Fprintf(l.writer, " %s=%v", k, meta[k]); err != nil {
			return fmt.Errorf("write log meta %q: %w", k, err)
		}
	}
	return nil
}

func (l *Logger) clone() *Logger {
	cp := *l
	if l.baseFields != nil {
		cp.baseFields = make(map[string]any, len(l.baseFields))
		for k, v := range l.baseFields {
			cp.baseFields[k] = v
		}
	}
	return &cp
}

// collectMeta merges base fields and a few context-derived IDs (if present).
func (l *Logger) collectMeta() map[string]any {
	out := map[string]any{}
	for k, v := range l.baseFields {
		out[k] = v
	}
	// Known context keys (optional): string IDs commonly used in tracing
	if l.ctx != nil {
		tryCopyString(l.ctx, "trace_id", out)
		tryCopyString(l.ctx, "span_id", out)
		tryCopyString(l.ctx, "correlation_id", out)
	}
	return out
}
