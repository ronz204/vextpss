package collectors

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ======================================
// Prompter reads data from an interactive terminal.
// A single buffered reader is shared to avoid consuming bytes across calls.
// ======================================
type Prompter struct {
	reader *bufio.Reader
}

func NewPrompter() *Prompter {
	return &Prompter{reader: bufio.NewReader(os.Stdin)}
}

// ======================================
// ReadLine reads a visible line from stdin, using label as the prompt text.
// ======================================
func (p *Prompter) ReadLine(label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := p.reader.ReadString('\n')
	return strings.TrimSpace(line), err
}

// ======================================
// ReadSecret reads a hidden line from the terminal (input is not echoed).
// ======================================
func (p *Prompter) ReadSecret(label string) ([]byte, error) {
	fmt.Printf("%s: ", label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return b, err
}
