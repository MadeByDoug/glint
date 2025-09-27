// internal/app/runtime/runner.go
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MrBigCode/glint/internal/app/config"
	model "github.com/MrBigCode/glint/internal/app/config/model"
	"github.com/MrBigCode/glint/internal/app/infra/output"
	customlog "github.com/MrBigCode/glint/internal/app/infra/output/log"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/fs"
	"github.com/MrBigCode/glint/internal/app/lint/plan"
)

// Options configures the runtime wiring for glint.
type Options struct {
	ErrorFormat string
	ErrorOutput string
	LogOutput   string
	DiagLevel   string
	LogLevel    string
}

// RunParams captures the inputs required to execute a lint run.
type RunParams struct {
	Dir          string
	Debug        bool
	ConfigPath   string
	EnvName      string
	CLIOverrides map[string]any
}

// Runner orchestrates the lifecycle of a glint invocation.
type Runner struct {
	logger *customlog.Logger
	rep    *reporting.Reporter
	router *output.Router
}

func New(opts *Options) (*Runner, error) {
	if opts == nil {
		return nil, fmt.Errorf("runtime options cannot be nil")
	}

	router, err := output.New(output.Config{
		ErrorFormat: opts.ErrorFormat,
		ErrorOutput: opts.ErrorOutput,
		LogOutput:   opts.LogOutput,
	})
	if err != nil {
		return nil, fmt.Errorf("open outputs: %w", err)
	}

	logger := customlog.New(opts.ErrorFormat, router.Log).
		SetLevel(customlog.ParseLevel(opts.LogLevel))

	var fmtter reporting.Formatter
	switch strings.ToLower(opts.ErrorFormat) {
	case "json":
		fmtter = reporting.NewJSONFormatter()
	default:
		fmtter = reporting.NewTextFormatter(router.IsDiagTTY)
	}

	thr, err := reporting.ParseThreshold(opts.DiagLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid diagnostics threshold: %w", err)
	}
	rep := reporting.New(fmtter, thr)

	return &Runner{logger: logger, rep: rep, router: router}, nil
}

func (r *Runner) Close() error {
	if r == nil || r.router == nil {
		return nil
	}
	if err := r.router.Close(); err != nil {
		return fmt.Errorf("close router: %w", err)
	}
	return nil
}

func (r *Runner) Run(params RunParams) (bool, error) {
	r.logger.Info("loading configuration")

	opts := model.Options{
		ConfigPath: params.ConfigPath,
		EnvName:    params.EnvName,
		EnvPrefix:  config.DefaultEnvPrefix,
		Defaults: map[string]any{
			"dir":   ".",
			"env":   "dev",
			"debug": false,
		},
	}

	appCfg, diags, err := config.Load(opts, params.CLIOverrides)
	if err != nil {
		r.rep.Emit(r.router.Diag, diags)
		return false, fmt.Errorf("load config: %w", err)
	}

	hadError := r.rep.Emit(r.router.Diag, diags)
	if hadError {
		return true, nil
	}

	r.configureDebug(appCfg, params.Debug)

	ruleError, err := r.executeLint(appCfg)
	if err != nil {
		return false, fmt.Errorf("run lint: %w", err)
	}

	return hadError || ruleError, nil
}

func (r *Runner) configureDebug(appCfg model.AppConfig, cliDebug bool) {
	if !appCfg.Debug && !cliDebug {
		return
	}

	if appCfg.Debug {
		r.logger = r.logger.With("env", appCfg.Env).With("dir", appCfg.Dir)
		r.logger.Info("debug enabled")
		r.dumpDebug("linter config (normalized):", appCfg.Linter, "failed to write debug linter config")
	}

	if cliDebug || appCfg.Debug {
		r.dumpDebug("loaded config (normalized):", appCfg, "failed to write debug loaded config")
	}
}

func (r *Runner) dumpDebug(label string, payload any, failure string) {
	if r.router == nil || r.router.Debug == nil {
		return
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return
	}

	if _, err := fmt.Fprintf(r.router.Debug, "%s\n%s\n", label, string(b)); err != nil {
		r.logger.Error(fmt.Sprintf("%s: %v", failure, err))
	}
}

func (r *Runner) executeLint(appCfg model.AppConfig) (bool, error) {
	checks, err := plan.FromConfig(appCfg.Linter)
	if err != nil {
		return false, fmt.Errorf("build plan: %w", err)
	}
	if len(checks) == 0 {
		return false, nil
	}

	reqs := lint.PlanRequirements(checks)
	tree, err := fs.BuildTreeFromFS(appCfg.Dir, fs.BuildOptions{
		IncludeFiles: reqs.IncludeFiles,
		DirsOnly:     !reqs.IncludeFiles,
	})
	if err != nil {
		return false, fmt.Errorf("build tree: %w", err)
	}

	if reqs.IncludeFiles && (reqs.NeedShallow || reqs.NeedContents) {
		if len(reqs.FileSelectors) > 0 && !reqs.NeedContents {
			if err := lint.EnrichTreeSelected(tree, appCfg.Dir, reqs.FileSelectors); err != nil {
				return false, fmt.Errorf("enrich tree (selected): %w", err)
			}
		} else {
			if err := lint.EnrichTree(tree, appCfg.Dir); err != nil {
				return false, fmt.Errorf("enrich tree: %w", err)
			}
		}
	}

	diags := lint.Lint(context.Background(), tree, checks...)
	return r.rep.Emit(r.router.Diag, diags), nil
}
