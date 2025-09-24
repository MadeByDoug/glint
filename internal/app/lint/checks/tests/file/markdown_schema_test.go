package tests

import (
	"context"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
	"github.com/MrBigCode/glint/internal/app/lint/checks/file"
	"github.com/MrBigCode/glint/internal/app/lint/checks/tests/testutil"
)

func repoRootPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		dir = filepath.Dir(dir)
	}
	return dir
}

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func markdownSchemaChecker(code string) lint.Checker {
	selector := lint.Selector{
		Kind:        "file",
		PathRegexes: []*regexp.Regexp{regexp.MustCompile(`^rfcs/[^/]+\.md$`)},
	}
	cfg := file.MarkdownSchemaConfig{SchemaPath: "schemas/rfc.schema.yaml"}
	return lint.WithSelector(selector, file.NewMarkdownSchemaCheck(cfg, code, reporting.SevError))
}

func buildRFCTree(t *testing.T, relPath, absPath string) func(*testutil.B) {
	return func(b *testutil.B) {
		root := repoRootPath(t)
		b.WithMeta("absPath", root).WithMeta("relPath", "")
		b.Dir("rfcs", func(b *testutil.B) {
			b.WithMeta("relPath", "rfcs").WithMeta("absPath", filepath.Join(root, "rfcs"))
			fileName := filepath.Base(relPath)
			b.File(fileName).
				WithMeta("relPath", relPath).
				WithMeta("absPath", absPath)
		})
	}
}

func TestMarkdownSchemaCheck_Passes(t *testing.T) {
	ctx := context.Background()

	rel := "rfcs/valid.md"
	abs := testdataPath(t, "valid.md")

	testutil.Given(t, "valid RFC document").
		Tree(buildRFCTree(t, rel, abs)).
		Checks(markdownSchemaChecker("RFC-SCHEMA-PASS")).
		WhenLint(ctx).
		Then().
		ExpectNone().
		Done()
}

func TestMarkdownSchemaCheck_Fails(t *testing.T) {
	ctx := context.Background()

	rel := "rfcs/invalid.md"
	abs := testdataPath(t, "invalid.md")

	testutil.Given(t, "invalid RFC document").
		Tree(buildRFCTree(t, rel, abs)).
		Checks(markdownSchemaChecker("RFC-SCHEMA-FAIL")).
		WhenLint(ctx).
		Then().
		ExpectContains(
			testutil.Expect{Code: "RFC-SCHEMA-FAIL/MD002", Msg: "bold text"},
			testutil.Expect{Code: "RFC-SCHEMA-FAIL/MD010", Msg: "expected a list"},
			testutil.Expect{Code: "RFC-SCHEMA-FAIL/MD100", Msg: "missing section heading"},
			testutil.Expect{Code: "RFC-SCHEMA-FAIL/MD102", Msg: "unexpected extra section"},
			testutil.Expect{Code: "RFC-SCHEMA-FAIL/MD105", Msg: "requires \"Phase 1: Scope & Goals\""},
		).
		ExpectHasSeverity(reporting.SevError).
		ExpectCount(6).
		Done()
}
