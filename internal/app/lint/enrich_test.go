// internal/app/lint/enrich_test.go
package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFilePathTrimsLeadingSlash(t *testing.T) {
	root := t.TempDir()

	got, err := resolveFilePath(root, "/foo/bar.txt")
	if err != nil {
		t.Fatalf("resolveFilePath returned error: %v", err)
	}

	want := filepath.Join(root, "foo", "bar.txt")
	if got != want {
		t.Fatalf("resolveFilePath = %q, want %q", got, want)
	}
}

func TestResolveFilePathPreventsEscapes(t *testing.T) {
	root := t.TempDir()

	_, err := resolveFilePath(root, "/../secret.txt")
	if err == nil {
		t.Fatalf("expected escape attempt to error")
	}
	if !strings.Contains(err.Error(), "escapes root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveFilePathRootRelatesToRootDir(t *testing.T) {
	root := t.TempDir()

	got, err := resolveFilePath(root, "/")
	if err != nil {
		t.Fatalf("resolveFilePath returned error: %v", err)
	}

	if got != filepath.Clean(root) {
		t.Fatalf("resolveFilePath root = %q, want %q", got, root)
	}
}
