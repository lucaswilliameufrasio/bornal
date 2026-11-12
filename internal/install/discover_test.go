package install

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFlavor(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "WTF"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("SET textLocale \"enUS\"\nSET agentUID \"wowf\"\n")
	if err := os.WriteFile(filepath.Join(dir, "WTF", "Config.wtf"), body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverFindsAFlavorFourLevelsDown(t *testing.T) {
	root := t.TempDir()
	hit := filepath.Join(root, "a", "b", "c", "d")
	writeFlavor(t, hit)
	miss := filepath.Join(root, "a", "b", "c", "e", "f")
	writeFlavor(t, miss)
	found := Discover([]string{root})
	if len(found) != 1 || found[0].Path != hit {
		t.Fatalf("found %+v", found)
	}
	if found[0].Locale != "enUS" || found[0].AgentUID != "wowf" {
		t.Fatalf("leitura %+v", found[0])
	}
}

func TestDiscoverSkipsHeavyAndHiddenNames(t *testing.T) {
	root := t.TempDir()
	writeFlavor(t, filepath.Join(root, "Data", "_forever_"))
	writeFlavor(t, filepath.Join(root, ".hidden", "_forever_"))
	if found := Discover([]string{root}); len(found) != 0 {
		t.Fatalf("found %+v", found)
	}
}

func TestRootsForDoesNotTouchTheDisk(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "HOME":
			return "/home/ada"
		case "USER":
			return "ada"
		case "BORNAL_WOW_PATH":
			return "/games/wow"
		case "ProgramFiles":
			return `C:\Program Files`
		default:
			return ""
		}
	}
	linux := RootsFor("linux", getenv)
	games := filepath.Join("/home/ada", "Games")
	if !containsAll(linux, []string{"/games/wow", games, "/mnt", "/media", filepath.Join("/run/media", "ada")}) {
		t.Fatalf("linux %v", linux)
	}
	mac := RootsFor("darwin", getenv)
	if !containsAll(mac, []string{"/Applications", "/Volumes", games}) || contains(mac, "/mnt") {
		t.Fatalf("darwin %v", mac)
	}
	windows := RootsFor("windows", getenv)
	if !containsAll(windows, []string{`C:\Program Files`, games}) || contains(windows, "/mnt") {
		t.Fatalf("windows %v", windows)
	}
}

func containsAll(got, want []string) bool {
	for _, item := range want {
		if !contains(got, item) {
			return false
		}
	}
	return true
}

func contains(got []string, want string) bool {
	for _, item := range got {
		if item == want {
			return true
		}
	}
	return false
}
