package formatters

import (
	"fmt"
	"os"
)

func Success(msg string) {
	fmt.Printf("[✓] %s\n", msg)
}

func Info(msg string) {
	fmt.Printf("[i] %s\n", msg)
}

func Error(msg string) {
	fmt.Fprintf(os.Stderr, "[X] Error: %s\n", msg)
}
