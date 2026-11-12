# bornal

Lê uma instalação local do World of Warcraft e mostra o resultado numa página em `127.0.0.1`.

Esta versão só lê. Não varre outros discos, não grava macro e não monta armory. O nome das habilidades, quando uma versão seguinte gravar macros, vem do `textLocale` do cliente. O texto do programa é português do Brasil ou inglês.

```sh
bornal doctor --path "/caminho/do/World of Warcraft"
bornal console --path "/caminho/do/World of Warcraft"
```

`--lang pt-BR` ou `--lang en` ganha de `LANG`, `LC_ALL` e `LC_MESSAGES`. Sem `--path`, os dois comandos usam `BORNAL_WOW_PATH`.

`bornal console` escuta em `127.0.0.1` numa porta livre e imprime a URL. Ctrl+C encerra. A página só responde GET vindo do loopback.

English: [README.md](README.md).

A porta de qualidade é `make test`.
