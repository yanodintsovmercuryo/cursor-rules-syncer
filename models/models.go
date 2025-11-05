package models

// OperationType represents the type of file operation
type OperationType string

const (
	OperationAdd    OperationType = "add"
	OperationDelete OperationType = "delete"
	OperationUpdate OperationType = "update"
)

// FileOperation represents a file operation with metadata
type FileOperation struct {
	Type         OperationType `json:"type"`
	SourcePath   string        `json:"source_path"`
	TargetPath   string        `json:"target_path"`
	RelativePath string        `json:"relative_path"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Operations []FileOperation `json:"operations"`
	HasChanges bool            `json:"has_changes"`
}

// IgnorePattern represents a compiled ignore pattern
type IgnorePattern struct {
	Pattern    string `json:"pattern"`
	Regex      string `json:"regex"`
	IsNegation bool   `json:"is_negation"`
}

// SyncOptions contains configuration for sync operations
type SyncOptions struct {
	RulesDir         string
	GitWithoutPush   bool
	OverwriteHeaders bool
	FilePatterns     string // Comma-separated file patterns to sync (e.g., "local_*.mdc,translate/*.md")
}
