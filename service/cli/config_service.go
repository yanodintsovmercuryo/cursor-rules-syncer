package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

// ConfigService handles configuration management
type ConfigService struct {
	output outputService
}

type outputService interface {
	PrintErrorf(format string, args ...interface{})
}

// NewConfigService creates a new ConfigService
func NewConfigService(output outputService) *ConfigService {
	return &ConfigService{
		output: output,
	}
}

// ShowConfig displays current configuration
func (s *ConfigService) ShowConfig() error {
	cfg, err := config.Load()
	if err != nil {
		s.output.PrintErrorf("Failed to load config: %v", err)
		return err
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

// UpdateConfig updates configuration based on CLI context
func (s *ConfigService) UpdateConfig(ctx *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		s.output.PrintErrorf("Failed to load config: %v", err)
		return err
	}

	// Process each flag
	if ctx.IsSet("rules-dir") {
		val := ctx.String("rules-dir")
		if err := cfg.Set("rules-dir", val); err != nil {
			s.output.PrintErrorf("Failed to set rules-dir: %v", err)
			return err
		}
		if val == "" {
			fmt.Println("Cleared rules-dir")
		} else {
			fmt.Printf("Set rules-dir to: %s\n", val)
		}
	}

	if ctx.IsSet("file-patterns") {
		val := ctx.String("file-patterns")
		if err := cfg.Set("file-patterns", val); err != nil {
			s.output.PrintErrorf("Failed to set file-patterns: %v", err)
			return err
		}
		if val == "" {
			fmt.Println("Cleared file-patterns")
		} else {
			fmt.Printf("Set file-patterns to: %s\n", val)
		}
	}

	if ctx.IsSet("overwrite-headers") {
		val := ctx.Bool("overwrite-headers")
		if err := cfg.Set("overwrite-headers", val); err != nil {
			s.output.PrintErrorf("Failed to set overwrite-headers: %v", err)
			return err
		}
		if val {
			fmt.Println("Set overwrite-headers to: true")
		} else {
			fmt.Println("Set overwrite-headers to: false")
		}
	}

	if ctx.IsSet("git-without-push") {
		val := ctx.Bool("git-without-push")
		if err := cfg.Set("git-without-push", val); err != nil {
			s.output.PrintErrorf("Failed to set git-without-push: %v", err)
			return err
		}
		if val {
			fmt.Println("Set git-without-push to: true")
		} else {
			fmt.Println("Set git-without-push to: false")
		}
	}

	if err := cfg.Save(); err != nil {
		s.output.PrintErrorf("Failed to save config: %v", err)
		return err
	}

	return nil
}

// HasConfigFlags checks if any config flags are set
func (s *ConfigService) HasConfigFlags(ctx *cli.Context) bool {
	return ctx.IsSet("rules-dir") || ctx.IsSet("file-patterns") || ctx.IsSet("overwrite-headers") || ctx.IsSet("git-without-push")
}
