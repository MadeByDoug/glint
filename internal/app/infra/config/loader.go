package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"

	"github.com/MadeByDoug/glint/internal/app/infra/logging"
)

// Config represents the structure of the .glint.yaml file.
// Add fields as the configuration grows; unknown keys are ignored.
type Config struct {
	Version int               `koanf:"version" yaml:"version"`
	Consts  map[string]string `koanf:"consts" yaml:"consts"`
	EnvVars []string          `koanf:"env-vars" yaml:"env-vars"`
	Rules   []RuleConfig      `koanf:"rules" yaml:"rules"`
	Policies map[string]CheckConfig `koanf:"policies" yaml:"policies"`
}

type RuleConfig struct {
	ID        string           `koanf:"id"`
	DependsOn []string         `koanf:"depends_on"`
	Selectors []SelectorConfig `koanf:"selectors"` // MOVED: Selectors are now part of the rule.
	Checks    []CheckConfig    `koanf:"checks"`    // CHANGED: Checks is now a slice of CheckConfig.
}

type SelectorConfig struct {
	Type     string      `koanf:"type" yaml:"type"`
	Selector interface{} `koanf:"selector" yaml:"selector"`
}

// CheckConfig defines a single, specific validation action to be performed.
type CheckConfig struct {
	ID        string                 `koanf:"id"`
	Uses      string                 `koanf:"uses"`      // The type of check, e.g., "markdown-schema".
	DependsOn []string               `koanf:"depends_on"`// For future dependency management.
	If        string                 `koanf:"if"`        // For future conditional execution.
	With      map[string]interface{} `koanf:"with"`      // Parameters for the check.
}

type SchemaRef struct {
	Schema string `koanf:"schema" yaml:"schema"`
}




// SelectorStrings returns the selector value as a slice of strings.
func (s SelectorConfig) SelectorStrings() ([]string, error) {
	switch v := s.Selector.(type) {
	case string:
		return []string{v}, nil
	case []string:
		return v, nil
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("selector list for type %s must contain only strings", s.Type)
			}
			out = append(out, str)
		}
		return out, nil
	default:
		if v == nil {
			return nil, fmt.Errorf("selector value for type %s is missing", s.Type)
		}
		return nil, fmt.Errorf("selector value for type %s has unsupported type %T", s.Type, v)
	}
}

// Load reads configuration from the provided path using Koanf.
// If path is empty, it defaults to ".glint.yaml" in the current directory.
func Load(path string) (*Config, error) {
	if path == "" {
		path = ".glint.yaml"
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("stat config file %q: %w", path, err)
	}

	// Expand environment variables in the raw YAML content first.
	expandedContent := os.ExpandEnv(string(content))

	// Now, load the expanded content.
	k := koanf.New(".")
	if err := k.Load(rawbytes.Provider([]byte(expandedContent)), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load config via koanf: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	devLog := logging.Get()
	devLog.Info().
		Str("config.path", path).
		Msg("configuration loaded and variables expanded")

	return &cfg, nil
}
