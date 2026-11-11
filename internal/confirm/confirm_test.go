package confirm

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func osChtimes(dir, token string, when time.Time) error {
	path := filepath.Join(dir, "console", token)
	return os.Chtimes(path, when, when)
}

func TestTokenIsSingleUseAndSideSpecific(t *testing.T) {
	dir := t.TempDir()
	token, err := Issue(dir, "cli", []byte("plan"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Take(dir, "mcp", token); err == nil {
		t.Fatal("o MCP aceitou o token da CLI")
	}
	body, err := Take(dir, "cli", token)
	if err != nil || string(body) != "plan" {
		t.Fatalf("take %s %v", body, err)
	}
	if _, err := Take(dir, "cli", token); err == nil {
		t.Fatal("token reutilizado")
	}
}

func TestExpiredTokenIsRefused(t *testing.T) {
	dir := t.TempDir()
	token, err := Issue(dir, "console", []byte("plan"))
	if err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-lifetime - time.Minute)
	if err := osChtimes(dir, token, past); err != nil {
		t.Fatal(err)
	}
	if _, err := Take(dir, "console", token); err == nil {
		t.Fatal("token vencido foi aceito")
	}
}
