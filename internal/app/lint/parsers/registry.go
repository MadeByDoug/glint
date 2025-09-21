// internal/app/lint/parsers/registry.go
package parsers

import (
	"path/filepath"
	"strings"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]Parser)
)

func canonical(ext string) string {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return strings.ToLower(ext)
}

// Register associates a file extension with a parser implementation.
func Register(ext string, parser Parser) {
	if parser == nil {
		return
	}
	ext = canonical(ext)
	if ext == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	registry[ext] = parser
}

// GetParser retrieves a parser by inspecting the file's extension.
func GetParser(filename string) Parser {
	mu.RLock()
	defer mu.RUnlock()
	return registry[canonical(filepath.Ext(filename))]
}
