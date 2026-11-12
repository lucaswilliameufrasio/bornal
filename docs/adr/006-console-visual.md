# ADR-006: Console visual

**Date:** 2026-11-12

**Status:** Accepted

**Evidence:** The owner approved the overview preview on 2026-11-12. Tokens, type, and the rejected directions are in `docs/design-system.md`.

## Context

The first console copied the git-mcp page, which measured jira-mcp. Later trials used a canvas palette and a brass frame meant to imitate the classic client. The owner rejected both. The client frame is painted texture, and those assets do not ship in this repo. Two private-server homepages were also shown and are not this product.

## Decision

`bornal console` uses a black page, Geist, and hairline panels. The word `bornal` is the mark. Indigo marks only the current page. One install makes the flavor the overview title. Numbers are type, not a chart. A green dot appears only when that install's game is closed, and the word is always present. The other pages keep this chrome and their existing forms.

## Alternatives

- Keep the git-mcp blue rail: rejected with the first look.
- Orange marketing palette from a design-data search: rejected. It did not describe this tool.
- Canvas, saddlebag, or a CSS imitation of the quest frame: rejected by the owner.
- Private-server homepage, with a hero, a player count, and a register form: rejected. Those are not bornal data.

## Consequences

Fonts are the SIL Open Font License files in `internal/console/font/`, inlined so the loopback server gains no font route. The content security policy allows `font-src data:`. Overview stays free of POST forms. A confirm on this side still does not confirm the CLI or the MCP server.
