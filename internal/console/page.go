// Package console renders the local overview page.
package console

import (
	"embed"
	"html/template"
	"io"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

//go:embed console.html console.css
var files embed.FS

// Link is one rail entry.
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

// Page is one console render.
type Page struct {
	HTMLLang    string
	Title       string
	Subtitle    string
	Nav         string
	Mode        string
	Heading     string
	Note        string
	Empty       string
	Message     string
	LabelPath   string
	LabelFlavor string
	LabelLocale string
	LabelAgent  string
	Missing     string
	Installs    []install.Install
	Links       []Link
	Rows        []Row
	Forms       []Form
	CSS         template.CSS
}

// NewPage fills the overview. running is the game-open line.
func NewPage(lang i18n.Language, installs []install.Install, empty, running string) Page {
	page := basePage(lang, i18n.NavInstalls(lang), "/web/overview")
	if running == "" {
		running = i18n.RunningUnchecked(lang)
	}
	page.Note = i18n.LocaleNote(lang) + " " + running
	page.Empty = empty
	page.Installs = installs
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
		HTMLLang:    htmlLang,
		Title:       title,
		Subtitle:    i18n.BrandSubtitle(lang),
		Nav:         i18n.NavInstalls(lang),
		Mode:        i18n.ModeLocal(lang),
		Heading:     title,
		LabelPath:   i18n.LabelPath(lang),
		LabelFlavor: i18n.LabelFlavor(lang),
		LabelLocale: i18n.LabelLocale(lang),
		LabelAgent:  i18n.LabelAgent(lang),
		Missing:     i18n.LabelMissing(lang),
		Links:       links,
	}
}

// Render writes the overview. CSS is inlined so the page has no second request.
func Render(w io.Writer, page Page) error {
	css, err := files.ReadFile("console.css")
	if err != nil {
		return err
	}
	page.CSS = template.CSS(css)
	tpl, err := template.ParseFS(files, "console.html")
	if err != nil {
		return err
	}
	return tpl.Execute(w, page)
}
