//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package output

import (
	"fmt"
	"io"
	"os"
)

// OutputService handles all output operations
type OutputService struct {
	stdout io.Writer
	stderr io.Writer
}

// NewOutputService creates a new OutputService
func NewOutputService() *OutputService {
	return &OutputService{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// NewOutputServiceWithWriters creates OutputService with custom writers
func NewOutputServiceWithWriters(stdout, stderr io.Writer) *OutputService {
	return &OutputService{
		stdout: stdout,
		stderr: stderr,
	}
}

// PrintInfo prints an informational message
func (s *OutputService) PrintInfo(message string) {
	fmt.Fprintln(s.stdout, message)
}

// PrintError prints an error message
func (s *OutputService) PrintError(message string) {
	fmt.Fprintln(s.stderr, message)
}

// PrintErrorf prints a formatted error message
func (s *OutputService) PrintErrorf(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, format+"\n", args...)
}

// PrintOperation prints a file operation with color coding
func (s *OutputService) PrintOperation(operationType, relativePath string) {
	colors := map[string]string{
		"add":    "\033[32m", // green
		"delete": "\033[31m", // red
		"update": "\033[33m", // yellow
		"reset":  "\033[0m",  // reset
	}

	symbols := map[string]string{
		"add":    "+",
		"delete": "-",
		"update": "*",
	}

	color := colors[operationType]
	if color == "" {
		color = colors["reset"]
	}
	reset := colors["reset"]
	symbol := symbols[operationType]
	if symbol == "" {
		symbol = "?"
	}

	fmt.Fprintf(s.stdout, "%s%s %s%s\n", color, symbol, relativePath, reset)
}

// PrintOperationWithTarget prints operation with additional target info
func (s *OutputService) PrintOperationWithTarget(operationType, relativePath, target string) {
	colors := map[string]string{
		"add":    "\033[32m", // green
		"delete": "\033[31m", // red
		"update": "\033[33m", // yellow
		"reset":  "\033[0m",  // reset
	}

	symbols := map[string]string{
		"add":    "+",
		"delete": "-",
		"update": "*",
	}

	color := colors[operationType]
	if color == "" {
		color = colors["reset"]
	}
	reset := colors["reset"]
	symbol := symbols[operationType]
	if symbol == "" {
		symbol = "?"
	}

	fmt.Fprintf(s.stdout, "%s%s %s (to %s)%s\n", color, symbol, relativePath, target, reset)
}

// PrintSuccess prints a success message
func (s *OutputService) PrintSuccess(message string) {
	s.PrintInfo("\033[32m" + message + "\033[0m")
}

// PrintWarning prints a warning message
func (s *OutputService) PrintWarning(message string) {
	s.PrintError("\033[33m" + message + "\033[0m")
}

// PrintWarningf prints a formatted warning message
func (s *OutputService) PrintWarningf(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, "\033[33m"+format+"\033[0m\n", args...)
}

// PrintFatal prints a fatal error and exits
func (s *OutputService) PrintFatal(message string) {
	s.PrintError("\033[31m" + message + "\033[0m")
	os.Exit(1)
}

// PrintFatalf prints a formatted fatal error and exits
func (s *OutputService) PrintFatalf(format string, args ...interface{}) {
	s.PrintErrorf("\033[31m"+format+"\033[0m", args...)
	os.Exit(1)
}
