// Package i18n chooses Brazilian Portuguese or English for the program's own text.
// Spell names are not translated here.
package i18n

import "strings"

// Language is the interface language, not the WoW client locale.
type Language int

const (
	// English is the fallback for anyone who does not read Portuguese.
	English Language = iota
	// Portuguese is Brazilian Portuguese, including for a pt_PT environment.
	Portuguese
)

// Detect follows acari. The first of LANG, LC_ALL, and LC_MESSAGES that
// starts with "pt" selects Portuguese. Any other value is skipped. When none
// matches, the language is English.
func Detect(getenv func(string) string) Language {
	if getenv == nil {
		return English
	}
	for _, key := range []string{"LANG", "LC_ALL", "LC_MESSAGES"} {
		value := strings.ToLower(strings.TrimSpace(getenv(key)))
		if strings.HasPrefix(value, "pt") {
			return Portuguese
		}
	}
	return English
}

// Parse overrides detection. Only pt-BR and en are accepted.
func Parse(value string) (Language, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pt-br", "pt_br", "pt":
		return Portuguese, true
	case "en", "en-us", "en_us":
		return English, true
	default:
		return English, false
	}
}

// Code is the value stored in the config file.
func Code(lang Language) string {
	if lang == Portuguese {
		return "pt-BR"
	}
	return "en"
}

func text(lang Language, portuguese, english string) string {
	if lang == Portuguese {
		return portuguese
	}
	return english
}

// Usage is the command list.
func Usage(lang Language) string {
	return text(lang,
		`bornal — lê instalações locais do World of Warcraft

Uso:
  bornal setup
  bornal doctor [--path DIR] [--lang pt-BR|en]
  bornal console [--path DIR] [--lang pt-BR|en]
  bornal mcp [--lang pt-BR|en]
  bornal help
  bornal version

--lang ganha desta execução. O idioma salvo ganha de LANG, LC_ALL e LC_MESSAGES.
O caminho segue --path, depois BORNAL_WOW_PATH, depois o config salvo.
No terminal, doctor e console perguntam quando não há caminho.
bornal mcp não pergunta neste fluxo. A ferramenta configuration devolve as perguntas.
Este corte só lê.
`,
		`bornal — reads local World of Warcraft installations

Usage:
  bornal setup
  bornal doctor [--path DIR] [--lang pt-BR|en]
  bornal console [--path DIR] [--lang pt-BR|en]
  bornal mcp [--lang pt-BR|en]
  bornal help
  bornal version

--lang wins for this run. A saved language wins over LANG, LC_ALL, and LC_MESSAGES.
The path follows --path, then BORNAL_WOW_PATH, then the saved config.
On a terminal, doctor and console ask when no path is known.
bornal mcp does not prompt on this stream. The configuration tool returns the questions.
This slice only reads.
`)
}

// UnknownCommand reports a command this binary does not have.
func UnknownCommand(lang Language, name string) string {
	return text(lang, "comando desconhecido: "+name, "unknown command: "+name)
}

// BadLang reports a --lang value outside the two supported choices.
func BadLang(lang Language, value string) string {
	return text(lang, "--lang inválido: "+value, "invalid --lang: "+value)
}

// NeedPath is used when no path is saved and this process has no terminal.
func NeedPath(lang Language) string {
	return text(lang,
		"Nenhum caminho salvo. Rode bornal setup, ou use --path ou BORNAL_WOW_PATH.",
		"No saved path. Run bornal setup, or pass --path or BORNAL_WOW_PATH.")
}

// NoneFound is used when the path contains no install.
func NoneFound(lang Language) string {
	return text(lang,
		"Nenhuma instalação nesse caminho.",
		"No installation at that path.")
}

// ReadFailed introduces a path that could not be read. The detail stays off the console page.
func ReadFailed(lang Language) string {
	return text(lang, "Não consegui ler o caminho.", "Could not read the path.")
}

// RunningUnchecked says this slice does not look at Wow.exe.
func RunningUnchecked(lang Language) string {
	return text(lang,
		"Jogo aberto: não verificado neste corte.",
		"Game running: not checked in this slice.")
}

// LocaleNote tells the reader that UI language and client locale differ on purpose.
func LocaleNote(lang Language) string {
	return text(lang,
		"O texto desta página segue o idioma do sistema. O nome das habilidades, quando existir, segue o locale do cliente.",
		"This page follows the system language. Spell names, when they exist, follow the client locale.")
}

// ConsoleOpened is printed before the URL.
func ConsoleOpened(lang Language) string {
	return text(lang, "console local aberto; Ctrl+C encerra", "local console open; Ctrl+C closes it")
}

// Brand subtitle under the wordmark.
func BrandSubtitle(lang Language) string {
	return text(lang, "instalações locais", "local installs")
}

// NavInstalls is the only page in this slice.
func NavInstalls(lang Language) string {
	return text(lang, "Instalações", "Installs")
}

// ModeLocal is the rail footer.
func ModeLocal(lang Language) string {
	return text(lang, "Modo local", "Local mode")
}

// Field labels.
func LabelPath(lang Language) string { return text(lang, "Caminho", "Path") }
func LabelFlavor(lang Language) string {
	return text(lang, "Flavor", "Flavor")
}
func LabelLocale(lang Language) string {
	return text(lang, "Locale do cliente", "Client locale")
}
func LabelAgent(lang Language) string { return text(lang, "agentUID", "agentUID") }
func LabelMissing(lang Language) string {
	return text(lang, "ausente", "missing")
}

// LanguageQuestion is bilingual because the language is not chosen yet.
func LanguageQuestion() string {
	return "Idioma da interface / Interface language\n  1  pt-BR\n  2  English\n"
}

// InstallQuestion introduces the numbered installs.
func InstallQuestion(lang Language) string {
	return text(lang,
		"Qual instalação do World of Warcraft?",
		"Which World of Warcraft installation?")
}

// NoInstalls introduces a typed path when the search found nothing.
func NoInstalls(lang Language) string {
	return text(lang,
		"Nenhuma instalação encontrada nesta busca.\nDigite o caminho da pasta do jogo.",
		"No installation found in this search.\nType the path of the game folder.")
}

// OtherPath is the last numbered choice.
func OtherPath(lang Language) string {
	return text(lang, "outro caminho", "another path")
}

// TypePath asks for a path after "another path".
func TypePath(lang Language) string {
	return text(lang, "Digite o caminho da pasta do jogo.", "Type the path of the game folder.")
}

// InvalidChoice is printed before the question is repeated.
func InvalidChoice(lang Language) string {
	return text(lang, "Não entendi. Tente de novo.", "Did not understand. Try again.")
}

// NotAnInstall rejects a path without WTF/Config.wtf.
func NotAnInstall(lang Language) string {
	return text(lang,
		"Esse caminho não tem uma instalação.",
		"That path has no installation.")
}

// SavedConfig confirms the file was written.
func SavedConfig(lang Language) string {
	return text(lang, "Configuração salva.", "Configuration saved.")
}

// SetupFailed is a question that ended without an answer.
func SetupFailed(lang Language) string {
	return text(lang, "Configuração interrompida.", "Setup stopped.")
}

// ConfigFailed hides the filesystem error behind the config file.
func ConfigFailed(lang Language) string {
	return text(lang, "Não consegui gravar a configuração.", "Could not write the configuration.")
}
