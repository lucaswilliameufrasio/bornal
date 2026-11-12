package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectReadsFlavorChild(t *testing.T) {
	root := t.TempDir()
	forever := filepath.Join(root, "_forever_", "WTF")
	if err := os.MkdirAll(forever, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "SET locale \"ptBR\"\nSET textLocale \"enUS\"\nSET agentUID \"wowf\"\n"
	if err := os.WriteFile(filepath.Join(forever, "Config.wtf"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	retail := filepath.Join(root, "_retail_", "WTF")
	if err := os.MkdirAll(retail, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(retail, "Config.wtf"), []byte("SET textLocale \"ptBR\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("len = %d", len(found))
	}
	if found[0].Flavor != "_forever_" || found[0].Locale != "enUS" || found[0].AgentUID != "wowf" {
		t.Fatalf("primeiro = %+v", found[0])
	}
	if found[1].Flavor != "_retail_" || found[1].Locale != "ptBR" || found[1].AgentUID != "" {
		t.Fatalf("segundo = %+v", found[1])
	}
}

func TestInspectAcceptsTheFlavorDirectoryItself(t *testing.T) {
	root := filepath.Join(t.TempDir(), "_forever_")
	if err := os.MkdirAll(filepath.Join(root, "WTF"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "WTF", "Config.wtf"), []byte("SET textLocale \"enUS\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	found, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Locale != "enUS" {
		t.Fatalf("%+v", found)
	}
}

func TestInspectEmptyDirectory(t *testing.T) {
	found, err := Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Fatalf("len = %d", len(found))
	}
}

func TestInspectMissingPath(t *testing.T) {
	if _, err := Inspect(filepath.Join(t.TempDir(), "ausente")); err == nil {
		t.Fatal("caminho ausente foi aceito")
	}
}
