// internal/app/lint/plan/factory.go
package plan

import (
	"encoding/json"
	"fmt"

	model "github.com/MrBigCode/glint/internal/app/config/model"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/checks/folder"
)

// Constructor defines the function signature for building a Checker.
type Constructor func(ruleID string, severity reporting.Severity, params json.RawMessage) (lint.Checker, error)

// registry holds all available check constructors.
var registry = make(map[string]Constructor)

func init() {
	// Register all known checks here.
	registry["folderName"] = newFolderNameCheck
}

// New creates a new Checker instance from the registry.
func New(chkType, ruleID string, severity reporting.Severity, params json.RawMessage) (lint.Checker, error) {
	constructor, ok := registry[chkType]
	if !ok {
		return nil, fmt.Errorf("unknown check type: %q", chkType)
	}
	return constructor(ruleID, severity, params)
}

// --- Constructors ---

func newFolderNameCheck(ruleID string, severity reporting.Severity, params json.RawMessage) (lint.Checker, error) {
	var p model.ChFolderName
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal folderName params: %w", err)
	}

	return folder.NewFolderNameCheck(
		folder.FolderNameConfig{
			Predicates:     toStringSlice(p.Predicates),
			Allow:          append([]string(nil), p.Allow...),
			Disallow:       append([]string(nil), p.Disallow...),
			Prefix:         strDeref(p.Prefix),
			Suffix:         strDeref(p.Suffix),
			ProhibitPrefix: strDeref(p.ProhibitPrefix),
			ProhibitSuffix: strDeref(p.ProhibitSuffix),
			Message:        messageDeref(p.Message),
		},
		ruleID,
		mapSeverity(p.Severity), // Note: Severity can be defined on the check itself
	), nil
}