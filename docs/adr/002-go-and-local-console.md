# ADR-002: Go binary and a separate local console

**Date:** 2026-11-11

**Status:** Accepted

**Evidence:** `git-mcp` and `jira-mcp` already ship this shape. `jira-mcp` ADR-004 records why the console is its own process.

## Context

The tool needs a stdio MCP server and a web panel with the same functions. The panel must not ride inside the MCP client's process, and its URL must not land in that client's log.

## Decision

The language is Go 1.27.4, the same pin as `git-mcp`. The MCP SDK, when the stdio command arrives, is `github.com/modelcontextprotocol/go-sdk` v1.8.0. This slice does not import it yet.

`bornal console` is a different process from any future stdio server. It binds IPv4 `127.0.0.1` on a port chosen by the operating system, prints the overview URL on its own stdout, and serves until interrupted. Pages are GET only in this slice. The handler rejects a non-loopback peer, a `Host` or `Origin` other than the bound address, a cross-site `Sec-Fetch-Site`, and any method other than GET. Responses send `no-store`, `nosniff`, `DENY` framing, and a closed content security policy. There is no local password. Anyone who can open ports on this machine can read the page.

## Alternatives

- Rust, `rmcp`, and Axum, as in `discord-mcp`: rejected. That stack fits a remote service. This one is local files and the console those Go servers already have.
- Starting the panel from the stdio process with an environment variable: rejected in `jira-mcp` ADR-004. The same rejection applies here.

## Consequences

Writes, when they exist, stay behind the same core the MCP tools call. A confirm in one door does not confirm the other. POST will need the same origin checks plus a same-page token. The encryption passphrase never becomes a tool argument.

## Review

Revisit if the console must serve more than one OS user on a shared host.
