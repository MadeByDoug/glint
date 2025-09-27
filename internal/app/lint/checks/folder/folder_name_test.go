// internal/app/lint/checks/folder/folder_name_test.go
package folder_test

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

const srcDir = "src"

type folderTestCase struct {
	name       string
	buildTree  func(*testutil.B)
	checker    lint.Checker
	expectNone bool
	expects    []testutil.Expect
}

type folderCaseConfig struct {
	name       string
	build      func(*testutil.B)
	selector   lint.Selector
	config     folder.FolderNameConfig
	code       string
	expectNone bool
	expects    []testutil.Expect
}

func runFolderNameCases(ctx context.Context, t *testing.T, cases []folderTestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := testutil.Given(t, tc.name).
				Tree(tc.buildTree).
				Checks(tc.checker).
				WhenLint(ctx).
				Then()
			if tc.expectNone {
				runner.ExpectNone()
			} else {
				runner.ExpectContains(tc.expects...)
			}
			runner.Done()
		})
	}
}

func newFolderCase(cfg *folderCaseConfig) folderTestCase {
	if cfg == nil {
		return folderTestCase{}
	}

	return folderTestCase{
		name:      cfg.name,
		buildTree: cfg.build,
		checker: lint.WithSelector(
			cfg.selector,
			folder.NewFolderNameCheck(&cfg.config, cfg.code, reporting.SeverityError),
		),
		expectNone: cfg.expectNone,
		expects:    cfg.expects,
	}
}

func buildSingleFolder(parent, child string) func(*testutil.B) {
	return func(b *testutil.B) {
		b.Dir(parent, func(b *testutil.B) { b.Dir(child) })
	}
}

func predicateRuleFailsCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "predicate rule fails on non-matching case",
		build:    buildSingleFolder(srcDir, "myComponent"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{Predicates: []string{"kebab"}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: "/src/myComponent: name 'myComponent' does not satisfy predicate: kebab"},
		},
	})
}

func disallowRuleFailsCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "disallow rule fails on matching pattern",
		build:    buildSingleFolder(srcDir, "gen-client"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{Disallow: []string{`^gen-.*$`}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: `name 'gen-client' matches disallowed pattern "^gen-.*$"`},
		},
	})
}

func prefixRuleFailsCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "prefix rule fails when prefix is missing",
		build:    buildSingleFolder(srcDir, "button"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{Prefix: []string{"comp-"}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: "name 'button' does not have required prefix 'comp-'"},
		},
	})
}

func suffixRuleFailsCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "suffix rule fails when suffix is missing",
		build:    buildSingleFolder("svc", "user"),
		selector: lint.Selector{PathRegexes: mustCompile(`^svc/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{Suffix: []string{"-svc"}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: "name 'user' does not have required suffix '-svc'"},
		},
	})
}

func prohibitPrefixCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "prohibitPrefix rule fails on forbidden prefix",
		build:    buildSingleFolder(srcDir, "_private"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{ProhibitPrefix: []string{"_"}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: "name '_private' has prohibited prefix '_'"},
		},
	})
}

func prohibitSuffixCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "prohibitSuffix rule fails on forbidden suffix",
		build:    buildSingleFolder(srcDir, "private_"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config:   folder.FolderNameConfig{ProhibitSuffix: []string{"_"}},
		code:     code,
		expects: []testutil.Expect{
			{Code: code, Msg: "name 'private_' has prohibited suffix '_'"},
		},
	})
}

func customMessageCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name:     "custom message is correctly prepended to diagnostic",
		build:    buildSingleFolder(srcDir, "MyComponent"),
		selector: lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
		config: folder.FolderNameConfig{
			Predicates: []string{"kebab"},
			Message:    "Please rename component before committing.",
		},
		code: code,
		expects: []testutil.Expect{
			{Code: code, Msg: "Please rename component before committing. (name 'MyComponent' does not satisfy predicate: kebab)"},
		},
	})
}

func emptyConfigCase(code string) folderTestCase {
	return newFolderCase(&folderCaseConfig{
		name: "empty config produces no diagnostics",
		build: func(b *testutil.B) {
			b.Dir("root", func(b *testutil.B) {
				b.Dir("anyName")
				b.Dir("Another")
			})
		},
		selector:   lint.Selector{PathRegexes: mustCompile(".*"), Kind: lint.SelectorKindFolder},
		config:     folder.FolderNameConfig{},
		code:       code,
		expectNone: true,
	})
}

// TestFolderName_CoreFunctionality verifies the basic wiring of each check parameter.
// It ensures that a simple case for each feature triggers the expected diagnostic.
func TestFolderName_CoreFunctionality(t *testing.T) {
	ctx := context.Background()
	const checkCode = "FOLDER-NAME-CORE-001"

	cases := []folderTestCase{
		predicateRuleFailsCase(checkCode),
		disallowRuleFailsCase(checkCode),
		prefixRuleFailsCase(checkCode),
		suffixRuleFailsCase(checkCode),
		prohibitPrefixCase(checkCode),
		prohibitSuffixCase(checkCode),
		customMessageCase(checkCode),
		emptyConfigCase(checkCode),
	}

	runFolderNameCases(ctx, t, cases)
}

// TestFolderName_Predicates tests the casing predicates in depth.
func TestFolderName_Predicates(t *testing.T) {
	ctx := context.Background()
	const checkCode = "FOLDER-NAME-PREDS-001"

	cases := []folderTestCase{
		{
			name: "kebab predicate fails on camelCase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("myComponent") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"kebab"}}, checkCode, reporting.SeverityError),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/myComponent: name 'myComponent' does not satisfy predicate: kebab"},
			},
		},
		{
			name: "kebab predicate allows digits",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("abc-123") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"kebab"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "lower predicate fails on uppercase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("API") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"lower"}}, checkCode, reporting.SeverityError),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/API: name 'API' does not satisfy predicate: lower"},
			},
		},
		{
			name: "lower predicate passes on correct case",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("api") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"lower"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "snake predicate fails on kebab",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("my-folder") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"snake"}}, checkCode, reporting.SeverityError),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/my-folder: name 'my-folder' does not satisfy predicate: snake"},
			},
		},
		{
			name: "snake predicate passes on snake_case",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("my_folder") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"snake"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "camel predicate passes on camelCase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("myComponent") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"camel"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "pascal predicate passes on PascalCase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("MyComponent") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"pascal"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "upper predicate fails on lowercase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("api") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"upper"}}, checkCode, reporting.SeverityError),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/api: name 'api' does not satisfy predicate: upper"},
			},
		},
		{
			name: "upper predicate passes on uppercase",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("API") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"upper"}}, checkCode, reporting.SeverityError),
			),
			expectNone: true,
		},
		{
			name: "all failing predicates are reported",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("Not-Kebab") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(&folder.FolderNameConfig{Predicates: []string{"kebab", "lower"}}, checkCode, reporting.SeverityError),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/Not-Kebab: name 'Not-Kebab' does not satisfy predicate: kebab"},
				{Code: checkCode, Msg: "/src/Not-Kebab: name 'Not-Kebab' does not satisfy predicate: lower"},
			},
		},
	}

	runFolderNameCases(ctx, t, cases)
}

// TestFolderName_Affixes tests prefix and suffix functionality.
func TestFolderName_Affixes(t *testing.T) {
	ctx := context.Background()
	const checkCode = "FOLDER-NAME-AFFIX-001"

	testCases := []struct {
		name      string
		buildTree func(*testutil.B)
		checker   lint.Checker
		expects   []testutil.Expect
	}{
		{
			name: "multiple affix violations on one folder produce multiple diagnostics",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("tmp_internal_tmp") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{
						Prefix:         []string{"lib-"},
						Suffix:         []string{"-svc"},
						ProhibitPrefix: []string{"tmp_"},
						ProhibitSuffix: []string{"_tmp"},
					},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "name 'tmp_internal_tmp' does not have required prefix 'lib-'"},
				{Code: checkCode, Msg: "name 'tmp_internal_tmp' does not have required suffix '-svc'"},
				{Code: checkCode, Msg: "name 'tmp_internal_tmp' has prohibited prefix 'tmp_'"},
				{Code: checkCode, Msg: "name 'tmp_internal_tmp' has prohibited suffix '_tmp'"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := testutil.Given(t, tc.name).Tree(tc.buildTree).Checks(tc.checker).WhenLint(ctx).Then()
			r.ExpectContains(tc.expects...)
			r.Done()
		})
	}
}

// TestFolderName_AllowDisallow tests the allow and disallow lists.
func TestFolderName_AllowDisallow(t *testing.T) {
	ctx := context.Background()
	const checkCode = "FOLDER-NAME-ALLOW-001"

	testCases := []struct {
		name       string
		buildTree  func(*testutil.B)
		checker    lint.Checker
		expectNone bool
		expects    []testutil.Expect
	}{
		{
			name: "disallow with regex matches sub-string pattern",
			buildTree: func(b *testutil.B) {
				b.Dir("api", func(b *testutil.B) {
					b.Dir("gen-client") // Fails
					b.Dir("generate")   // Passes
				})
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile(".*"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Disallow: []string{`^gen-.*$`}},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: `/api/gen-client: name 'gen-client' matches disallowed pattern "^gen-.*$"`},
			},
		},
		{
			name: "disallow with exact match does not match substring",
			buildTree: func(b *testutil.B) {
				b.Dir("model")  // Fails
				b.Dir("models") // Passes
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile(`^[^/]+$`), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Disallow: []string{`^model$`}},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: `/model: name 'model' matches disallowed pattern "^model$"`},
			},
		},
		{
			name: "disallow can be scoped by selector",
			buildTree: func(b *testutil.B) {
				b.Dir("model") // Fails (selected)
				b.Dir(srcDir, func(b *testutil.B) {
					b.Dir("model") // Passes (not selected)
				})
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile(`^[^/]+$`), Kind: lint.SelectorKindFolder}, // root-only
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{Disallow: []string{`^model$`}},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/model: name 'model' matches disallowed pattern \"^model$\""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := testutil.Given(t, tc.name).Tree(tc.buildTree).Checks(tc.checker).WhenLint(ctx).Then()
			if tc.expectNone {
				runner.ExpectNone()
			} else {
				runner.ExpectContains(tc.expects...)
			}
			runner.Done()
		})
	}
}

// TestFolderName_Precedence ensures rules are applied in the correct order.
// The expected order of operations is: 1) Allow, 2) Disallow, 3) Other checks.
func TestFolderName_Precedence(t *testing.T) {
	ctx := context.Background()
	const checkCode = "FOLDER-NAME-PRECEDENCE-001"

	cases := []folderTestCase{
		{
			name: "allow short-circuits all other failing rules",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("BAD_name") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile(`^src/[^/]+$`), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{
						Predicates:     []string{"kebab"},
						Disallow:       []string{`^BAD.*`},
						Prefix:         []string{"comp-"},
						ProhibitPrefix: []string{"BAD_"},
						Allow:          []string{`^BAD_name$`},
					},
					checkCode, reporting.SeverityError,
				),
			),
			expectNone: true,
		},
		{
			name: "disallow takes precedence over other failing rules",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) {
					b.Dir("gen-Bad") // Disallowed and violates kebab, but disallow is reported.
				})
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{
						Predicates: []string{"kebab"},
						Disallow:   []string{`^gen-.*$`},
					},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: `name 'gen-Bad' matches disallowed pattern "^gen-.*$"`},
			},
		},
		{
			name: "multiple non-precedence violations on a single folder are all reported",
			buildTree: func(b *testutil.B) {
				b.Dir(srcDir, func(b *testutil.B) { b.Dir("Foo") })
			},
			checker: lint.WithSelector(
				lint.Selector{PathRegexes: mustCompile("^src/[^/]+$"), Kind: lint.SelectorKindFolder},
				folder.NewFolderNameCheck(
					&folder.FolderNameConfig{
						Predicates: []string{"lower"},
						Prefix:     []string{"lib-"},
						Suffix:     []string{"-svc"},
					},
					checkCode, reporting.SeverityError,
				),
			),
			expects: []testutil.Expect{
				{Code: checkCode, Msg: "/src/Foo: name 'Foo' does not satisfy predicate: lower"},
				{Code: checkCode, Msg: "/src/Foo: name 'Foo' does not have required prefix 'lib-'"},
				{Code: checkCode, Msg: "/src/Foo: name 'Foo' does not have required suffix '-svc'"},
			},
		},
	}

	runFolderNameCases(ctx, t, cases)
}
