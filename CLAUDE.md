# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`gitwise` is a terminal-based git manager — a 3-panel TUI (Branches / Commits / Detail) that wraps the local `git` binary. It does not embed git or manage credentials. The module name is `github.com/user/gitwise`.

## Commands

```bash
go build ./...          # build everything
go run ./cmd/gitwise    # run the app (cmd/gitwise/main.go is the entry point)
go test ./...           # run all tests
go test ./internal/git  # run a single package's tests
go vet ./...            # lint
```

> `cmd/gitwise/main.go` is planned but not yet created. It should call `git.FindRepoRoot(cwd)`, construct `model.NewApp(repoRoot)`, and run `tea.NewProgram(app, tea.WithAltScreen())`.

## Architecture

### Layer 1: `internal/git/` — shell-out wrapper

- **`client.go`** — `Client` struct wraps `exec.Command("git", ...)` with `repoRoot` as working dir. All raw git operations live here, returning typed structs (`Branch`, `Commit`, `CommitDetail`) or errors. The `run`/`runLines` helpers capture stderr on failure.
- **`stash.go`** — `AutoStash`/`AutoUnstash` around destructive ops. Returns a `StashResult` with the stash ref so the pop can target the correct entry even if other stashes exist.
- **`workflows.go`** — `Workflows` composes `Client` into stash-aware high-level operations. Each method calls `stashAround(label, fn)` which handles the push/pop lifecycle and appends a note to the result message if changes were stashed.
- **`util.go`** — `FindRepoRoot` walks upward from a directory until it finds `.git`.

### Layer 2: `internal/model/` — Bubble Tea models

- **`app.go`** — Root `App` model. Owns the three panel models plus `focus panel` (0=branches, 1=commits, 2=detail). On `WindowSizeMsg` or focus change it calls `relayout()` → `ui.Widths()` → each panel's `SetSize()`. Git operations are dispatched as `tea.Cmd` goroutines and return `WorkflowResultMsg`. The `menuOpen bool` flag gates all key events to `ContextMenuModel` while a menu is active.
- **`branches.go`** — `BranchesModel` renders LOCAL/REMOTE section headers. On cursor move it fires `BranchSelectedMsg`, which `app.Update` handles by issuing a `loadCommitsCmd`.
- **`commits.go`** — `CommitsModel`. Has a multi-select mode toggled by `s`; `space` toggles individual commits; `SelectedHashes()` returns all marked hashes (falls back to cursor). On cursor move fires `CommitSelectedMsg` → `loadDetailCmd`.
- **`detail.go`** — `DetailModel` uses `bubbles/viewport` for scroll. Has three modes: `DetailModeView` (normal scroll), `DetailModeConfirm` (y/n prompt for destructive ops), `DetailModeAmend` (inline `textinput` for rewriting commit messages). Confirm and amend modes are entered via `AskConfirm()` and `StartAmend()`.
- **`contextmenu.go`** — `ContextMenuModel` is a full-screen overlay (rendered instead of the panels). `BranchMenuItems` and `CommitMenuItems` return context-sensitive `[]MenuItem`. Shortcut keys in the menu map directly to `MenuAction` constants.

### Layer 3: `internal/ui/` — rendering

- **`styles.go`** — All lipgloss styles and palette colors. Focused panels get `ColorAccent` border; inactive get `ColorDim`.
- **`statusbar.go`** — `RenderStatusBar` lays out keybinding hints on the left and a status/error message on the right. `BranchHints`, `CommitHints`, `DetailHints` are swapped in based on `app.focus`.
- **`layout.go`** — `Widths(termW, focus)` returns `[3]int` column widths. Base split is 22/35/43%; the focused panel gets a boost drawn proportionally from the other two.

### Message flow

```
KeyMsg → app.Update → panel.Update → returns tea.Cmd
  → goroutine: git.Workflows.SomeOp()
    → WorkflowResultMsg → app.Update → setStatus + refreshAll
```

`refreshAll` re-fetches branches and commits concurrently via `tea.Batch`.
