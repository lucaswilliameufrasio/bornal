// Package console renders the local console pages.
package console

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

//go:embed console.html console.css font/Geist-Regular.woff2 font/Geist-Medium.woff2 font/GeistMono-Regular.woff2
var files embed.FS

// Link is one navigation entry.
type Link struct {
	Href    string
	Label   string
	Current bool
}

// Row is one plain line on a page. Href, when set, is a same-origin link.
type Row struct {
	Name   string
	Detail string
	Href   string
}

// Form is one same-origin POST. The passphrase is typed by the browser and is not rendered back.
type Form struct {
	Kind     string
	Action   string
	CSRF     string
	Confirm  string
	Flavor   string
	Account  int
	Submit   string
	Kit      string
	Password bool
	Twice    bool
}

// Step is one stage of the write path shown on the overview.
type Step struct {
	Index  string
	Title  string
	Detail string
}

// Page is one console render.
type Page struct {
	HTMLLang     string
	Title        string
	Subtitle     string
	Nav          string
	Mode         string
	Heading      string
	Note         string
	Empty        string
	Message      string
	LabelPath    string
	LabelFlavor  string
	LabelLocale  string
	LabelAgent   string
	Missing      string
	FieldLabel   string
	FieldOut     string
	FieldEncrypt string
	FieldKit     string
	FieldReplace string
	FieldRealm   string
	FieldPass    string
	GameLabel    string
	GameValue    string
	GameSafe     bool
	ShowFlow     bool
	FlowTitle    string
	Steps        []Step
	Installs     []install.Install
	Links        []Link
	Rows         []Row
	Forms        []Form
	CSS          template.CSS
}

// NewPage fills the overview. running is the game-open line.
func NewPage(lang i18n.Language, installs []install.Install, empty, running string) Page {
	page := basePage(lang, i18n.NavInstalls(lang), "/web/overview")
	if running == "" {
		running = i18n.RunningUnchecked(lang)
	}
	page.Note = i18n.LocaleNote(lang)
	page.Empty = empty
	page.Installs = installs
	page.ShowFlow = true
	page.FlowTitle = i18n.FlowTitle(lang)
	page.Steps = []Step{
		{Index: "01", Title: i18n.FlowPreview(lang), Detail: i18n.FlowPreviewDetail(lang)},
		{Index: "02", Title: i18n.FlowToken(lang), Detail: i18n.FlowTokenDetail(lang)},
		{Index: "03", Title: i18n.FlowWrite(lang), Detail: i18n.FlowWriteDetail(lang)},
	}
	page.GameLabel = i18n.RunningLabel(lang)
	if len(installs) > 1 {
		flavor := strings.TrimSpace(installs[0].Flavor)
		if flavor != "" {
			page.GameLabel = page.GameLabel + " · " + flavor
		}
	}
	switch running {
	case i18n.RunningYes(lang):
		page.GameValue = i18n.RunningStateYes(lang)
	case i18n.RunningNo(lang):
		page.GameValue = i18n.RunningStateNo(lang)
		page.GameSafe = true
	default:
		page.GameValue = i18n.RunningStateUnchecked(lang)
	}
	return page
}

// Section starts a page that is not the overview.
func Section(lang i18n.Language, title, href, note string) Page {
	page := basePage(lang, title, href)
	page.Note = note
	return page
}

func basePage(lang i18n.Language, title, href string) Page {
	htmlLang := "en"
	if lang == i18n.Portuguese {
		htmlLang = "pt-BR"
	}
	links := []Link{
		{Href: "/web/overview", Label: i18n.NavInstalls(lang), Current: href == "/web/overview"},
		{Href: "/web/macros", Label: i18n.NavMacros(lang), Current: href == "/web/macros"},
		{Href: "/web/addons", Label: i18n.NavAddons(lang), Current: href == "/web/addons"},
		{Href: "/web/kits", Label: i18n.NavKits(lang), Current: href == "/web/kits"},
		{Href: "/web/rollback", Label: i18n.NavRollback(lang), Current: href == "/web/rollback"},
	}
	return Page{
		HTMLLang:     htmlLang,
		Title:        title,
		Subtitle:     i18n.BrandSubtitle(lang),
		Nav:          i18n.NavInstalls(lang),
		Mode:         i18n.ModeLocal(lang),
		Heading:      title,
		LabelPath:    i18n.LabelPath(lang),
		LabelFlavor:  i18n.LabelFlavor(lang),
		LabelLocale:  i18n.LabelLocale(lang),
		LabelAgent:   i18n.LabelAgent(lang),
		Missing:      i18n.LabelMissing(lang),
		FieldLabel:   i18n.FieldLabel(lang),
		FieldOut:     i18n.FieldOut(lang),
		FieldEncrypt: i18n.FieldEncrypt(lang),
		FieldKit:     i18n.FieldKit(lang),
		FieldReplace: i18n.FieldReplace(lang),
		FieldRealm:   i18n.FieldRealm(lang),
		FieldPass:    i18n.FieldPass(lang),
		Links:        links,
	}
}

// Render writes one page. CSS and the fonts are inlined, so the browser makes no second request.
func Render(w io.Writer, page Page) error {
	css, err := files.ReadFile("console.css")
	if err != nil {
		return err
	}
	sheet, err := withFonts(string(css))
	if err != nil {
		return err
	}
	page.CSS = template.CSS(sheet)
	tpl, err := template.ParseFS(files, "console.html")
	if err != nil {
		return err
	}
	return tpl.Execute(w, page)
}

func withFonts(sheet string) (string, error) {
	pairs := []struct {
		token string
		name  string
	}{
		{token: "__FONT_GEIST_400__", name: "font/Geist-Regular.woff2"},
		{token: "__FONT_GEIST_500__", name: "font/Geist-Medium.woff2"},
		{token: "__FONT_GEIST_MONO__", name: "font/GeistMono-Regular.woff2"},
	}
	for _, pair := range pairs {
		if !strings.Contains(sheet, pair.token) {
			return "", fmt.Errorf("console: css missing %s", pair.token)
		}
		raw, err := files.ReadFile(pair.name)
		if err != nil {
			return "", err
		}
		sheet = strings.ReplaceAll(sheet, pair.token, base64.StdEncoding.EncodeToString(raw))
	}
	return sheet, nil
}
