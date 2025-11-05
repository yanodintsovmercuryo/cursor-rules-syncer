# cursync

A CLI tool for synchronizing Cursor IDE rules between a source directory and git project's `.cursor/rules` directory. The tool provides bidirectional synchronization (`pull` and `push` commands) with support for file pattern filtering, YAML header preservation, automatic git commit management, and configuration management (`cfg` command) for default values.

## Running components

- **cursync** - CLI binary that provides `pull`, `push`, and `cfg` commands for rules synchronization and configuration management

## Dependencies

### Infrastructure

- **Git repository** - Required for detecting project root and committing changes
- **File system** - Read/write access to source and destination directories

### External API

- **Git CLI** - Required for git operations (commit, push)

## Features

### Pull Command

Pulls rules from a source directory to the current git project's `.cursor/rules` directory:

- Synchronizes files from source to project directory
- Deletes extra files in project directory that don't exist in source
- Supports file pattern filtering via `--file-patterns` flag or `CURSOR_RULES_PATTERNS` environment variable
- Preserves or overwrites YAML headers based on `--overwrite-headers` flag
- Skips copying identical files

### Push Command

Pushes rules from the current git project's `.cursor/rules` directory to the source directory:

- Synchronizes files from project to source directory
- Deletes extra files in source directory that don't exist in project
- Supports file pattern filtering
- Automatically commits changes to git repository
- Optional `--git-without-push` flag to commit without pushing

### Config Command

Manages default configuration values stored in `~/.config/cursync.toml`:

- **View configuration**: Run `cursync cfg` without flags to display current default values
- **Set defaults**: Use flags to set default values for any configuration option
- **Clear defaults**: 
  - For string flags: Set empty value (e.g., `--rules-dir=""`)
  - For bool flags: Set to `false` (e.g., `--overwrite-headers=false`)
- Configuration values are used as defaults when flags are not provided in `pull` and `push` commands

### Configuration

- **`--rules-dir` / `-d`** - Path to rules directory (overrides config file and `CURSOR_RULES_DIR` env var)
- **`--file-patterns` / `-p`** - Comma-separated file patterns (e.g., `local_*.mdc,translate/*.md`) (overrides config file and `CURSOR_RULES_PATTERNS` env var)
- **`--overwrite-headers` / `-o`** - Overwrite headers instead of preserving them
- **`--git-without-push` / `-w`** - Commit changes but don't push to remote (push command only)

**Priority order**: Command-line flags > Configuration file (`~/.config/cursync.toml`) > Environment variables

### Environment Variables

- **`CURSOR_RULES_DIR`** - Default source directory path for rules
- **`CURSOR_RULES_PATTERNS`** - Default file patterns for filtering

## Usage Examples

```bash
# Pull rules from source directory to project
cursync pull --rules-dir ~/my-rules

# Pull with file pattern filtering
cursync pull -d ~/my-rules -p "local_*.mdc"

# Push rules from project to source directory
cursync push --rules-dir ~/my-rules

# Push without git push
cursync push -d ~/my-rules -w

# Overwrite headers during sync
cursync pull -d ~/my-rules -o

# View current configuration
cursync cfg

# Set default rules directory
cursync cfg --rules-dir ~/my-rules

# Set default file patterns
cursync cfg --file-patterns "local_*.mdc,translate/*.md"

# Set default overwrite-headers flag
cursync cfg --overwrite-headers=true

# Clear default rules directory
cursync cfg --rules-dir=""

# Clear default overwrite-headers flag
cursync cfg --overwrite-headers=false
```

## Development

### Prerequisites

- Go 1.23+
- Task runner (for using Taskfile.yml)

### Building

```bash
task build
```

### Running Tests

```bash
task test
```

### Formatting and Linting

```bash
task fmt
task lint
```

### Generating Mocks

```bash
task generate
```

## Architecture & Flow

The tool follows a clean architecture pattern with separation of concerns:

- **`pkg/`** - Static utilities without dependencies (file operations, path utilities, git operations, output formatting)
- **`service/`** - Business logic with dependencies:
  - **`service/file/`** - File operations facade (comparator, copier, filter sub-services)
  - **`service/sync/`** - Main synchronization service orchestrating pull/push operations
- **`models/`** - Data structures and types

### Synchronization Flow

1. **Pull Flow:**
   - Get rules source directory from flag or env var
   - Detect git root directory
   - Find source files (with optional pattern filtering)
   - Clean up extra files in destination
   - Copy files maintaining directory structure
   - Skip identical files

2. **Push Flow:**
   - Get rules source directory from flag or env var
   - Detect git root directory
   - Verify project `.cursor/rules` directory exists
   - Find project files (with optional pattern filtering)
   - Clean up extra files in source directory
   - Copy files maintaining directory structure
   - Commit changes to git repository (with optional push)

## Troubleshooting Guide

### "rules directory not specified" error

Ensure either `--rules-dir` flag is provided or `CURSOR_RULES_DIR` environment variable is set.

### "failed to find git root" error

The tool searches recursively for either a `.git` directory or `.cursor` folder starting from the current directory. Ensure you're running the command from within a git repository or a directory containing a `.cursor` folder.

### "project rules directory not found" error (push command)

The push command requires the project's `.cursor/rules` directory to exist. Create it first if needed.

### Git commit failures

Check git repository status and ensure you have proper permissions. The tool will continue synchronization even if commit fails, but will display an error message.

