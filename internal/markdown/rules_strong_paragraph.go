package markdown

func init() {
	RegisterRule("strong_paragraph", newStrongParagraphRule)
}

type strongParagraphRule struct {
	required bool
}

func newStrongParagraphRule(step ContentStep) (ContentRule, error) {
	return &strongParagraphRule{required: step.Required}, nil
}

func (r *strongParagraphRule) Type() string { return "strong_paragraph" }

func (r *strongParagraphRule) Validate(rc RuleContext, secName string, body []NodeCursor) ([]Diagnostic, []NodeCursor) {
	if len(body) == 0 {
		return r.missingParagraphDiag(secName, body)
	}

	idx, para := firstNodeOfKind(body, kindParagraph)
	if para == nil {
		return r.missingParagraphDiag(secName, body)
	}

	if ok := paragraphStartsWithStrong(para); ok {
		return nil, body[idx+1:]
	}

	line := lineNumber(rc.Buf, para.Position())
	diag := Diagnostic{
		Code:     "MD002",
		Severity: severityError,
		Section:  secName,
		Message:  "first paragraph must begin with bold text",
		Line:     line,
	}
	return []Diagnostic{diag}, body
}

func (r *strongParagraphRule) missingParagraphDiag(secName string, body []NodeCursor) ([]Diagnostic, []NodeCursor) {
	if !r.required {
		return nil, body
	}
	diag := Diagnostic{
		Code:     "MD001",
		Severity: severityError,
		Section:  secName,
		Message:  "expected a paragraph starting with strong text",
		Line:     0,
	}
	return []Diagnostic{diag}, body
}

func paragraphStartsWithStrong(n NodeCursor) bool {
	if n == nil || n.Kind() != kindParagraph {
		return false
	}
	children := n.Children()
	return len(children) > 0 && children[0].Kind() == kindStrongText
}
