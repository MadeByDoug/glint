// internal/app/lint/runner.go
package lint

import (
	"context"
	"slices"
	"sort"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
)



// Lint is the single orchestration seam. Pure; no I/O.
func Lint(ctx context.Context, t *Tree, checks ...Checker) []reporting.Report {
	var out []reporting.Report
	for _, ch := range checks {
		out = append(out, ch.Apply(ctx, t)...)
	}
	// Deterministic ordering for stable tests (and nice output).
	sort.SliceStable(out, func(i, j int) bool {
		// 1) path (if embedded in Msg), 2) code, 3) severity
		// Keep simple first: by Code then Message.
		if out[i].Code != out[j].Code {
			return out[i].Code < out[j].Code
		}
		return out[i].Msg < out[j].Msg
	})
	return out
}

func (n *Node) Path() string {
	if n == nil {
		return ""
	}
	parts := []string{}
	for cur := n; cur != nil && cur.Name != ""; cur = cur.Parent {
		parts = append(parts, cur.Name)
	}
	slices.Reverse(parts)
	return "/" + strings.Join(parts, "/")
}
