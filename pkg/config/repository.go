package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	configDirName  = ".config"
	configFileName = "cursync.toml"
)

// ConfigRepositoryInterface defines interface for config repository
type ConfigRepositoryInterface interface {
	Load() (*Config, error)
	LoadOrDefault() *Config
	Save(cfg *Config) error
	Set(cfg *Config, key string, value interface{}) error
	Get(cfg *Config, key string) (interface{}, error)
	GetAll(cfg *Config) map[string]interface{}
}

// ConfigRepository handles loading and saving configuration
type ConfigRepository struct{}

// NewConfigRepository creates a new ConfigRepository
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// GetConfigPath returns the path to the config file
func (r *ConfigRepository) GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, configDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, configFileName), nil
}

// Load loads configuration from file
func (r *ConfigRepository) Load() (*Config, error) {
	configPath, err := r.GetConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	// If config file doesn't exist, return empty config
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// LoadOrDefault loads configuration or returns empty config on error
func (r *ConfigRepository) LoadOrDefault() *Config {
	cfg, err := r.Load()
	if err != nil {
		return &Config{}
	}
	return cfg
}

// Save saves configuration to file
func (r *ConfigRepository) Save(cfg *Config) error {
	configPath, err := r.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Set sets a configuration value by key
func (r *ConfigRepository) Set(cfg *Config, key string, value interface{}) error {
	switch key {
	case "rules-dir", "rules_dir":
		if val, ok := value.(string); ok {
			if val == "" {
				cfg.RulesDir = ""
			} else {
				cfg.RulesDir = val
			}
		}
	case "file-patterns", "file_patterns":
		if val, ok := value.(string); ok {
			if val == "" {
				cfg.FilePatterns = ""
			} else {
				cfg.FilePatterns = val
			}
		}
	case "overwrite-headers", "overwrite_headers":
		if val, ok := value.(bool); ok {
			cfg.OverwriteHeaders = val
		} else if valStr, ok := value.(string); ok && valStr == "" {
			cfg.OverwriteHeaders = false
		}
	case "git-without-push", "git_without_push":
		if val, ok := value.(bool); ok {
			cfg.GitWithoutPush = val
		} else if valStr, ok := value.(string); ok && valStr == "" {
			cfg.GitWithoutPush = false
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Get returns configuration value by key
func (r *ConfigRepository) Get(cfg *Config, key string) (interface{}, error) {
	switch key {
	case "rules-dir", "rules_dir":
		return cfg.RulesDir, nil
	case "file-patterns", "file_patterns":
		return cfg.FilePatterns, nil
	case "overwrite-headers", "overwrite_headers":
		return cfg.OverwriteHeaders, nil
	case "git-without-push", "git_without_push":
		return cfg.GitWithoutPush, nil
	default:
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
}

// GetAll returns all configuration as a map
func (r *ConfigRepository) GetAll(cfg *Config) map[string]interface{} {
	return map[string]interface{}{
		"rules_dir":         cfg.RulesDir,
		"file_patterns":     cfg.FilePatterns,
		"overwrite_headers": cfg.OverwriteHeaders,
		"git_without_push":  cfg.GitWithoutPush,
	}
}

// Load loads configuration from file (global function for backward compatibility)
func Load() (*Config, error) {
	repo := NewConfigRepository()
	return repo.Load()
}

// GetConfigPath returns the path to the config file (global function for backward compatibility)
func GetConfigPath() (string, error) {
	repo := NewConfigRepository()
	return repo.GetConfigPath()
}
