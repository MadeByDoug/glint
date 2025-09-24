package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/markdown"
)

// MarkdownSchemaConfig configures the markdown schema validator check.
type MarkdownSchemaConfig struct {
	SchemaPath string
}

// MarkdownSchemaCheck validates Markdown files against a compiled schema.
type MarkdownSchemaCheck struct {
	cfg      MarkdownSchemaConfig
	code     string
	severity reporting.Severity

	once      sync.Once
	mu        sync.Mutex
	compiled  *markdown.CompiledSchema
	loadErr   error
	schemaAbs string
}

// NewMarkdownSchemaCheck constructs the check instance.
func NewMarkdownSchemaCheck(cfg MarkdownSchemaConfig, code string, sev reporting.Severity) lint.Checker {
	return &MarkdownSchemaCheck{cfg: cfg, code: code, severity: sev}
}

func (c *MarkdownSchemaCheck) ID() string { return "check.markdownSchema" }

func (c *MarkdownSchemaCheck) Apply(context.Context, *lint.Tree) []reporting.Report {
	panic("internal error: MarkdownSchemaCheck must be wrapped in a selector")
}

func (c *MarkdownSchemaCheck) ApplyToNode(_ context.Context, n *lint.Node) []reporting.Report {
	if n.Kind != lint.File {
		return nil
	}

	abs := resolveAbsPath(n)
	if abs == "" {
		return []reporting.Report{reporting.Error(c.code, fmt.Sprintf("%s: unable to resolve absolute path", displayPath(n)))}
	}

	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		return nil
	}

	root := resolveRootPath(n)
	if root == "" {
		return []reporting.Report{reporting.Error(c.code, fmt.Sprintf("%s: unable to resolve repository root", displayPath(n)))}
	}

	if err := c.ensureSchema(root); err != nil {
		return []reporting.Report{reporting.Error(c.code, err.Error())}
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		return []reporting.Report{reporting.Error(c.code, fmt.Sprintf("%s: read file: %v", displayPath(n), err))}
	}

	if c.compiled == nil {
		return []reporting.Report{reporting.Error(c.code, "markdown schema not initialized")}
	}

	res := markdown.Validate(content, c.compiled)
	if res.OK {
		return nil
	}

	rel := displayPath(n)
	diags := make([]reporting.Report, 0, len(res.Diagnostics))
	for _, d := range res.Diagnostics {
		location := rel
		if d.Line > 0 {
			location = fmt.Sprintf("%s:%d", rel, d.Line)
		}
		msg := fmt.Sprintf("%s: %s", location, d.Message)
		code := c.code
		if d.Code != "" {
			code = fmt.Sprintf("%s/%s", c.code, d.Code)
		}
		sev := c.mergeSeverity(d.Severity)
		diags = append(diags, reporting.Report{Msg: msg, Code: code, Severity: sev})
	}

	return diags
}

func (c *MarkdownSchemaCheck) ensureSchema(rootAbs string) error {
	schemaAbs := c.resolveSchemaPath(rootAbs)
	c.once.Do(func() {
		compiled, err := markdown.LoadCompiled(schemaAbs)
		if err != nil {
			c.loadErr = fmt.Errorf("load schema %q: %w", c.cfg.SchemaPath, err)
			return
		}
		c.compiled = compiled
	})
	return c.loadErr
}

func (c *MarkdownSchemaCheck) resolveSchemaPath(rootAbs string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.schemaAbs != "" {
		return c.schemaAbs
	}
	if filepath.IsAbs(c.cfg.SchemaPath) {
		c.schemaAbs = c.cfg.SchemaPath
		return c.schemaAbs
	}
	c.schemaAbs = filepath.Join(rootAbs, filepath.FromSlash(c.cfg.SchemaPath))
	return c.schemaAbs
}

func (c *MarkdownSchemaCheck) mergeSeverity(s string) reporting.Severity {
	diag := c.severity
	switch strings.ToLower(s) {
	case "error":
		diag = reporting.SevError
	case "warn", "warning":
		diag = reporting.SevWarning
	case "note", "info":
		diag = reporting.SevNote
	default:
		return c.severity
	}

	if reporting.AtLeast(diag, c.severity) {
		return diag
	}
	return c.severity
}

func resolveAbsPath(n *lint.Node) string {
	if n == nil {
		return ""
	}
	if v, ok := n.Meta["absPath"].(string); ok && v != "" {
		return v
	}
	rel := strings.TrimPrefix(n.Path(), "/")
	if rel == "" {
		return ""
	}
	root := resolveRootPath(n)
	if root == "" {
		return ""
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

func resolveRootPath(n *lint.Node) string {
	for cur := n; cur != nil; cur = cur.Parent {
		if cur.Parent == nil {
			if v, ok := cur.Meta["absPath"].(string); ok {
				return v
			}
			return ""
		}
	}
	return ""
}

func displayPath(n *lint.Node) string {
	if n == nil {
		return ""
	}
	if rel, ok := n.Meta["relPath"].(string); ok && rel != "" {
		return rel
	}
	return strings.TrimPrefix(n.Path(), "/")
}

var _ lint.NodeChecker = (*MarkdownSchemaCheck)(nil)
