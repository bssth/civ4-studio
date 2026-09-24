package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const mainConfigFile = "config.json"
const configDirName = "civ4-studio"

type Config struct {
	GameDir  string `json:"game_dir"`
	Mod      string `json:"mod"`
	AutoSave bool   `json:"auto_save"`
}

// ConfigPath returns the config file location in the user config directory
// (e.g. %AppData%\civ4-studio\config.json). The game is usually installed into
// Program Files, so the working directory is often not writable.
func ConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return mainConfigFile
	}
	return filepath.Join(dir, configDirName, mainConfigFile)
}

// GetConfig returns the current configuration. If the configuration file does not exist, it will return the default configuration.
func GetConfig() (Config, bool) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		// Older versions kept config.json in the working directory
		data, err = os.ReadFile(mainConfigFile)
		if err != nil {
			return GetDefaultConfig(), false
		}
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
	data, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		return err
	}

	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetDefaultConfig returns the default configuration.
func GetDefaultConfig() Config {
	config := Config{
		AutoSave: true,
	}

	if _, err := os.Stat(steamDefaultGameDir); err == nil {
		config.GameDir = steamDefaultGameDir
	}

	return config
}
