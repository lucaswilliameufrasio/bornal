package addon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListReadsTOCAndDisabledState(t *testing.T) {
	root := t.TempDir()
	flavor := filepath.Join(root, "_forever_")
	addonDir := filepath.Join(flavor, "Interface", "AddOns", "Bag")
	if err := os.MkdirAll(addonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	toc := "## Interface: 16001\n## Title: Bag\n## Version: 1.2\n## Dependencies: MissingDep\n"
	if err := os.WriteFile(filepath.Join(addonDir, "Bag.toc"), []byte(toc), 0o644); err != nil {
		t.Fatal(err)
	}
	char := filepath.Join(flavor, "WTF", "Account", "acct-one", "91", "Hero")
	if err := os.MkdirAll(char, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(char, "AddOns.txt"), []byte("Bag: disabled\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(addonDir, filepath.Join(flavor, "Interface", "AddOns", "Linked")); err != nil {
		t.Fatal(err)
	}

	found, err := List(flavor)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("addons %+v", found)
	}
	if found[0].Folder != "Bag" || found[0].Interface != "16001" || !found[0].Disabled || !found[0].InterfaceOK {
		t.Fatalf("addon %+v", found[0])
	}
	if len(found[0].Missing) != 1 || found[0].Missing[0] != "MissingDep" {
		t.Fatalf("deps %+v", found[0].Missing)
	}
}

func TestForeverInterfaceMismatch(t *testing.T) {
	root := t.TempDir()
	flavor := filepath.Join(root, "_forever_")
	dir := filepath.Join(flavor, "Interface", "AddOns", "Old")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Old.toc"), []byte("## Interface: 11500\n## Title: Old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	found, err := List(flavor)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].InterfaceOK {
		t.Fatalf("addon %+v", found)
	}
}
