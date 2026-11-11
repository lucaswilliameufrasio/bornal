package cli

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lucaswilliameufrasio/bornal/internal/addon"
	"github.com/lucaswilliameufrasio/bornal/internal/confirm"
	panel "github.com/lucaswilliameufrasio/bornal/internal/console"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/kit"
	"github.com/lucaswilliameufrasio/bornal/internal/localconsole"
	"github.com/lucaswilliameufrasio/bornal/internal/pass"
	"github.com/lucaswilliameufrasio/bornal/internal/work"
)

// consoleActions serves macros, addons, kits, and rollback on the loopback console.
// Tokens issued here are refused by the CLI and the MCP server.
func consoleActions(root string, lang i18n.Language, getenv func(string) string) *localconsole.Actions {
	actions := &localconsole.Actions{}
	actions.Get = func(page, flavor string, account int, replace bool) (panel.Page, error) {
		built, err := renderConsole(root, lang, getenv, page, flavor, account, replace)
		return stamp(actions, built), err
	}
	actions.Post = func(page string, form localconsole.Form) (panel.Page, error) {
		built, err := submitConsole(root, lang, getenv, page, form)
		return stamp(actions, built), err
	}
	return actions
}

func stamp(actions *localconsole.Actions, page panel.Page) panel.Page {
	if actions == nil {
		return page
	}
	for i := range page.Forms {
		page.Forms[i].CSRF = actions.CSRF
	}
	return page
}

func renderConsole(root string, lang i18n.Language, getenv func(string) string, page, flavor string, account int, replace bool) (panel.Page, error) {
	switch page {
	case "macros":
		return macrosPage(root, lang, getenv, flavor, account, replace)
	case "addons":
		return addonsPage(root, lang, flavor)
	case "kits":
		return kitsPage(lang, getenv, flavor, account)
	case "rollback":
		return rollbackPage(root, lang, getenv, flavor, account)
	default:
		return panel.Page{}, errors.New("unknown page")
	}
}

func submitConsole(root string, lang i18n.Language, getenv func(string) string, page string, form localconsole.Form) (panel.Page, error) {
	switch page {
	case "macros":
		return submitMacros(root, lang, getenv, form)
	case "addons":
		return addonsPage(root, lang, form.Flavor)
	case "kits":
		return submitKits(root, lang, getenv, form)
	case "rollback":
		return submitRollback(root, lang, getenv, form)
	default:
		return panel.Page{}, errors.New("unknown page")
	}
}

func macrosPage(root string, lang i18n.Language, getenv func(string) string, flavor string, account int, replace bool) (panel.Page, error) {
	title := i18n.NavMacros(lang)
	g, err := gateFor("console", getenv)
	if err != nil {
		return noted(lang, "/web/macros", title, explain(lang, err)), nil
	}
	view, err := g.PreviewMacros(root, flavor, account, replace)
	if page, ok := choicePage(lang, "/web/macros", title, view, err); ok {
		return page, nil
	}
	if err != nil {
		return noted(lang, "/web/macros", title, explain(lang, err)), nil
	}
	page := panel.Section(lang, title, "/web/macros", i18n.LocaleNote(lang))
	var notes []string
	if !view.LocaleOK {
		notes = append(notes, i18n.MacroLocale(lang))
	}
	if view.Conflict {
		notes = append(notes, i18n.MacroConflict(lang))
	}
	page.Message = strings.Join(notes, " ")
	for _, item := range view.Items {
		page.Rows = append(page.Rows, panel.Row{Name: item.Name, Detail: string(item.Action)})
	}
	if view.Conflict && view.LocaleOK {
		q := url.Values{}
		if view.Flavor != "" {
			q.Set("flavor", view.Flavor)
		}
		n := account
		if n == 0 {
			n = 1
		}
		q.Set("account", strconv.Itoa(n))
		q.Set("replace", "1")
		page.Rows = append(page.Rows, panel.Row{Name: "replace", Href: hrefQuery("/web/macros", q)})
	}
	if view.Confirm != "" {
		n := account
		if n == 0 {
			n = 1
		}
		page.Forms = []panel.Form{{
			Kind:    "apply",
			Action:  "/web/macros",
			Confirm: view.Confirm,
			Flavor:  view.Flavor,
			Account: n,
			Submit:  i18n.SubmitApply(lang),
		}}
	}
	return page, nil
}

func submitMacros(root string, lang i18n.Language, getenv func(string) string, form localconsole.Form) (panel.Page, error) {
	title := i18n.NavMacros(lang)
	if form.Confirm == "" {
		return macrosPage(root, lang, getenv, form.Flavor, form.Account, form.Replace)
	}
	g, err := gateFor("console", getenv)
	if err != nil {
		return noted(lang, "/web/macros", title, explain(lang, err)), nil
	}
	if err := g.ApplyMacros(form.Confirm); err != nil {
		return noted(lang, "/web/macros", title, explain(lang, err)), nil
	}
	page, err := macrosPage(root, lang, getenv, form.Flavor, form.Account, false)
	if err != nil {
		return page, err
	}
	page.Message = i18n.MacroWrote(lang)
	return page, nil
}

func addonsPage(root string, lang i18n.Language, flavor string) (panel.Page, error) {
	title := i18n.NavAddons(lang)
	chosen, _, view, err := work.Resolve(root, flavor, 1)
	if page, ok := choicePage(lang, "/web/addons", title, view, err); ok {
		return page, nil
	}
	if err != nil && !errors.Is(err, work.ErrNeedAccount) {
		return noted(lang, "/web/addons", title, explain(lang, err)), nil
	}
	if chosen.Path == "" {
		return noted(lang, "/web/addons", title, explain(lang, err)), nil
	}
	found, err := addon.List(chosen.Path)
	if err != nil {
		return noted(lang, "/web/addons", title, i18n.ReadFailed(lang)), nil
	}
	page := panel.Section(lang, title, "/web/addons", i18n.LocaleNote(lang))
	if len(found) == 0 {
		page.Message = i18n.NoneFound(lang)
		return page, nil
	}
	for _, one := range found {
		state := "ok"
		if one.Disabled {
			state = "disabled"
		}
		if !one.InterfaceOK {
			state = "interface"
		}
		detail := strings.TrimSpace(one.Interface + " " + one.Version + " " + state)
		if len(one.Missing) > 0 {
			detail += " missing"
		}
		page.Rows = append(page.Rows, panel.Row{Name: one.Folder, Detail: detail})
	}
	return page, nil
}

func kitsPage(lang i18n.Language, getenv func(string) string, flavor string, account int) (panel.Page, error) {
	title := i18n.NavKits(lang)
	page := panel.Section(lang, title, "/web/kits", i18n.LocaleNote(lang))
	home := homeOf(getenv)
	last := ""
	if dir, err := confirm.Dir(getenv); err == nil {
		last = work.Last(dir)
	}
	var volumes []string
	if volumeRoots != nil {
		volumes = volumeRoots()
	}
	for _, hit := range kit.Look(kit.Places(home, volumes, last)) {
		name := hit.Label
		if name == "" {
			name = filepath.Base(hit.Path)
		}
		flag := "plain"
		if hit.Encrypted {
			flag = "encrypted"
		}
		page.Rows = append(page.Rows, panel.Row{Name: name, Detail: hit.Path + "  " + flag})
	}
	page.Forms = []panel.Form{
		{
			Kind:     "create",
			Action:   "/web/kits",
			Flavor:   flavor,
			Account:  account,
			Submit:   i18n.SubmitKit(lang),
			Password: true,
			Twice:    true,
		},
		{
			Kind:     "verify",
			Action:   "/web/kits",
			Submit:   i18n.SubmitVerify(lang),
			Password: true,
		},
		{
			Kind:     "restore",
			Action:   "/web/kits",
			Flavor:   flavor,
			Submit:   i18n.SubmitRestore(lang),
			Password: true,
		},
	}
	return page, nil
}

func submitKits(root string, lang i18n.Language, getenv func(string) string, form localconsole.Form) (panel.Page, error) {
	switch form.Kind {
	case "create":
		return createFrom(root, lang, getenv, form)
	case "verify":
		return verifyFrom(lang, form)
	case "restore":
		return restoreFrom(root, lang, getenv, form)
	default:
		return kitsPage(lang, getenv, form.Flavor, form.Account)
	}
}

func createFrom(root string, lang i18n.Language, getenv func(string) string, form localconsole.Form) (panel.Page, error) {
	title := i18n.NavKits(lang)
	chosen, account, view, err := work.Resolve(root, form.Flavor, form.Account)
	if page, ok := choicePage(lang, "/web/kits", title, view, err); ok {
		return page, nil
	}
	if err != nil {
		return noted(lang, "/web/kits", title, explain(lang, err)), nil
	}
	var secret []byte
	if form.Encrypt {
		if !secretsMatch(form.Passphrase, form.Again) {
			return noted(lang, "/web/kits", title, i18n.KitPass(lang)), nil
		}
		secret = append([]byte(nil), form.Passphrase...)
		defer pass.Wipe(secret)
	}
	dest := strings.TrimSpace(form.Out)
	if dest == "" {
		dir, dirErr := kitHome(getenv)
		if dirErr != nil {
			return noted(lang, "/web/kits", title, i18n.ReadFailed(lang)), nil
		}
		if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
			return noted(lang, "/web/kits", title, i18n.ReadFailed(lang)), nil
		}
		dest = filepath.Join(dir, kit.FileName(kit.FileLabel(form.Label, chosen.Flavor, time.Now().UTC())))
	}
	manifest, err := kit.Create(kit.CreateOptions{
		InstallPath: chosen.Path,
		Account:     account,
		Locale:      chosen.Locale,
		Build:       chosen.Build,
		Dest:        dest,
		Label:       form.Label,
		Encrypt:     form.Encrypt,
		Passphrase:  secret,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		return noted(lang, "/web/kits", title, explain(lang, err)), nil
	}
	if dir, dirErr := confirm.Dir(getenv); dirErr == nil {
		_ = work.Remember(dir, dest)
	}
	page := noted(lang, "/web/kits", title, i18n.KitWrote(lang))
	page.Rows = []panel.Row{{Name: filepath.Base(dest), Detail: dest}}
	for _, warning := range manifest.Warnings {
		page.Rows = append(page.Rows, panel.Row{Name: warning.File, Detail: strconv.Itoa(warning.Line)})
	}
	return page, nil
}

func verifyFrom(lang i18n.Language, form localconsole.Form) (panel.Page, error) {
	title := i18n.NavKits(lang)
	path := strings.TrimSpace(form.Kit)
	if path == "" {
		return noted(lang, "/web/kits", title, i18n.ReadFailed(lang)), nil
	}
	secret, encrypted, err := secretFrom(path, form.Passphrase)
	if len(secret) > 0 {
		defer pass.Wipe(secret)
	}
	if err != nil {
		return noted(lang, "/web/kits", title, passMessage(lang, err, encrypted, form.Passphrase)), nil
	}
	report, err := kit.Verify(path, secret, 0)
	if err != nil {
		return noted(lang, "/web/kits", title, passMessage(lang, err, encrypted, form.Passphrase)), nil
	}
	page := noted(lang, "/web/kits", title, report.Label)
	page.Rows = []panel.Row{{Name: report.Manifest.Flavor, Detail: report.Manifest.Locale}}
	for _, warning := range report.Manifest.Warnings {
		page.Rows = append(page.Rows, panel.Row{Name: warning.File, Detail: strconv.Itoa(warning.Line)})
	}
	return page, nil
}

func restoreFrom(root string, lang i18n.Language, getenv func(string) string, form localconsole.Form) (panel.Page, error) {
	title := i18n.NavKits(lang)
	path := strings.TrimSpace(form.Kit)
	if path == "" {
		return noted(lang, "/web/kits", title, i18n.ReadFailed(lang)), nil
	}
	secret, encrypted, err := secretFrom(path, form.Passphrase)
	if len(secret) > 0 {
		defer pass.Wipe(secret)
	}
	if err != nil {
		return noted(lang, "/web/kits", title, passMessage(lang, err, encrypted, form.Passphrase)), nil
	}
	g, err := gateFor("console", getenv)
	if err != nil {
		return noted(lang, "/web/kits", title, explain(lang, err)), nil
	}
	if form.Confirm == "" {
		view, report, err := g.PreviewRestore(root, form.Flavor, path, secret, form.Replace, form.RealmOK)
		if page, ok := choicePage(lang, "/web/kits", title, view, err); ok {
			return page, nil
		}
		if err != nil {
			return noted(lang, "/web/kits", title, passMessage(lang, err, encrypted, form.Passphrase)), nil
		}
		page := panel.Section(lang, title, "/web/kits", i18n.LocaleNote(lang))
		page.Rows = []panel.Row{{Name: report.Manifest.Flavor, Detail: report.Manifest.Locale}}
		for _, warning := range report.Manifest.Warnings {
			page.Rows = append(page.Rows, panel.Row{Name: warning.File, Detail: strconv.Itoa(warning.Line)})
		}
		page.Forms = []panel.Form{{
			Kind:     "restore",
			Action:   "/web/kits",
			Confirm:  view.Confirm,
			Flavor:   view.Flavor,
			Kit:      path,
			Submit:   i18n.SubmitRestore(lang),
			Password: encrypted,
		}}
		return page, nil
	}
	if _, err := g.Restore(form.Confirm, secret); err != nil {
		return noted(lang, "/web/kits", title, explain(lang, err)), nil
	}
	return noted(lang, "/web/kits", title, i18n.Restored(lang)), nil
}

func rollbackPage(root string, lang i18n.Language, getenv func(string) string, flavor string, account int) (panel.Page, error) {
	title := i18n.NavRollback(lang)
	g, err := gateFor("console", getenv)
	if err != nil {
		return noted(lang, "/web/rollback", title, explain(lang, err)), nil
	}
	view, err := g.PreviewRollback(root, flavor, account)
	if page, ok := choicePage(lang, "/web/rollback", title, view, err); ok {
		return page, nil
	}
	if err != nil {
		return noted(lang, "/web/rollback", title, explain(lang, err)), nil
	}
	page := panel.Section(lang, title, "/web/rollback", i18n.LocaleNote(lang))
	if view.Confirm != "" {
		n := account
		if n == 0 {
			n = 1
		}
		page.Forms = []panel.Form{{
			Kind:    "rollback",
			Action:  "/web/rollback",
			Confirm: view.Confirm,
			Flavor:  view.Flavor,
			Account: n,
			Submit:  i18n.SubmitRollback(lang),
		}}
	}
	return page, nil
}

func submitRollback(root string, lang i18n.Language, getenv func(string) string, form localconsole.Form) (panel.Page, error) {
	title := i18n.NavRollback(lang)
	if form.Confirm == "" {
		return rollbackPage(root, lang, getenv, form.Flavor, form.Account)
	}
	g, err := gateFor("console", getenv)
	if err != nil {
		return noted(lang, "/web/rollback", title, explain(lang, err)), nil
	}
	if err := g.Rollback(form.Confirm); err != nil {
		return noted(lang, "/web/rollback", title, explain(lang, err)), nil
	}
	return noted(lang, "/web/rollback", title, i18n.RolledBack(lang)), nil
}

func choicePage(lang i18n.Language, href, title string, view work.MacroView, err error) (panel.Page, bool) {
	if errors.Is(err, work.ErrNeedFlavor) {
		page := noted(lang, href, title, i18n.NeedFlavor(lang))
		page.Rows = flavorRows(href, view.Flavors)
		return page, true
	}
	if errors.Is(err, work.ErrNeedAccount) {
		page := noted(lang, href, title, i18n.NeedAccount(lang))
		page.Rows = accountRows(href, view.Flavor, view.Accounts)
		return page, true
	}
	return panel.Page{}, false
}

func flavorRows(pagePath string, names []string) []panel.Row {
	rows := make([]panel.Row, 0, len(names))
	for _, name := range names {
		q := url.Values{}
		q.Set("flavor", name)
		rows = append(rows, panel.Row{Name: name, Href: hrefQuery(pagePath, q)})
	}
	return rows
}

func accountRows(pagePath, flavor string, names []string) []panel.Row {
	rows := make([]panel.Row, 0, len(names))
	for i, name := range names {
		q := url.Values{}
		if flavor != "" {
			q.Set("flavor", flavor)
		}
		q.Set("account", strconv.Itoa(i+1))
		rows = append(rows, panel.Row{Name: name, Href: hrefQuery(pagePath, q)})
	}
	return rows
}

func noted(lang i18n.Language, href, title, message string) panel.Page {
	page := panel.Section(lang, title, href, i18n.LocaleNote(lang))
	page.Message = message
	return page
}

func hrefQuery(path string, values url.Values) string {
	if len(values) == 0 {
		return path
	}
	return path + "?" + values.Encode()
}

func homeOf(getenv func(string) string) string {
	if getenv == nil {
		return ""
	}
	home := getenv("HOME")
	if home == "" {
		home = getenv("USERPROFILE")
	}
	return home
}

func secretsMatch(a, b []byte) bool {
	if len(bytes.TrimSpace(a)) == 0 || len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

func secretFrom(path string, given []byte) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, false, err
	}
	if info.IsDir() {
		return nil, false, nil
	}
	_, _, encrypted, err := kit.Header(path)
	if err != nil {
		return nil, false, err
	}
	if !encrypted {
		return nil, false, nil
	}
	if len(bytes.TrimSpace(given)) == 0 {
		return nil, true, kit.ErrPass
	}
	return append([]byte(nil), given...), true, nil
}

func passMessage(lang i18n.Language, err error, encrypted bool, given []byte) string {
	if errors.Is(err, kit.ErrPass) && encrypted && len(bytes.TrimSpace(given)) == 0 {
		return i18n.KitNeedPass(lang)
	}
	return explain(lang, err)
}
