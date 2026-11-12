# Aceite

| Critério | Prova |
| --- | --- |
| `LANG`/`LC_ALL`/`LC_MESSAGES` começando com `pt` escolhe português do Brasil. O resto cai no inglês. `--lang` ganha desta execução. O idioma salvo ganha do ambiente. | `go test ./internal/i18n ./internal/cli` |
| `bornal setup` pergunta o idioma e a instalação e grava `config.json` com modo `0600`. Sem terminal, doctor recusa e cita `bornal setup`, sem varrer disco. | `go test ./internal/setup ./internal/config ./internal/cli` |
| A busca da pergunta acha um flavor a até quatro níveis e não entra em `Data` nem em pasta oculta. Os testes usam um diretório temporário. | `go test ./internal/install` |
| `--path` ganha de `BORNAL_WOW_PATH`, que ganha do caminho salvo. Um `_forever_/WTF/Config.wtf` mostra flavor, locale e `agentUID`. | `go test ./internal/install ./internal/cli` |
| `configuration` devolve as perguntas pendentes. `save_configuration` não grava um caminho que não é instalação. O stdio não pergunta. | `go test ./internal/server` |
| `bornal console` só responde GET de loopback, com `Host` e `Origin` do endereço impresso. | `go test ./internal/localconsole` |
| A porta de qualidade do repositório passa. | `make test` |

Fora deste corte: jogo aberto, macro, addon, kit, cifra e escrita no cliente.
