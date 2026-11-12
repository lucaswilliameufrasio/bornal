# console-visual

`bornal console` now uses the black page the owner approved on 2026-11-12. Geist is inlined. Indigo marks only the current page. One install shows the flavor as the overview title, three metrics, and the three write steps. The other pages use the same bar and the same panels. Behavior, POST names, and the per-side confirm token are unchanged.

## Why

The first console copied the git-mcp rail. Later trials were rejected. The approved preview is a black dashboard: the flavor and the game state are the figures, and the page does not imitate the game client.

## Acceptance

- One closed install renders `_forever_` as the title, `não` with the green dot, and "Para gravar". An open game renders `sim` with no dot. Covered by `internal/console` and checked on the rendered page at 1280 and 390.
- Several installs title the page "Instalações" and the game cell names the flavor that was scanned.
- Macros, addons, kits, and rollback omit the overview kicker and the step strip. Kits still posts label, destination, encrypt, kit, replace, realm, and the passphrase fields.
- Overview has no form. The content security policy still closes the page and allows `font-src data:`.
- `mise exec go@1.27.4 -- make test` passed on 2026-11-12: gofmt, `go mod verify`, `go vet`, `go test -count=1 ./...`, `go test -race -count=1 ./...`, and `go build ./cmd/bornal`.

## Limits

The macro row still prints the core action (`create`, `skip`, `replace`). The keyboard focus ring is specified and was not driven with Tab. Screenshots used synthetic installs. The live Forever install was not modified. No commit and no release tag. Version stays `0.1.0`.

## Evidence

Decision: `docs/adr/006-console-visual.md`. Tokens: `docs/design-system.md`.
