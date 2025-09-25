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
	matches, missing, orderingDiags := matchSchemaSections(rc, compiled, sections, used)
	diags := append([]Diagnostic(nil), orderingDiags...)

	if !compiled.Schema.Document.AllowExtraSections {
		diags = append(diags, extraSectionDiagnostics(rc, sections, used, md)...)
	}

	fulfilled, failures, ruleDiags := validateMatchedSections(rc, matches, md)
	diags = append(diags, ruleDiags...)
	diags = append(diags, dependencyDiagnostics(rc, matches, fulfilled, failures, md)...)
	diags = append(diags, missingSectionDiagnostics(missing, fulfilled)...)

	return diags
}

type matchedSection struct {
	section *CompiledSection
	block   SectionBlock
	idx     int
}

func matchSchemaSections(rc RuleContext, compiled *CompiledSchema, sections []SectionBlock, used []bool) ([]matchedSection, []Section, []Diagnostic) {
	lastIdx := -1
	var diags []Diagnostic
	var matches []matchedSection
	var missing []Section

	for i := range compiled.Sections {
		compSec := &compiled.Sections[i]
		sec := compSec.Section
		idx, block, ok := findSectionBlock(rc, sections, used, sec.Heading)
		if !ok {
			missing = append(missing, sec)
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
		matches = append(matches, matchedSection{section: compSec, block: block, idx: idx})
	}

	return matches, missing, diags
}

func validateMatchedSections(rc RuleContext, matches []matchedSection, md []byte) (fulfilled map[string]bool, failures map[string]bool, diags []Diagnostic) {
	fulfilled = make(map[string]bool)
	failures = make(map[string]bool)

	for _, m := range matches {
		matchDiags, hasError := validateMatchedSection(rc, m, md)
		if !hasError {
			fulfilled[m.section.Section.Name] = true
		}
		failures[m.section.Section.Name] = hasError
		diags = append(diags, matchDiags...)
	}

	return fulfilled, failures, diags
}

func validateMatchedSection(rc RuleContext, m matchedSection, md []byte) ([]Diagnostic, bool) {
	remaining := wrapNodes(m.block.Body, md)
	var diags []Diagnostic
	hasError := false

	for _, rule := range m.section.Rules {
		stepDiags, leftover := rule.Validate(rc, m.section.Section.Name, remaining)
		diags = append(diags, stepDiags...)
		if !hasError && containsError(stepDiags) {
			hasError = true
		}
		remaining = leftover
	}

	return diags, hasError
}

func dependencyDiagnostics(rc RuleContext, matches []matchedSection, fulfilled, failures map[string]bool, md []byte) []Diagnostic {
	var diags []Diagnostic
	for _, m := range matches {
		diags = append(diags, dependencyDiagnosticsForMatch(rc, m, fulfilled, failures, md)...)
	}
	return diags
}

func dependencyDiagnosticsForMatch(rc RuleContext, m matchedSection, fulfilled, failures map[string]bool, md []byte) []Diagnostic {
	deps := m.section.Section.Requires
	if len(deps) == 0 {
		return nil
	}
	line := headerLine(md, m.block.Heading)
	sectionName := m.section.Section.Name

	var diags []Diagnostic
	for _, req := range deps {
		if fulfilled[req] {
			continue
		}
		msg := fmt.Sprintf("section %q requires %q to be satisfied", sectionName, req)
		if failures[req] {
			msg = fmt.Sprintf("section %q requires %q to pass without errors", sectionName, req)
		}
		diags = append(diags, Diagnostic{
			Code:     "MD105",
			Severity: severityError,
			Section:  sectionName,
			Message:  msg,
			Line:     line,
		})
	}
	return diags
}

func missingSectionDiagnostics(missing []Section, fulfilled map[string]bool) []Diagnostic {
	var diags []Diagnostic
	for _, sec := range missing {
		if !dependenciesMet(sec.Requires, fulfilled) {
			continue
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

func dependenciesMet(deps []string, fulfilled map[string]bool) bool {
	if len(deps) == 0 {
		return true
	}
	for _, req := range deps {
		if !fulfilled[req] {
			return false
		}
	}
	return true
}

func containsError(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == severityError {
			return true
		}
	}
	return false
}

func headerLine(md []byte, heading *ast.Heading) int {
	if heading == nil {
		return 0
	}
	if lines := heading.Lines(); lines != nil && lines.Len() > 0 {
		return lineNumber(md, lines.At(0).Start)
	}
	return 0
}

func findSectionBlock(rc RuleContext, sections []SectionBlock, used []bool, heading SectionHeading) (int, SectionBlock, bool) {
	want := normalizeHeading(rc, heading.Text)
	for i, block := range sections {
		if used[i] || !headingMatches(rc, block.Heading, heading.Level, want) {
			continue
		}
		return i, block, true
	}
	return -1, SectionBlock{}, false
}

func headingMatches(rc RuleContext, h *ast.Heading, level int, want string) bool {
	if h == nil || h.Level != level {
		return false
	}
	got := normalizeHeading(rc, HeadingText(h, rc.Buf))
	return headingsEqual(rc, want, got)
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
		diags = append(diags, extraSectionDiagnostic(md, block))
	}
	return diags
}

func extraSectionDiagnostic(md []byte, block SectionBlock) Diagnostic {
	h := block.Heading
	if h == nil {
		return Diagnostic{Code: "MD102", Severity: severityError, Message: "unexpected extra section"}
	}
	name := HeadingText(h, md)
	line := headerLine(md, h)
	return Diagnostic{
		Code:     "MD102",
		Severity: severityError,
		Section:  name,
		Message:  fmt.Sprintf("unexpected extra section: %s", name),
		Line:     line,
	}
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
	walker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			b.Write(t.Segment.Value(c.buf))
		}
		return ast.WalkContinue, nil
	}
	if err := ast.Walk(c.node, walker); err != nil {
		return ""
	}
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
