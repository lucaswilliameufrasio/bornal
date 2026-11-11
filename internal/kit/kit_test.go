package kit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (string, string) {
	t.Helper()
	flavor := filepath.Join(t.TempDir(), "_forever_")
	account := "acct-one"
	accountDir := filepath.Join(flavor, "WTF", "Account", account)
	saved := filepath.Join(accountDir, "SavedVariables")
	char := filepath.Join(accountDir, "91", "Hero")
	if err := os.MkdirAll(saved, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(char, "SavedVariables"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(flavor, "WTF"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "SET textLocale \"enUS\"\nSET audioLocale \"enUS\"\nSET GxApi \"vulkan\"\n"
	if err := os.WriteFile(filepath.Join(flavor, "WTF", "Config.wtf"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir, "macros-cache.txt"), []byte("VER 3\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := []byte("password = \"super-secret-value\"\nsafe = 1\n")
	if err := os.WriteFile(filepath.Join(saved, "Secret.lua"), secret, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(saved, "Secret.lua.bak"), []byte("password = \"hidden-bak\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir, "chat-cache.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(accountDir, "Cache"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(saved, "Secret.lua"), filepath.Join(saved, "Link.lua")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(char, "SavedVariables", "Bag.lua"), []byte("Bag = {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return flavor, account
}

func TestCreateSkipsSecretsSymlinksAndGraphics(t *testing.T) {
	flavor, account := fixture(t)
	dest := filepath.Join(t.TempDir(), "forever.wowkit")
	manifest, err := Create(CreateOptions{
		InstallPath: flavor,
		Account:     account,
		Locale:      "enUS",
		Build:       "1.60.1",
		Dest:        dest,
		Now:         time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dest, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "super-secret-value") || strings.Contains(text, "hidden-bak") || strings.Contains(text, "vulkan") {
		t.Fatal("o manifesto guardou um valor que não devia")
	}
	if len(manifest.Warnings) != 1 || manifest.Warnings[0].Line != 1 || !strings.HasSuffix(manifest.Warnings[0].File, "Secret.lua") {
		t.Fatalf("aviso %+v", manifest.Warnings)
	}
	if _, err := os.Stat(filepath.Join(dest, "account", "SavedVariables", "Link.lua")); !os.IsNotExist(err) {
		t.Fatal("o kit copiou um symlink")
	}
	if _, err := os.Stat(filepath.Join(dest, "account", "SavedVariables", "Secret.lua.bak")); !os.IsNotExist(err) {
		t.Fatal("o kit copiou um .bak")
	}
	portable, err := os.ReadFile(filepath.Join(dest, "config", "portable.wtf"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(portable, []byte("textLocale")) || bytes.Contains(portable, []byte("GxApi")) {
		t.Fatalf("portable %s", portable)
	}
	report, err := Verify(dest, nil, 0)
	if err != nil || !report.Verified {
		t.Fatalf("verify %v %+v", err, report.Verified)
	}
}

func TestEncryptedHeaderHasNoAccountData(t *testing.T) {
	flavor, account := fixture(t)
	dest := filepath.Join(t.TempDir(), "forever.wowkit")
	pass := []byte("test-passphrase")
	if _, err := Create(CreateOptions{
		InstallPath: flavor,
		Account:     account,
		Locale:      "enUS",
		Dest:        dest,
		Encrypt:     true,
		Passphrase:  pass,
		Work:        10,
		Now:         time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	sep := bytes.Index(body, []byte("\n\n"))
	if sep < 0 {
		t.Fatal("sem cabeçalho")
	}
	header := string(body[:sep])
	if !strings.Contains(header, "WOWKIT1") || !strings.Contains(header, "kdf: age-scrypt") {
		t.Fatalf("cabeçalho %s", header)
	}
	if strings.Contains(header, account) || strings.Contains(header, "Hero") || strings.Contains(header, "super-secret-value") {
		t.Fatal("o cabeçalho tem dado da conta")
	}
	if _, err := Verify(dest, []byte("wrong-passphrase"), 0); err != ErrPass {
		t.Fatalf("senha errada %v", err)
	}
	marker := filepath.Join(flavor, "WTF", "Account", account, "macros-cache.txt")
	before, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(RestoreOptions{KitPath: dest, InstallPath: flavor, Locale: "enUS", Passphrase: []byte("wrong-passphrase")}); err != ErrPass {
		t.Fatalf("restore %v", err)
	}
	after, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("senha errada alterou o destino")
	}
	report, err := Verify(dest, []byte("test-passphrase"), 0)
	if err != nil || !report.Verified {
		t.Fatalf("verify %v", err)
	}
}

func TestTamperedRestoreLeavesDestinationUnchanged(t *testing.T) {
	flavor, account := fixture(t)
	dest := filepath.Join(t.TempDir(), "forever.wowkit")
	if _, err := Create(CreateOptions{InstallPath: flavor, Account: account, Locale: "enUS", Dest: dest}); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(flavor, "WTF", "Account", account, "macros-cache.txt")
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(dest, "account", "macros-cache.txt")
	if err := os.WriteFile(payload, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = Restore(RestoreOptions{KitPath: dest, InstallPath: flavor, Locale: "enUS"})
	if err != ErrHash {
		t.Fatalf("restore %v", err)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("kit alterado mudou o destino")
	}
	outside := filepath.Join(filepath.Dir(dest), "pwned.txt")
	manifestPath := filepath.Join(dest, "manifest.json")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest = bytes.Replace(manifest, []byte(`"files": [`), []byte(`"files": [{"path":"../pwned.txt","sha256":"abc","size":1},`), 1)
	if err := os.WriteFile(manifestPath, manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(RestoreOptions{KitPath: dest, InstallPath: flavor, Locale: "enUS"}); err == nil {
		t.Fatal("caminho fora do kit foi aceito")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatal("escreveu fora do kit")
	}
}

func TestRestoreRefusesAMissingAccountAndANewerFile(t *testing.T) {
	flavor, account := fixture(t)
	dest := filepath.Join(t.TempDir(), "forever.wowkit")
	if _, err := Create(CreateOptions{
		InstallPath: flavor,
		Account:     account,
		Locale:      "enUS",
		Build:       "1.60.1",
		Dest:        dest,
		Now:         time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "_forever_")
	if err := os.MkdirAll(filepath.Join(other, "WTF"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "WTF", "Config.wtf"), []byte("SET textLocale \"enUS\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(RestoreOptions{KitPath: dest, InstallPath: other, Locale: "enUS", Build: "1.60.1"}); err != ErrAccount {
		t.Fatalf("conta %v", err)
	}
	if _, err := os.Stat(filepath.Join(other, "WTF", "Account")); !os.IsNotExist(err) {
		t.Fatal("criou a pasta da conta")
	}
	target := filepath.Join(flavor, "WTF", "Account", account, "macros-cache.txt")
	future := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(target, future, future); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(RestoreOptions{KitPath: dest, InstallPath: flavor, Locale: "enUS", Build: "1.60.1"}); err != ErrNewer {
		t.Fatalf("newer %v", err)
	}
	if _, err := Restore(RestoreOptions{KitPath: dest, InstallPath: flavor, Locale: "enUS", Build: "1.60.1", Replace: true}); err != nil {
		t.Fatal(err)
	}
}

func TestFindIsShallow(t *testing.T) {
	volume := t.TempDir()
	dir := filepath.Join(volume, "WoW Kits")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	flavor, account := fixture(t)
	if _, err := Create(CreateOptions{InstallPath: flavor, Account: account, Locale: "enUS", Dest: filepath.Join(dir, "plain.wowkit")}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "hidden.wowkit"), []byte("WOWKIT1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "plain.wowkit"), filepath.Join(volume, "link.wowkit")); err != nil {
		t.Fatal(err)
	}
	hits := Look(Places("", []string{volume}, ""))
	if len(hits) != 1 || hits[0].Encrypted || strings.Contains(hits[0].Label, account) {
		t.Fatalf("hits %+v", hits)
	}
	if _, err := os.Stat(filepath.Join(volume, "link.wowkit")); err != nil {
		t.Fatal(err)
	}
}
