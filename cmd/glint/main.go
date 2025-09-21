// cmd/glint/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MrBigCode/glint/internal/app/config"
	"github.com/MrBigCode/glint/internal/app/infra/output"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/runtime"
)

type cliFlags struct {
	dir         string
	debug       bool
	configPath  string
	envName     string
	errorFormat string
	errorOutput string
	logOutput   string
	diagLevel   string
	logLevel    string
}

// main is the application entry point.
func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

func run() int {
	flags := parseFlags()

	runner, err := runtime.New(runtime.Options{
		ErrorFormat: flags.errorFormat,
		ErrorOutput: flags.errorOutput,
		LogOutput:   flags.logOutput,
		DiagLevel:   flags.diagLevel,
		LogLevel:    flags.logLevel,
	})
	if err != nil {
		return bootstrapFatal(flags.errorFormat, flags.errorOutput, "F-INIT-001", err.Error())
	}

	exitCode := execute(flags, runner)
	if closeErr := runner.Close(); closeErr != nil {
		if closeCode := bootstrapFatal(flags.errorFormat, flags.errorOutput, "F-EXIT-001", fmt.Sprintf("close outputs: %v", closeErr)); exitCode == 0 {
			exitCode = closeCode
		}
	}

	return exitCode
}

func execute(flags cliFlags, runner *runtime.Runner) int {
	cliOverrides := overlaysFromFlags(flags.dir, flags.debug, flags.envName)

	hadFatalDiagnostics, err := runner.Run(runtime.RunParams{
		Dir:          flags.dir,
		Debug:        flags.debug,
		ConfigPath:   flags.configPath,
		EnvName:      flags.envName,
		CLIOverrides: cliOverrides,
	})
	if err != nil {
		return bootstrapFatal(flags.errorFormat, flags.errorOutput, "F-RUN-001", err.Error())
	}

	if hadFatalDiagnostics {
		return 1
	}

	return 0
}

// parseFlags handles simple command-line arguments.
func parseFlags() cliFlags {
	dirFlag := flag.String("dir", ".", "Root directory to lint directory names in")
	debugFlag := flag.Bool("debug", false, "Print debug info")
	configPathFlag := flag.String("config", "", "Path to a glint config file")
	envNameFlag := flag.String("env", "", "Environment name for config loading (e.g., dev/stg/prod)")
	errorFormatFlag := flag.String("error-format", "text", "Error output format: text | json")
	errorOutputFlag := flag.String("error-output", "stderr", "Error output: stderr | stdout | <filename>")
	logOutputFlag := flag.String("log-output", "", "Log output: stderr | stdout | <filename> (empty/none disables)")

	diagLevelFlag := flag.String("diag-level", "error", "Diagnostic verbosity: error | warn | note")
	logLevelFlag := flag.String("log-level", "off", "Log verbosity: off | error | info")

	flag.Parse()

	return cliFlags{
		dir:         *dirFlag,
		debug:       *debugFlag,
		configPath:  *configPathFlag,
		envName:     *envNameFlag,
		errorFormat: *errorFormatFlag,
		errorOutput: *errorOutputFlag,
		logOutput:   *logOutputFlag,
		diagLevel:   *diagLevelFlag,
		logLevel:    *logLevelFlag,
	}
}

func bootstrapFatal(format, dest, code, msg string) int {
	// Use the application's own output routing and formatting for consistency,
	// even for fatal startup errors.
	router, err := output.New(output.Config{
		ErrorFormat: format,
		ErrorOutput: dest,
		LogOutput:   "none", // Disable logging for fatal bootstrap errors.
	})
	if err != nil {
		// If the router itself fails, fall back to a raw stderr print.
		fmt.Fprintf(os.Stderr, "error[%s]: %s (and could not init output: %v)\n", code, msg, err)
		return 1
	}
	defer func() {
		if err := router.Close(); err != nil {
			// Can't use the router, it might have failed. Print to stderr as a last resort.
			fmt.Fprintf(os.Stderr, "error[F-EXIT-002]: failed to close output router: %v\n", err)
		}
	}()

	var fmtter reporting.Formatter
	switch strings.ToLower(format) {
	case "json":
		fmtter = reporting.NewJSONFormatter()
	default:
		// isTTY doesn't matter as much for a single fatal error, but we can still check.
		fmtter = reporting.NewTextFormatter(router.IsDiagTTY)
	}

	rep := reporting.New(fmtter)
	rep.Print(router.Diag, []reporting.Report{
		reporting.Error(code, msg),
	})

	return 1
}

func overlaysFromFlags(dir string, debug bool, envName string) map[string]any {
	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })

	over := map[string]any{}
	if set["dir"] {
		over[config.KeyDir] = dir
	}
	if set["debug"] {
		over[config.KeyDebug] = debug
	}
	if set["env"] {
		over[config.KeyEnv] = envName
	}
	return over
}
