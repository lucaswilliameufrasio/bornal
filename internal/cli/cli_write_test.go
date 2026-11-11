package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/localconsole"
	"github.com/lucaswilliameufrasio/bornal/internal/running"
)

func stateEnv(t *testing.T) func(string) string {
	t.Helper()
	home := t.TempDir()
	return func(key string) string {
		switch key {
		case "XDG_CONFIG_HOME", "XDG_STATE_HOME", "HOME":
			return home
		default:
			return ""
		}
	}
}

func writeAccount(t *testing.T, root, locale, cache string) string {
	t.Helper()
	flavor := filepath.Join(root, "_forever_")
	account := filepath.Join(flavor, "WTF", "Account", "acct-one")
	if err := os.MkdirAll(filepath.Join(account, "SavedVariables"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "SET textLocale \"" + locale + "\"\nSET agentUID \"wowf\"\nSET engineSurveyPatch \"1.60.1\"\n"
	if err := os.WriteFile(filepath.Join(flavor, "WTF", "Config.wtf"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if cache == "" {
		cache = "VER 3\r\n0000000000000001 \"RS_MACRO\" \"667492\"\r\n/wave\r\nEND\r\n"
	}
	if err := os.WriteFile(filepath.Join(account, "macros-cache.txt"), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := []byte("password = \"hunter2-not-printed\"\n")
	if err := os.WriteFile(filepath.Join(account, "SavedVariables", "Bag.lua"), secret, 0o644); err != nil {
		t.Fatal(err)
	}
	return flavor
}

func cacheFile(flavor string) string {
	return filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt")
}

func confirmToken(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "confirm ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "confirm "))
		}
	}
	return ""
}

func TestMacrosPreviewAndApply(t *testing.T) {
	root := t.TempDir()
	flavor := writeAccount(t, root, "enUS", "")
	before, err := os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	getenv := stateEnv(t)
	code, out, errText := runCLI([]string{"macros", "preview", "--path", root, "--lang", "en"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("code %d %s", code, errText)
	}
	token := confirmToken(out)
	if !strings.Contains(out, "Rend  create") || len(token) != 32 {
		t.Fatalf("prévia: %s", out)
	}
	code, out, errText = runCLI([]string{"macros", "apply", "--path", root, "--lang", "en", "--confirm", token}, "", getenv, false)
	if code != 0 {
		t.Fatalf("apply %d %s", code, errText)
	}
	after, err := os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "/startattack") || !strings.Contains(string(after), "RS_MACRO") || !strings.Contains(string(after), "/wave") {
		t.Fatalf("cache: %s", after)
	}
	if string(after) == string(before) {
		t.Fatal("arquivo não mudou")
	}
	backup, err := os.ReadFile(cacheFile(flavor) + ".bornal-bak")
	if err != nil || string(backup) != string(before) {
		t.Fatalf("backup %v %q", err, backup)
	}
	code, out, errText = runCLI([]string{"macros", "preview", "--path", root, "--lang", "en"}, "", getenv, false)
	if code != 0 || confirmToken(out) != "" {
		t.Fatalf("segunda prévia %d %s %s", code, out, errText)
	}
}

func TestMacroConflictAndOtherLocaleDoNotWrite(t *testing.T) {
	root := t.TempDir()
	flavor := writeAccount(t, root, "enUS", "VER 3\r\n0000000000000002 \"Rend\" \"1\"\r\n/wave\r\nEND\r\n")
	before, err := os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	getenv := stateEnv(t)
	code, out, errText := runCLI([]string{"macros", "preview", "--path", root, "--lang", "en"}, "", getenv, false)
	if code != 0 || confirmToken(out) != "" || !strings.Contains(out, "different text") {
		t.Fatalf("conflito %d %s %s", code, out, errText)
	}
	after, err := os.ReadFile(cacheFile(flavor))
	if err != nil || string(after) != string(before) {
		t.Fatal("conflito alterou o arquivo")
	}

	other := t.TempDir()
	flavor = writeAccount(t, other, "ptBR", "")
	before, err = os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	code, out, errText = runCLI([]string{"macros", "preview", "--path", other, "--lang", "en"}, "", getenv, false)
	if code != 0 || confirmToken(out) != "" || !strings.Contains(out, "not applied") {
		t.Fatalf("locale %d %s %s", code, out, errText)
	}
	after, err = os.ReadFile(cacheFile(flavor))
	if err != nil || string(after) != string(before) {
		t.Fatal("locale alterou o arquivo")
	}
}

func TestApplyRefusesAnOpenGame(t *testing.T) {
	root := t.TempDir()
	flavor := writeAccount(t, root, "enUS", "")
	before, err := os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	getenv := stateEnv(t)
	code, out, errText := runCLI([]string{"macros", "preview", "--path", root, "--lang", "en"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("prévia %d %s", code, errText)
	}
	token := confirmToken(out)
	prev := processList
	processList = func() ([]running.Process, error) {
		return []running.Process{{Cmdline: flavor}}, nil
	}
	t.Cleanup(func() { processList = prev })
	code, _, errText = runCLI([]string{"macros", "apply", "--path", root, "--lang", "en", "--confirm", token}, "", getenv, false)
	if code == 0 || !strings.Contains(errText, "Game running: yes") {
		t.Fatalf("jogo aberto %d %s", code, errText)
	}
	after, err := os.ReadFile(cacheFile(flavor))
	if err != nil || string(after) != string(before) {
		t.Fatal("jogo aberto alterou o arquivo")
	}
}

func TestKitCreateVerifyAndRestore(t *testing.T) {
	root := t.TempDir()
	flavor := writeAccount(t, root, "enUS", "")
	before, err := os.ReadFile(cacheFile(flavor))
	if err != nil {
		t.Fatal(err)
	}
	getenv := stateEnv(t)
	dest := filepath.Join(t.TempDir(), "pack.wowkit")
	code, out, errText := runCLI([]string{"kit", "create", "--path", root, "--lang", "en", "--account", "1", "--out", dest}, "", getenv, false)
	if code != 0 {
		t.Fatalf("create %d %s %s", code, out, errText)
	}
	if !strings.Contains(out, dest) || strings.Contains(out, "hunter2-not-printed") || strings.Contains(out, "acct-one") {
		t.Fatalf("create vazou: %s", out)
	}
	if !strings.Contains(out, "SavedVariables/Bag.lua:") {
		t.Fatalf("aviso: %s", out)
	}
	code, out, errText = runCLI([]string{"kit", "verify", dest, "--lang", "en"}, "", getenv, false)
	if code != 0 || !strings.Contains(out, "_forever_") || strings.Contains(out, "hunter2-not-printed") {
		t.Fatalf("verify %d %s %s", code, out, errText)
	}
	changed := []byte("VER 3\r\n0000000000000001 \"RS_MACRO\" \"667492\"\r\n/sit\r\nEND\r\n")
	if err := os.WriteFile(cacheFile(flavor), changed, 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errText = runCLI([]string{"kit", "restore", dest, "--path", root, "--lang", "en", "--replace"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("restore prévia %d %s %s", code, out, errText)
	}
	token := confirmToken(out)
	if len(token) != 32 || strings.Contains(out, "hunter2-not-printed") || strings.Contains(out, "acct-one") {
		t.Fatalf("restore prévia: %s", out)
	}
	code, out, errText = runCLI([]string{"kit", "restore", dest, "--path", root, "--lang", "en", "--confirm", token, "--replace"}, "", getenv, false)
	if code != 0 {
		t.Fatalf("restore %d %s %s", code, out, errText)
	}
	after, err := os.ReadFile(cacheFile(flavor))
	if err != nil || string(after) != string(before) {
		t.Fatalf("restaurado %v %q", err, after)
	}
	code, out, errText = runCLI([]string{"kit", "find", filepath.Dir(dest), "--lang", "en"}, "", getenv, false)
	if code != 0 || !strings.Contains(out, dest) || strings.Contains(out, "acct-one") {
		t.Fatalf("find %d %s %s", code, out, errText)
	}
}

func TestEncryptWithoutAPromptWritesNothing(t *testing.T) {
	root := t.TempDir()
	writeAccount(t, root, "enUS", "")
	getenv := stateEnv(t)
	dest := filepath.Join(t.TempDir(), "secret.wowkit")
	code, out, errText := runCLI([]string{"kit", "create", "--path", root, "--lang", "en", "--encrypt", "--out", dest}, "", getenv, false)
	if code == 0 || !strings.Contains(errText, "console") {
		t.Fatalf("cifra %d %s %s", code, out, errText)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("arquivo cifrado apareceu sem senha")
	}
	if strings.Contains(out+errText, "hunter2-not-printed") {
		t.Fatal("segredo impresso")
	}
}

func TestConsoleApplyUsesItsOwnToken(t *testing.T) {
	root := t.TempDir()
	flavor := writeAccount(t, root, "enUS", "")
	getenv := stateEnv(t)
	actions := consoleActions(root, i18n.English, getenv)
	actions.CSRF = "csrf-test"
	page, err := actions.Get("macros", "", 0, false)
	if err != nil || len(page.Forms) != 1 || len(page.Forms[0].Confirm) != 32 {
		t.Fatalf("página %+v %v", page, err)
	}
	if page.Forms[0].CSRF != "csrf-test" {
		t.Fatal("csrf ausente")
	}
	token := page.Forms[0].Confirm
	code, _, errText := runCLI([]string{"macros", "apply", "--path", root, "--lang", "en", "--confirm", token}, "", getenv, false)
	if code == 0 {
		t.Fatal("token do console entrou pela cli")
	}
	if !strings.Contains(errText, "Could not read") {
		t.Fatalf("erro: %s", errText)
	}
	page, err = actions.Post("macros", localconsole.Form{Confirm: token})
	if err != nil || !strings.Contains(page.Message, "Macros written") {
		t.Fatalf("post %+v %v", page.Message, err)
	}
	after, err := os.ReadFile(cacheFile(flavor))
	if err != nil || !strings.Contains(string(after), "/startattack") {
		t.Fatalf("cache %v %s", err, after)
	}
	if strings.Contains(page.Message, token) {
		t.Fatal("token repetido na mensagem")
	}
}
