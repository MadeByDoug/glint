package markdown

import "fmt"

func init() {
	RegisterRule("list", newListRule)
}

type listRule struct {
	required bool
	items    ListItems
}

func newListRule(step ContentStep) (ContentRule, error) {
	if step.Items == nil {
		return nil, fmt.Errorf("list rule requires items configuration")
	}
	return &listRule{required: step.Required, items: *step.Items}, nil
}

func (r *listRule) Type() string { return "list" }

func (r *listRule) Validate(rc RuleContext, secName string, body []NodeCursor) ([]Diagnostic, []NodeCursor) {
	for i, node := range body {
		if node.Kind() != "list" {
			continue
		}

		items := node.Children()
		var diags []Diagnostic
		listLine := lineNumber(rc.Buf, node.Position())

		if r.items.Min > 0 && len(items) < r.items.Min {
			diags = append(diags, Diagnostic{
				Code:     "MD011",
				Severity: severityError,
				Section:  secName,
				Message:  fmt.Sprintf("list requires at least %d item(s)", r.items.Min),
				Line:     listLine,
			})
		}
		if r.items.Max > 0 && len(items) > r.items.Max {
			diags = append(diags, Diagnostic{
				Code:     "MD012",
				Severity: severityError,
				Section:  secName,
				Message:  fmt.Sprintf("list allows at most %d item(s)", r.items.Max),
				Line:     listLine,
			})
		}

		for idx, it := range items {
			re := r.validateItem(it.Text())
			if re != nil {
				line := lineNumber(rc.Buf, it.Position())
				diags = append(diags, Diagnostic{
					Code:     re.Code,
					Severity: re.Severity,
					Section:  secName,
					Message:  fmt.Sprintf("item %d: %s", idx+1, re.Message),
					Line:     line,
				})
			}
		}

		return diags, body[i+1:]
	}

	if r.required {
		return []Diagnostic{{
			Code:     "MD010",
			Severity: severityError,
			Section:  secName,
			Message:  "expected a list",
			Line:     0,
		}}, body
	}

	return nil, body
}

func (r *listRule) validateItem(text string) *Diagnostic {
	switch r.items.Of.Type {
	case "text":
		l := len(text)
		if r.items.Of.MinLen > 0 && l < r.items.Of.MinLen {
			return &Diagnostic{Code: "MD013", Severity: severityError, Message: fmt.Sprintf("length must be >= %d", r.items.Of.MinLen)}
		}
		if r.items.Of.MaxLen > 0 && l > r.items.Of.MaxLen {
			return &Diagnostic{Code: "MD014", Severity: severityError, Message: fmt.Sprintf("length must be <= %d", r.items.Of.MaxLen)}
		}
	case "email":
		if !isEmail(text) {
			return &Diagnostic{Code: "MD015", Severity: severityError, Message: "not a valid email"}
		}
	case "regex":
		re, err := compileCached(r.items.Of.Pattern)
		if err != nil {
			return &Diagnostic{Code: "MD016", Severity: severityError, Message: fmt.Sprintf("invalid regex: %v", err)}
		}
		if !re.MatchString(text) {
			return &Diagnostic{Code: "MD017", Severity: severityError, Message: "does not match pattern"}
		}
	default:
		return &Diagnostic{Code: "MD018", Severity: severityError, Message: fmt.Sprintf("unsupported list item type %q", r.items.Of.Type)}
	}
	return nil
}
