package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MouseWheelEnabled {
		t.Fatal("expected MouseWheelEnabled to be false by default")
	}
	if cfg.MouseWheelDelta != 1 {
		t.Fatalf("expected MouseWheelDelta to be 1, got %d", cfg.MouseWheelDelta)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "deepmd-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()
	if cfg.MouseWheelEnabled {
		t.Fatal("expected newly generated config to have MouseWheelEnabled = false")
	}

	// Modify and save
	cfg.MouseWheelEnabled = true
	cfg.MouseWheelDelta = 2
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify persistence
	loaded := LoadConfig()
	if !loaded.MouseWheelEnabled {
		t.Fatal("expected loaded config to have MouseWheelEnabled = true")
	}
	if loaded.MouseWheelDelta != 2 {
		t.Fatalf("expected MouseWheelDelta = 2, got %d", loaded.MouseWheelDelta)
	}

	// Check file on disk
	configPath := filepath.Join(tempDir, ".deepmd", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to exist at %s: %v", configPath, err)
	}
}
