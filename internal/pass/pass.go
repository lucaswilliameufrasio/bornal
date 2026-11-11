// Package pass reads a passphrase from a terminal or a local pinentry.
// The value is not taken from a flag, a URL, or a log line.
package pass

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// ErrCancel means the person dismissed the prompt.
var ErrCancel = errors.New("passphrase prompt cancelled")

// ErrNoPrompt means this process has nowhere safe to ask.
var ErrNoPrompt = errors.New("no local passphrase prompt")

// ErrMismatch means the two typed passphrases differed.
var ErrMismatch = errors.New("passphrase mismatch")

// ReadTerminal reads a passphrase without echo. in must be a terminal.
func ReadTerminal(in *os.File, prompt string, out io.Writer) ([]byte, error) {
	if in == nil || !term.IsTerminal(int(in.Fd())) {
		return nil, ErrNoPrompt
	}
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return nil, err
	}
	buf, err := term.ReadPassword(int(in.Fd()))
	if _, writeErr := fmt.Fprintln(out); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		Wipe(buf)
		return nil, err
	}
	if len(bytes.TrimSpace(buf)) == 0 {
		Wipe(buf)
		return nil, ErrCancel
	}
	return buf, nil
}

// Pair reads a passphrase twice and returns it only when both match.
func Pair(read func(string) ([]byte, error), first, second string) ([]byte, error) {
	if read == nil {
		return nil, ErrNoPrompt
	}
	a, err := read(first)
	if err != nil {
		return nil, err
	}
	b, err := read(second)
	if err != nil {
		Wipe(a)
		return nil, err
	}
	if string(a) != string(b) {
		Wipe(a)
		Wipe(b)
		return nil, ErrMismatch
	}
	Wipe(b)
	return a, nil
}

// Pinentry asks pinentry or zenity. The passphrase is not a process argument.
func Pinentry(prompt string) ([]byte, error) {
	for _, name := range []string{"pinentry", "pinentry-gnome3", "pinentry-qt", "pinentry-curses"} {
		bin, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		return runPinentry(bin, prompt)
	}
	bin, err := exec.LookPath("zenity")
	if err != nil {
		return nil, ErrNoPrompt
	}
	return runZenity(bin)
}

// ParsePinentry reads a pinentry transcript and returns only the D line.
func ParsePinentry(text string) ([]byte, error) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "D ") {
			value := []byte(strings.TrimPrefix(line, "D "))
			if len(bytes.TrimSpace(value)) == 0 {
				Wipe(value)
				return nil, ErrCancel
			}
			return value, nil
		}
		if line == "ERR" || strings.HasPrefix(line, "ERR ") {
			return nil, ErrCancel
		}
	}
	return nil, ErrCancel
}

func runPinentry(bin, prompt string) ([]byte, error) {
	cmd := exec.Command(bin)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = strings.NewReader("SETTITLE bornal\nSETDESC bornal\nSETPROMPT " + oneLine(prompt) + "\nGETPIN\nBYE\n")
	_ = cmd.Run()
	value, err := ParsePinentry(out.String())
	out.Reset()
	return value, err
}

func runZenity(bin string) ([]byte, error) {
	cmd := exec.Command(bin, "--password", "--title=bornal")
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrCancel
	}
	out = bytes.TrimRight(out, "\r\n")
	if len(out) == 0 {
		return nil, ErrCancel
	}
	return out, nil
}

func oneLine(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

// Wipe zeroes a passphrase buffer.
func Wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
