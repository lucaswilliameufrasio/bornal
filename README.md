# bornal

Reads a local World of Warcraft installation and shows it in a loopback page.

This version only reads. It does not write macros or build an armory. Spell names, when a later version writes macros, come from the client's `textLocale`. The program's own text is Brazilian Portuguese or English.

```sh
bornal setup
bornal doctor
bornal console
bornal mcp
```

`bornal setup` asks for the interface language and the install path, then stores them in `$XDG_CONFIG_HOME/bornal/config.json` (or `~/.config/bornal/config.json`). The path question runs a shallow search: a few known roots, at most four directory levels, skipping heavy names such as `Data`. `--path` still wins, then `BORNAL_WOW_PATH`, then the saved path.

On a terminal, `doctor` and `console` ask the same questions when no path is known. Without a terminal they tell you to run `bornal setup`.

`--lang pt-BR` or `--lang en` wins for that run. A saved language wins over `LANG`, `LC_ALL`, and `LC_MESSAGES`.

`bornal mcp` speaks MCP on stdin and stdout. It does not prompt. `configuration` returns the pending questions, and `save_configuration` stores the answers after the person has chosen. A path is kept only when `WTF/Config.wtf` is there.

`bornal console` binds `127.0.0.1` on a free port and prints the URL. Ctrl+C closes it. The page answers GET from loopback only.

Portuguese: [README.pt-BR.md](README.pt-BR.md).

The quality gate is `make test`.
