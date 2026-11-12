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

// Page is one overview render.
type Page struct {
	HTMLLang    string
	Title       string
	Subtitle    string
	Nav         string
	Mode        string
	Heading     string
	Note        string
	Empty       string
	LabelPath   string
	LabelFlavor string
	LabelLocale string
	LabelAgent  string
	Missing     string
	Installs    []install.Install
	CSS         template.CSS
}

// NewPage fills the labels for lang. empty is shown when Installs is empty.
func NewPage(lang i18n.Language, installs []install.Install, empty string) Page {
	htmlLang := "en"
	if lang == i18n.Portuguese {
		htmlLang = "pt-BR"
	}
	return Page{
		HTMLLang:    htmlLang,
		Title:       i18n.NavInstalls(lang),
		Subtitle:    i18n.BrandSubtitle(lang),
		Nav:         i18n.NavInstalls(lang),
		Mode:        i18n.ModeLocal(lang),
		Heading:     i18n.NavInstalls(lang),
		Note:        i18n.LocaleNote(lang) + " " + i18n.RunningUnchecked(lang),
		Empty:       empty,
		LabelPath:   i18n.LabelPath(lang),
		LabelFlavor: i18n.LabelFlavor(lang),
		LabelLocale: i18n.LabelLocale(lang),
		LabelAgent:  i18n.LabelAgent(lang),
		Missing:     i18n.LabelMissing(lang),
		Installs:    installs,
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
