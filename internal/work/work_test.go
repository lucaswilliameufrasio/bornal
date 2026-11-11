package work

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/macro"
)

func installFixture(t *testing.T, locale string) (string, string) {
	t.Helper()
	root := t.TempDir()
	flavor := filepath.Join(root, "_forever_")
	account := filepath.Join(flavor, "WTF", "Account", "acct-one")
	if err := os.MkdirAll(account, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "SET textLocale \"" + locale + "\"\nSET agentUID \"wowf\"\nSET engineSurveyPatch \"1.60.1\"\n"
	if err := os.WriteFile(filepath.Join(flavor, "WTF", "Config.wtf"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := "VER 3\r\n0000000000000001 \"RS_MACRO\" \"667492\"\r\n/wave\r\nEND\r\n"
	if err := os.WriteFile(filepath.Join(account, "macros-cache.txt"), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, flavor
}

func gate(t *testing.T, open bool, fail bool) Gate {
	t.Helper()
	return Gate{
		Side:     "cli",
		StateDir: t.TempDir(),
		Open: func(string) (bool, error) {
			if fail {
				return false, os.ErrInvalid
			}
			return open, nil
		},
	}
}

func TestApplyKeepsExistingMacrosAndRefusesAnOpenGame(t *testing.T) {
	root, flavor := installFixture(t, "enUS")
	g := gate(t, false, false)
	view, err := g.PreviewMacros(root, "", 0, false)
	if err != nil || view.Confirm == "" || !view.LocaleOK {
		t.Fatalf("preview %v %+v", err, view)
	}
	running := gate(t, true, false)
	running.StateDir = g.StateDir
	if err := running.ApplyMacros(view.Confirm); err != ErrRunning {
		t.Fatalf("jogo aberto %v", err)
	}
	if _, err := os.Stat(filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.bornal-bak")); !os.IsNotExist(err) {
		t.Fatal("gravou backup com o jogo aberto")
	}
	view, err = g.PreviewMacros(root, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	other := g
	other.Side = "mcp"
	if err := other.ApplyMacros(view.Confirm); err != ErrConfirm {
		t.Fatalf("outro lado %v", err)
	}
	view, err = g.PreviewMacros(root, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyMacros(view.Confirm); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "RS_MACRO") || !strings.Contains(text, "/startattack") || strings.Contains(strings.ToLower(text), "/cast attack") {
		t.Fatalf("cache %s", text)
	}
	if !strings.Contains(text, "\r\n") {
		t.Fatal("sem CRLF")
	}
	if _, err := os.Stat(path + ".old"); !os.IsNotExist(err) {
		t.Fatal("criou macros-cache.old")
	}
	if _, err := os.Stat(path + ".bornal-bak"); err != nil {
		t.Fatal(err)
	}
	macros, _, err := macro.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if macros[0].ID != "0000000000000001" || macros[0].Body != "/wave" {
		t.Fatalf("macro original %+v", macros[0])
	}
}

func TestConflictAndLocaleDoNotWrite(t *testing.T) {
	root, flavor := installFixture(t, "enUS")
	path := filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt")
	cache := "VER 3\r\n0000000000000002 \"Rend\" \"134400\"\r\n/wave\r\nEND\r\n"
	if err := os.WriteFile(path, []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	g := gate(t, false, false)
	view, err := g.PreviewMacros(root, "", 0, false)
	if err != nil || !view.Conflict || view.Confirm != "" {
		t.Fatalf("conflito %+v %v", view, err)
	}
	before, _ := os.ReadFile(path)
	view, err = g.PreviewMacros(root, "", 0, true)
	if err != nil || view.Confirm == "" {
		t.Fatalf("replace %v %+v", err, view)
	}
	changed := "VER 3\r\n0000000000000002 \"Rend\" \"134400\"\r\n/sit\r\nEND\r\n"
	if err := os.WriteFile(path, []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyMacros(view.Confirm); err != ErrChanged {
		t.Fatalf("hash %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != changed || string(after) == string(before) {
		t.Fatal("o apply gravou depois que o arquivo mudou")
	}
	ptRoot, ptFlavor := installFixture(t, "ptBR")
	pt := gate(t, false, false)
	ptView, err := pt.PreviewMacros(ptRoot, "", 0, false)
	if err != nil || ptView.LocaleOK || ptView.Confirm != "" {
		t.Fatalf("locale %+v %v", ptView, err)
	}
	ptPath := filepath.Join(ptFlavor, "WTF", "Account", "acct-one", "macros-cache.txt")
	ptBefore, _ := os.ReadFile(ptPath)
	if err := pt.ApplyMacros("0123456789abcdef0123456789abcdef"); err != ErrConfirm {
		t.Fatal(err)
	}
	ptAfter, _ := os.ReadFile(ptPath)
	if string(ptBefore) != string(ptAfter) {
		t.Fatal("locale recusou e mesmo assim gravou")
	}
}

func TestRollbackRestoresTheBackup(t *testing.T) {
	root, flavor := installFixture(t, "enUS")
	g := gate(t, false, false)
	view, err := g.PreviewMacros(root, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyMacros(view.Confirm); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt")
	backup, err := os.ReadFile(path + ".bornal-bak")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("VER 3\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rolled, err := g.PreviewRollback(root, "", 0)
	if err != nil || rolled.Confirm == "" {
		t.Fatalf("rollback %v", err)
	}
	if err := g.Rollback(rolled.Confirm); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(backup) {
		t.Fatal("rollback não devolveu o backup")
	}
}

func TestCheckFailureRefusesTheWrite(t *testing.T) {
	root, _ := installFixture(t, "enUS")
	g := gate(t, false, false)
	view, err := g.PreviewMacros(root, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	blind := g
	blind.Open = func(string) (bool, error) { return false, os.ErrInvalid }
	if err := blind.ApplyMacros(view.Confirm); err != ErrCheck {
		t.Fatalf("checagem %v", err)
	}
}
