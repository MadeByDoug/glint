# glint: A Directory Naming Linter

`glint` is a fast, configuration-driven linter for enforcing consistent naming conventions across your project's directories. It helps maintain a clean and predictable repository structure by ensuring directory names adhere to predefined rules. The CLI can output diagnostics as text or JSON, stream structured logs, and emit debug artifacts that inspect the derived configuration.

## Features

-   **Path-based Targeting**: Apply rules to specific parts of your project using regular-expression selectors (e.g., `^src/components/[^/]+$`).
-   **Casing Rules**: Enforce `kebab-case`, `snake_case`, `camelCase`, `PascalCase`, and more via predicates.
-   **Affix Rules**: Require or prohibit specific prefixes and suffixes (e.g., require `comp-` prefix, prohibit `_` prefix).
-   **Disallow Patterns**: Block specific directory names or names matching a regular expression (e.g., disallow `^legacy$`, `^temp-.*$`).
-   **Flexible Configuration**: Define rules in a project-root `.glint` (or `.glint.yaml`) file, override values via `GLINT_*` environment variables, or supply CLI flags for ad-hoc changes.
-   **Structured Diagnostics**: Choose between human-friendly text output or machine-readable JSON, and control how much detail is surfaced.
-   **Debuggable Runs**: Enable debug mode to capture the normalized configuration and linter plan for inspection.

## Configuration

`glint` looks for configuration in a `.glint` (optionally `.glint.yaml` / `.glint.yml`) file at the project root. You can provide caller defaults via the CLI, rely on environment variables (e.g., `GLINT_LINTER__RULES__0__SELECTOR__PATH=^src/[^/]+$`), or point directly at an alternate config file.

### Example Configuration

Here is an example that demonstrates several common rules for directory names:

```yaml
# .glint (or .glint.yaml)
linter:
  version: 1
  rules:
    - id: R101
      selector:
        # Select all directories directly under src/components
        kind: folder
        path:
          - "^src/components/[^/]+$"
      apply:
        checks:
          - folderName:
              # Enforce that component folders are in kebab-case.
              predicates:
                - "kebab"
              severity: "error"

    - id: R102
      selector:
        # Select every directory in the project
        kind: folder
        path:
          - ".+"
      apply:
        checks:
          - folderName:
              # Prohibit temporary or private folders from being committed.
              prohibitPrefix: "_"
              disallow:
                - "^temp-.*$"
                - "^tmp-.*$"
              severity: "warn"

    - id: W201
      selector:
        # Select directories at the root of the project
        kind: folder
        path:
          - "^[^/]+$"
      apply:
        checks:
          - folderName:
              # Discourage legacy patterns at the top level.
              disallow:
                - "^legacy$"
                - "^old$"
              message: "Avoid creating top-level legacy directories."
              severity: "warn"
```

## Usage

Download the source code repository and build with Go.

```bash
go build ./cmd/glint
```

To run the linter on the current project, navigate to the project root and execute:

```bash
glint --dir . \
      --config /path/to/config.yaml \
      --env prod \
      --debug \
      --diag-level warn \
      --error-format json \
      --error-output stdout \
      --log-output logs/glint.log
```

### Common CLI Flags

`--dir`
: Root directory whose tree will be linted.

`--config`
: Absolute or relative path to a config file; overrides the default `.glint` lookup.

`--env`
: Selects an environment-specific config overlay (e.g., `dev`, `stg`, `prod`).

`--debug`
: Writes normalized config and rule plan to the debug stream for inspection.

`--diag-level`
: Minimum severity that will be emitted (`error`, `warn`, `note`).

`--error-format`
: Output format for diagnostics (`text`, `json`).

`--error-output`
: Destination for diagnostics (`stderr`, `stdout`, or file path).

`--log-output`
: Destination for structured logs (`stderr`, `stdout`, `none`, or file path).

`--log-level`
: Logging verbosity (`off`, `error`, `info`).

`--error-output` and `--log-output` accept filesystem paths; `glint` cleans and validates these to prevent escaping the intended directory.

Environment variables prefixed with `GLINT_` mirror the configuration hierarchy. Double underscores (`__`) denote nested keys, and values are merged before CLI overrides are applied. CLI flags always win over environment variables, which themselves win over config files, allowing you to tweak behaviour without editing committed configuration.
