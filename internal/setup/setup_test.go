package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

func writeInstall(t *testing.T, parent, flavor string) string {
	t.Helper()
	dir := filepath.Join(parent, flavor, "WTF")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("SET textLocale \"enUS\"\nSET agentUID \"wowf\"\n")
	if err := os.WriteFile(filepath.Join(dir, "Config.wtf"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(parent, flavor)
}

func TestAskSelectsTheSecondInstall(t *testing.T) {
	root := t.TempDir()
	first := writeInstall(t, root, "_forever_")
	second := writeInstall(t, root, "_retail_")
	var out strings.Builder
	got, err := (Asker{
		In:              strings.NewReader("2\n2\n"),
		Out:             &out,
		DefaultLanguage: i18n.Portuguese,
		Discover: func([]string) []install.Install {
			return []install.Install{{Path: first, Flavor: "_forever_"}, {Path: second, Flavor: "_retail_"}}
		},
		Inspect: install.Inspect,
	}).Run()
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != "en" || got.InstallPath != second {
		t.Fatalf("got %+v", got)
	}
	if !strings.Contains(out.String(), "Interface language") || !strings.Contains(out.String(), second) {
		t.Fatalf("perguntas: %s", out.String())
	}
}

func TestAskRejectsAPathThenKeepsARealOne(t *testing.T) {
	root := t.TempDir()
	flavor := writeInstall(t, root, "_forever_")
	var out strings.Builder
	got, err := (Asker{
		In:              strings.NewReader("\n" + filepath.Join(root, "nope") + "\n" + flavor + "\n"),
		Out:             &out,
		DefaultLanguage: i18n.English,
		Discover: func([]string) []install.Install {
			return nil
		},
		Inspect: install.Inspect,
	}).Run()
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != "en" || got.InstallPath != flavor {
		t.Fatalf("got %+v texto %s", got, out.String())
	}
	if !strings.Contains(out.String(), "no installation") {
		t.Fatalf("aviso: %s", out.String())
	}
}

func TestAskStopsAtEOF(t *testing.T) {
	_, err := (Asker{
		In:              strings.NewReader(""),
		Out:             &strings.Builder{},
		DefaultLanguage: i18n.English,
		Discover: func([]string) []install.Install {
			t.Fatal("não deveria buscar depois do EOF da primeira pergunta")
			return nil
		},
		Inspect: install.Inspect,
	}).Run()
	if err == nil {
		t.Fatal("eof deveria parar")
	}
}
