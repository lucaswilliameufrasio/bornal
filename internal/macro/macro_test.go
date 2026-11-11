package macro

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKeepsAnExistingMacroAndPlansTheSet(t *testing.T) {
	body := "VER 3\r\n0000000000000001 \"RS_MACRO\" \"667492\"\r\n/say hi\r\nEND\r\n"
	got, err := Parse([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	plan, next := PlanSet(got, "enUS", false)
	if plan.Conflict || !plan.LocaleOK {
		t.Fatalf("plano %+v", plan)
	}
	if next[0].Name != "RS_MACRO" || next[0].ID != "0000000000000001" || next[0].Body != "/say hi" {
		t.Fatalf("macro existente %+v", next[0])
	}
	formatted, err := Format(next)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Parse(formatted)
	if err != nil || again[0].Body != "/say hi" {
		t.Fatalf("releitura %+v %v", again, err)
	}
	for _, one := range Catalog() {
		if strings.Contains(strings.ToLower(one.Body), "/cast attack") {
			t.Fatalf("%s contém /cast Attack", one.Name)
		}
		if one.Name != "Battle Shout" && !strings.Contains(one.Body, "/startattack") {
			t.Fatalf("%s não inicia o ataque", one.Name)
		}
	}
}

func TestConflictDoesNotReplaceUntilAsked(t *testing.T) {
	current := []Macro{{ID: "0000000000000002", Name: "Rend", Icon: "1", Body: "/wave"}}
	plan, next := PlanSet(current, "enUS", false)
	if !plan.Conflict || next[0].Body != "/wave" {
		t.Fatal("substituiu sem replace")
	}
	plan, next = PlanSet(current, "enUS", true)
	if plan.Conflict || !strings.Contains(next[0].Body, "/startattack") || next[0].ID != "0000000000000002" {
		t.Fatalf("replace %+v %+v", plan, next[0])
	}
}

func TestRejectsCastAttack(t *testing.T) {
	_, err := Format([]Macro{{ID: "0000000000000003", Name: "Bad", Icon: icon, Body: "/startattack\n/cast Attack"}})
	if err == nil {
		t.Fatal("aceitou /cast Attack")
	}
}

func TestWriteAndBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "macros-cache.txt")
	if err := os.WriteFile(path, []byte("VER 3\n0000000000000001 \"RS_MACRO\" \"1\"\n/say hi\nEND\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Backup(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bornal-bak"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "macros-cache.old")); err == nil {
		t.Fatal("usou macros-cache.old")
	}
	macros, _, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	_, next := PlanSet(macros, "enUS", false)
	if err := WriteFile(path, next); err != nil {
		t.Fatal(err)
	}
	bak, err := os.ReadFile(path + ".bornal-bak")
	if err != nil || !strings.Contains(string(bak), "RS_MACRO") {
		t.Fatal("backup perdeu o arquivo anterior")
	}
}

func TestNonEnglishLocaleIsNotAppliedAsEnglish(t *testing.T) {
	plan, _ := PlanSet(nil, "ptBR", false)
	if plan.LocaleOK {
		t.Fatal("ptBR foi tratado como inglês")
	}
}
