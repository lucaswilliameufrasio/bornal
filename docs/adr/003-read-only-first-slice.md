# ADR-003: First slice is read-only

**Date:** 2026-11-11

**Status:** Accepted. ADR-004 supersedes the required `--path` and the ban on a volume walk, and only for the first-run question.

**Evidence:** Macros, kits, and restore were designed. An armory was declined. Writing `macros-cache.txt` while `Wow.exe` is open is overwritten on exit.

## Context

The useful surface is large: discovery across disks, macros, addon audit, encrypted kits, and restore on another PC. Shipping the writes before the read path is tested puts real installations at risk.

## Decision

This version only looks. `bornal doctor` and `bornal console` accept `--path` or `BORNAL_WOW_PATH`. They do not scan every mounted volume. An install is a directory with `WTF/Config.wtf`, or a child of the given path that has one. The report shows flavor directory, `textLocale`, and `agentUID`. It does not say whether the game is running.

There is no armory, no item-name lookup, no CASC read, and no network call. SavedVariables are not executed. Account folders are not listed.

## Alternatives

- Walk every disk in the first binary: rejected. A fixture test must stay local, and a wrong root must not become a long scan.
- Resolve item names from the client database: rejected with the armory. Slot ids in addon files are not worth a second product.

## Consequences

The next slice can add a shallow volume search and a running-process check that understands a Wine path. Macro writes stay blocked until that check exists and the game for that directory is closed.

## Review

Revisit when the read-only report is enough to point `--path` at a real install without guessing.
