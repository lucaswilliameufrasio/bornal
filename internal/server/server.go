// Package server exposes the read-only configuration tools on stdio.
// The process does not prompt. The agent asks the person, then calls save_configuration.
package server

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucaswilliameufrasio/bornal/internal/config"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Deps are the files and the search behind the tools.
type Deps struct {
	Getenv         func(string) string
	Discover       func([]string) []install.Install
	Roots          func(func(string) string) []string
	Inspect        func(string) ([]install.Install, error)
	Language       i18n.Language
	LanguageLocked bool
}

// Choice is one answer id the agent can send back.
type Choice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Question is one pending first-run question.
type Question struct {
	ID      string   `json:"id"`
	Prompt  string   `json:"prompt"`
	Choices []Choice `json:"choices"`
	Default string   `json:"default,omitempty"`
}

// View is the saved setup plus the questions still open.
type View struct {
	Language    string     `json:"language,omitempty"`
	InstallPath string     `json:"install_path,omitempty"`
	Pending     []string   `json:"pending"`
	Questions   []Question `json:"questions"`
}

type saveIn struct {
	Language    *string `json:"language,omitempty"`
	InstallPath *string `json:"install_path,omitempty"`
}

// New builds the stdio server. version is the implementation version.
func New(deps Deps, version string) *mcp.Server {
	if version == "" {
		version = "0.1.0"
	}
	readOnly := true
	writable := false
	local := false
	srv := mcp.NewServer(&mcp.Implementation{Name: "bornal", Version: version}, nil)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "configuration",
		Description: "Read the saved interface language and World of Warcraft install path. When either is missing, returns questions for the person. Does not prompt on this stream and does not write the game.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    readOnly,
			DestructiveHint: &writable,
			OpenWorldHint:   &local,
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, View, error) {
		view, err := deps.view()
		if err != nil {
			return nil, View{}, err
		}
		return nil, view, nil
	})
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "save_configuration",
		Description: "Store the interface language and install path after the person has answered. language is pt-BR or en. install_path is kept only when it contains a World of Warcraft Config.wtf. Does not write the game.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &writable,
			OpenWorldHint:   &local,
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in saveIn) (*mcp.CallToolResult, View, error) {
		view, err := deps.save(in)
		if err != nil {
			return nil, View{}, err
		}
		return nil, view, nil
	})
	return srv
}

// Run serves MCP on the process stdin and stdout until ctx ends.
func Run(ctx context.Context, deps Deps, version string) error {
	return New(deps, version).Run(ctx, &mcp.StdioTransport{})
}

func (d Deps) view() (View, error) {
	file, err := config.Load(d.getenv())
	if err != nil {
		return View{}, errors.New("could not read configuration")
	}
	view := View{Pending: []string{}, Questions: []Question{}}
	textLang := d.Language
	if parsed, ok := i18n.Parse(file.Language); ok {
		view.Language = i18n.Code(parsed)
		if !d.LanguageLocked {
			textLang = parsed
		}
	} else {
		view.Pending = append(view.Pending, "language")
		view.Questions = append(view.Questions, languageQuestion(textLang))
	}
	if path, ok := d.knownInstall(file.InstallPath); ok {
		view.InstallPath = path
	} else {
		view.Pending = append(view.Pending, "install_path")
		view.Questions = append(view.Questions, d.installQuestion(textLang))
	}
	return view, nil
}

func (d Deps) save(in saveIn) (View, error) {
	if in.Language == nil && in.InstallPath == nil {
		return View{}, errors.New("nothing to save")
	}
	file, err := config.Load(d.getenv())
	if err != nil {
		return View{}, errors.New("could not read configuration")
	}
	if in.Language != nil {
		parsed, ok := i18n.Parse(*in.Language)
		if !ok {
			return View{}, errors.New("language must be pt-BR or en")
		}
		file.Language = i18n.Code(parsed)
	}
	if in.InstallPath != nil {
		path := strings.TrimSpace(*in.InstallPath)
		if path == "" {
			return View{}, errors.New("install_path is empty")
		}
		cleaned, ok := d.knownInstall(path)
		if !ok {
			return View{}, errors.New("install_path is not a World of Warcraft directory")
		}
		file.InstallPath = cleaned
	}
	if err := config.Save(d.getenv(), file); err != nil {
		return View{}, errors.New("could not save configuration")
	}
	return d.view()
}

func (d Deps) knownInstall(path string) (string, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false
	}
	inspect := d.Inspect
	if inspect == nil {
		inspect = install.Inspect
	}
	found, err := inspect(path)
	if err != nil || len(found) == 0 {
		return "", false
	}
	return filepath.Clean(path), true
}

func (d Deps) installQuestion(lang i18n.Language) Question {
	found := d.search()
	choices := make([]Choice, 0, len(found)+1)
	for _, one := range found {
		label := one.Path
		if one.Flavor != "" {
			label = one.Path + " (" + one.Flavor + ")"
		}
		choices = append(choices, Choice{ID: one.Path, Label: label})
	}
	choices = append(choices, Choice{ID: "other", Label: i18n.OtherPath(lang)})
	prompt := i18n.InstallQuestion(lang)
	if len(found) == 0 {
		prompt = i18n.NoInstalls(lang)
	}
	question := Question{
		ID:      "install_path",
		Prompt:  prompt,
		Choices: choices,
	}
	if len(found) == 1 {
		question.Default = found[0].Path
	}
	return question
}

func (d Deps) search() []install.Install {
	discover := d.Discover
	if discover == nil {
		discover = install.Discover
	}
	rootsFn := d.Roots
	if rootsFn == nil {
		rootsFn = install.DefaultRoots
	}
	found := discover(rootsFn(d.getenv()))
	sort.Slice(found, func(i, j int) bool {
		return found[i].Path < found[j].Path
	})
	return found
}

func (d Deps) getenv() func(string) string {
	if d.Getenv == nil {
		return func(string) string { return "" }
	}
	return d.Getenv
}

func languageQuestion(lang i18n.Language) Question {
	return Question{
		ID:     "language",
		Prompt: strings.TrimSpace(i18n.LanguageQuestion()),
		Choices: []Choice{
			{ID: "pt-BR", Label: "português do Brasil"},
			{ID: "en", Label: "English"},
		},
		Default: i18n.Code(lang),
	}
}
