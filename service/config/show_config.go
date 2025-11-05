package config

import (
	"fmt"
)

// ShowConfig displays current configuration
func (s *CfgService) ShowConfig() error {
	cfg, err := s.configRepository.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	all := s.configRepository.GetAll(cfg)
	hasAnyValue := false
	if val, ok := all["rules_dir"].(string); ok && val != "" {
		fmt.Printf("rules-dir: %s\n", val)
		hasAnyValue = true
	}
	if val, ok := all["file_patterns"].(string); ok && val != "" {
		fmt.Printf("file-patterns: %s\n", val)
		hasAnyValue = true
	}
	if val, ok := all["overwrite_headers"].(bool); ok && val {
		fmt.Printf("overwrite-headers: true\n")
		hasAnyValue = true
	}
	if val, ok := all["git_without_push"].(bool); ok && val {
		fmt.Printf("git-without-push: true\n")
		hasAnyValue = true
	}
	if !hasAnyValue {
		fmt.Println("No configuration values set.")
	}

	return nil
}
