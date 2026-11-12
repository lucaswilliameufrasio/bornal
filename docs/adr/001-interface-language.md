# ADR-001: Interface language

**Date:** 2026-11-11

**Status:** Accepted

**Evidence:** The product is used by the author in Brazilian Portuguese and by people who do not read it. Spell names must still match the WoW client.

## Context

A pt-BR-only interface leaves everyone else in front of text they cannot read. Translating spell names ourselves would break macros, because the client only casts the name it shows.

## Decision

The program's own text is Brazilian Portuguese or English. Detection follows acari: the first of `LANG`, `LC_ALL`, and `LC_MESSAGES` that starts with `pt` selects Portuguese; otherwise the language is English. `--lang pt-BR` and `--lang en` override the environment. A `pt_PT` system receives Brazilian Portuguese. There is no third copy.

WoW spell names, when a later slice writes macros, come from that install's `textLocale`. JSON keys in kits stay English.

## Alternatives

- Portuguese only: rejected because the tool is meant for other people too.
- Translate spell names in the UI language: rejected because the macro would not cast.
- A full catalog of locales now: rejected. Two languages match the existing acari contract.

## Consequences

Every new user-visible string needs both texts. The console and the CLI share that catalog. Client locale and UI language are allowed to differ, and the overview says so.

## Review

Revisit when a third UI language has a real user.
