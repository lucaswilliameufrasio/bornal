// Package cli is the command line of the bornal binary.
package cli

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/lucaswilliameufrasio/bornal/internal/localconsole"
)

// Version is the development slice. It is not a release.
const Version = "0.1.0"

// Run executes one command line, without the program name, and returns the exit code.
// Code 1 is a failed command and code 2 is a command line this binary does not know.
func Run(args []string, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, i18n.Usage(language(nil, getenv)))
		return 0
	}
	lang, rest, err := takeLang(args, getenv)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.BadLang(language(nil, getenv), err.Error()))
		return 1
	}
	if len(rest) == 0 {
		fmt.Fprint(stdout, i18n.Usage(lang))
		return 0
	}
	switch rest[0] {
	case "help", "--help", "-h":
		fmt.Fprint(stdout, i18n.Usage(lang))
		return 0
	case "version", "--version", "-version":
		fmt.Fprintf(stdout, "bornal %s\n", Version)
		return 0
	case "doctor":
		return doctor(rest[1:], stdout, stderr, lang, getenv)
	case "console":
		return console(rest[1:], stdout, stderr, lang, getenv)
	default:
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, rest[0]), i18n.Usage(lang))
		return 2
	}
}

func language(flag *string, getenv func(string) string) i18n.Language {
	if flag != nil {
		if lang, ok := i18n.Parse(*flag); ok {
			return lang
		}
	}
	return i18n.Detect(getenv)
}

func takeLang(args []string, getenv func(string) string) (i18n.Language, []string, error) {
	kept := make([]string, 0, len(args))
	var chosen *string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--lang" {
			if i+1 >= len(args) {
				return 0, nil, fmt.Errorf("%s", "")
			}
			i++
			value := args[i]
			if _, ok := i18n.Parse(value); !ok {
				return 0, nil, fmt.Errorf("%s", value)
			}
			chosen = &value
			continue
		}
		if strings.HasPrefix(arg, "--lang=") {
			value := strings.TrimPrefix(arg, "--lang=")
			if _, ok := i18n.Parse(value); !ok {
				return 0, nil, fmt.Errorf("%s", value)
			}
			chosen = &value
			continue
		}
		kept = append(kept, arg)
	}
	return language(chosen, getenv), kept, nil
}

func pathFrom(args []string, getenv func(string) string) (string, error) {
	var path string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing")
			}
			i++
			path = args[i]
		default:
			if strings.HasPrefix(args[i], "--path=") {
				path = strings.TrimPrefix(args[i], "--path=")
				continue
			}
			return "", fmt.Errorf("%s", args[i])
		}
	}
	if path == "" && getenv != nil {
		path = getenv("BORNAL_WOW_PATH")
	}
	return path, nil
}

func doctor(args []string, stdout, stderr io.Writer, lang i18n.Language, getenv func(string) string) int {
	path, err := pathFrom(args, getenv)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.UnknownCommand(lang, err.Error()))
		return 2
	}
	found, message, code := report(path, lang)
	if code != 0 {
		fmt.Fprintf(stderr, "bornal: %s\n", message)
		return code
	}
	fmt.Fprintln(stdout, message)
	for _, one := range found {
		locale := one.Locale
		if locale == "" {
			locale = i18n.LabelMissing(lang)
		}
		agent := one.AgentUID
		if agent == "" {
			agent = i18n.LabelMissing(lang)
		}
		fmt.Fprintf(stdout, "%s\n  %s: %s\n  %s: %s\n  %s: %s\n", one.Path, i18n.LabelFlavor(lang), one.Flavor, i18n.LabelLocale(lang), locale, i18n.LabelAgent(lang), agent)
	}
	fmt.Fprintln(stdout, i18n.RunningUnchecked(lang))
	return 0
}

func report(path string, lang i18n.Language) ([]install.Install, string, int) {
	if strings.TrimSpace(path) == "" {
		return nil, i18n.NeedPath(lang), 1
	}
	found, err := install.Inspect(path)
	if err != nil {
		return nil, i18n.ReadFailed(lang), 1
	}
	if len(found) == 0 {
		return nil, i18n.NoneFound(lang), 1
	}
	return found, "", 0
}

func console(args []string, stdout, stderr io.Writer, lang i18n.Language, getenv func(string) string) int {
	path, err := pathFrom(args, getenv)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.UnknownCommand(lang, err.Error()))
		return 2
	}
	web, err := localconsole.Start(func() localconsole.View {
		found, message, code := report(path, lang)
		empty := ""
		if code != 0 {
			empty = message
			found = nil
		}
		return localconsole.View{Language: lang, Installs: found, Empty: empty}
	}, func(serveErr error) {
		fmt.Fprintf(stderr, "bornal: %v\n", serveErr)
	})
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.ReadFailed(lang))
		return 1
	}
	defer web.Close()
	fmt.Fprintln(stdout, i18n.ConsoleOpened(lang))
	fmt.Fprintln(stdout, web.URL)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	return 0
}
