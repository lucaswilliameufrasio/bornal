# ADR-005: Macros, kits, and encryption

**Date:** 2026-11-12

**Status:** Accepted

**Evidence:** The read path and the first-run questions are in the public tree. The remaining product was already specified: macros, addon audit, portable kits, and optional encryption. An armory stays declined.

## Context

ADR-003 kept the first binary from writing `macros-cache.txt` while `Wow.exe` could overwrite the file on exit. The person asked to finish that product. The spell names of the built-in set are English because this client's `textLocale` is enUS, and the program does not translate spells.

## Decision

`bornal macros apply` writes account macros only when the `Wow.exe` for that install is closed. A Windows path inside a Wine command line counts as the same install. The preview and the write do not share a confirmation: the CLI, the MCP server, and the console each keep their own token. `confirm` is empty until the person accepts that preview.

The built-in warrior set uses `/startattack` and never `/cast Attack`. An existing macro with another body is left alone unless `replace` is set. Other macros in the file stay, including their ids. The backup beside the file is `macros-cache.bornal-bak`, never `macros-cache.old`. `rollback` restores that backup and also requires the game to be closed.

A kit is a directory ending in `.wowkit`, or one age-passphrase file with a clear `WOWKIT1` header. The header has the date, a label, and the kdf name. It has no character or account data. The passphrase is read from the terminal or from pinentry. It is not a flag, a tool argument, a URL, or a log line. The console sends it in a loopback POST body. SavedVariables are copied as text. They are not executed. A line that looks like a secret is reported as file and line number, and the value is not printed.

Restore checks every hash first, refuses a path that escapes the kit, and does not create an account folder that does not exist. The person logs in once so the client creates that folder. A character whose realm folder differs is not copied until the realm is chosen. Addon audit lists TOC fields only.

There is still no armory, no CASC read, and no network call to name items.

## Alternatives

- Write the cache while the game is open: rejected. The client overwrites it on exit.
- One confirmation shared by the panel and the MCP server: rejected. A preview on one side must not authorize the other.
- Put the passphrase in the MCP arguments: rejected. The agent transcript would keep it.

## Consequences

ADR-003's read-only limit is finished. The safety checks above stay. Apply still refuses when the process check cannot run.

## Review

Revisit when a second locale needs its own spell names, or when a kit must include addon code by default.
