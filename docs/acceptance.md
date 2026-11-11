# Aceite

| Critério | Prova |
| --- | --- |
| `LANG`/`LC_ALL`/`LC_MESSAGES` começando com `pt` escolhe português do Brasil. O resto cai no inglês. `--lang` ganha desta execução. O idioma salvo ganha do ambiente. | `go test ./internal/i18n ./internal/cli` |
| `bornal setup` pergunta o idioma e a instalação e grava `config.json` com modo `0600`. Sem terminal, doctor recusa e cita `bornal setup`, sem varrer disco. | `go test ./internal/setup ./internal/config ./internal/cli` |
| A busca da pergunta acha um flavor a até quatro níveis e não entra em `Data` nem em pasta oculta. Os testes usam um diretório temporário. | `go test ./internal/install` |
| `--path` ganha de `BORNAL_WOW_PATH`, que ganha do caminho salvo. Um `_forever_/WTF/Config.wtf` mostra flavor, locale e `agentUID`. | `go test ./internal/install ./internal/cli` |
| `configuration` devolve as perguntas pendentes. `save_configuration` não grava um caminho que não é instalação. O stdio não pergunta. | `go test ./internal/server` |
| `bornal console` só responde GET de loopback na visão geral, com `Host` e `Origin` do endereço impresso. As outras páginas aceitam POST da mesma origem, recusam query string e não devolvem a senha. | `go test ./internal/localconsole` |
| `doctor` diz se o jogo daquela instalação está aberto. A escrita é recusada com o jogo aberto ou quando a varredura falha. | `go test ./internal/cli ./internal/running ./internal/work` |
| A prévia das macros em inglês usa `/startattack`. Outro texto no mesmo nome é conflito até `--replace`. O locale que não é inglês não grava. O token da CLI não vale no console. | `go test ./internal/macro ./internal/cli ./internal/work` |
| O kit confere o hash antes de gravar. Senha errada ou arquivo adulterado não muda o destino. A senha não é campo de ferramenta nem flag. | `go test ./internal/kit ./internal/server ./internal/cli` |
| A porta de qualidade do repositório passa. | `make test` |

Fora desta versão: armory, leitura de CASC ou DB2, e chamada de rede para nomear itens. O código de addon não volta para o jogo.
