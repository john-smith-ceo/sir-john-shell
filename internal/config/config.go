package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UserConfig struct {
	User string `json:"user"`
}

const defaultUser = "Sir"

func Load() (UserConfig, error) {
	cfg := UserConfig{User: defaultUser}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil
	}

	if env := os.Getenv("SJS_USER"); env != "" {
		cfg.User = env
		return cfg, nil
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
	return cfg, nil
}
