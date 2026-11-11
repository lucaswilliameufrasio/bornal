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
  bornal macros preview [--path DIR] [--flavor NAME] [--account N] [--replace]
  bornal macros apply --confirm TOKEN
  bornal addons [--path DIR] [--flavor NAME]
  bornal kit create [--path DIR] [--flavor NAME] [--account N] [--out DIR] [--label TEXT] [--encrypt]
  bornal kit find [DIR]
  bornal kit verify PATH
  bornal kit restore PATH [--confirm TOKEN] [--replace] [--realm-ok]
  bornal rollback [--confirm TOKEN] [--path DIR] [--flavor NAME] [--account N]
  bornal help
  bornal version

--lang ganha desta execução. O idioma salvo ganha de LANG, LC_ALL e LC_MESSAGES.
O caminho segue --path, depois BORNAL_WOW_PATH, depois o config salvo.
No terminal, doctor e console perguntam quando não há caminho.
bornal mcp não pergunta neste fluxo. A ferramenta configuration devolve as perguntas.
A macro e o restore gravam só com o jogo fechado e com o token desta prévia.
A senha de um kit cifrado não entra em flag, argumento nem URL.
`,
		`bornal — reads local World of Warcraft installations

Usage:
  bornal setup
  bornal doctor [--path DIR] [--lang pt-BR|en]
  bornal console [--path DIR] [--lang pt-BR|en]
  bornal mcp [--lang pt-BR|en]
  bornal macros preview [--path DIR] [--flavor NAME] [--account N] [--replace]
  bornal macros apply --confirm TOKEN
  bornal addons [--path DIR] [--flavor NAME]
  bornal kit create [--path DIR] [--flavor NAME] [--account N] [--out DIR] [--label TEXT] [--encrypt]
  bornal kit find [DIR]
  bornal kit verify PATH
  bornal kit restore PATH [--confirm TOKEN] [--replace] [--realm-ok]
  bornal rollback [--confirm TOKEN] [--path DIR] [--flavor NAME] [--account N]
  bornal help
  bornal version

--lang wins for this run. A saved language wins over LANG, LC_ALL, and LC_MESSAGES.
The path follows --path, then BORNAL_WOW_PATH, then the saved config.
On a terminal, doctor and console ask when no path is known.
bornal mcp does not prompt on this stream. The configuration tool returns the questions.
Macros and restore write only while the game is closed, and only with that preview's token.
An encrypted kit passphrase is not a flag, an argument, or a URL.
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

// RunningUnchecked is used when the process scan cannot run.
func RunningUnchecked(lang Language) string {
	return text(lang, "Jogo aberto: não verificado", "Game running: not checked")
}

// RunningYes reports that Wow.exe for this install is open.
func RunningYes(lang Language) string {
	return text(lang, "Jogo aberto: sim", "Game running: yes")
}

// RunningNo reports that Wow.exe for this install is closed.
func RunningNo(lang Language) string {
	return text(lang, "Jogo aberto: não", "Game running: no")
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

// NavMacros is the macro page.
func NavMacros(lang Language) string { return text(lang, "Macros", "Macros") }

// NavAddons is the addon page.
func NavAddons(lang Language) string { return text(lang, "Addons", "Addons") }

// NavKits is the kit page.
func NavKits(lang Language) string { return text(lang, "Kits", "Kits") }

// NavRollback is the rollback page.
func NavRollback(lang Language) string { return text(lang, "Rollback", "Rollback") }

// NeedAccount asks for --account when several account folders exist.
func NeedAccount(lang Language) string {
	return text(lang, "Escolha a conta com --account N.", "Choose the account with --account N.")
}

// NeedFlavor asks for --flavor when several installs exist.
func NeedFlavor(lang Language) string {
	return text(lang, "Escolha o flavor com --flavor.", "Choose the flavor with --flavor.")
}

// MacroLocale refuses the English set on another client locale.
func MacroLocale(lang Language) string {
	return text(lang,
		"O conjunto em inglês não entra neste locale. A prévia mostra o texto, sem gravar.",
		"The English set is not applied on this locale. The preview shows the text and does not write.")
}

// MacroConflict asks for a new preview with --replace.
func MacroConflict(lang Language) string {
	return text(lang,
		"Uma macro com o mesmo nome tem outro texto. Rode de novo com --replace para trocar.",
		"A macro with the same name has different text. Preview again with --replace to change it.")
}

// MacroWrote confirms the cache was replaced.
func MacroWrote(lang Language) string {
	return text(lang, "Macros gravadas. O backup é macros-cache.bornal-bak.", "Macros written. The backup is macros-cache.bornal-bak.")
}

// MacroReady prints the confirm token label.
func MacroReady(lang Language) string {
	return text(lang, "confirm", "confirm")
}

// KitWrote confirms a kit path.
func KitWrote(lang Language) string {
	return text(lang, "Kit gravado.", "Kit written.")
}

// KitPass says the passphrase did not open the kit.
func KitPass(lang Language) string {
	return text(lang, "A senha não confere. Nada foi gravado.", "The passphrase does not match. Nothing was written.")
}

// KitNeedPass asks for a passphrase without accepting it as a flag.
func KitNeedPass(lang Language) string {
	return text(lang, "Este kit é cifrado. A senha é pedida no terminal ou no console.", "This kit is encrypted. The passphrase is asked in the terminal or the console.")
}

// UseConsole tells an MCP client where a passphrase can be typed.
func UseConsole(lang Language) string {
	return text(lang, "Sem prompt local. Use bornal console para cifrar.", "No local prompt. Use bornal console to encrypt.")
}

// Restored confirms a kit restore.
func Restored(lang Language) string {
	return text(lang, "Kit restaurado.", "Kit restored.")
}

// RolledBack confirms a macro rollback.
func RolledBack(lang Language) string {
	return text(lang, "Backup das macros restaurado.", "Macro backup restored.")
}

// SubmitApply is a console button.
func SubmitApply(lang Language) string { return text(lang, "Gravar macros", "Write macros") }

// SubmitKit is a console button.
func SubmitKit(lang Language) string { return text(lang, "Criar kit", "Create kit") }

// SubmitRestore is a console button.
func SubmitRestore(lang Language) string {
	return text(lang, "Restaurar kit", "Restore kit")
}

// SubmitRollback is a console button.
func SubmitRollback(lang Language) string {
	return text(lang, "Restaurar backup", "Restore backup")
}

// SubmitVerify is a console button.
func SubmitVerify(lang Language) string {
	return text(lang, "Conferir kit", "Check kit")
}

// NeedConfirm asks for a token from the same side.
func NeedConfirm(lang Language) string {
	return text(lang, "Rode a prévia e use o confirm desta mesma via.", "Run the preview and pass confirm from this same side.")
}
