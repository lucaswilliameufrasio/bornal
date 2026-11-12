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
  bornal doctor [--path DIR] [--lang pt-BR|en]
  bornal console [--path DIR] [--lang pt-BR|en]
  bornal help
  bornal version

--lang ganha de LANG, LC_ALL e LC_MESSAGES.
Sem --path, doctor e console usam BORNAL_WOW_PATH.
Este corte só lê. Não varre outros discos.
`,
		`bornal — reads local World of Warcraft installations

Usage:
  bornal doctor [--path DIR] [--lang pt-BR|en]
  bornal console [--path DIR] [--lang pt-BR|en]
  bornal help
  bornal version

--lang wins over LANG, LC_ALL, and LC_MESSAGES.
Without --path, doctor and console use BORNAL_WOW_PATH.
This slice only reads. It does not scan other disks.
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

// NeedPath is used when neither --path nor the environment variable is set.
func NeedPath(lang Language) string {
	return text(lang,
		"Nenhum caminho informado. Use --path ou BORNAL_WOW_PATH.",
		"No path given. Use --path or BORNAL_WOW_PATH.")
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
