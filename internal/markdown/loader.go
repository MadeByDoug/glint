package markdown

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// Normalizer normalizes heading strings prior to comparison.
type Normalizer func(string) string

func normalizer(mode string) Normalizer {
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case "NFC":
		return func(s string) string { return norm.NFC.String(s) }
	case "NFD":
		return func(s string) string { return norm.NFD.String(s) }
	default:
		return func(s string) string { return s }
	}
}

// LoadSchema reads a schema YAML file from disk.
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	return &schema, nil
}

// LoadCompiled loads and compiles a schema in one step.
func LoadCompiled(path string) (*CompiledSchema, error) {
	schema, err := LoadSchema(path)
	if err != nil {
		return nil, err
	}
	compiled, err := Compile(schema)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return compiled, nil
}
