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
		if r.required {
			return []Diagnostic{{
				Code:     "MD001",
				Severity: severityError,
				Section:  secName,
				Message:  "expected a paragraph starting with strong text",
				Line:     0,
			}}, body
		}
		return nil, body
	}

	for i, n := range body {
		if n.Kind() != "paragraph" {
			continue
		}
		children := n.Children()
		if len(children) == 0 || children[0].Kind() != "strong" {
			line := lineNumber(rc.Buf, n.Position())
			if r.required {
				return []Diagnostic{{
					Code:     "MD002",
					Severity: severityError,
					Section:  secName,
					Message:  "first paragraph must begin with bold text",
					Line:     line,
				}}, body
			}
			return []Diagnostic{{
				Code:     "MD002",
				Severity: severityError,
				Section:  secName,
				Message:  "first paragraph must begin with bold text",
				Line:     line,
			}}, body
		}
		return nil, body[i+1:]
	}

	if r.required {
		return []Diagnostic{{
			Code:     "MD001",
			Severity: severityError,
			Section:  secName,
			Message:  "expected a paragraph starting with strong text",
			Line:     0,
		}}, body
	}
	return nil, body
}
