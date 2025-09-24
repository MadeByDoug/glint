package markdown

// DocumentOptions capture global behavior modifiers for a schema.
type DocumentOptions struct {
	Normalization      string `yaml:"normalization"`
	CaseSensitive      bool   `yaml:"case_sensitive"`
	StrictOrder        bool   `yaml:"strict_order"`
	AllowExtraSections bool   `yaml:"allow_extra_sections"`
}

// Document describes the document-level metadata.
type Document struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	DocumentOptions
}

// ItemOf configures individual list item validation.
type ItemOf struct {
	Type    string `yaml:"type"`
	Pattern string `yaml:"pattern,omitempty"`
	MinLen  int    `yaml:"min_len,omitempty"`
	MaxLen  int    `yaml:"max_len,omitempty"`
}

// ListItems configures list validation boundaries and predicates.
type ListItems struct {
	Min int    `yaml:"min,omitempty"`
	Max int    `yaml:"max,omitempty"`
	Of  ItemOf `yaml:"of"`
}

// ContentStep represents one expected block within a section body.
type ContentStep struct {
	Type        string     `yaml:"type"`
	Required    bool       `yaml:"required,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Items       *ListItems `yaml:"items,omitempty"`
}

// SectionHeading binds schema sections to Markdown headings.
type SectionHeading struct {
	Level int    `yaml:"level"`
	Text  string `yaml:"text"`
}

// SectionContent defines the ordered content steps for a section.
type SectionContent struct {
	Sequence []ContentStep `yaml:"sequence"`
}

// Section describes a schema section.
type Section struct {
	Name     string         `yaml:"name"`
	Heading  SectionHeading `yaml:"heading"`
	Content  SectionContent `yaml:"content"`
	Requires []string       `yaml:"requires,omitempty"`
}

// Schema is the top-level YAML structure consumed by the validator.
type Schema struct {
	Version  int       `yaml:"version"`
	Document Document  `yaml:"document"`
	Sections []Section `yaml:"sections"`
}
