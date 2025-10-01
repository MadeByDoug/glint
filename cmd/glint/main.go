// cmd/glint/main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MadeByDoug/glint/internal/app/core/scheduler"
	"github.com/MadeByDoug/glint/internal/app/infra/config"
	"github.com/MadeByDoug/glint/internal/app/infra/logging"
	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
	"github.com/MadeByDoug/glint/internal/app/linter/runner"
	"github.com/MadeByDoug/glint/internal/app/linter/selector"
)

type cliFlags struct {
	dir         string
	configPath  string
	runtimeLogFormat      string
	runtimeLogLevel string
	runtimeLogSink string
	lintLogFormat      string
	lintLogLevel string
	lintLogSink string
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

	// Initialize both the runtime and lint loggers.
	if err := initialize_loggers(flags.runtimeLogFormat, flags.runtimeLogLevel, flags.runtimeLogSink, flags.lintLogFormat, flags.lintLogLevel, flags.lintLogSink); err != nil {
		return err
	}

	logger := logging.Get()
	reporter := reporting.Get()
	logger.Debug().Msg("Logging configured successfully.")

	// Load the linter configuration
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Schedule the rules running order.
	sortedRules, err := scheduler.SortRules(cfg.Rules)
	if err != nil {
		logger.Error().Err(err).Msg("failed to schedule rules")
	}
	logger.Info().Msg("rule execution plan has been created")

	// Load the selection context.
	selectionCtx, err := selector.NewContext(flags.dir, cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to build selection context")
	}
	logger.Info().Msg("rule selection context has been created")

	// Initialize the rule runner and run the rules
	linterRunner := runner.New(cfg, selectionCtx)
	logger.Debug().Msg("linter runner initialized")

	issues, err := linterRunner.Run(sortedRules)
	if err != nil {
		logger.Error().Err(err).Msg("linter execution failed")
		return fmt.Errorf("failed during linting run: %w", err)
	}

	logger.Info().Int("issues.count", len(issues)).Msg("linter run completed")

	// Report the findings and exit with an appropriate status code.
	if len(issues) > 0 {
		if err := reporter.Report(issues); err != nil {
			logger.Error().Err(err).Msg("failed to report issues")
			return fmt.Errorf("failed to report issues: %w", err)
		}
		return fmt.Errorf("%d issues found", len(issues))
	}

	logger.Info().Msg("no issues found")
	return nil

}

func initialize_loggers(runtimeLogFormat, runtimeLogLevel, runtimeLogSink, lintLogFormat, lintLogLevel, lintLogSink string) error {

	// Initialize runtime logger.
	if err := logging.Initialize(runtimeLogLevel, runtimeLogFormat, runtimeLogSink); err != nil {
		return fmt.Errorf("invalid 'log' flag provided: %w", err)
	}

	// Initialize the global reporter for user-facing output.
	if err := reporting.Initialize(lintLogFormat, lintLogLevel, lintLogSink); err != nil {
		return fmt.Errorf("invalid 'format' flag provided: %w", err)
	}

	return nil

}

// parseFlags handles simple command-line arguments.
func parseFlags() (cliFlags, error) {

	fs := flag.NewFlagSet("glint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dirFlag := fs.String("dir", ".", "Root directory to lint directory names in")
	configPathFlag := fs.String("config", "", "Path to config file (defaults to .glint.yaml in CWD)")
	runtimeLogFormatFlag := fs.String("runtime-log-format", "text", "Output format (text, json)")
	runtimeLogLevelFlag := fs.String("runtime-log-level", "error", "Log level (trace, debug, info, disabled)")
	runtimeLogSinkFlag := fs.String("runtime-log-sink", "stderr", "Log level (trace, debug, info, disabled)")
	lintLogFormatFlag := fs.String("lint-log-format", "text", "Output format (text, json)")
	lintLogLevelFlag := fs.String("lint-log-level", "error", "Log level (trace, debug, info, disabled)")
	lintLogSinkFlag := fs.String("lint-log-sink", "stdout", "Log level (trace, debug, info, disabled)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cliFlags{}, fmt.Errorf("parse args: %w", err)
	}

	return cliFlags{
		dir:                *dirFlag,
		configPath:         *configPathFlag,
		runtimeLogFormat:	*runtimeLogFormatFlag,
		runtimeLogLevel:    *runtimeLogLevelFlag,
		runtimeLogSink:     *runtimeLogSinkFlag,
		lintLogFormat:      *lintLogFormatFlag,
		lintLogLevel: 	 	*lintLogLevelFlag,
		lintLogSink: 	    *lintLogSinkFlag,
	}, nil
}
