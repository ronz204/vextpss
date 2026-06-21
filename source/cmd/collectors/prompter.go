package collectors

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	return strings.TrimSpace(line), err
}

func (p *Prompter) ReadSecret(label string) ([]byte, error) {
	fmt.Fprintf(p.out, "%s: ", label)
	if f, ok := p.in.(*os.File); ok {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.out)
		return b, err
	}
	line, err := p.reader.ReadString('\n')
	return []byte(strings.TrimSpace(line)), err
}
