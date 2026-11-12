# Design system do console

Fonte de verdade da interface de `bornal console`. Cor e medida vêm dos tokens. O que não está aqui não entra na página.

## Referência

| Ref | Fonte | Papel | Consultada |
| --- | --- | --- | --- |
| A | Console do git-mcp (`internal/console/console.css` e `docs/design-system.md`), que por sua vez mede o console do jira-mcp | Fundo, superfície, trilho, texto, foco, tipo e espaçamento | 2026-11-11 |

Os valores foram lidos do CSS do git-mcp, não estimados. A linha de commits e o diffstat ficam lá: pertencem ao git, não a uma instalação do jogo.

## Identidade

A marca é a palavra `bornal` em monoespaçada. Não há segunda cor de produto neste corte. Uma página só, Instalações.

## Tokens

Estão em `internal/console/console.css`, no bloco `:root`.

| Token | Valor | De |
| --- | --- | --- |
| `--color-bg` | `#edf3f8` | A |
| `--color-surface` | `#fff` | A |
| `--color-line` | `#cbd9e6` | A |
| `--color-rule` | `#93afc7` | A |
| `--color-text` | `#18304d` | A |
| `--color-muted` | `#50647a` | A |
| `--color-link` | `#1757a6` | A |
| `--color-rail` | `#163353` | A |
| `--color-rail-text` | `#d8e7f3` | A |
| `--color-rail-muted` | `#c4d8eb` | A |
| `--color-rail-current` | `#e9f3ff` | A |
| `--color-notice-bg` | `#fff4e9` | A |
| `--color-focus` | `#c27400` | A |
| `--font-body` | `"Trebuchet MS", Verdana, sans-serif` | A |
| `--font-mono` | `ui-monospace, "Cascadia Mono", "DejaVu Sans Mono", "Noto Sans Mono", monospace` | A |
| `--size-h1` | `clamp(2rem, 4vw, 3.1rem)` | A |
| `--rail-width` | `240px` | A |
| `--target-min` | `44px` | A |

Até 760px o trilho deixa de ser coluna e vira uma faixa no topo.
