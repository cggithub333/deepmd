---
title: "treemd: High-Performance Single-Binary Markdown Explorer & Reader CLI"
status: completed
created: 2026-09-23
author: jameskit
phases:
  - id: 1
    title: "Go Toolchain Setup & Project Scaffolding"
    status: completed
  - id: 2
    title: "Pure Go File Discovery ('fd' max-depth 5), Fuzzy Matching ('fzf') & Content Search"
    status: completed
  - id: 3
    title: "Glamour Markdown Rendering Engine & Terminal Styling"
    status: completed
  - id: 4
    title: "Interactive Bubble Tea Dual-Pane TUI, Live Preview & Fullscreen Reader"
    status: completed
  - id: 5
    title: "CLI Entrypoint, Direct File/Dump Modes & Single-Binary Release"
    status: completed
---

# treemd: High-Performance Single-Binary Markdown Explorer & Reader CLI

## Executive Summary

`treemd` is an ultra-fast, zero-dependency, single-binary Go CLI tool designed to find, explore, preview, and read Markdown files directly in the terminal.

It brings together key capabilities tailored to the developer's exact workflow:
1. **Automatic Calling-Path Discovery with Max Depth 5**: Automatically scans the current directory where `treemd` was executed (or a specified path argument) for all markdown files (`.md`, `.markdown`, `.mdx`), pruning directory traversal strictly at **maximum depth 5** (`maxDepth = 5`) and skipping `.git`, `node_modules`, `.cache`, and `.gitignore` entries using a pure Go walker (`charlievieth/fastwalk`).
2. **Instant Traverse & Live Preview**: As the user traverses files in the list using arrow keys or `j`/`k`, the right pane immediately renders a live Glamour ANSI preview of the selected markdown document.
3. **Press Enter to View Inside**: Pressing `Enter` opens the document in full-screen reader mode (Bubble Tea viewport pager) with vim-like scrolling (`j`/`k`, `d`/`u`, `g`/`G`).
4. **Search Inside Content**: Supports searching both file paths and **full-text content inside markdown files** in the explorer, as well as `/` in-file search with `n`/`N` jumping while reading inside.
5. **Pure Go Packages (Zero External CLI Binaries)**: Powered directly by Junegunn Choi's official `junegunn/fzf/src/algo` and Charm's `glamour`, compiling into a single standalone static binary with zero external dependencies (no external `glow`, `fd`, or `fzf` installed in the OS environment required).

## Architecture Topology

```
+-------------------------------------------------------------------------+
|                              treemd CLI                                 |
+--------------------+---------------------+------------------------------+
| Calling Path / CWD | Auto-discovery      | Max Depth = 5                |
| CLI Modes          | `treemd`            | Dual-Pane Interactive TUI    |
|                    | `treemd <path>`     | Scoped Directory TUI         |
|                    | `treemd <file.md>`  | Direct Fullscreen Reader     |
|                    | `treemd -p <file>`  | Dump rendered ANSI to stdout |
+--------------------+---------------------+------------------------------+
                                   |
         +-------------------------+-------------------------+
         |                                                   |
         v                                                   v
+-------------------------------+           +-------------------------------+
|     Discovery & Search        |           |      Rendering & Display      |
+-------------------------------+           +-------------------------------+
| 1. Fast Walk (depth <= 5)     |           | 1. Glamour Markdown Engine    |
|    - Concurrent multi-worker  |           |    - Stylesheet (Dark/Light)  |
|    - Respects .gitignore      |           |    - Word-wrapping & Pygments |
| 2. Fuzzy Finder (`fzf/algo`)  |           | 2. Bubble Tea TUI Framework   |
|    - Exact Junegunn algorithm |           |    - Left: Fuzzy file list    |
|    - Substring score ranking  |           |    - Right: Live preview      |
| 3. In-File Content Search     |           |    - Enter: Fullscreen Reader |
|    - Grep inside markdown     |           |    - `/`: In-document search  |
+-------------------------------+           +-------------------------------+
```

## Phase Breakdown

- **Phase 1: Go Toolchain Setup & Project Scaffolding** — Establish Go environment, initialize `go.mod`, directory layout, and core type definitions.
- **Phase 2: Pure Go File Discovery ('fd' max-depth 5), Fuzzy Matching ('fzf') & Content Search** — Implement high-throughput `.md` traversal with max depth 5, Junegunn fuzzy algorithm, and multi-file content search.
- **Phase 3: Glamour Markdown Rendering Engine & Terminal Styling** — Configure Glamour renderer, dark/light theme detection, custom style definitions, and dump-to-stdout mode.
- **Phase 4: Interactive Bubble Tea Dual-Pane TUI, Live Preview & Fullscreen Reader** — Build responsive split-pane interface, instant live preview on traverse, `Enter` to view inside, and `/` in-file search.
- **Phase 5: CLI Entrypoint, Direct File/Dump Modes & Single-Binary Release** — Integrate Cobra/pflag CLI commands, compile self-contained static binary, and verify end-to-end performance.
