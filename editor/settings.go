package editor

import (
	"encoding/json"
	"os"
)

const mainConfigFile = "config.json"

type Config struct {
	GameDir  string `json:"game_dir"`
	Mod      string `json:"mod"`
	AutoSave bool   `json:"auto_save"`
}

// GetConfig returns the current configuration. If the configuration file does not exist, it will return the default configuration.
func GetConfig() (Config, bool) {
	data, err := os.ReadFile(mainConfigFile)
	if err != nil {
		return GetDefaultConfig(), false
	}

	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return GetDefaultConfig(), false
	}

	return config, true
}

// SaveConfig saves the current configuration to the configuration file.
func SaveConfig() error {
	data, err := json.Marshal(GlobalConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(mainConfigFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetDefaultConfig returns the default configuration.
func GetDefaultConfig() Config {
	config := Config{
		AutoSave: true,
	}

	if config.GameDir == "" {
		if _, err := os.Stat(steamDefaultGameDir); err == nil {
			config.GameDir = steamDefaultGameDir
		}
	}

	return config
}
