# internal/ui — rendering layer

This package is a pure rendering and layout library. It holds no state, starts no goroutines, and imports nothing from `internal/git` or `internal/model`.

## Files

| File | Responsibility |
|------|---------------|
| `styles.go` | All lipgloss styles and palette color variables |
| `labels.go` | Every user-visible string (titles, hints, error formats, placeholders) |
| `layout.go` | `Widths()` — 3-panel column split; `InnerWidth()` — usable width after border/padding |
| `statusbar.go` | `HintSet` type, per-panel hint constructors, `RenderStatusBar` |
| `overlay.go` | `PlaceOverlay` — ANSI-aware compositing of two rendered strings |

## Key invariants

### All strings live in `labels.go`

Never hardcode a user-visible string in a `View()` method or a rendering function. Add it to `labels.go` and reference the constant. Format strings (containing `%s`, `%d`) live there too; they are passed to `fmt.Sprintf` at the call site.

### Inline ANSI styling breaks backgrounds — avoid it

Calling `lipgloss.Render()` on a substring and embedding the result inside a format string causes the background color to drop after the styled span. The reason: every `Render()` call closes with a full ANSI reset (`\033[m`), which clears any background that the outer `Render()` set — and lipgloss does not re-inject background codes mid-string.

**Do not do this:**
```go
bold := someStyle.Bold(true).Render(name)         // ends with \033[m
label := fmt.Sprintf("Merge %s into:", bold)      // " into:" loses background
outer := MenuBorderStyle.Render(label)             // too late — reset already fired
```

**Do this instead** — render the whole line as one unstyled string so the outer `Render()` applies the background uniformly:
```go
label := fmt.Sprintf("Merge %s into:", name)      // plain string, no ANSI
outer := MenuBorderStyle.Render(label)             // background applied cleanly
```

If a span genuinely needs a different color (not just bold), render each segment separately with the same background explicitly set on all of them.

### `MenuBorderStyle` is shared by all dialog overlays

Every dialog `View()` must return `ui.MenuBorderStyle.Render(body)` as its outermost render call. This ensures consistent border, padding, and background across confirm, amend, branch-picker, and context-menu dialogs.

### `PlaceOverlay` is ANSI-aware but not lipgloss-aware

`overlay.go` parses CSI escape sequences by byte scanning. It handles standard SGR codes but will silently corrupt output if other escape sequence types are introduced. Keep all styling within SGR/CSI.

## Adding new styles

Add to `styles.go`. Follow the existing pattern: colors as `var Color… = lipgloss.Color(…)`, derived styles as `var …Style = lipgloss.NewStyle()…`. Do not create styles inline in `View()` methods — styles defined inline are re-allocated on every render frame.

## Adding new strings

Add to `labels.go` under the appropriate section comment. Use `const` blocks, not `var`. Name format strings with a `Fmt` suffix (e.g. `ErrLoadCommitsFmt`).

## `Widths` and panel sizing

`Widths(termW, focus)` returns `[3]int` for `[branches, commits, detail]`. Base split is 22/35/43%; the focused panel gets a boost taken from the furthest panel. `InnerWidth(w)` subtracts the lipgloss border (1px each side) and padding (1 char each side) = 4 total.
