// internal/app/config/loader.go
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	model "github.com/MrBigCode/glint/internal/app/config/model"
	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	keyDelimiter     = "."
	doubleUnderscore = "__"
)

// Load resolves config using load(opts), applies CLI overrides, then validates & normalizes.
// All diagnostics (load + validate + normalize) are returned.
// internal/app/config/loader.go
func Load(opts model.Options, cliOverrides map[string]any) (model.AppConfig, []reporting.Report, error) {
	k := koanf.New(keyDelimiter)
	var diags []reporting.Report
	var appConfig model.AppConfig

	diags = append(diags, loadDefaults(k, opts)...)

    searchDirs, envName, dirDiags := resolveConfigSearch(k, opts)
    diags = append(diags, dirDiags...)
    diags = append(diags, loadConfigFiles(k, searchDirs, envName)...)

	if opts.ConfigPath != "" {
		diags = append(diags, loadYAML(k, opts.ConfigPath, true)...)
	}

	diags = append(diags, loadEnvOverrides(k, opts.EnvPrefix)...)
	diags = append(diags, applyCLIOverrides(k, cliOverrides)...)
	diags = append(diags, postProcessConfig(k)...)

	if err := unmarshalStrict(k, &appConfig); err != nil {
		diags = append(diags, reporting.Error("C-LOAD-040",
			fmt.Sprintf("unmarshal final config (unknown field?): %v", err)))
		return model.AppConfig{}, diags, nil
	}

	return appConfig, diags, nil
}

func resolveConfigSearch(k *koanf.Koanf, opts model.Options) ([]string, string, []reporting.Report) {
	var diags []reporting.Report

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		diags = append(diags, reporting.Warning("C-LOAD-010",
			fmt.Sprintf("unable to determine current working directory: %v", cwdErr)))
	}

	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		diags = append(diags, reporting.Warning("C-LOAD-011",
			fmt.Sprintf("unable to determine user home directory: %v", homeErr)))
	}

	dirs := make([]string, 0, 2)
	if homeErr == nil {
		dirs = append(dirs, home)
	}
	if cwdErr == nil {
		dirs = append(dirs, cwd)
	}

	envName := k.String(KeyEnv)
	if opts.EnvName != "" {
		envName = opts.EnvName
	}
	envName = strings.TrimSpace(envName)

	return dirs, envName, diags
}

func loadConfigFiles(k *koanf.Koanf, searchDirs []string, envName string) []reporting.Report {
    var diags []reporting.Report
    for _, dir := range searchDirs {
        for _, candidate := range configCandidates(dir, envName) {
            diags = append(diags, loadYAML(k, candidate, false)...)
        }
    }
    return diags
}

func configCandidates(dir, envName string) []string {
    base := filepath.Join(dir, ".glint")
    candidates := []string{
        base,
        base + ".yaml",
        base + ".yml",
    }

    if envName != "" {
        envBase := base + "." + envName
        candidates = append(candidates,
            envBase,
            envBase+".yaml",
            envBase+".yml",
        )
    }

    return candidates
}

func applyCLIOverrides(k *koanf.Koanf, overrides map[string]any) []reporting.Report {
	if len(overrides) == 0 {
		return nil
	}

	if err := k.Load(confmap.Provider(overrides, keyDelimiter), nil); err != nil {
		return []reporting.Report{reporting.Error("C-LOAD-030",
			fmt.Sprintf("apply CLI overrides: %v", err))}
	}
	return nil
}

// loadYAML loads a YAML config file. When required==true, a missing file is an error diagnostic.
func loadYAML(k *koanf.Koanf, path string, required bool) []reporting.Report {
	if diag := invalidYAMLExtension(path, required); diag != nil {
		return []reporting.Report{*diag}
	}

	exists, diags := ensureYAMLExists(path, required)
	if diags != nil {
		return diags
	}
	if !exists {
		return nil
	}

	return parseYAMLFile(k, path)
}

func invalidYAMLExtension(path string, required bool) *reporting.Report {
    if !required {
        return nil
    }

    ext := strings.ToLower(filepath.Ext(path))
    if ext == "" || ext == ".yaml" || ext == ".yml" {
        return nil
    }

    rep := reporting.Error("C-LOAD-020",
        fmt.Sprintf("unsupported config file extension for %q (only .yaml/.yml or none)", path))
    return &rep
}

func ensureYAMLExists(path string, required bool) (bool, []reporting.Report) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if required {
				rep := reporting.Error("C-LOAD-021",
					fmt.Sprintf("required config file not found: %s", path))
				return false, []reporting.Report{rep}
			}
			return false, nil
		}
		rep := reporting.Warning("C-LOAD-024",
			fmt.Sprintf("cannot stat config file %s: %v", path, err))
		return false, []reporting.Report{rep}
	}
	return true, nil
}

func parseYAMLFile(k *koanf.Koanf, path string) []reporting.Report {
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		rep := reporting.Error("C-LOAD-022",
			fmt.Sprintf("failed to load %s: %v", path, err))
		return []reporting.Report{rep}
	}
	rep := reporting.Note("C-LOAD-120",
		fmt.Sprintf("loaded config file: %s", path))
	return []reporting.Report{rep}
}

// loadEnvOverrides consumes GLINT_* environment variables (or the provided prefix)
// and maps double underscores to nested keys (e.g., GLINT_LINTER__RULES -> linter.rules).
func loadEnvOverrides(k *koanf.Koanf, prefix string) []reporting.Report {
	var diags []reporting.Report
	if prefix = strings.TrimSpace(prefix); prefix == "" {
		prefix = DefaultEnvPrefix
	}

	found := false
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, prefix) {
			found = true
			break
		}
	}
	if !found {
		return diags
	}

	transform := func(key string) string {
		key = strings.TrimPrefix(key, prefix)
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, doubleUnderscore, keyDelimiter)
		return key
	}

	if err := k.Load(env.Provider(prefix, keyDelimiter, transform), nil); err != nil {
		diags = append(diags, reporting.Error("C-LOAD-023",
			fmt.Sprintf("load env overrides with prefix %q: %v", prefix, err)))
	} else {
		diags = append(diags, reporting.Note("C-LOAD-121",
			fmt.Sprintf("loaded env overrides with prefix %q", prefix)))
	}
	return diags
}

// loadDefaults are always applied first; later loads override them.
func loadDefaults(k *koanf.Koanf, opts model.Options) []reporting.Report {
	var diags []reporting.Report

	cwd, _ := os.Getwd()
	defaults := map[string]any{
		KeyEnv:   "dev",
		KeyDir:   cwd,
		KeyDebug: false,
		KeyLinterRoot: map[string]any{
			"version": 1, // explicit baseline, rules may come from files
		},
	}
	if err := k.Load(confmap.Provider(defaults, keyDelimiter), nil); err != nil {
		diags = append(diags, reporting.Error("C-LOAD-000",
			fmt.Sprintf("load defaults: %v", err)))
	} else {
		diags = append(diags, reporting.Note("C-LOAD-100", "loaded default configuration"))
	}

	// Caller-supplied defaults (optional, still lowest precedence)
	if len(opts.Defaults) > 0 {
		if err := k.Load(confmap.Provider(opts.Defaults, keyDelimiter), nil); err != nil {
			diags = append(diags, reporting.Warning("C-LOAD-001",
				fmt.Sprintf("apply caller defaults: %v", err)))
		} else {
			diags = append(diags, reporting.Note("C-LOAD-101", "applied caller defaults"))
		}
	}
	return diags
}

func unmarshalStrict(k *koanf.Koanf, out *model.AppConfig) error {
	b, err := json.Marshal(k.Raw())
	if err != nil {
		return fmt.Errorf("marshal koanf tree: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("strict decode: %w", err)
	}
	return nil
}
