// internal/app/linter/check/markdown_schema.go
package check

import (
	"fmt"
	"path/filepath"

	"github.com/MadeByDoug/glint/internal/app/infra/logging"
	"github.com/MadeByDoug/glint/internal/app/infra/reporting"
	"github.com/MadeByDoug/glint/internal/app/linter/markdown"
	"github.com/MadeByDoug/glint/internal/app/linter/selector"
)

// MarkdownSchemaCheck implements the FileCheck interface. It validates Markdown
// files against a structural schema and orchestrates the execution of any
// attached policies.
type MarkdownSchemaCheck struct{}

// ExecuteOnFiles runs the markdown schema validation for a collection of artifacts.
func (c *MarkdownSchemaCheck) ExecuteOnFiles(ctx FileExecutionContext) ([]reporting.Issue, error) {
	params, err := c.parseParams(ctx.CheckConfig.With)
	if err != nil {
		return nil, fmt.Errorf("invalid params for check '%s': %w", ctx.CheckConfig.ID, err)
	}

	schemaPath := filepath.Join(ctx.ProjectRoot, params.SchemaPath)
	schema, err := markdown.LoadSchemaFromFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema '%s': %w", schemaPath, err)
	}

	var allIssues []reporting.Issue
	for _, artifact := range ctx.Artifacts {
		if artifact.Type != "file" {
			continue
		}

		issues, err := c.validateArtifact(ctx, artifact, schema, params.Section)
		if err != nil {
			// Report errors that prevent validation (e.g., file not readable).
			allIssues = append(allIssues, c.createIssue(ctx.CheckConfig.ID, artifact, 0, 0, err.Error()))
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

// markdownSchemaParams holds the validated configuration for the check.
type markdownSchemaParams struct {
	SchemaPath string
	Section    string
}

// parseParams validates the 'with' map from the check config.
func (c *MarkdownSchemaCheck) parseParams(with map[string]interface{}) (markdownSchemaParams, error) {
	params := markdownSchemaParams{
		Section: "complete", // Default value as per the RFC.
	}

	schemaPathVal, ok := with["schema"]
	if !ok {
		return params, fmt.Errorf("missing required parameter 'schema'")
	}
	schemaPathStr, ok := schemaPathVal.(string)
	if !ok {
		return params, fmt.Errorf("parameter 'schema' must be a string")
	}
	params.SchemaPath = schemaPathStr

	if sectionVal, ok := with["section"]; ok {
		sectionStr, ok := sectionVal.(string)
		if !ok {
			return params, fmt.Errorf("parameter 'section' must be a string")
		}
		params.Section = sectionStr
	}

	return params, nil
}

// validateArtifact handles the validation of a single file.
func (c *MarkdownSchemaCheck) validateArtifact(ctx FileExecutionContext, artifact selector.Artifact, schema *markdown.Schema, section string) ([]reporting.Issue, error) {
	doc, err := markdown.LoadFile(artifact.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read or parse file: %w", err)
	}

	validator := markdown.NewValidator(schema, doc.Src)
	validationErrors, policyRequests := validator.Validate(doc.Root, section)

	var issues []reporting.Issue

	// 1. Convert structural validation errors into issues.
	for _, valErr := range validationErrors {
		issues = append(issues, c.createIssue(ctx.CheckConfig.ID, artifact, valErr.Line, valErr.Column, valErr.Message))
	}

	// If there are structural errors, don't proceed to policy checks.
	if len(issues) > 0 {
		return issues, nil
	}

	// 2. Execute all requested policy checks.
	policyIssues, err := c.executePolicyRequests(ctx, artifact, doc, policyRequests)
	if err != nil {
		return nil, err
	}
	issues = append(issues, policyIssues...)

	return issues, nil
}

// executePolicyRequests iterates through policy requests from the validator and runs them.
func (c *MarkdownSchemaCheck) executePolicyRequests(ctx FileExecutionContext, artifact selector.Artifact, doc *markdown.Document, requests []markdown.PolicyRequest) ([]reporting.Issue, error) {
	var issues []reporting.Issue
	devLog := logging.Get()

	for _, req := range requests {
		for _, policyName := range req.PolicyNames {
			policyCfg, ok := ctx.Config.Policies[policyName]
			if !ok {
				msg := fmt.Sprintf("undefined policy '%s' referenced in schema", policyName)
				issues = append(issues, c.createIssue(ctx.CheckConfig.ID, artifact, req.Line, req.Column, msg))
				continue
			}

			devLog.Debug().
				Str("policy.name", policyName).
				Str("policy.uses", policyCfg.Uses).
				Str("file", artifact.Path).
				Msg("executing policy check")

			policyCheckImpl, err := New(policyCfg.Uses)
			if err != nil {
				return nil, fmt.Errorf("failed to create policy check '%s': %w", policyName, err)
			}

			// Use a type assertion to ensure the policy implements PolicyCheck.
			if policy, ok := policyCheckImpl.(PolicyCheck); ok {
				policyCtx := PolicyExecutionContext{
					ProjectRoot:  ctx.ProjectRoot,
					Config:       ctx.Config,
					CheckConfig:  policyCfg,
					TargetNode:   req.Node,
					TargetSource: doc.Src,
				}

				policyIssues, err := policy.ExecuteOnNode(policyCtx)
				if err != nil {
					return nil, fmt.Errorf("error executing policy '%s': %w", policyName, err)
				}

				// Add location info to issues returned from the policy.
				for i := range policyIssues {
					policyIssues[i].File = artifact.Path
					policyIssues[i].Line = req.Line
					policyIssues[i].Column = req.Column
				}
				issues = append(issues, policyIssues...)

			} else {
				msg := fmt.Sprintf("configuration error: policy '%s' (uses: %s) does not implement the required policy interface", policyName, policyCfg.Uses)
				issues = append(issues, c.createIssue(ctx.CheckConfig.ID, artifact, req.Line, req.Column, msg))
			}
		}
	}
	return issues, nil
}

// createIssue is a helper to build a consistent reporting.Issue.
func (c *MarkdownSchemaCheck) createIssue(ruleID string, artifact selector.Artifact, line, col int, msg string) reporting.Issue {
	return reporting.Issue{
		File:     artifact.Path,
		Line:     line,
		Column:   col,
		RuleID:   ruleID,
		Severity: reporting.SeverityError,
		Message:  msg,
	}
}