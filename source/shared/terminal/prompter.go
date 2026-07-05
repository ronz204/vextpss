package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type Prompter struct {
	in     io.Reader
	out    io.Writer
	reader *bufio.Reader
}

func NewPrompter() *Prompter {
	return NewPrompterFrom(os.Stdin, os.Stdout)
}

func NewPrompterFrom(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{
		in:     in,
		out:    out,
		reader: bufio.NewReader(in),
	}
}

func (p *Prompter) ReadLine(label string) (string, error) {
	fmt.Fprintf(p.out, "%s: ", label)
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (p *Prompter) ReadSecret(label string) ([]byte, error) {
	fmt.Fprintf(p.out, "%s: ", label)
	if f, ok := p.in.(*os.File); ok {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.out)
		return b, err
	}
	line, err := p.reader.ReadString('\n')
	fmt.Fprintln(p.out)
	return []byte(strings.TrimRight(line, "\r\n")), err
}

func (p *Prompter) ReadInteger(label string) (int, error) {
	raw, err := p.ReadLine(label)
	if err != nil {
		return 0, err
	}
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}
	return n, nil
}

func (p *Prompter) ReadLineOrKeep(label, current string) (string, error) {
	fmt.Fprintf(p.out, "%s [current: %s]: ", label, current)
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	v := strings.TrimRight(line, "\r\n")
	if v == "" {
		return current, nil
	}
	return v, nil
}

func (p *Prompter) ReadSecretOrKeep(label string, current []byte) ([]byte, error) {
	fmt.Fprintf(p.out, "%s [Enter to keep]: ", label)
	if f, ok := p.in.(*os.File); ok {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.out)
		if err != nil {
			return nil, err
		}
		if len(b) == 0 {
			return current, nil
		}
		return b, nil
	}
	line, err := p.reader.ReadString('\n')
	fmt.Fprintln(p.out)
	if err != nil {
		return nil, err
	}
	v := []byte(strings.TrimRight(line, "\r\n"))
	if len(v) == 0 {
		return current, nil
	}
	return v, nil
}

func (p *Prompter) ReadIntegerOrKeep(label string, current int) (int, error) {
	fmt.Fprintf(p.out, "%s [current: %d]: ", label, current)
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read input: %w", err)
	}
	raw := strings.TrimRight(line, "\r\n")
	if raw == "" {
		return current, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}
	return n, nil
}

func (p *Prompter) Confirm(prompt string) (bool, error) {
	fmt.Fprintf(p.out, "%s [y/N]: ", prompt)
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes", nil
}
