// Package cli is the command line of the bornal binary.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/lucaswilliameufrasio/bornal/internal/localconsole"
	"github.com/lucaswilliameufrasio/bornal/internal/server"
	"github.com/lucaswilliameufrasio/bornal/internal/setup"
)

// Version is the local build. A release build replaces it with the tag.
var Version = "0.1.0"

// search and searchRoots are replaced in tests so the suite never walks a disk.
var (
	search      = install.Discover
	searchRoots = install.DefaultRoots
)

// Run executes one command line, without the program name, and returns the exit code.
// Code 1 is a failed command and code 2 is a command line this binary does not know.
// interactive is true when doctor or console may ask on a terminal.
func Run(args []string, in io.Reader, stdout, stderr io.Writer, getenv func(string) string, interactive bool) int {
	if in == nil {
		in = strings.NewReader("")
	}
	chosen, rest, err := peelLang(args)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.BadLang(i18n.Detect(getenv), err.Error()))
		return 1
	}
	saved, loadErr := config.Load(getenv)
	lang := resolveLang(chosen, saved.Language, getenv)
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
	}
	if loadErr != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.ConfigFailed(lang))
		return 1
	}
	switch rest[0] {
	case "setup":
		return setupCommand(rest[1:], in, stdout, stderr, lang, getenv)
	case "doctor":
		return doctor(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	case "console":
		return console(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	case "mcp":
		return mcpCommand(rest[1:], stderr, lang, chosen != nil, getenv)
	case "macros":
		return macrosCommand(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	case "addons":
		return addonsCommand(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	case "kit":
		return kitCommand(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	case "rollback":
		return rollbackCommand(rest[1:], in, stdout, stderr, lang, chosen != nil, getenv, interactive, saved)
	default:
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, rest[0]), i18n.Usage(lang))
		return 2
	}
}

func resolveLang(flag *string, saved string, getenv func(string) string) i18n.Language {
	if flag != nil {
		if lang, ok := i18n.Parse(*flag); ok {
			return lang
		}
	}
	if lang, ok := i18n.Parse(saved); ok {
		return lang
	}
	return i18n.Detect(getenv)
}

func peelLang(args []string) (*string, []string, error) {
	kept := make([]string, 0, len(args))
	var chosen *string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--lang" {
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("%s", "")
			}
			i++
			value := args[i]
			if _, ok := i18n.Parse(value); !ok {
				return nil, nil, fmt.Errorf("%s", value)
			}
			chosen = &value
			continue
		}
		if strings.HasPrefix(arg, "--lang=") {
			value := strings.TrimPrefix(arg, "--lang=")
			if _, ok := i18n.Parse(value); !ok {
				return nil, nil, fmt.Errorf("%s", value)
			}
			chosen = &value
			continue
		}
		kept = append(kept, arg)
	}
	return chosen, kept, nil
}

func pathFrom(args []string, getenv func(string) string, saved string) (string, error) {
	var path string
	saw := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--path":
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing")
			}
			i++
			path = args[i]
			saw = true
		case strings.HasPrefix(args[i], "--path="):
			path = strings.TrimPrefix(args[i], "--path=")
			saw = true
		default:
			return "", fmt.Errorf("%s", args[i])
		}
	}
	if saw {
		return path, nil
	}
	if getenv != nil {
		if env := strings.TrimSpace(getenv("BORNAL_WOW_PATH")); env != "" {
			return env, nil
		}
	}
	return strings.TrimSpace(saved), nil
}

func doctor(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	lang, path, code := openPath(args, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	found, message, code := report(path, lang)
	if code != 0 {
		fmt.Fprintf(stderr, "bornal: %s\n", message)
		return code
	}
	for _, one := range found {
		locale := one.Locale
		if locale == "" {
			locale = i18n.LabelMissing(lang)
		}
		agent := one.AgentUID
		if agent == "" {
			agent = i18n.LabelMissing(lang)
		}
		fmt.Fprintf(stdout, "%s\n  %s: %s\n  %s: %s\n  %s: %s\n%s\n", one.Path, i18n.LabelFlavor(lang), one.Flavor, i18n.LabelLocale(lang), locale, i18n.LabelAgent(lang), agent, runningLine(lang, one.Path))
	}
	return 0
}

func console(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	lang, path, code := openPath(args, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	actions := consoleActions(path, lang, getenv)
	web, err := localconsole.Start(func() localconsole.View {
		found, message, code := report(path, lang)
		empty := ""
		if code != 0 {
			empty = message
			found = nil
		}
		running := ""
		if len(found) > 0 {
			running = runningLine(lang, found[0].Path)
		}
		return localconsole.View{Language: lang, Installs: found, Empty: empty, Running: running}
	}, actions, func(serveErr error) {
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

func setupCommand(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, getenv func(string) string) int {
	if len(args) > 0 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, args[0]), i18n.Usage(lang))
		return 2
	}
	file, code := askAndSave(in, stdout, stderr, lang, getenv)
	if code != 0 {
		return code
	}
	chosen := lang
	if parsed, ok := i18n.Parse(file.Language); ok {
		chosen = parsed
	}
	fmt.Fprintln(stdout, i18n.SavedConfig(chosen))
	return 0
}

func mcpCommand(args []string, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string) int {
	if len(args) > 0 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, args[0]), i18n.Usage(lang))
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := server.Run(ctx, server.Deps{
		Getenv:         getenv,
		Discover:       search,
		Roots:          searchRoots,
		Inspect:        install.Inspect,
		Language:       lang,
		LanguageLocked: langLocked,
		Open:           gameOpen,
		Volumes:        volumeRoots,
		Passphrase: func() ([]byte, error) {
			return askSecret("passphrase:")
		},
	}, Version)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.ReadFailed(lang))
		return 1
	}
	return 0
}

// openPath resolves the install. On a terminal it asks when nothing is known.
func openPath(args []string, in io.Reader, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) (i18n.Language, string, int) {
	path, err := pathFrom(args, getenv, saved.InstallPath)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.UnknownCommand(lang, err.Error()))
		return lang, "", 2
	}
	if strings.TrimSpace(path) != "" {
		return lang, path, 0
	}
	if !interactive {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.NeedPath(lang))
		return lang, "", 1
	}
	file, code := askAndSave(in, stderr, stderr, lang, getenv)
	if code != 0 {
		return lang, "", code
	}
	if !langLocked {
		if parsed, ok := i18n.Parse(file.Language); ok {
			lang = parsed
		}
	}
	fmt.Fprintln(stderr, i18n.SavedConfig(lang))
	return lang, file.InstallPath, 0
}

func askAndSave(in io.Reader, prompt, stderr io.Writer, lang i18n.Language, getenv func(string) string) (config.File, int) {
	file, err := (setup.Asker{
		In:              in,
		Out:             prompt,
		DefaultLanguage: lang,
		Discover:        search,
		Roots:           searchRoots(getenv),
		Inspect:         install.Inspect,
	}).Run()
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.SetupFailed(lang))
		return config.File{}, 1
	}
	if err := config.Save(getenv, file); err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.ConfigFailed(lang))
		return config.File{}, 1
	}
	return file, 0
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
