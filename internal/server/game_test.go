package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func writeMacroInstall(t *testing.T, locale string) string {
	t.Helper()
	flavor := filepath.Join(t.TempDir(), "_forever_")
	account := filepath.Join(flavor, "WTF", "Account", "acct-one")
	if err := os.MkdirAll(filepath.Join(account, "SavedVariables"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "SET textLocale \"" + locale + "\"\nSET agentUID \"wowf\"\nSET engineSurveyPatch \"1.60.1\"\n"
	if err := os.WriteFile(filepath.Join(flavor, "WTF", "Config.wtf"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := "VER 3\r\n0000000000000001 \"RS_MACRO\" \"667492\"\r\n/wave\r\nEND\r\n"
	if err := os.WriteFile(filepath.Join(account, "macros-cache.txt"), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := []byte("password = \"hunter2-not-printed\"\n")
	if err := os.WriteFile(filepath.Join(account, "SavedVariables", "Bag.lua"), secret, 0o644); err != nil {
		t.Fatal(err)
	}
	addon := filepath.Join(flavor, "Interface", "AddOns", "Bag")
	if err := os.MkdirAll(addon, 0o755); err != nil {
		t.Fatal(err)
	}
	toc := "## Interface: 16001\n## Title: Bag\n## Version: 1\n"
	if err := os.WriteFile(filepath.Join(addon, "Bag.toc"), []byte(toc), 0o644); err != nil {
		t.Fatal(err)
	}
	return flavor
}

func callRaw(t *testing.T, deps Deps, name string, args any) *mcp.CallToolResult {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(deps, "0.1.0").Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	session, err := mcp.NewClient(&mcp.Implementation{Name: "bornal-test", Version: "0"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func decodeResult(t *testing.T, result *mcp.CallToolResult) Result {
	t.Helper()
	body, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var view Result
	if err := json.Unmarshal(body, &view); err != nil {
		t.Fatal(err)
	}
	return view
}

func TestMacrosApplyAndAddonList(t *testing.T) {
	flavor := writeMacroInstall(t, "enUS")
	deps, getenv := testDeps(t, flavor)
	if err := config.Save(getenv, config.File{Language: "en", InstallPath: flavor}); err != nil {
		t.Fatal(err)
	}
	preview := callRaw(t, deps, "macros_preview", map[string]any{})
	if preview.IsError {
		t.Fatal(textOf(preview))
	}
	view := decodeResult(t, preview)
	if view.Confirm == "" || view.LocaleOK == nil || !*view.LocaleOK {
		t.Fatalf("prévia %+v", view)
	}
	applied := callRaw(t, deps, "macros_apply", map[string]any{"confirm": view.Confirm})
	if applied.IsError {
		t.Fatal(textOf(applied))
	}
	cache, err := os.ReadFile(filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cache), "/startattack") || !strings.Contains(string(cache), "RS_MACRO") {
		t.Fatalf("cache %s", cache)
	}
	listed := callRaw(t, deps, "addons_list", map[string]any{})
	if listed.IsError {
		t.Fatal(textOf(listed))
	}
	addons := decodeResult(t, listed)
	if len(addons.Addons) != 1 || addons.Addons[0].Folder != "Bag" {
		t.Fatalf("addons %+v", addons.Addons)
	}
	body, err := json.Marshal(listed.StructuredContent)
	if err != nil || strings.Contains(string(body), "acct-one") {
		t.Fatalf("addons vazou conta: %s", body)
	}
}

func TestOpenGameAndOtherLocaleDoNotWrite(t *testing.T) {
	flavor := writeMacroInstall(t, "enUS")
	deps, getenv := testDeps(t, flavor)
	deps.Open = func(string) (bool, error) { return true, nil }
	if err := config.Save(getenv, config.File{Language: "en", InstallPath: flavor}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt"))
	if err != nil {
		t.Fatal(err)
	}
	preview := callRaw(t, deps, "macros_preview", map[string]any{})
	view := decodeResult(t, preview)
	if view.Confirm == "" {
		t.Fatal("prévia com jogo aberto")
	}
	applied := callRaw(t, deps, "macros_apply", map[string]any{"confirm": view.Confirm})
	if !applied.IsError || !strings.Contains(textOf(applied), "game is running") {
		t.Fatalf("apply %s", textOf(applied))
	}
	after, err := os.ReadFile(filepath.Join(flavor, "WTF", "Account", "acct-one", "macros-cache.txt"))
	if err != nil || string(after) != string(before) {
		t.Fatal("jogo aberto alterou o arquivo")
	}

	other := writeMacroInstall(t, "ptBR")
	deps, getenv = testDeps(t, other)
	if err := config.Save(getenv, config.File{Language: "en", InstallPath: other}); err != nil {
		t.Fatal(err)
	}
	preview = callRaw(t, deps, "macros_preview", map[string]any{})
	view = decodeResult(t, preview)
	if preview.IsError || view.Confirm != "" || view.LocaleOK == nil || *view.LocaleOK {
		t.Fatalf("locale %+v %s", view, textOf(preview))
	}
}

func TestKitToolsDoNotTakeAPassphraseArgument(t *testing.T) {
	flavor := writeMacroInstall(t, "enUS")
	deps, getenv := testDeps(t, flavor)
	if err := config.Save(getenv, config.File{Language: "en", InstallPath: flavor}); err != nil {
		t.Fatal(err)
	}
	created := callRaw(t, deps, "kit_create", map[string]any{})
	if created.IsError {
		t.Fatal(textOf(created))
	}
	view := decodeResult(t, created)
	if view.Path == "" || strings.Contains(view.Path, "acct-one") {
		t.Fatalf("kit %+v", view)
	}
	raw, err := json.Marshal(created.StructuredContent)
	if err != nil || strings.Contains(string(raw), "hunter2-not-printed") || strings.Contains(string(raw), "test-secret") {
		t.Fatalf("payload %s", raw)
	}
	found := false
	for _, warning := range view.Warnings {
		if strings.Contains(warning.File, "Bag.lua") && warning.Line == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("avisos %+v", view.Warnings)
	}
}

func TestToolSchemasHaveNoPassphrase(t *testing.T) {
	ctx := context.Background()
	deps, _ := testDeps(t, t.TempDir())
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(deps, "0.1.0").Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	session, err := mcp.NewClient(&mcp.Implementation{Name: "bornal-test", Version: "0"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"macros_preview", "macros_apply", "addons_list", "kit_create", "kit_find", "kit_verify", "kit_restore", "rollback"}
	names := map[string]bool{}
	for _, tool := range listed.Tools {
		names[tool.Name] = true
		schema, ok := tool.InputSchema.(map[string]any)
		if !ok {
			t.Fatalf("%s schema %T", tool.Name, tool.InputSchema)
		}
		props, _ := schema["properties"].(map[string]any)
		for key := range props {
			switch strings.ToLower(key) {
			case "password", "passphrase", "secret":
				t.Fatalf("%s tem %s", tool.Name, key)
			}
		}
	}
	for _, name := range want {
		if !names[name] {
			t.Fatalf("falta %s", name)
		}
	}
}
