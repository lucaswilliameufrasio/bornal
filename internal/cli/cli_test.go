package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func envNone(string) string { return "" }

func TestHelpIsEnglishByDefault(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run(nil, &out, &err, envNone); code != 0 {
		t.Fatalf("code %d %s", code, err.String())
	}
	if !strings.Contains(out.String(), "reads local World of Warcraft") {
		t.Fatalf("ajuda: %s", out.String())
	}
}

func TestHelpFollowsPortuguese(t *testing.T) {
	var out, err bytes.Buffer
	getenv := func(key string) string {
		if key == "LANG" {
			return "pt_BR.UTF-8"
		}
		return ""
	}
	if code := Run([]string{"help"}, &out, &err, getenv); code != 0 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out.String(), "lê instalações locais") {
		t.Fatalf("ajuda: %s", out.String())
	}
}

func TestLangFlagWins(t *testing.T) {
	var out, err bytes.Buffer
	getenv := func(key string) string {
		if key == "LANG" {
			return "en_US.UTF-8"
		}
		return ""
	}
	if code := Run([]string{"--lang", "pt-BR", "help"}, &out, &err, getenv); code != 0 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out.String(), "lê instalações locais") {
		t.Fatalf("ajuda: %s", out.String())
	}
}

func TestDoctorReadsAFixture(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "_forever_", "WTF")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Config.wtf"), []byte("SET textLocale \"enUS\"\nSET agentUID \"wowf\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, err bytes.Buffer
	code := Run([]string{"doctor", "--path", root, "--lang", "en"}, &out, &err, envNone)
	if code != 0 {
		t.Fatalf("code %d %s", code, err.String())
	}
	body := out.String()
	for _, want := range []string{"_forever_", "enUS", "wowf", "not checked"} {
		if !strings.Contains(body, want) {
			t.Fatalf("ausente %s em %s", want, body)
		}
	}
}

func TestDoctorWithoutAPathDoesNotSucceed(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run([]string{"doctor", "--lang", "en"}, &out, &err, envNone); code != 1 {
		t.Fatalf("code %d %s", code, err.String())
	}
	if !strings.Contains(err.String(), "No path given") {
		t.Fatalf("erro: %s", err.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run([]string{"push"}, &out, &err, envNone); code != 2 {
		t.Fatalf("code %d", code)
	}
}

func TestBadLang(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run([]string{"--lang", "de", "doctor"}, &out, &err, envNone); code != 1 {
		t.Fatalf("code %d", code)
	}
}
