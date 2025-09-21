// internal/app/lint/parsers/parser.go
package parsers

// Parser defines the function signature for a shallow parser.
// It takes the raw content of a file and returns a map of extracted metadata.
type Parser func(content []byte) (map[string]any, error)
