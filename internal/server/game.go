package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lucaswilliameufrasio/bornal/internal/addon"
	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/confirm"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/lucaswilliameufrasio/bornal/internal/kit"
	"github.com/lucaswilliameufrasio/bornal/internal/macro"
	"github.com/lucaswilliameufrasio/bornal/internal/pass"
	"github.com/lucaswilliameufrasio/bornal/internal/running"
	"github.com/lucaswilliameufrasio/bornal/internal/work"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type macroIn struct {
	Flavor  string `json:"flavor,omitempty"`
	Account int    `json:"account,omitempty"`
	Replace bool   `json:"replace,omitempty"`
	Confirm string `json:"confirm,omitempty"`
}

type addonIn struct {
	Flavor string `json:"flavor,omitempty"`
}

type kitCreateIn struct {
	Flavor  string `json:"flavor,omitempty"`
	Account int    `json:"account,omitempty"`
	Label   string `json:"label,omitempty"`
	Out     string `json:"out,omitempty"`
	Encrypt bool   `json:"encrypt,omitempty"`
}

type kitFindIn struct {
	Path string `json:"path,omitempty"`
}

type kitPathIn struct {
	Path    string `json:"path,omitempty"`
	Flavor  string `json:"flavor,omitempty"`
	Confirm string `json:"confirm,omitempty"`
	Replace bool   `json:"replace,omitempty"`
	RealmOK bool   `json:"realm_ok,omitempty"`
}

// Result is a tool payload. It never contains a passphrase.
type Result struct {
	Items     []macro.Item  `json:"items,omitempty"`
	Confirm   string        `json:"confirm,omitempty"`
	LocaleOK  *bool         `json:"locale_ok,omitempty"`
	Conflict  bool          `json:"conflict,omitempty"`
	Flavor    string        `json:"flavor,omitempty"`
	Locale    string        `json:"locale,omitempty"`
	Build     string        `json:"build,omitempty"`
	Pending   []string      `json:"pending,omitempty"`
	Questions []Question    `json:"questions,omitempty"`
	Addons    []addon.Addon `json:"addons,omitempty"`
	Kits      []kit.Hit     `json:"kits,omitempty"`
	Path      string        `json:"path,omitempty"`
	Label     string        `json:"label,omitempty"`
	Encrypted bool          `json:"encrypted,omitempty"`
	Verified  bool          `json:"verified,omitempty"`
	Warnings  []kit.Warning `json:"warnings,omitempty"`
	Message   string        `json:"message,omitempty"`
}

func addGameTools(srv *mcp.Server, deps Deps) {
	local := false
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "macros_preview",
		Description: "Preview the English warrior macros for the saved install. Does not write. Returns a confirm token when a write is possible. Spell names stay in the client locale's language.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in macroIn) (*mcp.CallToolResult, Result, error) {
		return deps.previewMacros(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "macros_apply",
		Description: "Write the previewed macros when confirm is the token from macros_preview on this server. Refuses while that install's Wow.exe is open. Does not accept a passphrase.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in macroIn) (*mcp.CallToolResult, Result, error) {
		return deps.applyMacros(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "addons_list",
		Description: "List addon TOC fields for the saved install. Does not run Lua and does not install addons.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in addonIn) (*mcp.CallToolResult, Result, error) {
		return deps.listAddons(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "kit_create",
		Description: "Create a wow-kit from the saved install. encrypt asks a local prompt twice and has no passphrase argument. SavedVariables are copied as text.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in kitCreateIn) (*mcp.CallToolResult, Result, error) {
		return deps.createKit(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "kit_find",
		Description: "List wow-kit folders and files in the shallow search places. Does not decrypt and does not return account ids.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in kitFindIn) (*mcp.CallToolResult, Result, error) {
		return deps.findKits(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "kit_verify",
		Description: "Check kit hashes. An encrypted kit asks a local prompt and has no passphrase argument. Does not write the game.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in kitPathIn) (*mcp.CallToolResult, Result, error) {
		return deps.verifyKit(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "kit_restore",
		Description: "Preview or restore a kit. confirm must come from a previous kit_restore on this server. Refuses while Wow.exe for that install is open. Passphrase is a local prompt, never an argument.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in kitPathIn) (*mcp.CallToolResult, Result, error) {
		return deps.restoreKit(in)
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "rollback",
		Description: "Restore macros-cache.bornal-bak for the saved install. confirm comes from a previous rollback on this server. Refuses while that Wow.exe is open.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: &local, OpenWorldHint: &local},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in macroIn) (*mcp.CallToolResult, Result, error) {
		return deps.rollback(in)
	})
}

func (d Deps) previewMacros(in macroIn) (*mcp.CallToolResult, Result, error) {
	root, err := d.root()
	if err != nil {
		return nil, Result{}, err
	}
	g, err := d.gate()
	if err != nil {
		return nil, Result{}, err
	}
	view, err := g.PreviewMacros(root, in.Flavor, in.Account, in.Replace)
	if choice, ok := choiceResult(view, err); ok {
		return nil, choice, nil
	}
	if err != nil {
		return nil, Result{}, clean(err)
	}
	ok := view.LocaleOK
	return nil, Result{
		Items:    view.Items,
		Confirm:  view.Confirm,
		LocaleOK: &ok,
		Conflict: view.Conflict,
		Flavor:   view.Flavor,
		Locale:   view.Locale,
		Build:    view.Build,
	}, nil
}

func (d Deps) applyMacros(in macroIn) (*mcp.CallToolResult, Result, error) {
	g, err := d.gate()
	if err != nil {
		return nil, Result{}, err
	}
	if err := g.ApplyMacros(in.Confirm); err != nil {
		return nil, Result{}, clean(err)
	}
	return nil, Result{Message: "macros written"}, nil
}

func (d Deps) listAddons(in addonIn) (*mcp.CallToolResult, Result, error) {
	root, err := d.root()
	if err != nil {
		return nil, Result{}, err
	}
	chosen, view, err := pick(root, in.Flavor)
	if err != nil {
		if errors.Is(err, work.ErrNeedFlavor) {
			return nil, flavorQuestion(view), nil
		}
		return nil, Result{}, clean(err)
	}
	found, err := addon.List(chosen.Path)
	if err != nil {
		return nil, Result{}, errors.New("could not read addons")
	}
	return nil, Result{Addons: found, Flavor: chosen.Flavor}, nil
}

func (d Deps) createKit(in kitCreateIn) (*mcp.CallToolResult, Result, error) {
	root, err := d.root()
	if err != nil {
		return nil, Result{}, err
	}
	chosen, account, view, err := work.Resolve(root, in.Flavor, in.Account)
	if choice, ok := choiceResult(view, err); ok {
		return nil, choice, nil
	}
	if err != nil {
		return nil, Result{}, clean(err)
	}
	dest := in.Out
	if dest == "" {
		dir, err := d.kitDir()
		if err != nil {
			return nil, Result{}, err
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, Result{}, errors.New("could not create the kit directory")
		}
		dest = filepath.Join(dir, kit.FileName(kit.FileLabel(in.Label, chosen.Flavor, time.Now().UTC())))
	}
	var secret []byte
	if in.Encrypt {
		secret, err = d.passphrasePair()
		if err != nil {
			return nil, Result{}, err
		}
		defer pass.Wipe(secret)
	}
	manifest, err := kit.Create(kit.CreateOptions{
		InstallPath: chosen.Path,
		Account:     account,
		Locale:      chosen.Locale,
		Build:       chosen.Build,
		Dest:        dest,
		Label:       in.Label,
		Encrypt:     in.Encrypt,
		Passphrase:  secret,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		return nil, Result{}, clean(err)
	}
	if dir, dirErr := confirm.Dir(d.getenv()); dirErr == nil {
		_ = work.Remember(dir, dest)
	}
	return nil, Result{Path: dest, Warnings: manifest.Warnings, Message: "kit written", Encrypted: in.Encrypt}, nil
}

func (d Deps) findKits(in kitFindIn) (*mcp.CallToolResult, Result, error) {
	home := ""
	if d.Getenv != nil {
		home = d.Getenv("HOME")
	}
	last := ""
	if dir, err := confirm.Dir(d.getenv()); err == nil {
		last = work.Last(dir)
	}
	var volumes []string
	if d.Volumes != nil {
		volumes = d.Volumes()
	}
	places := kit.Places(home, volumes, last)
	if strings.TrimSpace(in.Path) != "" {
		places = append(places, in.Path)
	}
	return nil, Result{Kits: kit.Look(places)}, nil
}

func (d Deps) verifyKit(in kitPathIn) (*mcp.CallToolResult, Result, error) {
	if strings.TrimSpace(in.Path) == "" {
		return nil, Result{}, errors.New("path is empty")
	}
	secret, encrypted, err := d.secretFor(in.Path, true)
	if err != nil {
		return nil, Result{}, err
	}
	defer pass.Wipe(secret)
	report, err := kit.Verify(in.Path, secret, 0)
	if err != nil {
		if errors.Is(err, kit.ErrPass) && encrypted {
			return nil, Result{Encrypted: true, Label: report.Label}, errors.New("passphrase does not match")
		}
		return nil, Result{}, clean(err)
	}
	return nil, Result{Verified: true, Flavor: report.Manifest.Flavor, Locale: report.Manifest.Locale, Label: report.Label, Warnings: report.Manifest.Warnings, Encrypted: encrypted}, nil
}

func (d Deps) restoreKit(in kitPathIn) (*mcp.CallToolResult, Result, error) {
	if strings.TrimSpace(in.Path) == "" && in.Confirm == "" {
		return nil, Result{}, errors.New("path is empty")
	}
	g, err := d.gate()
	if err != nil {
		return nil, Result{}, err
	}
	if in.Confirm == "" {
		secret, _, err := d.secretFor(in.Path, true)
		if err != nil {
			return nil, Result{}, err
		}
		defer pass.Wipe(secret)
		root, err := d.root()
		if err != nil {
			return nil, Result{}, err
		}
		view, report, err := g.PreviewRestore(root, in.Flavor, in.Path, secret, in.Replace, in.RealmOK)
		if errors.Is(err, work.ErrNeedFlavor) {
			return nil, flavorQuestion(view), nil
		}
		if err != nil {
			return nil, Result{}, clean(err)
		}
		return nil, Result{Confirm: view.Confirm, Flavor: report.Manifest.Flavor, Locale: report.Manifest.Locale, Warnings: report.Manifest.Warnings}, nil
	}
	secret, _, err := d.secretFor(in.Path, true)
	if err != nil {
		return nil, Result{}, err
	}
	defer pass.Wipe(secret)
	if _, err := g.Restore(in.Confirm, secret); err != nil {
		return nil, Result{}, clean(err)
	}
	return nil, Result{Message: "kit restored"}, nil
}

func (d Deps) rollback(in macroIn) (*mcp.CallToolResult, Result, error) {
	g, err := d.gate()
	if err != nil {
		return nil, Result{}, err
	}
	if in.Confirm == "" {
		root, err := d.root()
		if err != nil {
			return nil, Result{}, err
		}
		view, err := g.PreviewRollback(root, in.Flavor, in.Account)
		if choice, ok := choiceResult(view, err); ok {
			return nil, choice, nil
		}
		if err != nil {
			return nil, Result{}, clean(err)
		}
		return nil, Result{Confirm: view.Confirm, Flavor: view.Flavor}, nil
	}
	if err := g.Rollback(in.Confirm); err != nil {
		return nil, Result{}, clean(err)
	}
	return nil, Result{Message: "macro backup restored"}, nil
}

func (d Deps) gate() (work.Gate, error) {
	dir, err := confirm.Dir(d.getenv())
	if err != nil {
		return work.Gate{}, errors.New("could not open confirm state")
	}
	openFn := d.Open
	if openFn == nil {
		openFn = running.Open
	}
	return work.Gate{Side: "mcp", StateDir: dir, Open: openFn}, nil
}

func (d Deps) root() (string, error) {
	file, err := config.Load(d.getenv())
	if err != nil {
		return "", errors.New("could not read configuration")
	}
	if strings.TrimSpace(file.InstallPath) == "" {
		return "", errors.New("no saved install path")
	}
	return file.InstallPath, nil
}

func (d Deps) kitDir() (string, error) {
	home := ""
	if d.Getenv != nil {
		home = d.Getenv("HOME")
		if home == "" {
			home = d.Getenv("USERPROFILE")
		}
	}
	if home == "" {
		return "", errors.New("no home directory for the kit")
	}
	return filepath.Join(home, ".local", "share", "bornal", "kits"), nil
}

func (d Deps) passphrasePair() ([]byte, error) {
	if d.Passphrase == nil {
		return nil, errors.New("no local prompt; use bornal console")
	}
	return pass.Pair(func(string) ([]byte, error) { return d.Passphrase() }, "a", "b")
}

func (d Deps) secretFor(path string, ask bool) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, false, errors.New("could not read the kit")
	}
	if info.IsDir() {
		return nil, false, nil
	}
	_, _, encrypted, err := kit.Header(path)
	if err != nil || !encrypted || !ask {
		return nil, encrypted, err
	}
	if d.Passphrase == nil {
		return nil, true, errors.New("no local prompt; use bornal console")
	}
	secret, err := d.Passphrase()
	return secret, true, err
}

func choiceResult(view work.MacroView, err error) (Result, bool) {
	if errors.Is(err, work.ErrNeedFlavor) {
		return flavorQuestion(view), true
	}
	if errors.Is(err, work.ErrNeedAccount) {
		choices := make([]Choice, 0, len(view.Accounts))
		for i, name := range view.Accounts {
			choices = append(choices, Choice{ID: strconv.Itoa(i + 1), Label: name})
		}
		return Result{Pending: []string{"account"}, Questions: []Question{{
			ID:      "account",
			Prompt:  "Which account folder should be used?",
			Choices: choices,
		}}}, true
	}
	return Result{}, false
}

func flavorQuestion(view work.MacroView) Result {
	choices := make([]Choice, 0, len(view.Flavors))
	for _, name := range view.Flavors {
		choices = append(choices, Choice{ID: name, Label: name})
	}
	return Result{Pending: []string{"flavor"}, Questions: []Question{{
		ID:      "flavor",
		Prompt:  "Which World of Warcraft flavor should be used?",
		Choices: choices,
	}}}
}

func pick(root, flavor string) (install.Install, work.MacroView, error) {
	chosen, _, view, err := work.Resolve(root, flavor, 1)
	if errors.Is(err, work.ErrNeedAccount) && chosen.Path != "" {
		return chosen, view, nil
	}
	return chosen, view, err
}

func clean(err error) error {
	switch {
	case errors.Is(err, work.ErrRunning):
		return errors.New("game is running")
	case errors.Is(err, work.ErrCheck):
		return errors.New("could not check whether the game is running")
	case errors.Is(err, work.ErrLocale):
		return errors.New("locale does not accept the english set")
	case errors.Is(err, work.ErrConflict):
		return errors.New("macro conflict")
	case errors.Is(err, work.ErrChanged):
		return errors.New("the file changed after the preview")
	case errors.Is(err, work.ErrConfirm):
		return errors.New("confirmation was refused")
	case errors.Is(err, work.ErrBackup):
		return errors.New("backup is missing")
	case errors.Is(err, kit.ErrPass):
		return errors.New("passphrase does not match")
	case errors.Is(err, kit.ErrAccount):
		return errors.New("account folder is missing")
	case errors.Is(err, kit.ErrNewer):
		return errors.New("destination is newer")
	case errors.Is(err, kit.ErrHash):
		return errors.New("kit hash does not match")
	case errors.Is(err, kit.ErrRealm):
		return errors.New("character realm differs")
	default:
		return errors.New("could not complete the request")
	}
}
