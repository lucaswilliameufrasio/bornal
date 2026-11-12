# first-run-questions

Date: 2026-11-11

## What changed

`bornal setup` asks for the interface language and the World of Warcraft install, then writes `config.json` (mode 0600). `doctor` and `console` use `--path`, then `BORNAL_WOW_PATH`, then that file. On a terminal they ask when all three are empty. Without a terminal they name `bornal setup`.

`bornal mcp` exposes `configuration` and `save_configuration`. The stdio stream does not prompt. A path is stored only when `WTF/Config.wtf` is present.

The question's search is capped: known roots, four directory levels, 2000 entries, skipping heavy and hidden names. Decision: `docs/adr/004-first-run-questions.md`.

## Why

Flags are a poor first run. The MCP process has no terminal, so the agent asks the person and then stores the same answers.

## Verification

`mise exec go@1.27.4 -- make test` passed on 2026-11-11: gofmt, `go mod verify`, `go vet`, `go test -count=1`, `go test -race -count=1`, and `go build -o /tmp/bornal ./cmd/bornal`.

The tests use a temporary WTF and a stubbed search. They do not walk `/mnt` or a live install.

## Limitations

The passphrase and the account are not questions yet. The search does not read Battle.net `product.db`. This slice still does not write the game, check whether `Wow.exe` is running, or build an armory.
