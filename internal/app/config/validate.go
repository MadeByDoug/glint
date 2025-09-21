// internal/app/config/validate.go
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/knadh/koanf/v2"
)

// postProcessConfig normalizes and lightly validates config in-place.
func postProcessConfig(k *koanf.Koanf) []reporting.Report {
	var diags []reporting.Report
	diags = append(diags, normalizeDir(k)...)
	diags = append(diags, normalizeEnv(k)...)
	diags = append(diags, ensureLinterVersion(k)...)
	return diags
}

func normalizeDir(k *koanf.Koanf) []reporting.Report {
	dir := strings.TrimSpace(k.String(KeyDir))
	if dir == "" {
		return nil
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return []reporting.Report{reporting.Warning("C-NORM-001",
			fmt.Sprintf("could not determine absolute path for dir %q: %v", dir, err))}
	}

	if abs == dir {
		return nil
	}

	if err := k.Set(KeyDir, abs); err != nil {
		return []reporting.Report{reporting.Error("C-NORM-002",
			fmt.Sprintf("failed to set normalized dir: %v", err))}
	}

	return []reporting.Report{reporting.Note("C-NORM-100", "normalized dir to absolute path")}
}

func normalizeEnv(k *koanf.Koanf) []reporting.Report {
	env := strings.TrimSpace(k.String(KeyEnv))
	if env == "" {
		return nil
	}

	lc := strings.ToLower(env)
	if lc == env {
		return nil
	}

	if err := k.Set(KeyEnv, lc); err != nil {
		return []reporting.Report{reporting.Error("C-NORM-003",
			fmt.Sprintf("failed to set normalized env: %v", err))}
	}

	return []reporting.Report{reporting.Note("C-NORM-101", "normalized env to lowercase")}
}

func ensureLinterVersion(k *koanf.Koanf) []reporting.Report {
	if k.Int(KeyLinterVersion) != 0 {
		return nil
	}

	if err := k.Set(KeyLinterVersion, 1); err != nil {
		return []reporting.Report{reporting.Error("C-NORM-004",
			fmt.Sprintf("failed to set default linter.version: %v", err))}
	}

	return []reporting.Report{reporting.Note("C-NORM-110", "defaulted linter.version to 1")}
}
