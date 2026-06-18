package formatters

import (
	"fmt"
	"os"
)

// Success prints a success line to stdout.
func Success(msg string) {
	fmt.Printf("[✓] %s\n", msg)
}

// Info prints a neutral informational line to stdout.
func Info(msg string) {
	fmt.Printf("[i] %s\n", msg)
}

// Error prints a formatted error line to stderr.
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "[X] Error: %s\n", msg)
}
