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
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/v2 v2.3.0
)

require (
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
)

require (
	github.com/knadh/koanf/providers/rawbytes v1.0.0
	github.com/rs/zerolog v1.34.0
	github.com/yuin/goldmark v1.7.13
	go.starlark.net v0.0.0-20250906160240-bf296ed553ea
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)
