package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"

	"pleat/internal/analysis"
)

const (
	defaultMaxStatements    = 6
	defaultReturnBlockLimit = 2
)

func defaultSettings() analysis.Settings {
	return analysis.Settings{
		StatementRun:     analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ControlCuddle:    analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.Low},
		MaxStatements:    defaultMaxStatements,
		ReturnBlockLimit: defaultReturnBlockLimit,
	}
}

type ruleConfig struct {
	Enabled    *bool   `toml:"enabled"`
	Level      *string `toml:"level"`
	Max        *int    `toml:"max"`
	BlockLimit *int    `toml:"block_limit"`
}

type configFile struct {
	Rules map[string]ruleConfig `toml:"rules"`
}

func loadSettings(cwd, explicitPath string) (analysis.Settings, error) {
	settings := defaultSettings()
	path := explicitPath
	if path == "" {
		path = ".pleat.toml"
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if explicitPath == "" && os.IsNotExist(err) {
			return settings, nil
		}
		return settings, fmt.Errorf("read configuration %q: %w", path, err)
	}
	var config configFile
	metadata, err := toml.Decode(string(data), &config)
	if err != nil {
		return settings, fmt.Errorf("configuration %q: %w", path, err)
	}
	if unknown := metadata.Undecoded(); len(unknown) > 0 {
		return settings, fmt.Errorf("configuration %q: unknown setting %q", path, unknown[0].String())
	}
	ids := make([]string, 0, len(config.Rules))
	for id := range config.Rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		override := config.Rules[id]
		var rule *analysis.RuleSettings
		switch id {
		case analysis.MaxStatementRun:
			rule = &settings.StatementRun
		case analysis.ControlCuddle:
			rule = &settings.ControlCuddle
		case analysis.ReturnSeparation:
			rule = &settings.ReturnSeparation
		default:
			return settings, fmt.Errorf("unknown rule %q", id)
		}
		if override.Enabled != nil {
			rule.Enabled = *override.Enabled
		}
		if override.Level != nil {
			level := analysis.Level(*override.Level)
			if level != analysis.Low && level != analysis.Medium && level != analysis.High {
				return settings, fmt.Errorf("rule %q: level must be low, medium, or high", id)
			}
			rule.Level = level
		}
		if override.Max != nil {
			if id != analysis.MaxStatementRun || *override.Max < 1 {
				return settings, fmt.Errorf("max must be positive and belongs only to %s", analysis.MaxStatementRun)
			}
			settings.MaxStatements = *override.Max
		}
		if override.BlockLimit != nil {
			if id != analysis.ReturnSeparation || *override.BlockLimit < 0 {
				return settings, fmt.Errorf("block_limit must be nonnegative and belongs only to %s", analysis.ReturnSeparation)
			}
			settings.ReturnBlockLimit = *override.BlockLimit
		}
	}
	return settings, nil
}
