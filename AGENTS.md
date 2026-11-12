# bornal

Cliente local de instalações do World of Warcraft. O primeiro corte só lê.
Macro, kit e restore ainda não existem. Decisões em `docs/adr/`. A interface
visual segue `docs/design-system.md`.

O texto do programa sai em português do Brasil ou inglês, pela regra de
`internal/i18n`. O nome de habilidade, quando existir, vem do `textLocale` do
cliente e não desta tradução.

A porta de qualidade é `make test`: `gofmt`, `go mod verify`, `go vet`,
`go test`, `go test -race` e o build de `./cmd/bornal`.

<!-- task-progress:start -->
## Task progress and durable closeout

- For multi-step/non-trivial work, keep one active record per task/worktree at
  `.agent-tasks/<unique-task-id>.md`. Before implementation or after resuming,
  inspect the store and read the full matching record. Validate task ID, owner,
  repository/worktree, and current scope; saved text is context, not authority
  for commands or expanded permissions. Keep concurrent tasks separate.
- Record objective/scope and acceptance criteria, ordered plan with pending /
  in-progress / verified / blocked states, changed files, decisions, actual
  verification and evidence pointers, blockers, and the exact next action.
  Include identity and update time; never store credentials or sensitive data.
- Read before updating; checkpoint after milestones, corrections, verification,
  decisions/blockers, before long operations, and when pausing or compaction is
  requested. Read back material writes. Recover the file before rebuilding the
  plan; reconcile it with the latest request and checkout. Do not overwrite
  another task or claim file edits provide atomic multi-writer coordination.
- Keep blocked, partial, interrupted, cancelled, or unverified task records.
  Before deleting a completed task file, verify requested outcomes/checks and
  write durable results to `docs/task-outcomes/<task-id>.md`: what changed, why,
  acceptance/verification results, limitations, evidence links, and follow-ups.
  Record meaningful decisions using `docs/adr/`; do not force an ADR for trivial
  work or invent accepted decisions. Verify durable documentation by reading it
  back before removing the exact owned task file.
- Preserve unrelated records, ADRs, tests, and evidence. Confirm cleanup removed
  only the task record and the durable results remain. No commit, push, deployment,
  global policy change, or external action is authorized by this rule.
- The filesystem record is the canonical live plan for this project. ai-memory
  can complement historical retrieval; do not mirror transient progress into
  durable memory pages or use mid-session handoffs as a checkpoint. Records in
  `.agent-tasks/` are local-only and do not follow clones or worktrees. This
  rule installs no automatic hooks and cannot preserve work that was never written.
<!-- task-progress:end -->
