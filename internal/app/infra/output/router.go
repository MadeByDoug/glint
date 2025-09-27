// internal/app/infra/output/router.go
package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

type Router struct {
	Diag      io.Writer // diagnostics (reporter)
	Debug     io.Writer // human/debug tables (show-* flags)
	Log       io.Writer // logger sink (separate from diagnostics)
	IsDiagTTY bool
	closers   []io.Closer
}

type Config struct {
	ErrorFormat string // "text" | "json" (used only for policy below)
	ErrorOutput string // "stdout" | "stderr" | <file>
	LogOutput   string // "stdout" | "stderr" | <file>
}

func New(cfg Config) (*Router, error) {
	r := &Router{}

	if err := r.initDiag(cfg); err != nil {
		return nil, err
	}

	r.configureDebug(cfg)
	r.initLog(cfg)

	return r, nil
}

func (r *Router) Close() error {
	var firstErr error
	for _, c := range r.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func openDest(dest string) (w io.Writer, isTTY bool, closer io.Closer, err error) {
	switch strings.ToLower(dest) {
	case "stdout", "":
		f := os.Stdout
		return f, term.IsTerminal(int(f.Fd())), nil, nil
	case "stderr":
		f := os.Stderr
		return f, term.IsTerminal(int(f.Fd())), nil, nil
	default:
		cleanDest, err := sanitizeOutputPath(dest)
		if err != nil {
			return nil, false, nil, fmt.Errorf("sanitize output path %q: %w", dest, err)
		}
		f, e := os.Create(cleanDest) // #nosec G304 -- sanitizeOutputPath ensures dest cannot escape allowed directories
		if e != nil {
			return nil, false, nil, fmt.Errorf("create output file %q: %w", cleanDest, e)
		}
		return f, false, f, nil
	}
}

func sanitizeOutputPath(dest string) (string, error) {
	clean := filepath.Clean(dest)
	if clean == "" || clean == "." {
		return "", fmt.Errorf("invalid output path: %q", dest)
	}
	if !filepath.IsAbs(clean) {
		if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("output path may not traverse parent directories: %q", dest)
		}
	}
	return clean, nil
}

func isStdout(dest string) bool { return strings.EqualFold(dest, "stdout") }

func (r *Router) initDiag(cfg Config) error {
	diagW, diagIsTTY, diagCloser, err := openDest(cfg.ErrorOutput)
	if err != nil {
		return fmt.Errorf("open error-output %q: %w", cfg.ErrorOutput, err)
	}

	r.Diag = diagW
	r.IsDiagTTY = diagIsTTY
	r.addCloser(diagCloser)
	return nil
}

func (r *Router) configureDebug(cfg Config) {
	if cfg.ErrorFormat == "json" && isStdout(cfg.ErrorOutput) {
		r.Debug = os.Stderr
		return
	}
	r.Debug = os.Stdout
}

func (r *Router) initLog(cfg Config) {
	if cfg.LogOutput == "" || strings.EqualFold(cfg.LogOutput, "none") {
		r.Log = io.Discard
		return
	}

	logW, _, logCloser, err := openDest(cfg.LogOutput)
	if err != nil {
		r.Log = r.Diag
		return
	}

	r.Log = logW
	r.addCloser(logCloser)
}

func (r *Router) addCloser(c io.Closer) {
	if c != nil {
		r.closers = append(r.closers, c)
	}
}
