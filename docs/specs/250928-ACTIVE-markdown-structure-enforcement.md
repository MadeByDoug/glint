# RFC: Markdown Document Structure Validation

- Start Date: 2025-09-28
- RFC Status: Draft

## Summary

This RFC formalizes the `markdown-schema` validator, a check that enforces the structural integrity of Markdown documents. It validates the sequence of top-level blocks (headings, lists, code blocks, etc.) against a declarative YAML schema.

Key features include support for reusable rule definitions, sectioned validation to support a document's lifecycle (e.g., `draft` vs. `complete`), and detailed error reporting with line and column numbers.

## Motivation

Maintaining structural consistency across key documents—such as RFCs, architectural design records, and tutorials—is critical for readability and process automation. Without an automated check, these documents are prone to structural drift, making them difficult to parse programmatically and for humans to navigate.

This validator addresses these issues by:
- **Enforcing Consistency:** Ensures all documents of a certain type share the same fundamental structure.
- **Automating Quality Gates:** Prevents structurally incomplete or incorrect documents from progressing in a workflow.
- **Supporting Document Lifecycles:** Allows for different structural requirements at different stages of a document's evolution, such as requiring "Summary" and "Motivation" sections for a `draft` before allowing a "Design" section.

## Design

The validator's design consists of three main parts: the schema definition, the validation logic, and the user-facing CLI integration.

### Schema Configuration (`schema.yaml`)

Validation rules are defined in a YAML file. This schema-driven approach allows us to define and modify complex document structures without changing the validator's Go code.

The schema file has two primary top-level keys:

1.  **`definitions`**: An optional map of named, reusable rule sets. Any rule set defined here can be referenced from within a structure using `$ref: "#/definitions/definition_name"`. This is ideal for common patterns, like a document's metadata block.

2.  **`sections`**: A map defining the valid structures for different stages of a document's lifecycle. Each section (e.g., `draft`, `complete`) contains a `structure` key, which holds the list of rules.

A **Rule** is an object with the following keys:
- `type` (string): The required Markdown block type. Supported types include `heading`, `list`, `paragraph`, and `code_block`.
- `level` (int, for `heading`): The required heading level (e.g., `1` for H1).
- `text` (string, for `heading`): The exact required text of the heading.
- `prefix` (string, for `heading`): A required prefix for the heading text.
- `item_count` (int, for `list`): The exact number of items required in a list.
- `items` (array<Rule>, for `list`): A list of rules to validate the text content of each list item. Currently, only `prefix` is supported for item rules.

### Validation Logic

The validation process follows a strict, sequential algorithm:

1.  The validator is invoked with a path to a Markdown file and a section name.
2.  The Markdown file is parsed into an Abstract Syntax Tree (AST).
3.  The validator retrieves the `structure` rule set corresponding to the requested section from the schema.
4.  It iterates through the top-level child nodes of the AST and the schema rules in lockstep.
5.  For each node, it checks if the node's type and properties match the corresponding rule.
6.  Validation **stops at the first mismatch**, ensuring clear and focused error feedback.
7.  The validator will fail if it runs out of nodes before rules are exhausted (document is incomplete) or if it runs out of rules but nodes still remain (document has unexpected trailing content).

### Reporting

Errors are reported to the user via the existing reporting infrastructure, supporting both `text` and `json` formats. Each validation issue includes:
- The file path.
- The 1-based line and column number where the error was detected.
- A descriptive message explaining the failure (e.g., "expected a 'Heading' but found a 'List'").

### CLI Integration

The validator is controlled via a command-line flag:
- `--section <name>`: Specifies which validation section from the schema to apply. If omitted, it defaults to `complete`.

Example invocation:

```shell
# Validate a document against its initial draft structure
go run ./cmd/glint --section=draft rfc/my-new-idea.md
```

## Examples

### 1\. Example Schema

A schema defining two sections for an RFC document.

**/`rfc/rfc.schema.yaml`/**

```yaml
definitions:
  rfc_meta:
    - type: list
      item_count: 3
      items:
        - { prefix: "Start Date:" }
        - { prefix: "RFC Status:" }
        - { prefix: "Owners:" }

sections:
  draft:
    structure:
      - { type: heading, level: 1 }
      - { $ref: "#/definitions/rfc_meta" }
      - { type: heading, level: 2, text: "Summary" }
      - { type: paragraph }
      - { type: heading, level: 2, text: "Motivation" }
      - { type: paragraph }

  complete:
    structure:
      # ... includes all draft sections plus Design, Examples, etc.
      - { type: heading, level: 1 }
      - { $ref: "#/definitions/rfc_meta" }
      - { type: heading, level: 2, text: "Summary" }
      - { type: paragraph }
      - { type: heading, level: 2, text: "Motivation" }
      - { type: paragraph }
      - { type: heading, level: 2, text: "Design" }
      - { type: paragraph }
```

### 2\. Example Usage

Consider a new RFC that only contains the summary and motivation.

**/`rfc/new-feature.md`/**

```md
# RFC: A New Feature

- Start Date: 2025-09-28
- RFC Status: Draft
- Owners: @some-user

## Summary

A brief summary of the new feature.

## Motivation

Why this feature is important.
```

  - **Running validation for the `draft` section will PASS:**
    `go run ./cmd/glint --section=draft rfc/new-feature.md`

  - **Running validation for the `complete` section will FAIL** with an error like:
    `ERROR rfc/new-feature.md:14 > incomplete document for section 'complete': expected more content based on schema rule=markdown.schema.structure`

## Non-Goals

  - This validator does not check inline content (e.g., broken links, text formatting within a paragraph).
  - It is not a stylistic linter; it does not check for line length, grammar, or spelling.
  - It does not support "autofix" capabilities for structural errors.
  - It does not validate the content of code blocks.

## Open Questions

  - Should validation continue after the first error to report all structural issues at once?
  - Should we implement the placeholder inline validation rules (`contains`, `max_length`) that exist in the schema struct?
  - Should we add support for more complex rules, such as optional blocks or a "one-of" choice between block types?