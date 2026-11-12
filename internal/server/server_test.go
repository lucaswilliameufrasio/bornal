package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func writeInstall(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "_forever_", "WTF")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("SET textLocale \"enUS\"\nSET agentUID \"wowf\"\n")
	if err := os.WriteFile(filepath.Join(dir, "Config.wtf"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(dir)
}

func testDeps(t *testing.T, flavor string) (Deps, func(string) string) {
	t.Helper()
	home := t.TempDir()
	getenv := func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return home
		}
		return ""
	}
	deps := Deps{
		Getenv:   getenv,
		Language: i18n.English,
		Discover: func([]string) []install.Install {
			return []install.Install{{Path: flavor, Flavor: "_forever_", Locale: "enUS"}}
		},
		Roots:   func(func(string) string) []string { return nil },
		Inspect: install.Inspect,
	}
	return deps, getenv
}

func call(t *testing.T, deps Deps, name string, args any) View {
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
	if result.IsError {
		t.Fatalf("ferramenta %s: %s", name, textOf(result))
	}
	return decode(t, result)
}

func callError(t *testing.T, deps Deps, name string, args any) string {
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
	if !result.IsError {
		t.Fatal("esperava erro de ferramenta")
	}
	return textOf(result)
}

func TestConfigurationListsPendingQuestions(t *testing.T) {
	flavor := writeInstall(t)
	deps, _ := testDeps(t, flavor)
	view := call(t, deps, "configuration", map[string]any{})
	if !contains(view.Pending, "language") || !contains(view.Pending, "install_path") {
		t.Fatalf("pending %+v", view.Pending)
	}
	if !hasQuestion(view, "language") || !hasQuestion(view, "install_path") {
		t.Fatalf("perguntas %+v", view.Questions)
	}
	installQ := question(view, "install_path")
	if installQ.Default != flavor || !hasChoice(installQ, flavor) || !hasChoice(installQ, "other") {
		t.Fatalf("escolhas %+v", installQ)
	}
	if question(view, "language").Default != "en" {
		t.Fatalf("idioma %+v", question(view, "language"))
	}
}

func TestSaveConfigurationKeepsOnlyARealInstall(t *testing.T) {
	flavor := writeInstall(t)
	deps, getenv := testDeps(t, flavor)
	bad := callError(t, deps, "save_configuration", map[string]any{"install_path": filepath.Join(t.TempDir(), "vazio")})
	if !strings.Contains(bad, "not a World of Warcraft directory") {
		t.Fatalf("erro: %s", bad)
	}
	stored, err := config.Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if stored != (config.File{}) {
		t.Fatalf("gravou %+v", stored)
	}

	lang := "pt-BR"
	view := call(t, deps, "save_configuration", map[string]any{"language": lang})
	if contains(view.Pending, "language") || !contains(view.Pending, "install_path") {
		t.Fatalf("pending %+v", view.Pending)
	}
	done := call(t, deps, "save_configuration", map[string]any{"install_path": flavor})
	if len(done.Pending) != 0 || done.Language != "pt-BR" || done.InstallPath != flavor {
		t.Fatalf("view %+v", done)
	}
	stored, err = config.Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Language != "pt-BR" || stored.InstallPath != flavor {
		t.Fatalf("arquivo %+v", stored)
	}
}

func TestToolDescriptionsAreEnglish(t *testing.T) {
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
	names := map[string]string{}
	for _, tool := range listed.Tools {
		names[tool.Name] = tool.Description
	}
	for _, name := range []string{"configuration", "save_configuration"} {
		text := names[name]
		if text == "" || strings.Contains(strings.ToLower(text), "pergunt") {
			t.Fatalf("%s descrição %q", name, text)
		}
	}
}

func decode(t *testing.T, result *mcp.CallToolResult) View {
	t.Helper()
	body, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var view View
	if err := json.Unmarshal(body, &view); err != nil {
		t.Fatal(err)
	}
	return view
}

func textOf(result *mcp.CallToolResult) string {
	parts := make([]string, 0)
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func hasQuestion(view View, id string) bool {
	return question(view, id).ID == id
}

func question(view View, id string) Question {
	for _, item := range view.Questions {
		if item.ID == id {
			return item
		}
	}
	return Question{}
}

func hasChoice(item Question, id string) bool {
	for _, choice := range item.Choices {
		if choice.ID == id {
			return true
		}
	}
	return false
}
