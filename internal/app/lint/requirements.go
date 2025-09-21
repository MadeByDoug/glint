// internal/app/lint/requirements.go
package lint

// Requirements describes what the current plan needs from the pipeline.
type Requirements struct {
	IncludeDirs  bool
	IncludeFiles bool
	NeedShallow  bool // e.g., language sniffers, cheap metadata
	NeedContents bool // full file reads (for content checks)
}

// PlanRequirements inspects compiled checks to compute the minimal needs.
func PlanRequirements(checks []Checker) Requirements {
	var r Requirements
	for _, chk := range checks {
		inspect := applySelection(&r, chk)
		applyNeeds(&r, inspect)
	}
	// Folders always imply dirs
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
	case "folder":
		r.IncludeDirs = true
	case "file":
		r.IncludeFiles = true
	}

	return selected.inner
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
