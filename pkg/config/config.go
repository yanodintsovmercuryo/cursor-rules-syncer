package config

// Config holds configuration values
type Config struct {
	RulesDir         string `toml:"rules_dir,omitempty"`
	FilePatterns     string `toml:"file_patterns,omitempty"`
	OverwriteHeaders bool   `toml:"overwrite_headers,omitempty"`
	GitWithoutPush   bool   `toml:"git_without_push,omitempty"`
}
