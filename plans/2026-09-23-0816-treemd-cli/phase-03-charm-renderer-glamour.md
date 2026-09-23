---
phase: 3
title: "Glamour Markdown Rendering Engine & Terminal Styling"
status: completed
priority: P1
effort: "25m"
dependencies: [1]
---

# Phase 3: Glamour Markdown Rendering Engine & Terminal Styling

## Overview
Integrate Charm's `glamour` library (`github.com/charmbracelet/glamour`) to render Markdown into ANSI-styled terminal output with custom word-wrapping, automatic dark/light background detection, and syntax-highlighted codeblocks.

## Requirements
- Functional:
  - Render raw Markdown strings or files into styled terminal text.
  - Dynamically wrap rendered content according to terminal or viewport width.
  - Detect dark vs light terminal backgrounds (or respect `--theme` flag with choices: `auto`, `dark`, `light`, `dracula`, `notty`).
- Non-functional:
  - Sub-millisecond render time for typical preview snippets (≤ 500 lines).
  - Preview size ceiling: limit preview rendering to first 64KB / 500 lines with `[Truncated preview — press Enter to view full document]` to avoid OOM on giant files.
  - Graceful fallback for non-UTF8 or malformed Markdown.

## Architecture
- `internal/render/glamour.go`: Wraps `glamour.TermRenderer` with options for width, style, preview line capping, and base URL.
- Exposes:
  - `Render(content string, width int, style string) (string, error)`
  - `RenderPreview(path string, width int, style string, maxLines int) (string, error)`
  - `RenderFile(path string, width int, style string) (string, error)`

## Related Code Files
- Create: `go-tools/treemd/internal/render/glamour.go`
- Create: `go-tools/treemd/internal/render/render_test.go`

## Implementation Steps

### Step 1: Add Glamour Dependency & Renderer Wrapper
- **Files**: `go-tools/treemd/internal/render/glamour.go`
- **Action**: Add `github.com/charmbracelet/glamour` and implement `Renderer` with dynamic width wrapping and preview size truncation limits.
- **Verification Command**: `export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH" && go test -v ./internal/render`
- **Expected Output**: PASS with ANSI escape sequence assertions for headings, bold, and codeblocks.

## Success Criteria
- [ ] Markdown documents render with beautiful Charm terminal styling.
- [ ] Dynamic width adjustments prevent awkward line breaks or overflow.
