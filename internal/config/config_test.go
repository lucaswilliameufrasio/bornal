package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	home := t.TempDir()
	getenv := func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return home
		}
		return ""
	}
	want := File{Language: "pt-BR", InstallPath: "/jogos/World of Warcraft"}
	if err := Save(getenv, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v", got)
	}
	info, err := os.Stat(filepath.Join(home, "bornal", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm %o", info.Mode().Perm())
	}
}

func TestLoadMissing(t *testing.T) {
	getenv := func(string) string { return "" }
	got, err := Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if got != (File{}) {
		t.Fatalf("got %+v", got)
	}
}
