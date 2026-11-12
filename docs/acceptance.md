# Aceite do corte só de leitura

| Critério | Prova |
| --- | --- |
| `LANG`/`LC_ALL`/`LC_MESSAGES` começando com `pt` escolhe português do Brasil. O resto cai no inglês. `--lang` ganha do ambiente. | `go test ./internal/i18n ./internal/cli` |
| `--path` com um `_forever_/WTF/Config.wtf` mostra flavor, locale e `agentUID`. Sem caminho, o doctor termina com erro e não varre disco. | `go test ./internal/install ./internal/cli` |
| `bornal console` só responde GET de loopback, com `Host` e `Origin` do endereço impresso. | `go test ./internal/localconsole` |
| A porta de qualidade do repositório passa. | `make test` |

Fora deste corte: busca em outros discos, jogo aberto, macro, addon, kit, cifra e MCP stdio.
