// go.mod
module github.com/MadeByDoug/glint

go 1.23.0

toolchain go1.24.6

// The replace directive tells the Go toolchain that any import path starting
// with "github.com/MadeByDoug/glint" should be resolved from the local filesystem
// (the "." indicates the current directory, where go.mod is located).
// This prevents any network activity for your local module.
replace github.com/MadeByDoug/glint => .

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
)

require (
	github.com/rs/zerolog v1.34.0
	golang.org/x/sys v0.35.0 // indirect
)
