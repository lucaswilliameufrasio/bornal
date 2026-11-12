# Design system do console

Status: aprovado em 2026-11-12 pelo dono do repositório. Decisão: `docs/adr/006-console-visual.md`.

Fonte de verdade da interface de `bornal console`. Cor e medida vêm dos tokens. O que não está aqui não entra na página.

## 1. Propósito e audiência

O bornal lê uma instalação local, mostra o que uma gravação mudaria e só grava com o Wow.exe daquela instalação fechado. A pessoa usa o console no mesmo computador, no navegador, em `127.0.0.1`. A primeira leitura da visão geral tem de dizer qual instalação é, se o jogo está aberto e que a gravação pede prévia, token e jogo fechado.

A marca é a palavra `bornal`. O cliente de jogo não é a moldura da página.

## 2. Auditoria

O console anterior copiava o git-mcp, que media o jira-mcp: trilho azul, cara de painel genérico. Uma paleta de lona e uma moldura de latão tentaram parecer o cliente clássico. As duas foram recusadas. Gradiente e serifa de sistema imitam o site de servidor privado, e a moldura real do cliente é textura pintada, que este repositório não embute.

O que fica é o conteúdo: flavor, caminho, locale, agentUID, jogo aberto ou fechado, e o formulário que já existia. Não há herói, notícia, contagem de jogadores nem cadastro.

## 3. Conceito e princípios

Conceito: um painel preto em que o número é o dado. O flavor e o estado do jogo ocupam o lugar que um gráfico ocuparia.

1. O dado manda. O título da visão geral, com uma instalação, é o flavor, então a página nunca abre com um slogan.
2. Uma cor de acento marca a página atual, então o índigo não pinta fundo nem botão.
3. Cor não substitui a palavra. O ponto verde só acompanha "não" quando aquele jogo está fechado, então jogo aberto e jogo não verificado não ganham ponto.
4. A página não inventa métrica. Com várias instalações, a célula de jogo aberto nomeia o flavor que foi lido, então não parece que todas foram checadas.
5. A prévia não grava. A visão geral não tem formulário, então o token de um lado não aparece como botão de escrita.

## 4. Tokens

### Cor

Medidos em 2026-11-12 contra o fundo indicado. Texto corrido precisa de 4,5:1. Componente essencial precisa de 3:1.

| Token | Valor | Papel | Contraste |
| --- | --- | --- | --- |
| `--color-bg` | `#000` | Fundo | — |
| `--color-surface` | `#0a0a0a` | Superfície reservada | — |
| `--color-text` | `#ededed` | Texto e fundo do botão | 17,94:1 no `#000` |
| `--color-muted` | `#a1a1a1` | Rótulo, nota, nav inativa | 8,13:1 no `#000`, 7,66:1 no `#0a0a0a` |
| `--color-dim` | `#888` | Caminho e modo | 5,92:1 no `#000` |
| `--color-line` | `#1a1a1a` | Divisor decorativo | 1,21:1. Não é o contorno do campo |
| `--color-rule` | `#2e2e2e` | Fio do painel | 1,55:1. Agrupa; o foco é que mostra o campo ativo |
| `--color-control` | `#5c5c5c` | Borda do campo | 3,14:1 no `#000` |
| `--color-link` | `#cdddff` | Índice do passo e foco | 15,38:1 no `#000` |
| `--color-accent` | `#5266eb` | Sublinhado da página atual, 1px | 4,46:1. Abaixo de 4,5:1, então não é texto |
| `--color-safe` | `#3dd68c` | Ponto de jogo fechado | 11,20:1, e mesmo assim não é texto |

O botão é `#000` em `#ededed`, 17,94:1. Branco em `#5266eb` daria 4,71:1 e não é usado.

### Tipo

| Estilo | Tamanho / entrelinha | Peso | Uso |
| --- | --- | --- | --- |
| Marca | 16px / 1,5 | 500, tracking −0,03em | Palavra `bornal` |
| Título | 56px / 1, 40px abaixo de 760px | 500, tracking −0,045em | Flavor ou nome da página |
| Número | 28px / 1 | 500, tracking −0,03em | Valor da métrica |
| Passo | 20px / 1,5 | 500 | Nome do passo |
| Corpo | 16px / 1,5 | 400 | Texto |
| Nota | 14px / 1,5, até 62ch | 400 | Nota de locale |
| Caminho | 13px, mono | 400 | Caminho, quebra em qualquer ponto |

Geist para o texto, Geist Mono para caminho e índice. A Arcadia do Mercury não é pública. A licença SIL OFL está em `internal/console/font/OFL.txt`. O nome Geist não é renomeado.

### Espaço, forma, profundidade

Barra com 64px de altura mínima. Miolo até 920px, com 56px de respiro no desktop e 16px na lateral abaixo de 760px. Painel, métrica e passo usam borda de 1px, raio 0, sem sombra. O alvo mínimo é 44px. Abaixo de 760px a barra vira grelha: marca e modo na primeira linha, navegação na segunda, e as faixas empilham.

### Movimento

Não há animação. `prefers-reduced-motion` não tem o que desligar. O foco aparece na hora.

## 5. Ícone, imagem e honestidade

Não há conjunto de ícones. O único sinal gráfico é o ponto de 8px, com `aria-hidden`, ao lado da palavra. Não há ilustração, retrato, textura do cliente nem captura de outro produto. Dado ausente usa o texto de falta já traduzido. Lista vazia diz o que não foi encontrado.

## 6. Voz

Português do Brasil ou inglês, pela regra de `internal/i18n`. Chave de formulário e de manifesto fica em inglês. A linha da macro mostra a ação que o núcleo já devolve: `create`, `skip` ou `replace`. O nome de habilidade vem do `textLocale` do cliente. O mesmo fato usa a mesma palavra: prévia, token, gravar, kit, jogo aberto. Sem urgência fabricada e sem texto que culpe a pessoa por não gravar.

## 7. Componentes

- Navegação: cinco links, alvo 44px. A página atual tem texto `#ededed` e sublinhado índigo. As outras ficam em `#a1a1a1` e passam a `#ededed` no hover. Foco: 2px `#cdddff`, offset 3px.
- Métrica: rótulo mudo e valor em 28px. Três células na visão geral de uma instalação. Uma célula, na largura do conteúdo até 320px, quando não há flavor como título.
- Passo: índice mono `#cdddff`, título, detalhe mudo. Só na visão geral.
- Lista: nome e detalhe, fio entre itens. Link, quando existe, usa `#cdddff`.
- Formulário: rótulo mudo, campo com borda `#5c5c5c`, hover na borda `#888`, senha com `autocomplete="off"`. A caixa de seleção tem alvo na linha de 44px e `accent-color` índigo. O botão é texto preto em fundo `#ededed`, alvo 44px. Não há estado de carregamento: a resposta troca a página.
- Mensagem: a mesma faixa do painel para vazio, erro e confirmação.

## 8. Páginas

A visão geral com uma instalação mostra o kicker "instalações locais", o flavor como `h1`, o caminho, três métricas e a faixa "Para gravar". Com nenhuma ou com várias, o `h1` é "Instalações". Várias instalações repetem cada flavor numa faixa, e a célula de jogo aberto leva o flavor lido. Macros, addons, kits e rollback usam a barra e o painel, sem kicker e sem a faixa de passos. A visão geral não tem `POST`.

## 9. Engajamento

Não há laço de retenção. A página não pede notificação, não conta visita e não empurra uma gravação. Cifrar o kit é opção desmarcada.

## 10. Plataformas

Web local, desktop e largura de telefone. Não há app nativo nem TUI desta página. O terminal continua com as frases longas de `bornal doctor`.

## 11. Diretrizes

1. MUST usar só os tokens deste arquivo para cor, tipo e alvo.
2. MUST manter o índigo fora de texto pequeno e de fundo de botão.
3. MUST mostrar a palavra do estado do jogo. O ponto verde só entra com o jogo fechado.
4. MUST nomear o flavor lido quando a visão geral lista mais de uma instalação.
5. MUST NOT colocar formulário na visão geral.
6. MUST NOT inventar número, jogador, notícia ou textura do cliente.
7. MUST NOT buscar fonte em CDN. A fonte entra como data URI, com a licença OFL no repositório.
8. MUST NOT tratar um token de confirmação como exemplo público.

## 12. Anti-padrões

- Trilho azul do git-mcp: parecia outro produto. A barra preta substitui.
- Paleta laranja de busca de estilo: não descrevia o bornal.
- Lona, pesponto e moldura de latão: a cópia em CSS ficou falsa. O dado em tipo grande substitui.
- Home de servidor privado: herói e contagem de jogadores não são dados daqui.
- Ícone de qualidade de item como legenda de criar, pular ou conflito: a palavra já diz o estado.

## 13. Lista antes de entregar

- [x] Tokens nomeados no CSS
- [x] Contraste medido para texto e para a borda do campo
- [x] Vazio, várias instalações, jogo aberto e jogo fechado renderizados
- [x] Foco visível no CSS e alvo de 44px
- [x] Sem animação
- [x] Conteúdo sintético na prova visual, sem métrica inventada
- [x] Desktop 1280 e largura 390 conferidos na visão geral

## 14. Fontes

| Fonte | O que informou | Consultado | Adotado |
| --- | --- | --- | --- |
| Painel da Vercel, [fudge.design/share/vercel.com-design](https://fudge.design/share/vercel.com-design) | Fundo `#000`, texto `#ededed`, mudo `#a1a1a1`, fio `#1a1a1a` e `#2e2e2e`, número como tipo | 2026-11-12 | Sim, menos o azul de ação `#006efe` (4,66:1), que não entra |
| Mercury, índigo e tipo em [shadcn.io/design/mercury](https://www.shadcn.io/design/mercury), a partir de mercury.com | Acento `#5266eb`, peso de título perto de 500, tracking negativo. A Arcadia não é pública | 2026-11-12 | O índigo e o jeito do número. O fundo violeta `#171721` não: o preview aprovado é `#000` |
| Geist, SIL OFL, copyright 2023 Vercel com basement.studio. Arquivos em `internal/console/font/` | Família que substitui a Arcadia | 2026-11-12 | Sim, como data URI |
| WCAG 2.2, [alvo mínimo 2.5.8](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) | Texto 4,5:1 (1.4.3), componente 3:1 (1.4.11). O alvo AA é 24px. O critério 2.5.5, citado nessa página, usa 44px | 2026-11-12 | Os limiares. O console fica no alvo de 44px. Nenhuma paleta |
| git-mcp / jira-mcp | Console anterior | 2026-11-11 | Não. Substituído neste corte |
| Busca de paleta ui-ux-pro-max | Sugestão laranja de marketing | 2026-11-12 | Não |
| Paleta de lona e moldura de latão | Tentativas locais | 2026-11-12 | Não. Recusadas pelo dono |
| Homes de servidor privado mostradas na conversa | Referência que a pessoa gosta em outro gênero | 2026-11-12 | Não. Não são a ferramenta |
| Pika Backup, Déjà Dup, Terraform plan/apply | O fluxo de prévia, token e restauração, já decidido no produto | 2026-11-12 | Não a identidade visual |

## 15. Mudanças

| Data | Mudança | Decisão |
| --- | --- | --- |
| 2026-11-11 | Console com os tokens do git-mcp | Corte anterior |
| 2026-11-12 | Painel preto, Geist, índigo só na página atual | ADR-006 |
