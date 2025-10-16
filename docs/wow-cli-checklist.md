# WOW CLI Implementation Checklist

## Environment Setup
- [ ] Confirm Go 1.22 toolchain.
- [ ] Create module (`go mod init github.com/<user>/wow`), add dependencies.
- [ ] Add `make lint`, `make test` helpers (go fmt, go test).

## Milestone 1 – Foundations
- [ ] Implement config package (path resolution, directory creation, Linux-focused defaults).
- [ ] Add key parser with validation + tests.
- [ ] Initialize SQLite (migration on startup, connection management).
- [ ] Stub command interface and dispatcher skeleton.

## Milestone 2 – Core CRUD
- [ ] Build storage layer: atomic file writes, reads, deletes.
- [ ] Implement metadata repository with transactions.
- [ ] Implement auto-key generator (epoch seconds with collision suffix) and plug into save flow.
- [ ] Finish `save` command with tag normalization and type detection tests.
- [ ] Finish `get` command with not-found handling.
- [ ] Wire default dispatch (pipe → save, tty → get).

## Milestone 3 – Extended Commands
- [ ] `ls` command with plain + pretty renderers, verbose flag, tests.
- [ ] `open` command (URL opener abstraction, WOW_PAGER → PAGER → less → stdout fallback).
- [ ] `edit` command (editor resolution, modified timestamp updates).
- [ ] `rm` command (file→metadata deletion order, error surfacing).

## Milestone 4 – UX Polish
- [ ] Global `--plain` flag wiring; auto-detect TTY.
- [ ] Helpful errors (`wow rm missing/key`).
- [ ] CLI help/usage text and README snippets.
- [ ] Optional `--desc`, `--tags` flag validation (max length, formatting).

## QA & Release
- [ ] Integration tests using temp homes.
- [ ] Smoke tests for piping scenarios (`wow ls --plain | jq` compatibility).
- [ ] Manual edit/browser verification on target OS.
- [ ] Document known limitations and open questions.

## Definition of Done
- [ ] All checklist items complete or tracked as backlog.
- [ ] `go test ./...` green.
- [ ] Basic user guide drafted from blueprint.
- [ ] Known issues captured in backlog (GitHub issues or TODO file).
