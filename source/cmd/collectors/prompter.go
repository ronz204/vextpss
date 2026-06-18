package collectors

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter is the low-level input contract for reading raw data from the terminal.
// Collectors use this interface — adapters never interact with it directly.
type Prompter interface {
	ReadLine(label string) (string, error)
	ReadSecret(label string) ([]byte, error)
}

// TerminalPrompter implements Prompter using interactive terminal prompts.
// A single buffered reader is shared to avoid consuming bytes across calls.
type TerminalPrompter struct {
	reader *bufio.Reader
}

func NewTerminalPrompter() *TerminalPrompter {
	return &TerminalPrompter{reader: bufio.NewReader(os.Stdin)}
}

// ================================
// ReadLine reads a visible line from stdin, using label as the prompt text.
// ================================
func (p *TerminalPrompter) ReadLine(label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := p.reader.ReadString('\n')
	return strings.TrimSpace(line), err
}

// ================================
// ReadSecret reads a hidden line from the terminal (input is not echoed).
// ================================
func (p *TerminalPrompter) ReadSecret(label string) ([]byte, error) {
	fmt.Printf("%s: ", label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return b, err
}
