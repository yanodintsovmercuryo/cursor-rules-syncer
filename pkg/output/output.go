package output

import (
	"fmt"
	"io"
	"os"
)

// Output handles all output operations
type Output struct {
	stdout io.Writer
	stderr io.Writer
}

// NewOutput creates a new Output instance
func NewOutput() *Output {
	return &Output{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// NewOutputWithWriters creates Output with custom writers
func NewOutputWithWriters(stdout, stderr io.Writer) *Output {
	return &Output{
		stdout: stdout,
		stderr: stderr,
	}
}

// PrintInfo prints info message
func (o *Output) PrintInfo(message string) {
	fmt.Fprintln(o.stdout, message)
}

// PrintError prints error message
func (o *Output) PrintError(message string) {
	fmt.Fprintln(o.stderr, message)
}

// PrintErrorf prints formatted error message
func (o *Output) PrintErrorf(format string, args ...interface{}) {
	fmt.Fprintf(o.stderr, format+"\n", args...)
}

// PrintOperation prints file operation with color coding
func (o *Output) PrintOperation(operationType, relativePath string) {
	colors := map[string]string{
		"add":    "\033[32m",
		"delete": "\033[31m",
		"update": "\033[33m",
		"reset":  "\033[0m",
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

	fmt.Fprintf(o.stdout, "%s%s %s%s\n", color, symbol, relativePath, reset)
}

// PrintOperationWithTarget prints operation with additional target information
func (o *Output) PrintOperationWithTarget(operationType, relativePath, target string) {
	colors := map[string]string{
		"add":    "\033[32m",
		"delete": "\033[31m",
		"update": "\033[33m",
		"reset":  "\033[0m",
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

	fmt.Fprintf(o.stdout, "%s%s %s (to %s)%s\n", color, symbol, relativePath, target, reset)
}

// PrintSuccess prints success message
func (o *Output) PrintSuccess(message string) {
	o.PrintInfo("\033[32m" + message + "\033[0m")
}

// PrintWarning prints warning message
func (o *Output) PrintWarning(message string) {
	o.PrintError("\033[33m" + message + "\033[0m")
}

// PrintWarningf prints formatted warning message
func (o *Output) PrintWarningf(format string, args ...interface{}) {
	fmt.Fprintf(o.stderr, "\033[33m"+format+"\033[0m\n", args...)
}

// PrintFatal prints fatal error and exits program
func (o *Output) PrintFatal(message string) {
	o.PrintError("\033[31m" + message + "\033[0m")
	os.Exit(1)
}

// PrintFatalf prints formatted fatal error and exits program
func (o *Output) PrintFatalf(format string, args ...interface{}) {
	o.PrintErrorf("\033[31m"+format+"\033[0m", args...)
	os.Exit(1)
}
