# bornal

bornal looks after the macros and settings of a World of Warcraft install on this machine. It shows what a write would change, then writes only while that install's `Wow.exe` is closed.

```sh
bornal setup
bornal doctor
bornal console
```

`setup` asks for the interface language and the install path. It stores `language` and `install_path` in `$XDG_CONFIG_HOME/bornal/config.json`, or `~/.config/bornal/config.json`. Nothing else goes in that file.

`doctor` prints the flavor, the client locale, the `agentUID`, and whether that install's game is open.

The path comes from `--path`, then `BORNAL_WOW_PATH`, then the saved path. On a terminal, `doctor` and `console` ask when no path is known. The search stays shallow: a few known roots, four directory levels, and it skips heavy names such as `Data`. Without a terminal, run `bornal setup`. One install is used as it is. Several installs, or several account folders, are a question, not a guess.

`--lang pt-BR` or `--lang en` wins for that run. A saved language wins over `LANG`, `LC_ALL`, and `LC_MESSAGES`. Spell names come from the client's `textLocale`.

## Macros

The built-in set is an English warrior. It is written only when the client locale is `enUS`, `enGB`, or `en`.

```sh
bornal macros preview [--path DIR] [--flavor NAME] [--account N] [--replace]
bornal macros apply --confirm TOKEN
bornal rollback [--confirm TOKEN]
```

`preview` lists each macro as create, replace, skip, or conflict. `apply` writes with the token from that preview. The CLI, the MCP server, and the console do not accept each other's tokens. A macro with the same name and a different body waits for `--replace`. Every other macro stays, including its id. The backup beside the file is `macros-cache.bornal-bak`. `rollback` restores that backup, again with the game closed.

A Windows path inside a Wine command line counts as the same install. If the process check cannot run, the write is refused.

## Addons

```sh
bornal addons [--path DIR] [--flavor NAME]
```

This reads each TOC: title, version, interface, dependencies, and the disabled flag. It does not install addons, and it does not write addon code back into the game.

## Kits

A wow-kit is a directory ending in `.wowkit`, or one encrypted file.

```sh
bornal kit create [--out DIR] [--label TEXT] [--encrypt]
bornal kit find [DIR]
bornal kit verify PATH
bornal kit restore PATH [--confirm TOKEN] [--replace] [--realm-ok]
```

`create` copies account macros, bindings, edit files, SavedVariables, character files, an addon index, and a short `portable.wtf`. Graphics options stay on the machine. `--encrypt` asks for the passphrase twice, in the terminal or in pinentry. It is not a flag value, a tool argument, or a URL. An encrypted kit is one file with a clear `WOWKIT1` header: date, label, and kdf. The header has no character or account data. A line that looks like a secret is reported as a file and a line number.

`find` lists kits without opening them. `verify` checks hashes. `restore` checks hashes before it writes, refuses an account folder that does not exist, and asks when a character's realm differs.

## MCP and the console

```sh
bornal mcp
bornal console
```

`bornal mcp` speaks MCP on stdin and stdout. It does not prompt on that stream. `configuration` returns the pending questions, and `save_configuration` stores the answers. The game tools are `macros_preview`, `macros_apply`, `addons_list`, `kit_create`, `kit_find`, `kit_verify`, `kit_restore`, and `rollback`. None of them take a passphrase. When encryption needs a prompt, use `bornal console`.

`bornal console` listens on `127.0.0.1` and prints the URL. Ctrl+C closes it. The page shows the install, the macro preview, the addon audit, the kits, and the rollback. The overview answers GET from loopback only. The other pages accept a same-origin POST. The passphrase is a form field, not a query parameter.

## Not in this version

There is no armory, no CASC or DB2 read, and no network call to name items.

Releases use `vX.Y.Z`. The steps are in [docs/releases.md](docs/releases.md).

Portuguese: [README.pt-BR.md](README.pt-BR.md).

The quality gate is `make test`.
