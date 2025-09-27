// internal/app/lint/selector_test.go
package lint_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/checks/folder"
	"github.com/MrBigCode/glint/internal/app/lint/testutil"
)

func mustCompile(patterns ...string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, p := range patterns {
		out = append(out, regexp.MustCompile(p))
	}
	return out
}

func TestSelectorEdgeCases(t *testing.T) {
	ctx := context.Background()
	const checkCode = "SELECTOR-EDGE-001"

	testCases := []struct {
		name       string
		buildTree  func(*testutil.B)
		checker    lint.Checker
		expectNone bool
		expects    []testutil.Expect
	}{
		{
			name: "empty PathRegexes selects nothing",
			buildTree: func(b *testutil.B) {
				b.Dir("src", func(b *testutil.B) { b.Dir("BadName") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: nil, Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Predicates: []string{"lower"}},
					checkCode, reporting.SeverityError,
				),
			),
			expectNone: true,
		},
		{
			name: "kind=file selects nothing for folders",
			buildTree: func(b *testutil.B) {
				b.Dir("src", func(b *testutil.B) { b.Dir("BadName") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile(".*"), Kind: lint.SelectorKindFile},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Predicates: []string{"lower"}},
					checkCode, reporting.SeverityError,
				),
			),
			expectNone: true,
		},
		{
			name: "meta filters nodes by exact match",
			buildTree: func(b *testutil.B) {
				b.Dir("src", func(b *testutil.B) {
					b.Dir("Svc").WithMeta("layer", "svc")
					b.Dir("Lib").WithMeta("layer", "lib")
				})
			},
			checker: lint.WithSelector(
				lint.Selector{
					PathRegexes: mustCompile(`^src/[^/]+$`),
					Kind:        lint.SelectorKindFolder,
					Meta:        map[string]string{"layer": "svc"},
				},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Predicates: []string{"lower"}},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/Svc: name 'Svc' does not satisfy predicate: lower"},
			},
		},
		{
			name: "multiple regex act as union",
			buildTree: func(b *testutil.B) {
				b.Dir("pkg", func(b *testutil.B) { b.Dir("A") })   // fails lower predicate
				b.Dir("cmd", func(b *testutil.B) { b.Dir("B") })   // fails lower predicate
				b.Dir("docs", func(b *testutil.B) { b.Dir("ok") }) // not selected
			},
			checker: lint.WithSelector(
				lint.Selector{
					PathRegexes: mustCompile(`^pkg/[^/]+$`, `^cmd/[^/]+$`),
					Kind:        lint.SelectorKindFolder,
				},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Predicates: []string{"lower"}},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/pkg/A: name 'A' does not satisfy predicate: lower"},
				{Code: checkCode, Msg: "/cmd/B: name 'B' does not satisfy predicate: lower"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := testutil.Given(t, tc.name).
				Tree(tc.buildTree).
				Checks(tc.checker).
				WhenLint(ctx).
				Then()
			if tc.expectNone {
				r.ExpectNone()
			} else {
				r.ExpectContains(tc.expects...)
			}
			r.Done()
		})
	}
}
