# bornal

Lê uma instalação local do World of Warcraft e mostra o resultado numa página em `127.0.0.1`.

Esta versão só lê. Não grava macro e não monta armory. O nome das habilidades, quando uma versão seguinte gravar macros, vem do `textLocale` do cliente. O texto do programa é português do Brasil ou inglês.

```sh
bornal setup
bornal doctor
bornal console
bornal mcp
```

`bornal setup` pergunta o idioma da interface e o caminho da instalação, e grava em `$XDG_CONFIG_HOME/bornal/config.json` (ou `~/.config/bornal/config.json`). A pergunta do caminho faz uma busca rasa: algumas raízes conhecidas, no máximo quatro níveis, pulando nomes pesados como `Data`. `--path` ainda ganha, depois `BORNAL_WOW_PATH`, depois o caminho salvo.

No terminal, `doctor` e `console` fazem as mesmas perguntas quando não há caminho. Sem terminal, eles pedem para rodar `bornal setup`.

`--lang pt-BR` ou `--lang en` ganha daquela execução. O idioma salvo ganha de `LANG`, `LC_ALL` e `LC_MESSAGES`.

`bornal mcp` fala MCP em stdin e stdout. Não pergunta nesse fluxo. `configuration` devolve as perguntas pendentes, e `save_configuration` grava as respostas depois que a pessoa escolheu. Um caminho só fica salvo quando existe `WTF/Config.wtf`.

`bornal console` escuta em `127.0.0.1` numa porta livre e imprime a URL. Ctrl+C encerra. A página só responde GET vindo do loopback.

English: [README.md](README.md).

A porta de qualidade é `make test`.
