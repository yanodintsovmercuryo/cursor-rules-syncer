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

### Pull Rules

To pull the latest rules from your central `CURSOR_RULES_DIR` into the current project:

```bash
cursor-rules-syncer pull
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. Create a `.cursor/rules` directory in the project root if it doesn't exist.
4. Copy all `.mdc` files from `CURSOR_RULES_DIR` to `.cursor/rules`, preserving headers of existing files in the project.
5. Delete any extra files in the project that don't exist in the source.

### Push Rules

To push local rule changes from the current project's `.cursor/rules` directory back to your central `CURSOR_RULES_DIR`:

```bash
cursor-rules-syncer push
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. Copy all `.mdc` files from the project's `.cursor/rules` directory to `CURSOR_RULES_DIR`, preserving headers of existing files in the central repository.
4. Delete any extra files in the central repository that don't exist in the project.
5. Change to the `CURSOR_RULES_DIR` Git repository.
6. Execute `git add .`.
7. Execute `git commit -m "Sync cursor rules: updated from project [current_project_name]"`.
8. Execute `git push` (only if `origin` remote exists).

## Features

*   **Smart Synchronization:** Only copies files that have actually changed, reducing unnecessary operations.
*   **Header Preservation:** Preserves YAML frontmatter (header block between `---` lines) of existing files.
*   **Color-coded Output:** Shows operation status with colored indicators:
    *   🟢 `+` - Added files
    *   🟡 `*` - Updated files  
    *   🔴 `-` - Deleted files
*   **Safe Operations:** Only shows updates when content actually differs.
*   **Auto-cleanup:** Removes extra files in destination that don't exist in source.
*   **Git Integration:** Automatically commits and pushes changes when using `push` command.

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

