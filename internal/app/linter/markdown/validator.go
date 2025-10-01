// internal/app/linter/markdown/validator.go
package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type PolicyRequest struct {
	PolicyNames []string
	Node        ast.Node
	Line        int
	Column      int
}

// ValidationError is a custom error that includes the location of the error.
type ValidationError struct {
	Line    int
	Column  int
	Message string
}

// Error implements the standard error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// Validator holds the schema and source content for validation.
type Validator struct {
	Schema *Schema
	Source []byte
}

// NewValidator creates and configures a new schema validator.
func NewValidator(schema *Schema, source []byte) *Validator {
	return &Validator{
		Schema: schema,
		Source: source,
	}
}

// Validate traverses the document's top-level nodes against the schema's structure rules.
func (v *Validator) Validate(root ast.Node, sectionName string) ([]*ValidationError, []PolicyRequest) {
	var errors []*ValidationError
	var requests []PolicyRequest

	section, exists := v.Schema.Sections[sectionName]
	if !exists {
		errors = append(errors, &ValidationError{Message: fmt.Sprintf("schema error: section '%s' is not defined", sectionName)})
		return errors, requests
	}
	rules := section.Structure

	currentNode := root.FirstChild()

	for _, rule := range rules {
		if currentNode == nil {
			errors = append(errors, &ValidationError{Message: fmt.Sprintf("incomplete document for section '%s': expected a '%s' based on the schema", sectionName, rule.Type)})
			return errors, requests // Stop if structure is broken
		}

		if rule.Type == "freeform" {
			var freeformErr *ValidationError
			currentNode, freeformErr = v.validateFreeform(rule, currentNode)
			if freeformErr != nil {
				errors = append(errors, freeformErr)
				return errors, requests
			}
			// After validateFreeform, currentNode is pointing at the 'until' heading
			// (or is nil if it consumed to the end). The next rule will validate it.
			continue
		}

		// Handle rule references ($ref)
		if rule.Ref != "" {
			var subErrors []*ValidationError
			var subRequests []PolicyRequest
			currentNode, subErrors, subRequests = v.validateRef(rule, currentNode)
			errors = append(errors, subErrors...)
			requests = append(requests, subRequests...)
			if len(subErrors) > 0 {
				return errors, requests // Stop on first structural failure in a reference
			}
			continue // Move to the next rule
		}

		// Handle standard rules
		validationErr := v.validateNode(rule, currentNode)
		if validationErr != nil {
			errors = append(errors, validationErr)
			return errors, requests // Stop on first structural failure
		}

		// If structurally valid, check for policies
		if len(rule.Policies) > 0 {
			line, col := positionFor(currentNode, v.Source)
			requests = append(requests, PolicyRequest{
				PolicyNames: rule.Policies,
				Node:        currentNode,
				Line:        line,
				Column:      col,
			})
		}

		currentNode = currentNode.NextSibling()
	}

	if currentNode != nil {
		line, col := positionFor(currentNode, v.Source)
		errors = append(errors, &ValidationError{
			Line:    line,
			Column:  col,
			Message: fmt.Sprintf("unexpected content of type '%s' found after the document structure defined in the schema", getNodeTypeName(currentNode)),
		})
	}

	return errors, requests
}

// validateRef handles a $ref rule by validating it as a sub-sequence.
func (v *Validator) validateRef(rule Rule, startNode ast.Node) (ast.Node, []*ValidationError, []PolicyRequest) {
	refName := strings.TrimPrefix(rule.Ref, "#/definitions/")
	defRules, ok := v.Schema.Definitions[refName]
	if !ok {
		return startNode, []*ValidationError{{
			Message: fmt.Sprintf("schema error: definition '%s' not found", refName),
		}}, nil
	}
	return v.validateSequence(defRules, startNode)
}

// validateSequence validates a series of rules against a series of nodes.
func (v *Validator) validateSequence(rules []Rule, startNode ast.Node) (ast.Node, []*ValidationError, []PolicyRequest) {
	var errors []*ValidationError
	var requests []PolicyRequest
	currentNode := startNode

	for _, rule := range rules {
		if currentNode == nil {
			errors = append(errors, &ValidationError{Message: "incomplete document section: expected more content"})
			return nil, errors, requests
		}
		err := v.validateNode(rule, currentNode)
		if err != nil {
			errors = append(errors, err)
			return currentNode, errors, requests // Stop on first error in sequence
		}

		if len(rule.Policies) > 0 {
			line, col := positionFor(currentNode, v.Source)
			requests = append(requests, PolicyRequest{
				PolicyNames: rule.Policies,
				Node:        currentNode,
				Line:        line,
				Column:      col,
			})
		}

		currentNode = currentNode.NextSibling()
	}
	return currentNode, errors, requests
}

// validateNode dispatches validation to the correct function based on rule type.
func (v *Validator) validateNode(rule Rule, node ast.Node) *ValidationError {
	switch rule.Type {
	case "heading":
		return v.validateHeading(rule, node)
	case "list":
		return v.validateList(rule, node)
	case "paragraph":
		return v.validateParagraph(rule, node)
	case "code_block":
		return v.validateCodeBlock(rule, node)
	default:
		return &ValidationError{Message: fmt.Sprintf("unknown rule type in schema: '%s'", rule.Type)}
	}
}

// validateFreeform consumes nodes until it finds the terminating 'until' heading.
// It returns the terminating node so the main validation loop can resume from there.
func (v *Validator) validateFreeform(rule Rule, startNode ast.Node) (ast.Node, *ValidationError) {
	isConsumingToEnd := rule.Until == ""
	blockCount := 0
	currentNode := startNode

	for currentNode != nil {
		// Check if the current node is the terminator heading.
		if !isConsumingToEnd {
			if heading, ok := currentNode.(*ast.Heading); ok {
				headingText := string(heading.Text(v.Source))
				if headingText == rule.Until {
					// Found the end. Return this node so the main loop can validate it.
					return currentNode, nil
				}
			}
		}

		// Check if the block limit has been exceeded.
		if rule.MaxBlocks > 0 && blockCount >= rule.MaxBlocks {
			line, col := positionFor(currentNode, v.Source)
			return currentNode, &ValidationError{
				Line:    line,
				Column:  col,
				Message: fmt.Sprintf("freeform section exceeds maximum of %d blocks", rule.MaxBlocks),
			}
		}

		// Check if the block type is allowed.
		if len(rule.AllowedTypes) > 0 {
			nodeSchemaType := mapAstTypeToSchemaType(currentNode)
			isAllowed := false
			for _, allowedType := range rule.AllowedTypes {
				if nodeSchemaType == allowedType {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				line, col := positionFor(currentNode, v.Source)
				return currentNode, &ValidationError{
					Line:    line,
					Column:  col,
					Message: fmt.Sprintf("block type '%s' is not allowed in this freeform section", getNodeTypeName(currentNode)),
				}
			}
		}

		blockCount++
		currentNode = currentNode.NextSibling()
	}

	if !isConsumingToEnd {
		// If the loop finishes, we were looking for a terminator but didn't find one.
		return nil, &ValidationError{
			Message: fmt.Sprintf("incomplete document: freeform section did not end with a heading '%s'", rule.Until),
		}
	}

	// If we were consuming to the end, reaching nil is the correct termination.
	return nil, nil
}

// --- Specific Node Validators ---

func (v *Validator) validateHeading(rule Rule, node ast.Node) *ValidationError {
	heading, ok := node.(*ast.Heading)
	line, col := positionFor(node, v.Source)
	if !ok {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected a 'Heading' but found a '%s'", getNodeTypeName(node))}
	}

	if rule.Level != 0 && heading.Level != rule.Level {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected heading of level %d but got level %d", rule.Level, heading.Level)}
	}

	headingText := string(heading.Text(v.Source))
	if rule.Prefix != "" && !strings.HasPrefix(headingText, rule.Prefix) {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected heading to have prefix '%s', but got '%s'", rule.Prefix, headingText)}
	}

	if rule.Text != "" && headingText != rule.Text {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected heading to be '%s', but got '%s'", rule.Text, headingText)}
	}

	return nil
}

func (v *Validator) validateList(rule Rule, node ast.Node) *ValidationError {
	list, ok := node.(*ast.List)
	line, col := positionFor(node, v.Source)
	if !ok {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected a 'List' but found a '%s'", getNodeTypeName(node))}
	}

	if rule.ItemCount != 0 && list.ChildCount() != rule.ItemCount {
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected list to have %d items, but it has %d", rule.ItemCount, list.ChildCount())}
	}

	if len(rule.Items) > 0 {
		return v.validateListItems(rule.Items, list)
	}

	return nil
}

func (v *Validator) validateListItems(itemRules []Rule, list *ast.List) *ValidationError {
	if list.ChildCount() != len(itemRules) {
		line, col := positionFor(list, v.Source)
		return &ValidationError{
			Line:    line,
			Column:  col,
			Message: fmt.Sprintf("schema definition mismatch: expected %d list item rules, but got %d items in document", len(itemRules), list.ChildCount()),
		}
	}

	currentItemNode := list.FirstChild()
	for i, itemRule := range itemRules {
		itemText := strings.TrimSpace(string(currentItemNode.Text(v.Source)))
		if itemRule.Prefix != "" && !strings.HasPrefix(itemText, itemRule.Prefix) {
			line, col := positionFor(currentItemNode, v.Source)
			return &ValidationError{
				Line:    line,
				Column:  col,
				Message: fmt.Sprintf("expected list item %d to have prefix '%s', but got '%s'", i+1, itemRule.Prefix, itemText),
			}
		}
		currentItemNode = currentItemNode.NextSibling()
	}
	return nil
}

func (v *Validator) validateParagraph(rule Rule, node ast.Node) *ValidationError {
	if _, ok := node.(*ast.Paragraph); !ok {
		line, col := positionFor(node, v.Source)
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected a 'Paragraph' but found a '%s'", getNodeTypeName(node))}
	}
	return nil
}

func (v *Validator) validateCodeBlock(rule Rule, node ast.Node) *ValidationError {
	switch node.(type) {
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return nil
	default:
		line, col := positionFor(node, v.Source)
		return &ValidationError{Line: line, Column: col, Message: fmt.Sprintf("expected a 'CodeBlock' or 'FencedCodeBlock' but found a '%s'", getNodeTypeName(node))}
	}
}

// --- Helper Functions ---

// positionFor uses goldmark's text reader to translate a node's byte offset
// into 1-based line and column numbers.
func positionFor(node ast.Node, source []byte) (int, int) {
	if node == nil || len(source) == 0 {
		return 1, 1
	}

	segments := node.Lines()
	if segments == nil || segments.Len() == 0 {
		if child := node.FirstChild(); child != nil {
			return positionFor(child, source)
		}
		return 1, 1
	}

	seg := segments.At(0)
	if seg.Start < 0 {
		return 1, 1
	}

	// Advance a reader to the offset to find the line and column.
	reader := text.NewReader(source)
	reader.Advance(seg.Start)

	// Use the reader's current position to compute line and column.
	line, _ := reader.Position()
	column := reader.LineOffset()

	// Return 1-based line and column numbers for user-facing reports.
	return line + 1, column + 1
}

func getNodeTypeName(node ast.Node) string {
	typeName := fmt.Sprintf("%T", node)
	return strings.TrimPrefix(typeName, "*ast.")
}


// mapAstTypeToSchemaType converts an AST node type to its corresponding schema string.
func mapAstTypeToSchemaType(node ast.Node) string {
	switch node.(type) {
	case *ast.Heading:
		return "heading"
	case *ast.List:
		return "list"
	case *ast.Paragraph:
		return "paragraph"
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return "code_block"
	default:
		// Fallback for other potential top-level types.
		return strings.ToLower(getNodeTypeName(node))
	}
}