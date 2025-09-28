// cmd/glint/main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MadeByDoug/glint/internal/app/infra/logging"
	"github.com/MadeByDoug/glint/internal/app/linter/reporting"
)

type cliFlags struct {
	dir         string
	configPath  string
	format      string
	devLogLevel string
}

// main is the application entry point.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {

	// Parse command-line flags.
	flags, err := parseFlags()
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Initialize the developer logger.
	if err := logging.Initialize(flags.devLogLevel, flags.format); err != nil {
		return fmt.Errorf("invalid 'dev-log' flag provided: %w", err)
	}

	// From this point on, the logger is configured and ready.
	devLog := logging.Get()
	devLog.Info().Msg("logger is configured, starting application execution")

	reporter, err := reporting.NewReporter(flags.format)
	if err != nil {
		return fmt.Errorf("invalid format specified: %w", err)
	}

	// sample linter reports
	results := []reporting.Issue{
		{File: "main.go", Line: 42, Column: 5, RuleID: "unused-var", Severity: reporting.SeverityWarning, Message: "Variable 'x' is unused"},
		{File: "utils.go", Line: 101, Column: 1, RuleID: "high-complexity", Severity: reporting.SeverityError, Message: "Cyclomatic complexity is too high (15)"},
	}

	if err := reporter.Report(os.Stdout, results); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	devLog.Debug().
		Str("dir", flags.dir).
		Str("configPath", flags.configPath).
		Msg("parsed flags")

	// If everything succeeds, return nil for no error.
	return nil
}

// parseFlags handles simple command-line arguments.
func parseFlags() (cliFlags, error) {

	// Use flag.NewFlagSet to allow for better testing and error handling.
	fs := flag.NewFlagSet("glint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dirFlag := fs.String("dir", ".", "Root directory to lint directory names in")
	configPathFlag := fs.String("config", "", "Path to a glint config file")

	formatFlag := fs.String("format", "text", "Output format (text, json)")
	devLogLevelFlag := fs.String("dev-log", "disabled", "Log level for development (trace, debug, info, disabled)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cliFlags{}, fmt.Errorf("parse args: %w", err)
	}

	return cliFlags{
		dir:         *dirFlag,
		configPath:  *configPathFlag,
		format:      *formatFlag,
		devLogLevel: *devLogLevelFlag,
	},

	nil

}

