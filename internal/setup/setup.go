// Package setup asks the first-run questions on a terminal.
package setup

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

const maxAttempts = 5

// ErrStopped means the person did not answer, or the input ended.
var ErrStopped = errors.New("setup stopped")

// Asker reads answers and checks that a path is an install.
type Asker struct {
	In              io.Reader
	Out             io.Writer
	DefaultLanguage i18n.Language
	Discover        func([]string) []install.Install
	Roots           []string
	Inspect         func(string) ([]install.Install, error)
}

// Run asks for the language, then the install path.
func (a Asker) Run() (config.File, error) {
	if a.In == nil || a.Out == nil || a.Inspect == nil {
		return config.File{}, ErrStopped
	}
	reader := bufio.NewReader(a.In)
	lang, err := a.language(reader)
	if err != nil {
		return config.File{}, err
	}
	path, err := a.install(reader, lang)
	if err != nil {
		return config.File{}, err
	}
	return config.File{Language: i18n.Code(lang), InstallPath: path}, nil
}

func (a Asker) language(reader *bufio.Reader) (i18n.Language, error) {
	fallback := a.DefaultLanguage
	if fallback != i18n.Portuguese {
		fallback = i18n.English
	}
	mark := "2"
	if fallback == i18n.Portuguese {
		mark = "1"
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if _, err := fmt.Fprintf(a.Out, "%s[%s]\n> ", i18n.LanguageQuestion(), mark); err != nil {
			return 0, err
		}
		line, err := readLine(reader)
		if err != nil {
			return 0, ErrStopped
		}
		if line == "" {
			return fallback, nil
		}
		switch line {
		case "1":
			return i18n.Portuguese, nil
		case "2":
			return i18n.English, nil
		}
		if lang, ok := i18n.Parse(line); ok {
			return lang, nil
		}
		if _, err := fmt.Fprintln(a.Out, i18n.InvalidChoice(fallback)); err != nil {
			return 0, err
		}
	}
	return 0, ErrStopped
}

func (a Asker) install(reader *bufio.Reader, lang i18n.Language) (string, error) {
	found := a.found()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := a.writeInstallPrompt(found, lang); err != nil {
			return "", err
		}
		line, err := readLine(reader)
		if err != nil {
			return "", ErrStopped
		}
		path, status := a.choose(line, found)
		switch status {
		case choicePath:
			return path, nil
		case choiceType:
			return a.typed(reader, lang)
		case choiceMissing:
			if _, err := fmt.Fprintln(a.Out, i18n.NotAnInstall(lang)); err != nil {
				return "", err
			}
		default:
			if _, err := fmt.Fprintln(a.Out, i18n.InvalidChoice(lang)); err != nil {
				return "", err
			}
		}
	}
	return "", ErrStopped
}

func (a Asker) found() []install.Install {
	if a.Discover == nil {
		return nil
	}
	found := a.Discover(a.Roots)
	sort.Slice(found, func(i, j int) bool {
		return found[i].Path < found[j].Path
	})
	return found
}

func (a Asker) writeInstallPrompt(found []install.Install, lang i18n.Language) error {
	if len(found) == 0 {
		_, err := fmt.Fprintf(a.Out, "%s\n> ", i18n.NoInstalls(lang))
		return err
	}
	if _, err := fmt.Fprintln(a.Out, i18n.InstallQuestion(lang)); err != nil {
		return err
	}
	for i, one := range found {
		if _, err := fmt.Fprintf(a.Out, "  %d  %s\n", i+1, one.Path); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(a.Out, "  %d  %s\n", len(found)+1, i18n.OtherPath(lang)); err != nil {
		return err
	}
	if len(found) == 1 {
		_, err := fmt.Fprintf(a.Out, "[1]\n> ")
		return err
	}
	_, err := fmt.Fprintf(a.Out, "> ")
	return err
}

type choice int

const (
	choiceAgain choice = iota
	choicePath
	choiceType
	choiceMissing
)

// choose maps one answer to an install, a request to type a path, or a miss.
func (a Asker) choose(line string, found []install.Install) (string, choice) {
	if len(found) == 0 {
		if line == "" {
			return "", choiceAgain
		}
		cleaned, ok := a.accept(line)
		if !ok {
			return "", choiceMissing
		}
		return cleaned, choicePath
	}
	if line == "" && len(found) == 1 {
		cleaned, ok := a.accept(found[0].Path)
		if !ok {
			return "", choiceMissing
		}
		return cleaned, choicePath
	}
	if line == "" {
		return "", choiceAgain
	}
	number, err := strconv.Atoi(line)
	if err != nil {
		cleaned, ok := a.accept(line)
		if !ok {
			return "", choiceMissing
		}
		return cleaned, choicePath
	}
	if number >= 1 && number <= len(found) {
		cleaned, ok := a.accept(found[number-1].Path)
		if !ok {
			return "", choiceMissing
		}
		return cleaned, choicePath
	}
	if number == len(found)+1 {
		return "", choiceType
	}
	return "", choiceAgain
}

func (a Asker) typed(reader *bufio.Reader, lang i18n.Language) (string, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if _, err := fmt.Fprintf(a.Out, "%s\n> ", i18n.TypePath(lang)); err != nil {
			return "", err
		}
		line, err := readLine(reader)
		if err != nil {
			return "", ErrStopped
		}
		cleaned, ok := a.accept(line)
		if ok {
			return cleaned, nil
		}
		if _, err := fmt.Fprintln(a.Out, i18n.NotAnInstall(lang)); err != nil {
			return "", err
		}
	}
	return "", ErrStopped
}

func (a Asker) accept(path string) (string, bool) {
	path = strings.Trim(strings.TrimSpace(path), `"'`)
	if path == "" {
		return "", false
	}
	found, err := a.Inspect(path)
	if err != nil || len(found) == 0 {
		return "", false
	}
	return filepath.Clean(path), true
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return line, nil
		}
		return "", err
	}
	return line, nil
}
