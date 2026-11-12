# bornal

Reads a local World of Warcraft installation and shows it on a loopback page.

This version asks where the game is, reports that install, and stores only its own configuration. It does not write the game.

```sh
bornal setup
bornal doctor
bornal console
bornal mcp
```

`bornal setup` asks for the interface language and the install path, then stores them in `$XDG_CONFIG_HOME/bornal/config.json` (or `~/.config/bornal/config.json`). The path question runs a shallow search: a few known roots, at most four directory levels, skipping heavy names such as `Data`. `--path` still wins, then `BORNAL_WOW_PATH`, then the saved path.

On a terminal, `doctor` and `console` ask the same questions when no path is known. Without a terminal they tell you to run `bornal setup`.

`--lang pt-BR` or `--lang en` wins for that run. A saved language wins over `LANG`, `LC_ALL`, and `LC_MESSAGES`. Spell names, when a later version writes macros, come from the client's `textLocale`.

`bornal mcp` speaks MCP on stdin and stdout. It does not prompt. `configuration` returns the pending questions, and `save_configuration` stores the answers after the person has chosen. A path is kept only when `WTF/Config.wtf` is there.

`bornal console` binds `127.0.0.1` on a free port and prints the URL. Ctrl+C closes it. The page answers GET from loopback only.

## Not in this version

Macros, addon audit, backup kits, restore, encryption, an armory, and a check for whether the game is open were designed and are not in the binary yet.

Releases use `vX.Y.Z`. The steps are in [docs/releases.md](docs/releases.md).

Portuguese: [README.pt-BR.md](README.pt-BR.md).

The quality gate is `make test`.
