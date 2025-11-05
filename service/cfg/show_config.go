package cfg

import (
	"fmt"
)

// ShowConfig displays current configuration
func (s *CfgService) ShowConfig() error {
	cfg, err := s.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	all := cfg.GetAll()
	hasAnyValue := false
	if val := all["rules_dir"].(string); val != "" {
		fmt.Printf("rules-dir: %s\n", val)
		hasAnyValue = true
	}
	if val := all["file_patterns"].(string); val != "" {
		fmt.Printf("file-patterns: %s\n", val)
		hasAnyValue = true
	}
	if val := all["overwrite_headers"].(bool); val {
		fmt.Printf("overwrite-headers: true\n")
		hasAnyValue = true
	}
	if val := all["git_without_push"].(bool); val {
		fmt.Printf("git-without-push: true\n")
		hasAnyValue = true
	}
	if !hasAnyValue {
		fmt.Println("No configuration values set.")
	}

	return nil
}
