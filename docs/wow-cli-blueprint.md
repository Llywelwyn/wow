# WOW CLI – Developer Blueprint

## Goals
- Provide a fast CLI for capturing and retrieving code snippets and bookmarks.
- Work seamlessly in pipe-driven workflows on Linux/macOS terminals.
- Keep storage human-inspectable (files) with searchable metadata in SQLite.

## Non-Goals
- No cloud sync or collaboration in v1.
- No Windows or BSD support in v1; focus on Linux terminals and design with future expansion in mind.
- No interactive TUI beyond stub plumbing.
- No binary distribution tooling yet; local builds only.

## Primary Scenarios
- Save code or URLs directly from stdout with a declarative key.
- Retrieve snippets for scripting (`wow key | grep foo`).
- Launch bookmarks in a browser (`wow open urls/github`).
- Update snippets with the user’s `$EDITOR`.
- List snippets in either machine-friendly TSV or styled tables.

## CLI Surface
| Command | Example | Notes |
| --- | --- | --- |
| save (implicit) | `echo "..." \| wow --tags "golang,utils"` | Default when stdin is piped; auto-generates key unless explicitly provided. |
| get (implicit) | `wow go/foo` | Default when stdin is a TTY. |
| open | `wow open urls/github` | URLs open via `xdg-open`/`open`; text mirrors `get`. |
| edit | `wow edit go/foo` | Launches `$EDITOR`, refreshes `modified`. |
| ls | `wow ls -v --plain` | Supports verbose mode and forced plain output. |
| rm | `wow rm go/foo` | Removes file + metadata transactionally. |

### Command Behavior Details
- **save**: read stdin, reject empty stream, ensure parent dirs exist, auto-generate key when none supplied (Unix epoch seconds under `auto/<ts>` namespace, retry with suffix on collision), write file with `0600`, detect type (`url` vs `text`), normalize tags (comma-separated, deduped, lowercase), persist metadata. On DB failure, remove the file; on file failure, abort before DB.
- **get**: resolve path from key, read bytes, write to stdout. On missing snippet, return `SnippetNotFound`.
- **open**: dispatch by snippet type. For `url`, run Linux opener (`xdg-open`) with future hook for macOS/others. For `text`, launch viewer preference order: `$WOW_PAGER`, `$PAGER`, then `less`; if all missing/unavailable, stream to stdout. Surface opener exit status to user.
- **edit**: resolve path, ensure it exists, launch `$EDITOR` (fallback `nano`), stat file to detect changes, update `modified` timestamp; if file touched, optionally re-run type detection.
- **rm**: verify existence via metadata lookup, delete file first, then metadata; if metadata delete fails, warn and surface error without losing track of file removal.
- **ls**: query ordered results with optional filters (future). Render plain TSV (`key\tcreated\t...`) or pretty Lipgloss table. In non-TTY or when `--plain` is set, force plain.

## Platform Scope
- Target Linux shells for v1. Keep interfaces abstracted (opener, path rules) so macOS/Windows support can follow without major refactors.

## Storage & Metadata
- Base directory: `$WOW_HOME` env override, otherwise `$XDG_DATA_HOME/wow` or `~/.wow`. Create at startup if missing.
- Files: stored under `<base>/<key>`, where `key` segments map to nested directories. Reject `..`, absolute paths, or characters other than `[A-Za-z0-9._-]` plus `/`.
- Metadata DB: `~/.wow/meta.db` using SQLite. Single table `snippets` (schema from original doc) with indexes on `created` and `tags`.
- Tags: stored as comma-separated list with canonical ordering for stable output. Parsing utility handles splitting and trimming.
- Auto keys live under `auto/<unix-epoch>`; collisions append `-1`, `-2`, etc. Users can still provide explicit keys to override.

## Architecture
- `cmd/wow/main.go`: argument parsing, environment setup, command dispatch.
- `internal/command`: thin command structs implementing `Execute([]string) error`, responsible for flag parsing and invoking services.
- `internal/core`: business logic orchestrating storage, metadata, and UI (new layer to avoid bloated command files).
- `internal/storage`: filesystem helpers (ensuring dirs, atomic writes) and SQLite repository.
- `internal/ui`: output mode detection, plain/pretty renderers, future interactive stub.
- `internal/model`: domain structs plus validation helpers.
- Shared `config` package for environment resolution.

## Execution Flow Highlights
1. **Startup**: resolve base paths, initialize directories, open SQLite connection, migrate schema.
2. **Dispatch**: inspect CLI args + stdin mode to choose command; each command requests an output mode when needed.
3. **Operations**: wrap filesystem + DB actions in transactional helpers (best effort; SQLite transaction where cross-cutting).
4. **Output**: UI layer takes domain models and prints according to mode.

## Error Handling & Logging
- Commands return rich errors (wrapped with `%w`); main prints `error: <message>` to stderr and exits `1`.
- Use `errors.Is` for well-known cases (`ErrNotFound`, `ErrInvalidKey`, `ErrEditorFailed`).
- Optional `WOW_DEBUG=1` enables verbose diagnostics to stderr.

## Dependencies & Tooling
- Go 1.22.
- `github.com/mattn/go-sqlite3` (CGO).
- `golang.org/x/term` for TTY detection.
- `github.com/charmbracelet/lipgloss` for pretty output (optional stub if CGO disabled).
- `github.com/charmbracelet/bubbletea` kept as stub dependency for future interactive mode.

## Testing & QA
- Unit tests for key parsing, auto-key generation, storage, metadata repository, tag normalization.
- Integration tests using temp directories & SQLite in-memory DB.
- CLI smoke tests via `go test` with `exec.Command` + golden outputs.
- Manual checks for editor integration and browser opening.

## Delivery Roadmap
1. Bootstrap module, config, and storage scaffolding.
2. Implement save/get pipelines with tests.
3. Add list rendering and metadata formatting.
4. Layer open/edit/rm commands.
5. Polish output modes and error ergonomics.
6. Finalize CLI help, documentation, packaging notes.

## Open Questions
- Should we support auto-generated keys when piping with `--auto`?
- Do we need Windows support (different opener, path rules)?
- Should `open` launch text snippets in `$PAGER` instead of stdout?
- Is there any need for search/filter flags in `ls` for v1?
- How do we handle concurrent saves (file locking vs last-write wins)?
