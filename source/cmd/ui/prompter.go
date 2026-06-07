package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter abstracts all user I/O so handlers can be tested without real stdin/stdout.
type Prompter interface {
	ReadLine(prompt string) (string, error)
	ReadPassword(prompt string) ([]byte, error)
	Confirm(prompt string) (bool, error)
	Zero(b []byte)
}

// CLIPrompter is the production Prompter that reads from the real terminal.
type CLIPrompter struct{}

func (p *CLIPrompter) ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (p *CLIPrompter) ReadPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return password, err
}

func (p *CLIPrompter) Confirm(prompt string) (bool, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func (p *CLIPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

