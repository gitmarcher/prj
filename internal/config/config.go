package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Source struct {
	Base string `mapstructure:"base"`
}

type Config struct {
	Workspace        string            `mapstructure:"workspace"`
	Sources          map[string]Source `mapstructure:"sources"`
	KnowledgeAgeDays int               `mapstructure:"knowledge_age_days"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	cfgDir := filepath.Join(home, ".config", "prj")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgDir)

	v.SetDefault("workspace", filepath.Join(home, "Projects"))
	v.SetDefault("knowledge_age_days", 7)

	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		if err := writeDefaultConfig(cfgFile, home); err != nil {
			return nil, err
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	// Expand ~ in workspace path
	if len(cfg.Workspace) >= 2 && cfg.Workspace[:2] == "~/" {
		cfg.Workspace = filepath.Join(home, cfg.Workspace[2:])
	}

	return &cfg, nil
}

func (c *Config) ResolveCloneURL(source, repo string) (string, error) {
	src, ok := c.Sources[source]
	if !ok {
		return "", fmt.Errorf("source %q not found in config; add it to ~/.config/prj/config.yaml", source)
	}
	return fmt.Sprintf("%s/%s.git", src.Base, repo), nil
}

func writeDefaultConfig(path, home string) error {
	content := `# prj configuration
workspace: ~/Projects

sources:
  # example:
  #   carousell:
  #     base: git@github.com:carousell
`
	return os.WriteFile(path, []byte(content), 0o644)
}
