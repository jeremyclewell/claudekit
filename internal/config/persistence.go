package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const persistenceFile = ".claudekit.json"

// GetPersistenceFilePath returns the full path to the persistence config file.
func GetPersistenceFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, persistenceFile), nil
}

// Load reads the persistence config from .claudekit.json if it exists.
func Load() (*PersistenceConfig, error) {
	path, err := GetPersistenceFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &PersistenceConfig{}, nil
		}
		return nil, err
	}

	var cfg PersistenceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the current config to .claudekit.json for future runs.
func Save(cfg Config) error {
	path, err := GetPersistenceFilePath()
	if err != nil {
		return err
	}

	persistCfg := PersistenceConfig{
		ProjectName:     cfg.ProjectName,
		Language:        cfg.Language,
		Subagents:       cfg.Subagents,
		Hooks:           cfg.Hooks,
		SlashCommands:   cfg.SlashCommands,
		MCPs:            cfg.MCPs,
		IncludeCLAUDE:   cfg.IncludeCLAUDE,
		IncludeAgents:   cfg.IncludeAgents,
		IncludeHooks:    cfg.IncludeHooks,
		IncludeExamples: cfg.IncludeExamples,
	}

	data, err := json.MarshalIndent(persistCfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
