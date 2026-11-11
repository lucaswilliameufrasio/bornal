# ADR-004: Ask on first use

**Date:** 2026-11-11

**Status:** Accepted

**Evidence:** The choices left open while designing discovery and language (which install, which UI language) should be asked, not required as flags the person already knows.

## Context

`--path` and `LANG` are enough for a test. They are a poor first run. A person with two clients, or a client on another disk, should see the options. The MCP process has no terminal: its stdin is the protocol.

## Decision

`bornal setup` asks two things and stores the answers in the config file (`$XDG_CONFIG_HOME/bornal/config.json`, or `~/.config/bornal/config.json`).

1. Interface language, `pt-BR` or `en`. The question itself is bilingual. The detected language is the default if the person presses Enter.
2. Install path. A shallow search offers the directories it found. The person can type another path. The path is kept only when `WTF/Config.wtf` is there.

`bornal doctor` and `bornal console` use `--path`, then `BORNAL_WOW_PATH`, then the saved path. On a terminal, if all three are empty, they ask the same questions and then continue. Without a terminal they refuse and name `bornal setup`.

`--lang` still wins for one invocation. After setup, the saved language wins over `LANG`.

The stdio server does not prompt. `configuration` returns the pending questions. `save_configuration` stores the same file after the agent has asked the person. A path that is not an install is not saved.

The search used by the question looks at a few known roots and at most four directory levels under mounted volumes, skips heavy names such as `Data` and `.git`, and stops after 2000 directories. It is not a full-disk walk.

## Alternatives

- Keep requiring `--path`: rejected. That is what the first binary did, and it is the wrong first use.
- Prompt inside the MCP stdio stream: rejected. That stream is the protocol, same reason the console is a separate process.
- Ask for the kit passphrase or the account now: rejected. The passphrase belongs to the moment a kit is created. The account is only a question when a later slice finds more than one.

## Consequences

ADR-001 still holds: two UI languages, spell names from the client. The saved language is a new step in front of detection. ADR-005 finishes writes. A full-disk scan stays out. The search above is the exception, and only to fill the question.

## Review

Revisit when a third question has a real default that should be saved, such as the kit export directory.
