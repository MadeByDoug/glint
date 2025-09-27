// internal/app/lint/requirements.go
package lint

// Requirements describes what the current plan needs from the pipeline.
type Requirements struct {
	IncludeDirs   bool
	IncludeFiles  bool
	NeedShallow   bool // e.g., language sniffers, cheap metadata
	NeedContents  bool // full file reads (for content checks)
	FileSelectors []Selector
}

// PlanRequirements inspects compiled checks to compute the minimal needs.
func PlanRequirements(checks []Checker) Requirements {
	var r Requirements
	for _, chk := range checks {
		inspect := applySelection(&r, chk)
		applyNeeds(&r, inspect)
	}
	if r.NeedShallow || r.NeedContents {
		r.IncludeFiles = true
	}
	// Access to files always requires directory traversal.
	if r.IncludeFiles {
		r.IncludeDirs = true
	}
	return r
}

func applySelection(r *Requirements, chk Checker) Checker {
	selected, ok := chk.(*SelectedCheck)
	if !ok {
		return chk
	}

	switch selected.selector.Kind {
	case SelectorKindFolder:
		r.IncludeDirs = true
	case SelectorKindFile:
		r.IncludeFiles = true
		r.FileSelectors = appendSelector(r.FileSelectors, selected.selector)
	}

	return selected.inner
}

func appendSelector(existing []Selector, sel Selector) []Selector {
	for _, s := range existing {
		if selectorEqual(s, sel) {
			return existing
		}
	}
	return append(existing, sel)
}

func selectorEqual(a, b Selector) bool {
	if a.Kind != b.Kind {
		return false
	}
	if len(a.Meta) != len(b.Meta) {
		return false
	}
	for k, v := range a.Meta {
		if b.Meta[k] != v {
			return false
		}
	}
	if len(a.PathRegexes) != len(b.PathRegexes) {
		return false
	}
	for i := range a.PathRegexes {
		if a.PathRegexes[i].String() != b.PathRegexes[i].String() {
			return false
		}
	}
	return true
}

func applyNeeds(r *Requirements, chk Checker) {
	if nc, ok := chk.(NeedsShallowMeta); ok && nc.NeedsShallowMeta() {
		r.NeedShallow = true
	}
	if nc, ok := chk.(NeedsFileContents); ok && nc.NeedsFileContents() {
		r.NeedContents = true
	}
}

// Marker interfaces a checker can implement to declare needs.

type NeedsShallowMeta interface {
	NeedsShallowMeta() bool
}

type NeedsFileContents interface {
	NeedsFileContents() bool
}
