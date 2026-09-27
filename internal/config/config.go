package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UserConfig struct {
	User       string         `json:"user"`
	FontFamily string         `json:"fontFamily,omitempty"`
	Theme      map[string]any `json:"theme,omitempty"`
	UI         map[string]any `json:"ui,omitempty"`
}

const defaultUser = "Sir"

func Load() (UserConfig, error) {
	cfg := UserConfig{User: defaultUser}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil
	}

	envUser := os.Getenv("SJS_USER")
	if envUser != "" {
		cfg.User = envUser
	}

	path := filepath.Join(home, ".sir-john-shell", "config")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.User == "" {
		cfg.User = defaultUser
	}
	if envUser != "" {
		cfg.User = envUser
	}
	return cfg, nil
}

// Save persists the shell profile while keeping the config file private.
func Save(cfg UserConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".sir-john-shell")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "config")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
