// internal/app/config/model/app.go
package model

// AppConfig defines the top-level settings.
type AppConfig struct {
    Env    string           `koanf:"env"`
    Dir    string           `koanf:"dir"`
    Debug  bool             `koanf:"debug"`
    Linter LinterConfig     `koanf:"linter"`
}

// Options control how config is resolved.
type Options struct {
	ConfigPath string
	EnvName    string
	Defaults   map[string]any
	EnvPrefix  string
}
