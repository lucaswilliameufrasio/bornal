# writes-kits

The public binary used to only read an install. It now previews and writes the English warrior macro set, lists addon TOC fields, and creates, finds, verifies, and restores a wow-kit. Optional age encryption asks for the passphrase locally. The same actions exist on the CLI, the MCP server, and the loopback console. Each side keeps its own confirm token.

## Why

The person asked to finish the product that ADR-005 already accepted. Writes stay blocked while that install's Wow.exe is open, including a Wine path, and when the process check cannot run.

## Acceptance

- Apply refuses an open game and a failed process check. Covered by `internal/work`, `internal/cli`, and `internal/server` tests.
- A CLI token does not authorize the console. Covered by `TestConsoleApplyUsesItsOwnToken`.
- The English set uses `/startattack`. A different body is a conflict until replace. Other macros stay. The backup is `macros-cache.bornal-bak`.
- The built-in set is not applied for a non-English locale.
- Kit restore checks hashes first. A bad passphrase or a tampered kit leaves the destination unchanged. Covered by `internal/kit`.
- The passphrase is not a flag, a tool argument, or a query string. Secret-shaped SavedVariables lines are file and line only.
- `mise exec go@1.27.4 -- make test` passed on 2026-11-12: gofmt, `go mod verify`, `go vet`, `go test -count=1 ./...`, `go test -race -count=1 ./...`, and `go build ./cmd/bornal`.

## Limits

There is no armory and no CASC, DB2, or network item lookup. Addon code is not written back into the game. The live Forever install was not modified. No release tag was cut. Version stays `0.1.0`.

## Evidence

Decisions: `docs/adr/005-writes-kits-and-encryption.md`. ADR-003's status line points at that limit being finished. Acceptance rows: `docs/acceptance.md`.
