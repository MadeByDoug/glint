package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Result captures the outcome of validating a document.
type Result struct {
	Diagnostics []Diagnostic
	OK          bool
}

// Validate runs the compiled schema against the supplied Markdown bytes.
func Validate(md []byte, compiled *CompiledSchema) Result {
	if compiled == nil {
		return Result{Diagnostics: []Diagnostic{{
			Code:     "MD900",
			Severity: severityError,
			Message:  "no schema provided",
		}}}
	}

	doc, err := ParseMarkdown(md)
	if err != nil {
		return Result{Diagnostics: []Diagnostic{{
			Code:     "MD901",
			Severity: severityError,
			Message:  fmt.Sprintf("parse markdown: %v", err),
		}}, OK: false}
	}

	sections := CollectSections(doc)
	rc := RuleContext{
		CaseSensitive: compiled.Schema.Document.CaseSensitive,
		Normalize:     compiled.Normalizer,
		Buf:           md,
	}

	diags := validateSections(rc, compiled, sections, md)
	return Result{Diagnostics: diags, OK: len(filterErrors(diags)) == 0}
}

func validateSections(rc RuleContext, compiled *CompiledSchema, sections []SectionBlock, md []byte) []Diagnostic {
	used := make([]bool, len(sections))
	var diags []Diagnostic

	lastIdx := -1
	type matched struct {
		section *CompiledSection
		block   SectionBlock
		idx     int
	}

	var matches []matched
	missingSections := make([]Section, 0)
	for i := range compiled.Sections {
		compSec := &compiled.Sections[i]
		sec := compSec.Section
		idx, block, ok := findSectionBlock(rc, sections, used, sec.Heading)
		if !ok {
			missingSections = append(missingSections, sec)
			continue
		}
		used[idx] = true

		if compiled.Schema.Document.StrictOrder && idx <= lastIdx {
			diags = append(diags, Diagnostic{
				Code:     "MD101",
				Severity: severityError,
				Section:  sec.Name,
				Message:  "section appears out of order",
			})
		}
		lastIdx = idx
		matches = append(matches, matched{section: compSec, block: block, idx: idx})
	}

	if !compiled.Schema.Document.AllowExtraSections {
		diags = append(diags, extraSectionDiagnostics(rc, sections, used, md)...)
	}

	fulfilled := make(map[string]bool)
	sectionHasErrors := make(map[string]bool)

	for _, m := range matches {
		body := wrapNodes(m.block.Body, md)
		remaining := body
		hasError := false
		for _, rule := range m.section.Rules {
			stepDiags, leftover := rule.Validate(rc, m.section.Section.Name, remaining)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				if !hasError {
					for _, sd := range stepDiags {
						if sd.Severity == severityError {
							hasError = true
							break
						}
					}
				}
			}
			remaining = leftover
		}
		if !hasError {
			fulfilled[m.section.Section.Name] = true
		}
		sectionHasErrors[m.section.Section.Name] = hasError
	}

	// dependency checks
	for _, m := range matches {
		depends := m.section.Section.Requires
		if len(depends) == 0 {
			continue
		}
		for _, req := range depends {
			if fulfilled[req] {
				continue
			}
			line := 0
			if lines := m.block.Heading.Lines(); lines != nil && lines.Len() > 0 {
				line = lineNumber(md, lines.At(0).Start)
			}
			msg := fmt.Sprintf("section %q requires %q to be satisfied", m.section.Section.Name, req)
			if sectionHasErrors[req] {
				msg = fmt.Sprintf("section %q requires %q to pass without errors", m.section.Section.Name, req)
			}
			diags = append(diags, Diagnostic{
				Code:     "MD105",
				Severity: severityError,
				Section:  m.section.Section.Name,
				Message:  msg,
				Line:     line,
			})
		}
	}

	for _, sec := range missingSections {
		deps := sec.Requires
		if len(deps) > 0 {
			ready := true
			for _, req := range deps {
				if !fulfilled[req] {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
		}
		diags = append(diags, Diagnostic{
			Code:     "MD100",
			Severity: severityError,
			Section:  sec.Name,
			Message:  fmt.Sprintf("missing section heading %q (level %d)", sec.Heading.Text, sec.Heading.Level),
			Line:     0,
		})
	}

	return diags
}

func findSectionBlock(rc RuleContext, sections []SectionBlock, used []bool, heading SectionHeading) (int, SectionBlock, bool) {
	want := normalizeHeading(rc, heading.Text)
	for i, block := range sections {
		if used[i] {
			continue
		}
		h := block.Heading
		if h == nil || h.Level != heading.Level {
			continue
		}
		got := normalizeHeading(rc, HeadingText(h, rc.Buf))
		if headingsEqual(rc, want, got) {
			return i, block, true
		}
	}
	return -1, SectionBlock{}, false
}

func normalizeHeading(rc RuleContext, s string) string {
	if rc.Normalize != nil {
		s = rc.Normalize(s)
	}
	return s
}

func headingsEqual(rc RuleContext, want, got string) bool {
	if rc.CaseSensitive {
		return want == got
	}
	return strings.EqualFold(want, got)
}

func extraSectionDiagnostics(rc RuleContext, sections []SectionBlock, used []bool, md []byte) []Diagnostic {
	var diags []Diagnostic
	for i, block := range sections {
		if used[i] {
			continue
		}
		h := block.Heading
		if h == nil {
			continue
		}
		name := HeadingText(h, md)
		line := 0
		if lines := h.Lines(); lines != nil && lines.Len() > 0 {
			line = lineNumber(md, lines.At(0).Start)
		}
		diags = append(diags, Diagnostic{
			Code:     "MD102",
			Severity: severityError,
			Section:  name,
			Message:  fmt.Sprintf("unexpected extra section: %s", name),
			Line:     line,
		})
	}
	return diags
}

func filterErrors(ds []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range ds {
		if d.Severity == severityError {
			out = append(out, d)
		}
	}
	return out
}

// --- node cursor implementation ---

type nodeCursor struct {
	node ast.Node
	buf  []byte
}

func (c nodeCursor) Kind() string {
	switch n := c.node.(type) {
	case *ast.Paragraph:
		return "paragraph"
	case *ast.List:
		return "list"
	case *ast.ListItem:
		return "listitem"
	case *ast.Text:
		return "text"
	case *ast.Emphasis:
		if n.Level == 2 {
			return "strong"
		}
		return "emphasis"
	default:
		return "other"
	}
}

func (c nodeCursor) Text() string {
	var b strings.Builder
	ast.Walk(c.node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			b.Write(t.Segment.Value(c.buf))
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

func (c nodeCursor) Children() []NodeCursor {
	var out []NodeCursor
	for child := c.node.FirstChild(); child != nil; child = child.NextSibling() {
		out = append(out, nodeCursor{node: child, buf: c.buf})
	}
	return out
}

func (c nodeCursor) Position() int {
	if c.node == nil {
		return -1
	}
	lines := c.node.Lines()
	if lines == nil || lines.Len() == 0 {
		return -1
	}
	seg := lines.At(0)
	return seg.Start
}

func wrapNodes(nodes []ast.Node, buf []byte) []NodeCursor {
	out := make([]NodeCursor, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, nodeCursor{node: n, buf: buf})
	}
	return out
}
