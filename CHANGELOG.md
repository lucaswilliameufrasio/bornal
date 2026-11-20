# Changelog

Mudanças do bornal. As entradas saem do histórico Git e são revisadas no
pull request de release antes da tag.

## [0.1.2] - 2026-11-20


### CI / Release

- **release:** List what shipped and publish installers


The release page was only a static header, so the tag did not say what
changed. GoReleaser now groups the commits since the previous tag,
publishes one SHA-256 per file, and attaches the installers. Prepare
Release starts CI because the bot token does not.


## [0.1.1] - 2026-11-20


### Bug fixes

- Show each install's own game state and localized macro actions


The overview used one running check for every install. Each install now
keeps its own result, and a missing check stays unchecked. Preview words
follow the interface language. The MCP action field stays English.


## [0.1.0] - 2026-11-12


### CI / Release

- Add release workflows and state the current slice


Prepare Release opens a changelog pull request. A stable tag runs CI
and publishes GoReleaser archives. The README lists what this binary
does not do yet.



### Documentation

- Record the language, console split, and read-only slice


The first binary only reports an install given with --path. UI text is
Brazilian Portuguese or English. Spell names stay with the client locale.


- Rewrite the readme


Lead with what bornal does, then give each door its own commands.



### Features

- Report a local install from doctor and the loopback console


bornal doctor and bornal console read WTF/Config.wtf and refuse to scan
other disks. The page accepts GET from 127.0.0.1 only.


- Ask language and install on first use


The first CLI run stores the interface language and the World of Warcraft
folder. The stdio server returns the same questions and does not prompt
on the protocol.


- Write macros and kits with the game closed


The CLI, the MCP server, and the loopback console each confirm their own preview. Writes stay refused while that install's Wow.exe is open.


- Restyle the local console


The blue rail and the brass frame were rejected. The page is the black
dashboard that was approved, with Geist inlined and indigo only on the
current item.


