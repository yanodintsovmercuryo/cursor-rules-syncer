# Cursor Rules Syncer

A command-line tool to help synchronize [Cursor](https://cursor.sh) IDE's AI rules (`.mdc` files) between a central rules repository and your individual project repositories.

It allows you to maintain a canonical set of rules in one place and easily `pull` them into any project, or `push` local project-specific rule modifications back to the central repository.

## Prerequisites

*   Go 1.21 or higher installed.
*   Git installed and configured in your PATH.
*   An environment variable `CURSOR_RULES_DIR` set to the absolute path of your central Git repository for Cursor rules.
    
    Example:
    ```bash
    export CURSOR_RULES_DIR="/Users/yourname/path/to/your/central-cursor-rules-repo"
    ```
    It's recommended to add this export to your shell's configuration file (e.g., `.zshrc`, `.bashrc`).

## Installation

To install the `cursor-rules-syncer` CLI, run:

```bash
go install github.com/yanodintsovmercuryo/cursor-rules-syncer@latest
```

This will download the source code, compile it, and place the executable in your `$GOPATH/bin` or `$GOBIN` directory. Ensure this directory is in your system's `PATH`.

Alternatively, you can build from source:
1. Clone this repository.
2. Navigate to the repository directory.
3. Run `go build .` or `task build`. This will create a `cursor-rules-syncer` executable.

## Usage

Once installed and `CURSOR_RULES_DIR` is set, you can use the tool from any directory within a Git project.

### Command Line Options

#### Global Flags
*   `--rules-dir <path>` - Specify rules directory path (overrides `CURSOR_RULES_DIR` environment variable)
*   `--ignore-files <file1,file2>` - Comma-separated list of files to ignore during sync
*   `--overwrite-headers` - Overwrite YAML headers instead of preserving them (default: preserve headers)

#### Push-specific Flags
*   `--git-without-push` - Commit changes but don't push to remote repository

### Pull Rules

To pull the latest rules from your central `CURSOR_RULES_DIR` into the current project:

```bash
cursor-rules-syncer pull
```

You can also specify a different rules directory:

```bash
cursor-rules-syncer pull --rules-dir /path/to/other/rules
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. Create a `.cursor/rules` directory in the project root if it doesn't exist.
4. **Recursively** copy all files from `CURSOR_RULES_DIR` to `.cursor/rules`, preserving directory structure and headers of existing `.mdc` files in the project (unless `--overwrite-headers` is used).
5. Delete any extra files in the project that don't exist in the source.

### Push Rules

To push local rule changes from the current project's `.cursor/rules` directory back to your central `CURSOR_RULES_DIR`:

```bash
cursor-rules-syncer push
```

To commit changes without pushing to remote:

```bash
cursor-rules-syncer push --git-without-push
```

To overwrite headers instead of preserving them:

```bash
cursor-rules-syncer push --overwrite-headers
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. **Recursively** copy all files from the project's `.cursor/rules` directory to `CURSOR_RULES_DIR`, preserving directory structure and headers of existing `.mdc` files in the central repository (unless `--overwrite-headers` is used).
4. Delete any extra files in the central repository that don't exist in the project.
5. Change to the `CURSOR_RULES_DIR` Git repository.
6. Execute `git add .`.
7. Execute `git commit -m "Sync cursor rules: updated from project [current_project_name]"`.
8. Execute `git push` (only if `origin` remote exists).

## Features

*   **Smart Synchronization:** Only copies files that have actually changed, reducing unnecessary operations.
*   **Recursive Directory Support:** Processes all files in subdirectories, preserving directory structure.
*   **Header Preservation:** Preserves YAML frontmatter (header block between `---` lines) of existing `.mdc` files by default, with option to overwrite using `--overwrite-headers`. Non-.mdc files are copied as-is.
*   **Advanced File Filtering:** Support for `.ruleignore` file with gitignore-style patterns (wildcards, negation with `!`, directory patterns) and `--ignore-files` flag.
*   **Color-coded Output:** Shows operation status with colored indicators:
    *   🟢 `+` - Added files
    *   🟡 `*` - Updated files  
    *   🔴 `-` - Deleted files
*   **Safe Operations:** Only shows updates when content actually differs.
*   **Auto-cleanup:** Removes extra files in destination that don't exist in source.
*   **Git Integration:** Automatically commits and pushes changes when using `push` command.

## File Filtering

### .ruleignore File

You can create a `.ruleignore` file in your rules source directory to specify files that should be excluded from synchronization. The format follows gitignore-style patterns with support for wildcards, directory patterns, and negation:

```
# Comments start with #
# Empty lines are ignored

# Ignore specific files
secret-rules.mdc
experimental.mdc

# Ignore all files in specific directories
temp/
drafts/

# Ignore files with specific patterns
*.backup.mdc
test-*.mdc

# Use wildcards for subdirectories
**/private/**
**/temp/**/*.mdc

# Negation patterns (include files that would otherwise be ignored)
!important.mdc
!**/public/**/*.mdc
```

### Command Line Filtering

You can also specify files to ignore using the `--ignore-files` flag:

```bash
cursor-rules-syncer pull --ignore-files "secret.mdc,temp.mdc"
cursor-rules-syncer push --ignore-files "experimental.mdc"
```

### Conflict Detection

When using `pull`, if any ignored files exist in the destination project, the operation will fail with an error. This prevents accidental conflicts. You must either:
- Remove the conflicting files from the project
- Update your `.ruleignore` file to exclude them

## Header Preservation

The tool identifies a header as the content between two `---` lines at the very beginning of an `.mdc` file. For example:

```yaml
---
description: My rule description
globs: 
  - "src/**/*.js"
alwaysApply: false
---

This is the actual rule content that will be executed by Cursor...
```

When a file is copied:
*   If the destination file exists and has such a header, this header is kept.
*   The content of the source file (excluding its own header, if `existingHeader` was preserved from destination) is then appended after the preserved header in the destination file.
*   If the destination file does not exist or does not have a valid header, the entire source file (including its header, if any) is copied.

---

## Development

This section is for developers who want to contribute to or modify the project.

### Build Requirements

- [Task](https://taskfile.dev/) - Task runner for development commands
- Go 1.21+
- Git

### Development Commands

#### `task deps`
Installs development dependencies (`goimports`, `golangci-lint`) to `.build/deps/`.
Runs automatically when other commands need dependencies and only re-runs if dependency files are missing.

#### `task fmt` 
Formats code using `goimports` and `gofmt`:
- Fixes imports
- Formats code according to Go standards

#### `task lint`
Runs `golangci-lint` with automatic fixes for common issues.

#### `task build`
Builds the application to `.build/cursor-rules-syncer`.

#### `task test` 
Runs all project tests.

#### `task tidy`
Cleans up and updates `go.mod`.

### Development Workflow

1. Install dependencies: `task deps`
2. Make your changes
3. Format and lint: `task fmt && task lint`
4. Test: `task test`
5. Build: `task build`

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is open-source and available under the [MIT License](LICENSE).

