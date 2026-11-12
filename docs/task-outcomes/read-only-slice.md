# read-only-slice

Date: 2026-11-11

## What changed

The first binary reports a local install from `bornal doctor` and from the loopback console. The page language follows the acari rule, with `--lang` as an override. Spell names stay with the client locale. Decisions: `docs/adr/001-interface-language.md`, `docs/adr/002-go-and-local-console.md`, `docs/adr/003-read-only-first-slice.md`.

## Why

Writes to a live install were deferred until the read path and the loopback page were tested.

## Verification

`mise exec go@1.27.4 -- make test` passed again on 2026-11-11 while the first-run questions landed. The same gate covers this slice.

## Limitations

ADR-004 later allows a capped search, and only to fill the first-run question. Game-open detection, macros, addons, kits, and encryption remain out of this slice.
