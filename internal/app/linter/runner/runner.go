// internal/app/linter/runner/runner.go
package runner

import (
	"fmt"

	"github.com/MadeByDoug/glint/internal/app/infra/config"
	"github.com/MadeByDoug/glint/internal/app/infra/logging"
	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
	"github.com/MadeByDoug/glint/internal/app/linter/check"
	"github.com/MadeByDoug/glint/internal/app/linter/selector"
)

// Runner orchestrates the execution of linting rules and their checks.
type Runner struct {
	selectionCtx selector.Context
	cfg          *config.Config
}

// New creates a new Runner.
func New(cfg *config.Config, selectionCtx selector.Context) *Runner {
	return &Runner{
		cfg:          cfg,
		selectionCtx: selectionCtx,
	}
}

// Run executes the sorted rules and returns a collection of all issues found.
func (r *Runner) Run(rules []config.RuleConfig) ([]reporting.Issue, error) {
	devLog := logging.Get()
	var allIssues []reporting.Issue

	artifacts := make([]selector.Artifact, 0, len(r.selectionCtx.Artifacts))
	for _, art := range r.selectionCtx.Artifacts {
		artifacts = append(artifacts, art)
	}

	if len(artifacts) == 0 {
		devLog.Warn().Msg("no artifacts were selected for linting; exiting early")
		return nil, nil
	}

	for _, rule := range rules {
		devLog.Debug().Str("rule.id", rule.ID).Msg("executing rule")

		for _, checkCfg := range rule.Checks {
			devLog.Debug().
				Str("rule.id", rule.ID).
				Str("check.id", checkCfg.ID).
				Str("check.uses", checkCfg.Uses).
				Msg("evaluating check")

			checkImpl, err := check.New(checkCfg.Uses)
			if err != nil {
				return nil, fmt.Errorf("failed to create check for rule '%s': %w", rule.ID, err)
			}

			// Use a type assertion to ensure the check is a file-based check.
			if fileCheck, ok := checkImpl.(check.FileCheck); ok {
				fileCtx := check.FileExecutionContext{
					ProjectRoot: r.selectionCtx.Root,
					Config:      r.cfg,
					CheckConfig: checkCfg,
					Artifacts:   artifacts,
				}

				issues, err := fileCheck.ExecuteOnFiles(fileCtx)
				if err != nil {
					return nil, fmt.Errorf(
						"error executing check '%s' in rule '%s': %w",
						checkCfg.ID, rule.ID, err,
					)
				}
				allIssues = append(allIssues, issues...)
			} else {
				// A check at this top level must be a FileCheck.
				return nil, fmt.Errorf(
					"configuration error: check '%s' (uses: %s) in rule '%s' is not a file-based check",
					checkCfg.ID, checkCfg.Uses, rule.ID,
				)
			}
		}
	}

	return allIssues, nil
}

// collectArtifacts converts the artifacts map from the context into a slice.
func (r *Runner) collectArtifacts() []selector.Artifact {
	artifacts := make([]selector.Artifact, 0, len(r.selectionCtx.Artifacts))
	for _, art := range r.selectionCtx.Artifacts {
		artifacts = append(artifacts, art)
	}
	return artifacts
}
