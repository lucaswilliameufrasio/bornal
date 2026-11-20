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
	closed := NewPage(i18n.Portuguese, one, "", []string{i18n.RunningNo(i18n.Portuguese)})
	if !closed.ShowFlow || !closed.GameSafe || closed.GameValue != "não" {
		t.Fatalf("closed game = flow %v safe %v value %q", closed.ShowFlow, closed.GameSafe, closed.GameValue)
	}
	if closed.GameLabel != "Jogo aberto" {
		t.Fatalf("one install label = %q", closed.GameLabel)
	}
	open := NewPage(i18n.Portuguese, one, "", []string{i18n.RunningYes(i18n.Portuguese)})
	if open.GameSafe || open.GameValue != "sim" {
		t.Fatalf("open game = safe %v value %q", open.GameSafe, open.GameValue)
	}
	unchecked := NewPage(i18n.Portuguese, nil, "", nil)
	if unchecked.GameSafe || unchecked.GameValue != "não verificado" {
		t.Fatalf("unchecked = safe %v value %q", unchecked.GameSafe, unchecked.GameValue)
	}
}

func TestEachInstallKeepsItsOwnGameState(t *testing.T) {
	installs := []install.Install{{Flavor: "_forever_"}, {Flavor: "_classic_"}}
	page := NewPage(i18n.Portuguese, installs, "", []string{
		i18n.RunningNo(i18n.Portuguese),
		i18n.RunningYes(i18n.Portuguese),
	})
	if page.GameValue != "" {
		t.Fatalf("page cell = %q", page.GameValue)
	}
	if page.Heading != i18n.NavInstalls(i18n.Portuguese) {
		t.Fatalf("heading = %q", page.Heading)
	}
	if len(page.Shown) != 2 {
		t.Fatalf("shown = %d", len(page.Shown))
	}
	if !page.Shown[0].GameSafe || page.Shown[0].GameValue != "não" {
		t.Fatalf("first = safe %v %q", page.Shown[0].GameSafe, page.Shown[0].GameValue)
	}
	if page.Shown[1].GameSafe || page.Shown[1].GameValue != "sim" {
		t.Fatalf("second = safe %v %q", page.Shown[1].GameSafe, page.Shown[1].GameValue)
	}
	var buf bytes.Buffer
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	body := buf.String()
	if strings.Count(body, "class=\"dot\"") != 1 {
		t.Fatalf("dots = %d\n%s", strings.Count(body, "class=\"dot\""), body)
	}

	short := NewPage(i18n.Portuguese, installs, "", []string{i18n.RunningNo(i18n.Portuguese)})
	if short.Shown[0].GameValue != "não" || short.Shown[1].GameSafe || short.Shown[1].GameValue != "não verificado" {
		t.Fatalf("missing line first %q second safe %v %q", short.Shown[0].GameValue, short.Shown[1].GameSafe, short.Shown[1].GameValue)
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
