package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents persistent deepmd user settings in ~/.deepmd/config.json.
type Config struct {
	// MouseWheelEnabled controls whether mouse wheel scrolling is active.
	// Defaults to false to prevent erratic scrolling in terminal emulators.
	MouseWheelEnabled bool `json:"mouse_wheel_enabled"`

	// MouseWheelDelta is the number of lines scrolled per wheel notch.
	// Defaults to 1 for slow, smooth, controlled scrolling.
	MouseWheelDelta int `json:"mouse_wheel_delta"`

	// Theme is the Glamour markdown rendering theme (dark, light, dracula, notty, auto).
	Theme string `json:"theme"`
}

// DefaultConfig returns the default configuration for deepmd.
func DefaultConfig() Config {
	return Config{
		MouseWheelEnabled: false, // Disabled by default per user preference
		MouseWheelDelta:   1,     // Slow 1-line scroll when enabled
		Theme:             "dark",
	}
}

// GetConfigPath returns the path to ~/.deepmd/config.json.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".deepmd")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig loads configuration from ~/.deepmd/config.json, creating it if absent.
func LoadConfig() Config {
	cfg := DefaultConfig()
	path, err := GetConfigPath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Auto-persist default config so user can easily inspect and edit it
		_ = SaveConfig(cfg)
		return cfg
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	if cfg.MouseWheelDelta <= 0 {
		cfg.MouseWheelDelta = 1
	}

	return cfg
}

// SaveConfig writes the configuration to ~/.deepmd/config.json.
func SaveConfig(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if cfg.MouseWheelDelta <= 0 {
		cfg.MouseWheelDelta = 1
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
