package markdown

import "fmt"

// CompiledSchema holds the parsed schema and pre-built content rules.
type CompiledSchema struct {
	Schema     *Schema
	Normalizer Normalizer
	Sections   []CompiledSection
}

// CompiledSection pairs a schema section with its content rules.
type CompiledSection struct {
	Section Section
	Rules   []ContentRule
}

// Compile prepares reusable rule instances for a schema.
func Compile(schema *Schema) (*CompiledSchema, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema: nil")
	}

	norm := normalizer(schema.Document.Normalization)
	compiled := &CompiledSchema{
		Schema:     schema,
		Normalizer: norm,
	}

	compiled.Sections = make([]CompiledSection, 0, len(schema.Sections))
	for _, sec := range schema.Sections {
		rules := make([]ContentRule, 0, len(sec.Content.Sequence))
		for _, step := range sec.Content.Sequence {
			rule, err := BuildRule(step)
			if err != nil {
				return nil, fmt.Errorf("section %q: %w", sec.Name, err)
			}
			rules = append(rules, rule)
		}
		compiled.Sections = append(compiled.Sections, CompiledSection{Section: sec, Rules: rules})
	}

	return compiled, nil
}
