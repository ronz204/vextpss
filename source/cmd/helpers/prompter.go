package helpers

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

// MockPrompter is a test double that returns pre-set values without touching stdin.
type MockPrompter struct {
	LineResponse    string
	PasswordBytes   []byte
	ConfirmResponse bool
	Err             error
}

func (m *MockPrompter) ReadLine(_ string) (string, error)     { return m.LineResponse, m.Err }
func (m *MockPrompter) ReadPassword(_ string) ([]byte, error) { return m.PasswordBytes, m.Err }
func (m *MockPrompter) Confirm(_ string) (bool, error)        { return m.ConfirmResponse, m.Err }
func (m *MockPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
