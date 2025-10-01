// internal/app/linter/markdown/schema.go
package markdown

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Rule defines a single validation check, which can be a standard rule,
// or a reference to a definition.
type Rule struct {
	// For standard validation rules
	Type      string   `yaml:"type,omitempty"`
	Level     int      `yaml:"level,omitempty"`
	Prefix    string   `yaml:"prefix,omitempty"`
	Text      string   `yaml:"text,omitempty"`
	ItemCount int      `yaml:"item_count,omitempty"`
	Items     []Rule   `yaml:"items,omitempty"`
	Policies  []string `yaml:"policies,omitempty"`

	// For freeform sections
	Until        string   `yaml:"until,omitempty"`
	MaxBlocks    int      `yaml:"max_blocks,omitempty"`
	AllowedTypes []string `yaml:"allowed_types,omitempty"`

	// For referencing a block from the 'definitions' section.
	// Example: $ref: "#/definitions/rfc_meta"
	Ref string `yaml:"$ref,omitempty"`

	// --- Fields for future inline validation rules (currently parsed but not validated) ---
	Contains  []Rule `yaml:"contains,omitempty"`
	Allowed   *bool  `yaml:"allowed,omitempty"` // Use a pointer to distinguish between false and not set
	MaxLength int    `yaml:"max_length,omitempty"`
}

type Section struct {
	Structure []Rule `yaml:"structure"`
}

// Schema is the top-level structure of a validation file.
type Schema struct {
	Definitions map[string][]Rule  `yaml:"definitions"`
	Sections map[string]Section   `yaml:"sections"`
}

// LoadSchemaFromFile parses a YAML schema file into our structs.
func LoadSchemaFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}