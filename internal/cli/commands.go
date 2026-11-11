package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucaswilliameufrasio/bornal/internal/addon"
	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/confirm"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/kit"
	"github.com/lucaswilliameufrasio/bornal/internal/pass"
	"github.com/lucaswilliameufrasio/bornal/internal/running"
	"github.com/lucaswilliameufrasio/bornal/internal/work"
)

var (
	processList = running.List
	volumeRoots = defaultVolumes
	askSecret   = defaultAsk
)

func defaultAsk(prompt string) ([]byte, error) {
	buf, err := pass.ReadTerminal(os.Stdin, prompt+" ", os.Stderr)
	if err == nil {
		return buf, nil
	}
	if !errors.Is(err, pass.ErrNoPrompt) {
		return nil, err
	}
	return pass.Pinentry(prompt)
}

func defaultVolumes() []string {
	body, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		mount := fields[1]
		if mount != "/" && mount != "/home" && !strings.HasPrefix(mount, "/mnt/") && !strings.HasPrefix(mount, "/media/") && !strings.HasPrefix(mount, "/run/media/") {
			continue
		}
		if seen[mount] {
			continue
		}
		seen[mount] = true
		out = append(out, mount)
	}
	return out
}

func gameOpen(path string) (bool, error) {
	if processList == nil {
		return false, os.ErrInvalid
	}
	procs, err := processList()
	if err != nil {
		return false, err
	}
	return running.Matches(path, procs), nil
}

func runningLine(lang i18n.Language, path string) string {
	open, err := gameOpen(path)
	if err != nil {
		return i18n.RunningUnchecked(lang)
	}
	if open {
		return i18n.RunningYes(lang)
	}
	return i18n.RunningNo(lang)
}

func gateFor(side string, getenv func(string) string) (work.Gate, error) {
	dir, err := confirm.Dir(getenv)
	if err != nil {
		return work.Gate{}, err
	}
	return work.Gate{Side: side, StateDir: dir, Open: gameOpen}, nil
}

func commandFail(stderr io.Writer, lang i18n.Language, err error) int {
	fmt.Fprintf(stderr, "bornal: %s\n", explain(lang, err))
	return 1
}

func explain(lang i18n.Language, err error) string {
	switch {
	case errors.Is(err, work.ErrRunning):
		return i18n.RunningYes(lang)
	case errors.Is(err, work.ErrCheck):
		return i18n.RunningUnchecked(lang)
	case errors.Is(err, work.ErrLocale):
		return i18n.MacroLocale(lang)
	case errors.Is(err, work.ErrConflict):
		return i18n.MacroConflict(lang)
	case errors.Is(err, work.ErrChanged), errors.Is(err, work.ErrConfirm):
		return i18n.ReadFailed(lang)
	case errors.Is(err, work.ErrNeedFlavor):
		return i18n.NeedFlavor(lang)
	case errors.Is(err, work.ErrNeedAccount):
		return i18n.NeedAccount(lang)
	case errors.Is(err, work.ErrBackup):
		return i18n.ReadFailed(lang)
	case errors.Is(err, kit.ErrPass), errors.Is(err, pass.ErrMismatch), errors.Is(err, pass.ErrCancel):
		return i18n.KitPass(lang)
	case errors.Is(err, pass.ErrNoPrompt):
		return i18n.UseConsole(lang)
	case errors.Is(err, kit.ErrAccount):
		return i18n.NeedAccount(lang)
	case errors.Is(err, kit.ErrNewer), errors.Is(err, kit.ErrRealm), errors.Is(err, kit.ErrHash), errors.Is(err, kit.ErrFlavor), errors.Is(err, kit.ErrLocale), errors.Is(err, kit.ErrBuild), errors.Is(err, kit.ErrEscape):
		return i18n.ReadFailed(lang)
	default:
		return i18n.ReadFailed(lang)
	}
}

func printChoice(stdout, stderr io.Writer, lang i18n.Language, view work.MacroView, err error) int {
	if errors.Is(err, work.ErrNeedFlavor) {
		for _, name := range view.Flavors {
			fmt.Fprintln(stdout, name)
		}
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.NeedFlavor(lang))
		return 1
	}
	if errors.Is(err, work.ErrNeedAccount) {
		for i, name := range view.Accounts {
			fmt.Fprintf(stdout, "%d  %s\n", i+1, name)
		}
		fmt.Fprintf(stderr, "bornal: %s\n", i18n.NeedAccount(lang))
		return 1
	}
	return commandFail(stderr, lang, err)
}

func macrosCommand(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, "macros"), i18n.Usage(lang))
		return 2
	}
	opt, err := parseOptions(args[1:])
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, err.Error()), i18n.Usage(lang))
		return 2
	}
	root, lang, code := rootFrom(opt, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	g, err := gateFor("cli", getenv)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	switch args[0] {
	case "preview":
		view, err := g.PreviewMacros(root, opt.Flavor, opt.Account, opt.Replace)
		if err != nil {
			return printChoice(stdout, stderr, lang, view, err)
		}
		printItems(stdout, view)
		if !view.LocaleOK {
			fmt.Fprintln(stdout, i18n.MacroLocale(lang))
		}
		if view.Conflict {
			fmt.Fprintln(stdout, i18n.MacroConflict(lang))
		}
		if view.Confirm != "" {
			fmt.Fprintf(stdout, "%s %s\n", i18n.MacroReady(lang), view.Confirm)
		}
		return 0
	case "apply":
		if opt.Confirm == "" {
			fmt.Fprintf(stderr, "bornal: %s\n", i18n.NeedConfirm(lang))
			return 1
		}
		if err := g.ApplyMacros(opt.Confirm); err != nil {
			return commandFail(stderr, lang, err)
		}
		fmt.Fprintln(stdout, i18n.MacroWrote(lang))
		return 0
	default:
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, args[0]), i18n.Usage(lang))
		return 2
	}
}

func addonsCommand(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	opt, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, err.Error()), i18n.Usage(lang))
		return 2
	}
	root, lang, code := rootFrom(opt, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	chosen, _, view, err := work.Resolve(root, opt.Flavor, 1)
	if err != nil && !errors.Is(err, work.ErrNeedAccount) {
		if errors.Is(err, work.ErrNeedFlavor) || errors.Is(err, work.ErrNeedAccount) {
			return printChoice(stdout, stderr, lang, view, err)
		}
		return commandFail(stderr, lang, err)
	}
	if chosen.Path == "" {
		return printChoice(stdout, stderr, lang, view, err)
	}
	found, err := addon.List(chosen.Path)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	if len(found) == 0 {
		fmt.Fprintln(stdout, i18n.NoneFound(lang))
		return 0
	}
	for _, one := range found {
		state := "ok"
		if one.Disabled {
			state = "disabled"
		}
		if !one.InterfaceOK {
			state = "interface"
		}
		fmt.Fprintf(stdout, "%s  %s  %s  %s\n", one.Folder, one.Interface, one.Version, state)
	}
	return 0
}

func kitCommand(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, "kit"), i18n.Usage(lang))
		return 2
	}
	switch args[0] {
	case "create":
		return kitCreate(args[1:], in, stdout, stderr, lang, langLocked, getenv, interactive, saved)
	case "find":
		return kitFind(args[1:], stdout, stderr, lang, getenv)
	case "verify":
		return kitVerify(args[1:], stdout, stderr, lang)
	case "restore":
		return kitRestore(args[1:], in, stdout, stderr, lang, langLocked, getenv, interactive, saved)
	default:
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, args[0]), i18n.Usage(lang))
		return 2
	}
}

func kitCreate(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	opt, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, err.Error()), i18n.Usage(lang))
		return 2
	}
	root, lang, code := rootFrom(opt, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	chosen, account, view, err := work.Resolve(root, opt.Flavor, opt.Account)
	if err != nil {
		return printChoice(stdout, stderr, lang, view, err)
	}
	dest := opt.Out
	if dest == "" {
		dir, err := kitHome(getenv)
		if err != nil {
			return commandFail(stderr, lang, err)
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return commandFail(stderr, lang, err)
		}
		dest = filepath.Join(dir, kit.FileName(kit.FileLabel(opt.Label, chosen.Flavor, time.Now().UTC())))
	}
	var secret []byte
	if opt.Encrypt {
		secret, err = pass.Pair(askSecret, "passphrase:", "passphrase again:")
		if err != nil {
			return commandFail(stderr, lang, err)
		}
		defer pass.Wipe(secret)
	}
	manifest, err := kit.Create(kit.CreateOptions{
		InstallPath: chosen.Path,
		Account:     account,
		Locale:      chosen.Locale,
		Build:       chosen.Build,
		Dest:        dest,
		Label:       opt.Label,
		Encrypt:     opt.Encrypt,
		Passphrase:  secret,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	if dir, dirErr := confirm.Dir(getenv); dirErr == nil {
		_ = work.Remember(dir, dest)
	}
	fmt.Fprintln(stdout, i18n.KitWrote(lang))
	fmt.Fprintln(stdout, dest)
	for _, warning := range manifest.Warnings {
		fmt.Fprintf(stdout, "%s:%d\n", warning.File, warning.Line)
	}
	return 0
}

func kitFind(args []string, stdout, stderr io.Writer, lang i18n.Language, getenv func(string) string) int {
	opt, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, err.Error()), i18n.Usage(lang))
		return 2
	}
	var extra []string
	if opt.SawPath {
		extra = append(extra, opt.Path)
	}
	extra = append(extra, opt.Rest...)
	home := ""
	if getenv != nil {
		home = getenv("HOME")
		if home == "" {
			home = getenv("USERPROFILE")
		}
	}
	last := ""
	if dir, err := confirm.Dir(getenv); err == nil {
		last = work.Last(dir)
	}
	roots := volumeRoots
	if roots == nil {
		roots = func() []string { return nil }
	}
	hits := kit.Look(append(kit.Places(home, roots(), last), extra...))
	for _, hit := range hits {
		flag := "plain"
		if hit.Encrypted {
			flag = "encrypted"
		}
		fmt.Fprintf(stdout, "%s  %s  %s  %d\n", hit.Path, hit.Label, flag, hit.Size)
	}
	return 0
}

func kitVerify(args []string, stdout, stderr io.Writer, lang i18n.Language) int {
	opt, err := parseOptions(args)
	if err != nil || len(opt.Rest) != 1 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, "verify"), i18n.Usage(lang))
		return 2
	}
	secret, encrypted, err := passphraseFor(opt.Rest[0], false)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	defer pass.Wipe(secret)
	report, err := kit.Verify(opt.Rest[0], secret, 0)
	if err != nil {
		if errors.Is(err, kit.ErrPass) && encrypted && len(secret) == 0 {
			fmt.Fprintln(stdout, i18n.KitNeedPass(lang))
			return 1
		}
		return commandFail(stderr, lang, err)
	}
	fmt.Fprintf(stdout, "%s  %s  %s\n", report.Manifest.Flavor, report.Manifest.Locale, report.Label)
	for _, warning := range report.Manifest.Warnings {
		fmt.Fprintf(stdout, "%s:%d\n", warning.File, warning.Line)
	}
	return 0
}

func kitRestore(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	opt, err := parseOptions(args)
	if err != nil || len(opt.Rest) != 1 {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, "restore"), i18n.Usage(lang))
		return 2
	}
	root, lang, code := rootFrom(opt, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	g, err := gateFor("cli", getenv)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	secret, _, err := passphraseFor(opt.Rest[0], true)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	defer pass.Wipe(secret)
	if opt.Confirm == "" {
		view, report, err := g.PreviewRestore(root, opt.Flavor, opt.Rest[0], secret, opt.Replace, opt.RealmOK)
		if err != nil {
			if errors.Is(err, work.ErrNeedFlavor) {
				return printChoice(stdout, stderr, lang, view, err)
			}
			return commandFail(stderr, lang, err)
		}
		fmt.Fprintf(stdout, "%s  %s\n", report.Manifest.Flavor, report.Manifest.Locale)
		for _, warning := range report.Manifest.Warnings {
			fmt.Fprintf(stdout, "%s:%d\n", warning.File, warning.Line)
		}
		fmt.Fprintf(stdout, "%s %s\n", i18n.MacroReady(lang), view.Confirm)
		return 0
	}
	if _, err := g.Restore(opt.Confirm, secret); err != nil {
		return commandFail(stderr, lang, err)
	}
	fmt.Fprintln(stdout, i18n.Restored(lang))
	return 0
}

func rollbackCommand(args []string, in io.Reader, stdout, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) int {
	opt, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "bornal: %s\n\n%s", i18n.UnknownCommand(lang, err.Error()), i18n.Usage(lang))
		return 2
	}
	root, lang, code := rootFrom(opt, in, stderr, lang, langLocked, getenv, interactive, saved)
	if code != 0 {
		return code
	}
	g, err := gateFor("cli", getenv)
	if err != nil {
		return commandFail(stderr, lang, err)
	}
	if opt.Confirm == "" {
		view, err := g.PreviewRollback(root, opt.Flavor, opt.Account)
		if err != nil {
			return printChoice(stdout, stderr, lang, view, err)
		}
		fmt.Fprintf(stdout, "%s %s\n", i18n.MacroReady(lang), view.Confirm)
		return 0
	}
	if err := g.Rollback(opt.Confirm); err != nil {
		return commandFail(stderr, lang, err)
	}
	fmt.Fprintln(stdout, i18n.RolledBack(lang))
	return 0
}

func rootFrom(opt options, in io.Reader, stderr io.Writer, lang i18n.Language, langLocked bool, getenv func(string) string, interactive bool, saved config.File) (string, i18n.Language, int) {
	if opt.SawPath {
		return opt.Path, lang, 0
	}
	next, path, code := openPath(nil, in, stderr, lang, langLocked, getenv, interactive, saved)
	return path, next, code
}

func printItems(stdout io.Writer, view work.MacroView) {
	for _, item := range view.Items {
		fmt.Fprintf(stdout, "%s  %s\n", item.Name, item.Action)
	}
}

func kitHome(getenv func(string) string) (string, error) {
	home := ""
	if getenv != nil {
		home = getenv("HOME")
		if home == "" {
			home = getenv("USERPROFILE")
		}
	}
	if home == "" {
		return "", errors.New("home")
	}
	return filepath.Join(home, ".local", "share", "bornal", "kits"), nil
}

func passphraseFor(path string, ask bool) ([]byte, bool, error) {
	_, _, encrypted, err := kit.Header(path)
	if err != nil {
		return nil, false, err
	}
	info, statErr := os.Lstat(path)
	if statErr == nil && info.IsDir() {
		encrypted = false
	}
	if !encrypted || !ask {
		return nil, encrypted, nil
	}
	secret, err := askSecret("passphrase:")
	if err != nil {
		return nil, true, err
	}
	return secret, true, nil
}
