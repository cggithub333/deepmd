---
phase: 4
title: "Interactive Bubble Tea Dual-Pane TUI, Live Preview & Content Search"
status: completed
priority: P1
effort: "45m"
dependencies: [2, 3]
---

# Phase 4: Interactive Bubble Tea Dual-Pane TUI, Live Preview & Content Search

## Overview
Develop the interactive terminal user interface using Charm's Bubble Tea (`tea.Model`), Bubbles (`viewport`, `textinput`), and Lip Gloss, featuring instant preview on file traverse, full-screen reader on `Enter`, and content searching.

## Requirements
- Functional:
  - **Dual-Pane Layout**:
    - **Left Pane (35-40% width)**:
      - Search query bar at top: live fuzzy path filter or toggle content search (`Ctrl+f` or `/` prefix).
      - Scrollable list of discovered markdown files (relative path, match highlights, match line preview if content search active).
      - Status bar: total files, matching count, active search mode (`PATH` vs `CONTENT`).
    - **Right Pane (60-65% width)**:
      - Live Glamour-rendered markdown preview that updates **immediately** as the user traverses up/down the file list.
  - **Traversal & Navigation**:
    - `Up` / `Down` or `k` / `j` or `Ctrl+p` / `Ctrl+n`: Traverse files; right preview instantly updates.
    - `Ctrl+f`: Toggle between searching file names/paths and searching **inside file contents**.
    - `Enter`: Open inside the file into full-screen interactive reader mode (Bubble Tea viewport pager).
  - **Full-Screen Reader Mode (Inside File)**:
    - Expanded full-width viewport rendered with Charm Glamour.
    - Keybindings: `j`/`k` or `Down`/`Up` to scroll, `d`/`u` for half-page, `g`/`G` for top/bottom.
    - In-file search: Press `/` to search within the document, jump to next (`n`) or previous (`N`) match with highlighted lines.
    - `Esc` or `Backspace`: Return cleanly to dual-pane file explorer.
    - `q` or `Ctrl+c`: Quit cleanly.
  - Dynamic window resize (`tea.WindowSizeMsg`) recalculating split widths and re-rendering Glamour markdown to fit exact terminal columns.

## Architecture
- `internal/tui/app.go`: Root Bubble Tea model managing state (`StateDualPane`, `StateReader`, `StateSearchMode`).
- `internal/tui/list_pane.go`: File list component with search input and live debounce for content search.
- `internal/tui/preview_pane.go`: Live preview viewport with Glamour async render caching.
- `internal/tui/reader_pane.go`: Fullscreen document viewer with in-document search and scroll handling.
- `internal/tui/styles.go`: Lip Gloss border, padding, and accent color definitions.

## Related Code Files
- Create: `go-tools/treemd/internal/tui/app.go`
- Create: `go-tools/treemd/internal/tui/list_pane.go`
- Create: `go-tools/treemd/internal/tui/preview_pane.go`
- Create: `go-tools/treemd/internal/tui/reader_pane.go`
- Create: `go-tools/treemd/internal/tui/styles.go`

## Implementation Steps

### Step 1: Create Lip Gloss Styles & Layout Calculator
- **Files**: `go-tools/treemd/internal/tui/styles.go`, `go-tools/treemd/internal/tui/app.go`
- **Action**: Define borders, colors (cyan, magenta, dim gray), responsive pane dimension calculations, and a minimum dimension guard (<60 cols or <15 rows) preventing layout panic.

### Step 2: Implement Dual-Pane Model with Async Debounced Preview
- **Files**: `go-tools/treemd/internal/tui/app.go`, `go-tools/treemd/internal/tui/list_pane.go`, `go-tools/treemd/internal/tui/preview_pane.go`
- **Action**: Handle keyboard navigation. On cursor move, dispatch asynchronous `tea.Cmd` returning `previewRenderedMsg { path string, content string, gen int64 }` with a 50ms debounce and generation counter to discard stale in-flight renders during rapid traversal.

### Step 3: Implement Content Search Toggle & In-File Search
- **Files**: `go-tools/treemd/internal/tui/list_pane.go`, `go-tools/treemd/internal/tui/reader_pane.go`
- **Action**: Support `Ctrl+f` to search inside file content, and `/` inside reader mode for in-file match navigation (`n`/`N`).

### Step 4: Implement Full-Screen Viewport Reader on Enter
- **Files**: `go-tools/treemd/internal/tui/reader_pane.go`, `go-tools/treemd/internal/tui/app.go`
- **Action**: Transition to full-screen viewport on `Enter`, hook vim-like keybindings, preserve active cursor index in the file list, and restore dual-pane seamlessly on `Esc`.

## Success Criteria
- [ ] Traversing files instantly updates the right preview pane with Glamour styling.
- [ ] Pressing `Enter` opens the file in full-screen reader mode with smooth scrolling.
- [ ] Users can search both file paths and inside markdown file content.
