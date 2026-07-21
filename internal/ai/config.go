package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".phanes-dna")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadUserConfig() (Config, error) {
	var cfg Config
	p, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg, err
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg, nil
}

func SaveUserConfig(cfg Config) error {
	p, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
