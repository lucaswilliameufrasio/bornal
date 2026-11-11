# bornal

Lê uma instalação local do World of Warcraft, mostra um conjunto de macros de guerreiro em inglês, audita o TOC dos addons e grava um wow-kit portátil.

Uma escrita nessa instalação só acontece com o `Wow.exe` dela fechado. Um caminho do Windows na linha de comando do Wine conta como a mesma instalação. Se a checagem do processo não roda, a escrita é recusada.

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

`bornal setup` pergunta o idioma da interface e o caminho da instalação, e grava em `$XDG_CONFIG_HOME/bornal/config.json` (ou `~/.config/bornal/config.json`). O arquivo guarda só `language` e `install_path`. A pergunta do caminho faz uma busca rasa: algumas raízes conhecidas, no máximo quatro níveis, pulando nomes pesados como `Data`. `--path` ainda ganha, depois `BORNAL_WOW_PATH`, depois o caminho salvo.

No terminal, `doctor` e `console` fazem as mesmas perguntas quando não há caminho. Sem terminal, eles pedem para rodar `bornal setup`.

`--lang pt-BR` ou `--lang en` ganha daquela execução. O idioma salvo ganha de `LANG`, `LC_ALL` e `LC_MESSAGES`. O nome das habilidades vem do `textLocale` do cliente. O conjunto embutido está em inglês e só é gravado em `enUS`, `enGB` ou `en`.

`bornal doctor` mostra flavor, locale, `agentUID` e se o jogo daquela instalação está aberto.

`bornal macros preview` lista create, replace, skip e conflict. `apply` grava só com o token daquela prévia. A CLI, o servidor MCP e o console não aceitam o token um do outro. As outras macros ficam, com os ids delas. O mesmo nome com outro texto espera `--replace`. O backup ao lado do arquivo é `macros-cache.bornal-bak`. `rollback` restaura esse backup, também com o jogo fechado.

`bornal addons` lista título, versão, interface, dependências e o flag de desativado do TOC. Não instala addon.

`bornal kit create` copia macros da conta, bindings, arquivos de edição, SavedVariables, arquivos de personagem, um índice de addons e um `portable.wtf` curto. Opções gráficas ficam na máquina. `--encrypt` pede a senha no terminal, ou no pinentry, duas vezes. Ela não é flag, argumento de ferramenta nem URL. Um kit cifrado é um arquivo com cabeçalho claro `WOWKIT1`: data, rótulo e kdf. O cabeçalho não tem personagem nem conta. Uma linha com cara de segredo aparece como arquivo e número da linha. `kit find` lista kits sem abri-los. `kit verify` confere os hashes. `kit restore` confere os hashes antes de gravar, recusa uma pasta de conta que não existe e pergunta quando o reino do personagem difere.

`bornal mcp` fala MCP em stdin e stdout. Não pergunta nesse fluxo. `configuration` devolve as perguntas pendentes, e `save_configuration` grava as respostas depois que a pessoa escolheu. As ferramentas do jogo são `macros_preview`, `macros_apply`, `addons_list`, `kit_create`, `kit_find`, `kit_verify`, `kit_restore` e `rollback`. Nenhuma recebe senha. Quando a cifra precisa de um prompt e não há um, use `bornal console`.

`bornal console` escuta em `127.0.0.1` numa porta livre e imprime a URL. Ctrl+C encerra. A visão geral só responde GET vindo do loopback. Macros, addons, kits e rollback aceitam POST da mesma origem. A senha é um campo do formulário, não um parâmetro da URL.

## Fora desta versão

Não há armory, nem leitura de CASC ou DB2, nem chamada de rede para nomear itens. O código de addon não é gravado de volta no jogo.

As releases usam `vX.Y.Z`. Os passos estão em [docs/releases.md](docs/releases.md).

English: [README.md](README.md).

A porta de qualidade é `make test`.
