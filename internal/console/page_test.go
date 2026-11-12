package console

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

func TestRunningDotOnlyWhenThatGameIsClosed(t *testing.T) {
	one := []install.Install{{Flavor: "_forever_", Locale: "enUS", AgentUID: "wowf"}}
	closed := NewPage(i18n.Portuguese, one, "", i18n.RunningNo(i18n.Portuguese))
	if !closed.ShowFlow || !closed.GameSafe || closed.GameValue != "não" {
		t.Fatalf("closed game = flow %v safe %v value %q", closed.ShowFlow, closed.GameSafe, closed.GameValue)
	}
	if closed.GameLabel != "Jogo aberto" {
		t.Fatalf("one install label = %q", closed.GameLabel)
	}
	open := NewPage(i18n.Portuguese, one, "", i18n.RunningYes(i18n.Portuguese))
	if open.GameSafe || open.GameValue != "sim" {
		t.Fatalf("open game = safe %v value %q", open.GameSafe, open.GameValue)
	}
	unchecked := NewPage(i18n.Portuguese, nil, "", "")
	if unchecked.GameSafe || unchecked.GameValue != "não verificado" {
		t.Fatalf("unchecked = safe %v value %q", unchecked.GameSafe, unchecked.GameValue)
	}
}

func TestSeveralInstallsNameTheScannedFlavor(t *testing.T) {
	page := NewPage(i18n.Portuguese, []install.Install{{Flavor: "_forever_"}, {Flavor: "_classic_"}}, "", i18n.RunningNo(i18n.Portuguese))
	if page.GameLabel != "Jogo aberto · _forever_" {
		t.Fatalf("label = %q", page.GameLabel)
	}
	if page.Heading != i18n.NavInstalls(i18n.Portuguese) {
		t.Fatalf("heading = %q", page.Heading)
	}
}

func TestSectionOmitsTheOverviewKicker(t *testing.T) {
	page := Section(i18n.Portuguese, i18n.NavMacros(i18n.Portuguese), "/web/macros", i18n.LocaleNote(i18n.Portuguese))
	if page.ShowFlow {
		t.Fatal("section shows the write flow")
	}
	var buf bytes.Buffer
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	body := buf.String()
	if strings.Contains(body, "instalações locais") {
		t.Fatal("section repeats the overview kicker")
	}
	if !strings.Contains(body, "<h1>Macros</h1>") {
		t.Fatal("section heading missing")
	}
}
