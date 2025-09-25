package markdown

import "fmt"

const (
	kindListNode   = "list"
	kindParagraph  = "paragraph"
	kindStrongText = "strong"

	itemTypeText  = "text"
	itemTypeEmail = "email"
	itemTypeRegex = "regex"
)

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
	idx, node := firstNodeOfKind(body, kindListNode)
	if node == nil {
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

	diags := r.validateListNode(rc, secName, node)
	return diags, body[idx+1:]
}

func (r *listRule) validateListNode(rc RuleContext, secName string, node NodeCursor) []Diagnostic {
	items := node.Children()
	diags := r.validateItemCount(secName, node, items, rc)
	for idx, it := range items {
		if diag := r.validateListItem(secName, idx, it, rc); diag != nil {
			diags = append(diags, *diag)
		}
	}
	return diags
}

func (r *listRule) validateItemCount(secName string, node NodeCursor, items []NodeCursor, rc RuleContext) []Diagnostic {
	listLine := lineNumber(rc.Buf, node.Position())
	var diags []Diagnostic
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
	return diags
}

func (r *listRule) validateListItem(secName string, idx int, item NodeCursor, rc RuleContext) *Diagnostic {
	re := r.validateItemText(item.Text())
	if re == nil {
		return nil
	}
	line := lineNumber(rc.Buf, item.Position())
	return &Diagnostic{
		Code:     re.Code,
		Severity: re.Severity,
		Section:  secName,
		Message:  fmt.Sprintf("item %d: %s", idx+1, re.Message),
		Line:     line,
	}
}

func (r *listRule) validateItemText(text string) *Diagnostic {
	switch r.items.Of.Type {
	case itemTypeText:
		return r.validatePlainText(text)
	case itemTypeEmail:
		return validateEmail(text)
	case itemTypeRegex:
		return validateRegex(text, r.items.Of.Pattern)
	default:
		return &Diagnostic{Code: "MD018", Severity: severityError, Message: fmt.Sprintf("unsupported list item type %q", r.items.Of.Type)}
	}
}

func (r *listRule) validatePlainText(text string) *Diagnostic {
	length := len(text)
	if r.items.Of.MinLen > 0 && length < r.items.Of.MinLen {
		return &Diagnostic{Code: "MD013", Severity: severityError, Message: fmt.Sprintf("length must be >= %d", r.items.Of.MinLen)}
	}
	if r.items.Of.MaxLen > 0 && length > r.items.Of.MaxLen {
		return &Diagnostic{Code: "MD014", Severity: severityError, Message: fmt.Sprintf("length must be <= %d", r.items.Of.MaxLen)}
	}
	return nil
}

func validateEmail(text string) *Diagnostic {
	if isEmail(text) {
		return nil
	}
	return &Diagnostic{Code: "MD015", Severity: severityError, Message: "not a valid email"}
}

func validateRegex(text, pattern string) *Diagnostic {
	re, err := compileCached(pattern)
	if err != nil {
		return &Diagnostic{Code: "MD016", Severity: severityError, Message: fmt.Sprintf("invalid regex: %v", err)}
	}
	if !re.MatchString(text) {
		return &Diagnostic{Code: "MD017", Severity: severityError, Message: "does not match pattern"}
	}
	return nil
}

func firstNodeOfKind(body []NodeCursor, kind string) (int, NodeCursor) {
	for i, node := range body {
		if node.Kind() == kind {
			return i, node
		}
	}
	return -1, nil
}
