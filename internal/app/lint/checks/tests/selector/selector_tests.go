// internal/app/lint/checks/tests/selector/selector_test.go
package selector

import (
	"context"
	"regexp"
	"testing"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/checks/folder"
	"github.com/MrBigCode/glint/internal/app/lint/checks/tests/testutil"
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

	testCases := []selectorTestCase{
		{
			name: "empty PathRegexes selects nothing",
			buildTree: func(b *testutil.B) {
				b.Dir("src", func(b *testutil.B) { b.Dir("BadName") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: nil, Kind: "folder"},
				folder.NewFolderNameCheck(
					folder.FolderNameConfig{Predicates: []string{"lowercase"}},
					checkCode, reporting.SevError,
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
				lint.Selector{PathRegexes: mustCompile(".*"), Kind: "file"},
				folder.NewFolderNameCheck(
					folder.FolderNameConfig{Predicates: []string{"lowercase"}},
					checkCode, reporting.SevError,
				),
			),
			expectNone: true,
		},
		{
			name: "multiple regex act as union",
			buildTree: func(b *testutil.B) {
				b.Dir("pkg", func(b *testutil.B) { b.Dir("A") })  // fails lowercase
				b.Dir("cmd", func(b *testutil.B) { b.Dir("B") })  // fails lowercase
				b.Dir("docs", func(b *testutil.B) { b.Dir("ok") }) // not selected
			},
			checker: lint.WithSelector(
				lint.Selector{
					PathRegexes: mustCompile(`^pkg/[^/]+$`, `^cmd/[^/]+$`),
					Kind:        "folder",
				},
				folder.NewFolderNameCheck(
					folder.FolderNameConfig{Predicates: []string{"lowercase"}},
					checkCode, reporting.SevError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/pkg/A: name 'A' does not satisfy predicate: lowercase"},
				{Code: checkCode, Msg: "/cmd/B: name 'B' does not satisfy predicate: lowercase"},
			},
		},
	}
	runSelectorCases(t, ctx, testCases)
}

type selectorTestCase struct {
	name       string
	buildTree  func(*testutil.B)
	checker    lint.Checker
	expectNone bool
	expects    []testutil.Expect
}

func runSelectorCases(t *testing.T, ctx context.Context, cases []selectorTestCase) {
	t.Helper()
	for _, tc := range cases {
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
