// go.mod
module github.com/MrBigCode/glint

go 1.23.0

toolchain go1.24.6

// The replace directive tells the Go toolchain that any import path starting
// with "github.com/MrBigCode/glint" should be resolved from the local filesystem
// (the "." indicates the current directory, where go.mod is located).
// This prevents any network activity for your local module.
replace github.com/MrBigCode/glint => .

require golang.org/x/term v0.34.0

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
)

require (
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/confmap v1.0.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.0
	github.com/knadh/koanf/v2 v2.2.2
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	golang.org/x/sys v0.35.0 // indirect
)
