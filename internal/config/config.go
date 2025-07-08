package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	var path string
	home, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}

	path = filepath.Join(home, configFileName)

	return path, nil
}

func Read(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return err
	}
	return nil
}

func write(cfg *Config) error {
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *Config) SetUser(name string) error {
	cfg.CurrentUserName = name
	return write(cfg)
}
