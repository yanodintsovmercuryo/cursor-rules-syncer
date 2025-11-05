package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

func TestConfigPath(t *testing.T) {
	repo := config.NewConfigRepository()
	path, err := repo.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".config", "cursync.toml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestConfigLoadSave(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory for this test
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	repo := config.NewConfigRepository()
	cfg := &config.Config{
		RulesDir:         "/path/to/rules",
		FilePatterns:     "*.mdc",
		OverwriteHeaders: true,
		GitWithoutPush:   false,
	}

	if err := repo.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if diff := cmp.Diff(cfg, loaded); diff != "" {
		t.Errorf("Config mismatch (-want +got):\n%s", diff)
	}
}

func TestConfigLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	repo := config.NewConfigRepository()
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load non-existent config: %v", err)
	}

	if cfg.RulesDir != "" || cfg.FilePatterns != "" || cfg.OverwriteHeaders || cfg.GitWithoutPush {
		t.Errorf("Expected empty config, got %+v", cfg)
	}
}

func TestConfigSetGet(t *testing.T) {
	repo := config.NewConfigRepository()
	cfg := &config.Config{}

	// Test string values
	if err := repo.Set(cfg, "rules-dir", "/test/path"); err != nil {
		t.Fatalf("Failed to set rules-dir: %v", err)
	}
	val, err := repo.Get(cfg, "rules-dir")
	if err != nil {
		t.Fatalf("Failed to get rules-dir: %v", err)
	}
	if val != "/test/path" {
		t.Errorf("Expected /test/path, got %v", val)
	}

	// Test clearing string value
	if setErr := repo.Set(cfg, "rules-dir", ""); setErr != nil {
		t.Fatalf("Failed to clear rules-dir: %v", setErr)
	}
	val, err = repo.Get(cfg, "rules-dir")
	if err != nil {
		t.Fatalf("Failed to get rules-dir: %v", err)
	}
	if val != "" {
		t.Errorf("Expected empty string, got %v", val)
	}

	// Test bool values
	if setErr := repo.Set(cfg, "overwrite-headers", true); setErr != nil {
		t.Fatalf("Failed to set overwrite-headers: %v", setErr)
	}
	val, err = repo.Get(cfg, "overwrite-headers")
	if err != nil {
		t.Fatalf("Failed to get overwrite-headers: %v", err)
	}
	if val != true {
		t.Errorf("Expected true, got %v", val)
	}

	if setErr := repo.Set(cfg, "overwrite-headers", false); setErr != nil {
		t.Fatalf("Failed to set overwrite-headers to false: %v", setErr)
	}
	val, err = repo.Get(cfg, "overwrite-headers")
	if err != nil {
		t.Fatalf("Failed to get overwrite-headers: %v", err)
	}
	if val != false {
		t.Errorf("Expected false, got %v", val)
	}

	// Test invalid key
	if err := repo.Set(cfg, "invalid-key", "value"); err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestConfigGetAll(t *testing.T) {
	repo := config.NewConfigRepository()
	cfg := &config.Config{
		RulesDir:         "/test/rules",
		FilePatterns:     "*.mdc",
		OverwriteHeaders: true,
		GitWithoutPush:   false,
	}

	all := repo.GetAll(cfg)
	expected := map[string]interface{}{
		"rules_dir":         "/test/rules",
		"file_patterns":     "*.mdc",
		"overwrite_headers": true,
		"git_without_push":  false,
	}

	if diff := cmp.Diff(expected, all); diff != "" {
		t.Errorf("GetAll mismatch (-want +got):\n%s", diff)
	}
}
