// Package work is the shared preview and write path for the CLI, MCP, and console.
// A token from one side is refused by the others.
package work

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/confirm"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/lucaswilliameufrasio/bornal/internal/kit"
	"github.com/lucaswilliameufrasio/bornal/internal/macro"
	"github.com/lucaswilliameufrasio/bornal/internal/running"
)

var (
	ErrNeedFlavor  = errors.New("choose a flavor")
	ErrNeedAccount = errors.New("choose an account")
	ErrRunning     = errors.New("game is running")
	ErrCheck       = errors.New("could not check whether the game is running")
	ErrLocale      = errors.New("locale does not accept the english set")
	ErrConflict    = errors.New("macro conflict")
	ErrChanged     = errors.New("the file changed after the preview")
	ErrConfirm     = errors.New("confirmation was refused")
	ErrBackup      = errors.New("backup is missing")
)

// Gate is one caller. Side is cli, mcp, or console.
type Gate struct {
	Side     string
	StateDir string
	Open     func(string) (bool, error)
}

// MacroView is a macro preview. Confirm is empty until a write is possible.
type MacroView struct {
	Items    []macro.Item `json:"items"`
	LocaleOK bool         `json:"locale_ok"`
	Conflict bool         `json:"conflict"`
	Confirm  string       `json:"confirm,omitempty"`
	Flavor   string       `json:"flavor,omitempty"`
	Locale   string       `json:"locale,omitempty"`
	Build    string       `json:"build,omitempty"`
	Flavors  []string     `json:"flavors,omitempty"`
	Accounts []string     `json:"accounts,omitempty"`
	SHA256   string       `json:"sha256,omitempty"`
}

type ticket struct {
	Kind        string `json:"kind"`
	Install     string `json:"install"`
	Account     string `json:"account"`
	Replace     bool   `json:"replace"`
	SHA256      string `json:"sha256"`
	BackupSHA   string `json:"backup_sha256,omitempty"`
	Seal        string `json:"seal,omitempty"`
	Locale      string `json:"locale"`
	Build       string `json:"build"`
	Kit         string `json:"kit,omitempty"`
	AcceptRealm bool   `json:"accept_realm,omitempty"`
	KitReplace  bool   `json:"kit_replace,omitempty"`
}

// Accounts lists account folders. The names are not stored in app config.
func Accounts(flavor string) ([]string, error) {
	dir := filepath.Join(flavor, "WTF", "Account")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, statErr := os.Lstat(filepath.Join(dir, entry.Name()))
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// Resolve picks one flavor and one account. Several matches return the choices and an error.
func Resolve(root, flavor string, account int) (install.Install, string, MacroView, error) {
	found, err := install.Inspect(root)
	if err != nil {
		return install.Install{}, "", MacroView{}, err
	}
	names := make([]string, 0, len(found))
	for _, one := range found {
		names = append(names, one.Flavor)
	}
	var chosen install.Install
	switch {
	case len(found) == 0:
		return install.Install{}, "", MacroView{}, errors.New("no installation")
	case len(found) == 1 && flavor == "":
		chosen = found[0]
	default:
		for _, one := range found {
			if one.Flavor == flavor || filepath.Clean(one.Path) == filepath.Clean(flavor) {
				chosen = one
				break
			}
		}
		if chosen.Path == "" {
			return install.Install{}, "", MacroView{Flavors: names}, ErrNeedFlavor
		}
	}
	accounts, err := Accounts(chosen.Path)
	if err != nil {
		return install.Install{}, "", MacroView{}, err
	}
	view := MacroView{Flavor: chosen.Flavor, Locale: chosen.Locale, Build: chosen.Build, Accounts: accounts}
	if len(accounts) == 0 {
		return chosen, "", view, ErrNeedAccount
	}
	if account == 0 && len(accounts) == 1 {
		account = 1
	}
	if account < 1 || account > len(accounts) {
		return chosen, "", view, ErrNeedAccount
	}
	return chosen, accounts[account-1], view, nil
}

// PreviewMacros plans the English set. It does not write the cache.
func (g Gate) PreviewMacros(root, flavor string, account int, replace bool) (MacroView, error) {
	chosen, name, view, err := Resolve(root, flavor, account)
	if err != nil {
		return view, err
	}
	current, sum, err := macro.Read(cachePath(chosen.Path, name))
	if err != nil {
		return view, err
	}
	plan, _ := macro.PlanSet(current, chosen.Locale, replace)
	view.Items = plan.Items
	view.LocaleOK = plan.LocaleOK
	view.Conflict = plan.Conflict
	view.SHA256 = sum
	view.Flavor = chosen.Flavor
	view.Locale = chosen.Locale
	view.Build = chosen.Build
	view.Accounts = nil
	view.Flavors = nil
	if !plan.LocaleOK || plan.Conflict || !needsWrite(plan.Items) {
		return view, nil
	}
	token, err := g.issue(ticket{
		Kind:    "macros",
		Install: chosen.Path,
		Account: name,
		Replace: replace,
		SHA256:  sum,
		Locale:  chosen.Locale,
		Build:   chosen.Build,
	})
	if err != nil {
		return view, err
	}
	view.Confirm = token
	return view, nil
}

// ApplyMacros writes the planned set when the token, the file, and the closed game still agree.
func (g Gate) ApplyMacros(token string) error {
	body, err := g.take(token)
	if err != nil {
		return ErrConfirm
	}
	var plan ticket
	if err := json.Unmarshal(body, &plan); err != nil || plan.Kind != "macros" {
		return ErrConfirm
	}
	if err := g.sameFile(plan); err != nil {
		return err
	}
	current, _, err := macro.Read(cachePath(plan.Install, plan.Account))
	if err != nil {
		return err
	}
	nextPlan, next := macro.PlanSet(current, plan.Locale, plan.Replace)
	if !nextPlan.LocaleOK {
		return ErrLocale
	}
	if nextPlan.Conflict {
		return ErrConflict
	}
	if !needsWrite(nextPlan.Items) {
		return nil
	}
	path := cachePath(plan.Install, plan.Account)
	if err := macro.Backup(path); err != nil {
		return err
	}
	return macro.WriteFile(path, next)
}

// PreviewRollback issues a token for the macro backup beside the cache.
func (g Gate) PreviewRollback(root, flavor string, account int) (MacroView, error) {
	chosen, name, view, err := Resolve(root, flavor, account)
	if err != nil {
		return view, err
	}
	path := cachePath(chosen.Path, name)
	_, sum, err := macro.Read(path)
	if err != nil {
		return view, err
	}
	backup, err := os.ReadFile(path + ".bornal-bak")
	if err != nil {
		if os.IsNotExist(err) {
			return view, ErrBackup
		}
		return view, err
	}
	token, err := g.issue(ticket{
		Kind:      "rollback",
		Install:   chosen.Path,
		Account:   name,
		SHA256:    sum,
		BackupSHA: hashBytes(backup),
		Locale:    chosen.Locale,
		Build:     chosen.Build,
	})
	if err != nil {
		return view, err
	}
	view.Confirm = token
	view.Accounts = nil
	view.Flavors = nil
	return view, nil
}

// Rollback restores macros-cache.bornal-bak over the cache.
func (g Gate) Rollback(token string) error {
	body, err := g.take(token)
	if err != nil {
		return ErrConfirm
	}
	var plan ticket
	if err := json.Unmarshal(body, &plan); err != nil || plan.Kind != "rollback" {
		return ErrConfirm
	}
	if err := g.sameFile(plan); err != nil {
		return err
	}
	path := cachePath(plan.Install, plan.Account)
	backup, err := os.ReadFile(path + ".bornal-bak")
	if err != nil {
		return ErrBackup
	}
	if hashBytes(backup) != plan.BackupSHA {
		return ErrChanged
	}
	tmp := path + ".bornal-tmp"
	if err := os.WriteFile(tmp, backup, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// PreviewRestore checks the kit and issues a token. It does not write the game.
func (g Gate) PreviewRestore(root, flavor, kitPath string, passphrase []byte, replace, acceptRealm bool) (MacroView, kit.Report, error) {
	chosen, view, err := pickFlavor(root, flavor)
	if err != nil {
		return view, kit.Report{}, err
	}
	report, err := kit.Verify(kitPath, passphrase, 0)
	if err != nil {
		return view, report, err
	}
	if report.Manifest.AccountID == "" {
		return view, report, kit.ErrAccount
	}
	token, err := g.issue(ticket{
		Kind:        "restore",
		Install:     chosen.Path,
		Account:     report.Manifest.AccountID,
		SHA256:      hashBytes(mustManifest(kitPath, report)),
		Seal:        kit.Seal(report.Manifest),
		Locale:      chosen.Locale,
		Build:       chosen.Build,
		Kit:         kitPath,
		AcceptRealm: acceptRealm,
		KitReplace:  replace,
	})
	if err != nil {
		return view, report, err
	}
	view.Confirm = token
	view.Locale = chosen.Locale
	view.Build = chosen.Build
	view.Flavor = chosen.Flavor
	return view, report, nil
}

// Restore writes a kit only when the token still matches and the game is closed.
func (g Gate) Restore(token string, passphrase []byte) (kit.RestoreReport, error) {
	body, err := g.take(token)
	if err != nil {
		return kit.RestoreReport{}, ErrConfirm
	}
	var plan ticket
	if err := json.Unmarshal(body, &plan); err != nil || plan.Kind != "restore" {
		return kit.RestoreReport{}, ErrConfirm
	}
	if err := g.closed(plan.Install); err != nil {
		return kit.RestoreReport{}, err
	}
	live, err := reread(plan.Install)
	if err != nil {
		return kit.RestoreReport{}, err
	}
	if live.Locale != plan.Locale || live.Build != plan.Build {
		return kit.RestoreReport{}, ErrChanged
	}
	report, err := kit.Verify(plan.Kit, passphrase, 0)
	if err != nil {
		return kit.RestoreReport{}, err
	}
	if kit.Seal(report.Manifest) != plan.Seal || report.Manifest.AccountID != plan.Account {
		return kit.RestoreReport{}, ErrChanged
	}
	return kit.Restore(kit.RestoreOptions{
		KitPath:     plan.Kit,
		InstallPath: plan.Install,
		Locale:      live.Locale,
		Build:       live.Build,
		Passphrase:  passphrase,
		Replace:     plan.KitReplace,
		AcceptRealm: plan.AcceptRealm,
	})
}

func (g Gate) sameFile(plan ticket) error {
	if err := g.closed(plan.Install); err != nil {
		return err
	}
	live, err := reread(plan.Install)
	if err != nil {
		return err
	}
	if live.Locale != plan.Locale || live.Build != plan.Build {
		return ErrChanged
	}
	_, sum, err := macro.Read(cachePath(plan.Install, plan.Account))
	if err != nil {
		return err
	}
	if sum != plan.SHA256 {
		return ErrChanged
	}
	return nil
}

func (g Gate) closed(installPath string) error {
	openFn := g.Open
	if openFn == nil {
		openFn = running.Open
	}
	open, err := openFn(installPath)
	if err != nil {
		return ErrCheck
	}
	if open {
		return ErrRunning
	}
	return nil
}

func (g Gate) issue(plan ticket) (string, error) {
	if g.StateDir == "" || g.Side == "" {
		return "", errors.New("confirm state is missing")
	}
	body, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	return confirm.Issue(g.StateDir, g.Side, body)
}

func (g Gate) take(token string) ([]byte, error) {
	if g.StateDir == "" || g.Side == "" {
		return nil, ErrConfirm
	}
	return confirm.Take(g.StateDir, g.Side, token)
}

func reread(flavor string) (install.Install, error) {
	found, err := install.Inspect(flavor)
	if err != nil {
		return install.Install{}, err
	}
	for _, one := range found {
		if filepath.Clean(one.Path) == filepath.Clean(flavor) {
			return one, nil
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	return install.Install{}, ErrChanged
}

func cachePath(flavor, account string) string {
	return filepath.Join(flavor, "WTF", "Account", account, "macros-cache.txt")
}

func needsWrite(items []macro.Item) bool {
	for _, item := range items {
		if item.Action == macro.Create || item.Action == macro.Replace {
			return true
		}
	}
	return false
}

func hashBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func pickFlavor(root, flavor string) (install.Install, MacroView, error) {
	found, err := install.Inspect(root)
	if err != nil {
		return install.Install{}, MacroView{}, err
	}
	names := make([]string, 0, len(found))
	for _, one := range found {
		names = append(names, one.Flavor)
	}
	if len(found) == 0 {
		return install.Install{}, MacroView{}, errors.New("no installation")
	}
	if len(found) == 1 && flavor == "" {
		return found[0], MacroView{Flavor: found[0].Flavor, Locale: found[0].Locale, Build: found[0].Build}, nil
	}
	for _, one := range found {
		if one.Flavor == flavor || filepath.Clean(one.Path) == filepath.Clean(flavor) {
			return one, MacroView{Flavor: one.Flavor, Locale: one.Locale, Build: one.Build}, nil
		}
	}
	return install.Install{}, MacroView{Flavors: names}, ErrNeedFlavor
}

func mustManifest(path string, report kit.Report) []byte {
	if report.Encrypted {
		body, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		return body
	}
	body, err := os.ReadFile(filepath.Join(path, "manifest.json"))
	if err != nil {
		return nil
	}
	return body
}

// Remember stores the last kit path. It does not store an account id.
func Remember(dir, path string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "last-export"), []byte(strings.TrimSpace(path)+"\n"), 0o600)
}

// Last reads the last kit path.
func Last(dir string) string {
	body, err := os.ReadFile(filepath.Join(dir, "last-export"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}
