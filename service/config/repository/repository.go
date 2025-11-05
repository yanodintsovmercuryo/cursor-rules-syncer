package repository

import (
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

// ConfigRepository handles loading and saving configuration
type ConfigRepository struct {
}

// NewConfigRepository creates a new ConfigRepository
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// Load loads configuration from file
func (m *ConfigRepository) Load() (*config.Config, error) {
	return config.Load()
}

// LoadOrDefault loads configuration or returns empty config on error
func (m *ConfigRepository) LoadOrDefault() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		return &config.Config{}
	}
	return cfg
}
