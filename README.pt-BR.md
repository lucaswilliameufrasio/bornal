# bornal

O bornal cuida das macros e das configurações de uma instalação do World of Warcraft nesta máquina. Ele mostra o que a gravação mudaria e só grava com o `Wow.exe` daquela instalação fechado.

```sh
bornal setup
bornal doctor
bornal console
```

`setup` pergunta o idioma da interface e o caminho da instalação. Grava `language` e `install_path` em `$XDG_CONFIG_HOME/bornal/config.json`, ou em `~/.config/bornal/config.json`. O arquivo não guarda mais nada.

`doctor` mostra o flavor, o locale do cliente, o `agentUID` e se o jogo daquela instalação está aberto.

O caminho vem de `--path`, depois de `BORNAL_WOW_PATH`, depois do caminho salvo. No terminal, `doctor` e `console` perguntam quando não há caminho. A busca é rasa: algumas raízes conhecidas, quatro níveis de diretório, e pula nomes pesados como `Data`. Sem terminal, rode `bornal setup`. Uma instalação é usada direto. Várias instalações, ou várias pastas de conta, viram pergunta. O programa não escolhe sozinho.

`--lang pt-BR` ou `--lang en` vale para aquela execução. O idioma salvo ganha de `LANG`, `LC_ALL` e `LC_MESSAGES`. O nome das habilidades vem do `textLocale` do cliente.

## Macros

O conjunto embutido é de guerreiro, em inglês. Ele só é gravado quando o locale do cliente é `enUS`, `enGB` ou `en`.

```sh
bornal macros preview [--path DIR] [--flavor NAME] [--account N] [--replace]
bornal macros apply --confirm TOKEN
bornal rollback [--confirm TOKEN]
```

`preview` lista cada macro como create, replace, skip ou conflict. `apply` grava com o token daquela prévia. A CLI, o servidor MCP e o console não aceitam o token um do outro. O mesmo nome com outro texto espera `--replace`. As outras macros ficam, com os ids delas. O backup ao lado do arquivo é `macros-cache.bornal-bak`. `rollback` restaura esse backup, também com o jogo fechado.

Um caminho do Windows na linha de comando do Wine conta como a mesma instalação. Se a checagem do processo não roda, a gravação é recusada.

## Addons

```sh
bornal addons [--path DIR] [--flavor NAME]
```

Lê o TOC de cada addon: título, versão, interface, dependências e a marca de desativado. Não instala addon e não grava código de addon de volta no jogo.

## Kits

Um wow-kit é um diretório que termina em `.wowkit`, ou um arquivo cifrado.

```sh
bornal kit create [--out DIR] [--label TEXT] [--encrypt]
bornal kit find [DIR]
bornal kit verify PATH
bornal kit restore PATH [--confirm TOKEN] [--replace] [--realm-ok]
```

`create` copia as macros da conta, bindings, arquivos de edição, SavedVariables, arquivos de personagem, um índice de addons e um `portable.wtf` curto. As opções gráficas ficam na máquina. `--encrypt` pede a senha duas vezes, no terminal ou no pinentry. Ela não entra como valor de flag, argumento de ferramenta nem URL. O kit cifrado é um arquivo com cabeçalho claro `WOWKIT1`: data, rótulo e kdf. O cabeçalho não tem personagem nem conta. Uma linha com cara de segredo aparece como arquivo e número da linha.

`find` lista os kits sem abri-los. `verify` confere os hashes. `restore` confere os hashes antes de gravar, recusa uma pasta de conta que não existe e pergunta quando o reino do personagem é outro.

## MCP e o console

```sh
bornal mcp
bornal console
```

`bornal mcp` fala MCP em stdin e stdout. Não pergunta nesse fluxo. `configuration` devolve as perguntas pendentes, e `save_configuration` grava as respostas. As ferramentas do jogo são `macros_preview`, `macros_apply`, `addons_list`, `kit_create`, `kit_find`, `kit_verify`, `kit_restore` e `rollback`. Nenhuma recebe senha. Quando a cifra precisa de um prompt, use `bornal console`.

`bornal console` escuta em `127.0.0.1` e imprime a URL. Ctrl+C encerra. A página mostra a instalação, a prévia das macros, a auditoria dos addons, os kits e o rollback. A visão geral só responde GET vindo do loopback. As outras páginas aceitam POST da mesma origem. A senha é um campo do formulário, não um parâmetro da URL.

## Fora desta versão

Não há armory, nem leitura de CASC ou DB2, nem chamada de rede para nomear itens.

As releases usam `vX.Y.Z`. Os passos estão em [docs/releases.md](docs/releases.md).

English: [README.md](README.md).

A porta de qualidade é `make test`.
