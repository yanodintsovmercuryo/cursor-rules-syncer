# Cursor Rules Syncer

`cursor-rules-syncer` is a command-line tool to help synchronize [Cursor](https://cursor.sh) IDE's AI rules (`.mdc` files) between a central rules repository and your individual project repositories.

It allows you to maintain a canonical set of rules in one place and easily `pull` them into any project, or `push` local project-specific rule modifications back to the central repository.

## Features

*   **Pull rules:** Copies `.mdc` rule files from a central directory (specified by the `CURSOR_RULES_DIR` environment variable) to your current Git project's `.cursor/rules` directory.
*   **Push rules:** Copies `.mdc` rule files from your current Git project's `.cursor/rules` directory to the central directory and then attempts to `git add`, `git commit`, and `git push` the changes in the central rules repository.
*   **Header Preservation:** When copying files, the tool attempts to preserve the YAML frontmatter (header block between `---` lines) of the destination files if they already exist. This ensures that rule descriptions, globs, and `alwaysApply` settings are not lost.
*   **Safe Push:** The `push` command will only attempt `git push` if a remote named `origin` is configured in the central rules repository. It also handles cases where there are no changes to commit.

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

To build from source locally:
1. Clone this repository.
2. Navigate to the repository directory.
3. Run `go build .`. This will create a `cursor-rules-syncer` executable in the current directory.

## Usage

Once installed and `CURSOR_RULES_DIR` is set, you can use the tool from any directory within a Git project.

### `pull`

To pull the latest rules from your central `CURSOR_RULES_DIR` into the current project:

```bash
cursor-rules-syncer pull
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. Create a `.cursor/rules` directory in the project root if it doesn't exist.
4. Copy all `.mdc` files from `CURSOR_RULES_DIR` to `.cursor/rules`, preserving headers of existing files in the project.

### `push`

To push local rule changes from the current project's `.cursor/rules` directory back to your central `CURSOR_RULES_DIR`:

```bash
cursor-rules-syncer push
```

This will:
1. Read the `CURSOR_RULES_DIR` environment variable.
2. Find the root of the current Git project.
3. Copy all `.mdc` files from the project's `.cursor/rules` directory to `CURSOR_RULES_DIR`, preserving headers of existing files in the central repository.
4. Change to the `CURSOR_RULES_DIR` Git repository.
5. Execute `git add .`.
6. Execute `git commit -m "Sync cursor rules: updated from project [current_project_name]"`.
7. Execute `git push` (only if `origin` remote exists).

## How Header Preservation Works

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

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is open-source and available under the [MIT License](LICENSE). (You would need to add a LICENSE file for this link to work).


---
description: Central Rule 2
globs:
    - "*.ts"
alwaysApply: false
---
console.warn("This is central rule 2");