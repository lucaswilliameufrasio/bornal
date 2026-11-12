package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

func envNone(string) string { return "" }

func runCLI(args []string, in string, getenv func(string) string, interactive bool) (int, string, string) {
	var out, err bytes.Buffer
	code := Run(args, strings.NewReader(in), &out, &err, getenv, interactive)
	return code, out.String(), err.String()
}

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

func configEnv(t *testing.T, extra func(string) string) func(string) string {
	t.Helper()
	home := t.TempDir()
	return func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return home
		}
		if extra != nil {
			return extra(key)
		}
		return ""
	}
}

func TestHelpIsEnglishByDefault(t *testing.T) {
	code, out, errText := runCLI(nil, "", envNone, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, "reads local World of Warcraft") {
		t.Fatalf("ajuda: %s", out)
	}
}

func TestHelpFollowsPortuguese(t *testing.T) {
	getenv := func(key string) string {
		if key == "LANG" {
			return "pt_BR.UTF-8"
		}
		return ""
	}
	code, out, _ := runCLI([]string{"help"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out, "lê instalações locais") {
		t.Fatalf("ajuda: %s", out)
	}
}

func TestLangFlagWins(t *testing.T) {
	getenv := func(key string) string {
		if key == "LANG" {
			return "en_US.UTF-8"
		}
		return ""
	}
	code, out, _ := runCLI([]string{"--lang", "pt-BR", "help"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out, "lê instalações locais") {
		t.Fatalf("ajuda: %s", out)
	}
}

func TestSavedLanguageWinsOverTheEnvironment(t *testing.T) {
	getenv := configEnv(t, func(key string) string {
		if key == "LANG" {
			return "en_US.UTF-8"
		}
		return ""
	})
	if err := config.Save(getenv, config.File{Language: "pt-BR"}); err != nil {
		t.Fatal(err)
	}
	code, out, errText := runCLI([]string{"help"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, "lê instalações locais") {
		t.Fatalf("ajuda: %s", out)
	}
}

func TestLangFlagWinsOverSavedLanguage(t *testing.T) {
	getenv := configEnv(t, nil)
	if err := config.Save(getenv, config.File{Language: "pt-BR"}); err != nil {
		t.Fatal(err)
	}
	code, out, errText := runCLI([]string{"--lang", "en", "help"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, "reads local World of Warcraft") {
		t.Fatalf("ajuda: %s", out)
	}
}

func TestDoctorReadsAFixture(t *testing.T) {
	root := t.TempDir()
	writeInstall(t, root, "_forever_")
	code, out, errText := runCLI([]string{"doctor", "--path", root, "--lang", "en"}, "", envNone, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	for _, want := range []string{"_forever_", "enUS", "wowf", "not checked"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ausente %s em %s", want, out)
		}
	}
}

func TestDoctorWithoutAPathDoesNotSucceed(t *testing.T) {
	called := false
	prev := search
	search = func([]string) []install.Install {
		called = true
		return nil
	}
	t.Cleanup(func() { search = prev })
	code, _, errText := runCLI([]string{"doctor", "--lang", "en"}, "", envNone, false)
	if code != 1 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(errText, "bornal setup") {
		t.Fatalf("erro: %s", errText)
	}
	if called {
		t.Fatal("buscou disco sem terminal")
	}
}

func TestDoctorUsesTheSavedPath(t *testing.T) {
	root := t.TempDir()
	flavor := writeInstall(t, root, "_forever_")
	getenv := configEnv(t, nil)
	if err := config.Save(getenv, config.File{Language: "en", InstallPath: root}); err != nil {
		t.Fatal(err)
	}
	code, out, errText := runCLI([]string{"doctor"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, flavor) {
		t.Fatalf("saída: %s", out)
	}
}

func TestPathFlagBeatsTheEnvironmentAndTheSavedPath(t *testing.T) {
	good := t.TempDir()
	bad := t.TempDir()
	writeInstall(t, good, "_forever_")
	getenv := configEnv(t, func(key string) string {
		if key == "BORNAL_WOW_PATH" {
			return bad
		}
		return ""
	})
	if err := config.Save(getenv, config.File{InstallPath: bad}); err != nil {
		t.Fatal(err)
	}
	code, out, errText := runCLI([]string{"doctor", "--path", good, "--lang", "en"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, "_forever_") {
		t.Fatalf("saída: %s", out)
	}
}

func TestEnvironmentBeatsTheSavedPath(t *testing.T) {
	good := t.TempDir()
	bad := t.TempDir()
	writeInstall(t, good, "_forever_")
	getenv := configEnv(t, func(key string) string {
		if key == "BORNAL_WOW_PATH" {
			return good
		}
		return ""
	})
	if err := config.Save(getenv, config.File{InstallPath: bad}); err != nil {
		t.Fatal(err)
	}
	code, out, errText := runCLI([]string{"doctor", "--lang", "en"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, good) {
		t.Fatalf("saída: %s", out)
	}
}

func TestDoctorAsksOnATerminal(t *testing.T) {
	root := t.TempDir()
	flavor := writeInstall(t, root, "_forever_")
	getenv := configEnv(t, nil)
	prev := search
	search = func([]string) []install.Install {
		return []install.Install{{Path: flavor, Flavor: "_forever_"}}
	}
	t.Cleanup(func() { search = prev })
	code, out, errText := runCLI([]string{"doctor"}, "1\n1\n", getenv, true)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	if !strings.Contains(out, "não verificado") || !strings.Contains(out, flavor) {
		t.Fatalf("saída: %s", out)
	}
	stored, err := config.Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Language != "pt-BR" || stored.InstallPath != flavor {
		t.Fatalf("config %+v", stored)
	}
}

func TestSetupSavesAnswers(t *testing.T) {
	root := t.TempDir()
	flavor := writeInstall(t, root, "_forever_")
	getenv := configEnv(t, nil)
	prev := search
	search = func([]string) []install.Install {
		return nil
	}
	t.Cleanup(func() { search = prev })
	code, out, errText := runCLI([]string{"setup", "--lang", "en"}, "\n"+flavor+"\n", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s\n%s", code, errText, out)
	}
	if !strings.Contains(out, "Configuration saved.") {
		t.Fatalf("saída: %s", out)
	}
	stored, err := config.Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Language != "en" || stored.InstallPath != flavor {
		t.Fatalf("config %+v", stored)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, _ := runCLI([]string{"push"}, "", envNone, false)
	if code != 2 {
		t.Fatalf("code %d", code)
	}
}

func TestMcpRejectsExtraArguments(t *testing.T) {
	code, _, errText := runCLI([]string{"mcp", "nope", "--lang", "en"}, "", envNone, false)
	if code != 2 {
		t.Fatalf("code %d %s", code, errText)
	}
}

func TestBadLang(t *testing.T) {
	code, _, _ := runCLI([]string{"--lang", "de", "doctor"}, "", envNone, false)
	if code != 1 {
		t.Fatalf("code %d", code)
	}
}
