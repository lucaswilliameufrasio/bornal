# bornal

Reads a local World of Warcraft installation and shows it in a loopback page.

This version only reads. It does not scan other disks, write macros, or build an armory. Spell names, when a later version writes macros, come from the client's `textLocale`. The program's own text is Brazilian Portuguese or English.

```sh
bornal doctor --path "/path/to/World of Warcraft"
bornal console --path "/path/to/World of Warcraft"
```

`--lang pt-BR` or `--lang en` overrides `LANG`, `LC_ALL`, and `LC_MESSAGES`. Without `--path`, both commands use `BORNAL_WOW_PATH`.

`bornal console` binds `127.0.0.1` on a free port and prints the URL. Ctrl+C closes it. The page answers GET from loopback only.

Portuguese: [README.pt-BR.md](README.pt-BR.md).

The quality gate is `make test`.
