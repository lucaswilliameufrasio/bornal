# bornal

Reads a local World of Warcraft installation, previews an English warrior macro set, audits addon TOC files, and writes a portable wow-kit.

A write into that install happens only while its `Wow.exe` is closed. A Windows path inside a Wine command line counts as the same install. If the process check cannot run, the write is refused.

```sh
bornal setup
bornal doctor
bornal console
bornal mcp
bornal macros preview
bornal macros apply --confirm TOKEN
bornal addons
bornal kit create
bornal kit find
bornal kit verify PATH
bornal kit restore PATH
bornal rollback
```

`bornal setup` asks for the interface language and the install path, then stores them in `$XDG_CONFIG_HOME/bornal/config.json` (or `~/.config/bornal/config.json`). The file keeps `language` and `install_path` only. The path question runs a shallow search: a few known roots, at most four directory levels, skipping heavy names such as `Data`. `--path` still wins, then `BORNAL_WOW_PATH`, then the saved path.

On a terminal, `doctor` and `console` ask the same questions when no path is known. Without a terminal they tell you to run `bornal setup`.

`--lang pt-BR` or `--lang en` wins for that run. A saved language wins over `LANG`, `LC_ALL`, and `LC_MESSAGES`. Spell names come from the client's `textLocale`. The built-in set is English and is applied only for `enUS`, `enGB`, or `en`.

`bornal doctor` prints flavor, locale, `agentUID`, and whether that install's game is open.

`bornal macros preview` lists create, replace, skip, and conflict. `apply` writes only with the token from that preview. The CLI, the MCP server, and the console do not accept each other's tokens. Other macros stay, including their ids. The same name with a different body waits for `--replace`. The backup beside the file is `macros-cache.bornal-bak`. `rollback` restores that backup, also with the game closed.

`bornal addons` lists TOC title, version, interface, dependencies, and the disabled flag. It does not install addons.

`bornal kit create` copies account macros, bindings, edit files, SavedVariables, character files, an addon index, and a short `portable.wtf`. Graphics options stay on the machine. `--encrypt` asks for a passphrase in the terminal, or in pinentry, twice. It is not a flag, a tool argument, or a URL. An encrypted kit is one file with a clear `WOWKIT1` header: date, label, and kdf. The header has no character or account data. A line that looks like a secret is reported as file and line. `kit find` lists kits without opening them. `kit verify` checks hashes. `kit restore` checks hashes before it writes, refuses an account folder that does not exist, and asks when a character realm differs.

`bornal mcp` speaks MCP on stdin and stdout. It does not prompt on that stream. `configuration` returns the pending questions, and `save_configuration` stores the answers after the person has chosen. The game tools are `macros_preview`, `macros_apply`, `addons_list`, `kit_create`, `kit_find`, `kit_verify`, `kit_restore`, and `rollback`. None of them take a passphrase. When encryption needs a prompt and none is available, use `bornal console`.

`bornal console` binds `127.0.0.1` on a free port and prints the URL. Ctrl+C closes it. The overview answers GET from loopback only. Macros, addons, kits, and rollback accept a same-origin POST. The passphrase is a form field on that page, not a query parameter.

## Not in this version

There is no armory, no CASC or DB2 read, and no network call to name items. Addon code is not written back into the game.

Releases use `vX.Y.Z`. The steps are in [docs/releases.md](docs/releases.md).

Portuguese: [README.pt-BR.md](README.pt-BR.md).

The quality gate is `make test`.
