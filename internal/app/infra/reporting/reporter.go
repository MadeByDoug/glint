package reporting

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	once     sync.Once
	reporter Reporter
)

// Initialize configures the global reporter based on the format.
// This should be called once at application startup.
func Initialize(format string, lintLogLevel string, lintLogSink string) error {

	var writer io.Writer
	switch lintLogSink {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	level, err := zerolog.ParseLevel(lintLogLevel)
	if err != nil {
		return fmt.Errorf("parse lint log level %q: %w", lintLogLevel, err)
	}

	baseReporter := BaseReporter{writer: writer, level: level}

	switch format {
	case "text":
		reporter = &TextReporter{BaseReporter: baseReporter}
		return nil
	case "json":
		reporter = &JsonReporter{BaseReporter: baseReporter}
		return nil
	default:
		// The NopReporter remains if the format is invalid.
		return fmt.Errorf("unknown report format: %q", format)
	}
}

// Get returns the configured reporter instance.
func Get() Reporter {
	return reporter
}

// NopReporter is a reporter that does nothing. It is used as a safe
// default before the reporter is initialized.
type NopReporter struct{}

// Report does nothing and returns no error.
func (r *NopReporter) Report(issues []Issue) error {
	return nil
}

// Severity defines the level of a lint issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue represents a single finding by the linter.
type Issue struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column"` // Optional, but good practice
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// Reporter defines the interface for any output formatter.
// It takes the destination writer and the list of issues to report.
type Reporter interface {
	Report(issues []Issue) error
}

// severityToLevel maps the application's severity levels to zerolog's levels.
func severityToLevel(severity Severity) zerolog.Level {
	switch severity {
	case SeverityError:
		return zerolog.ErrorLevel
	case SeverityWarning:
		return zerolog.WarnLevel
	case SeverityInfo:
		return zerolog.InfoLevel
	default:
		return zerolog.NoLevel
	}
}