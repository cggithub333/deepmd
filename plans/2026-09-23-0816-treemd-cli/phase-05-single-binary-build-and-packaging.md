---
phase: 5
title: "CLI Entrypoint, Direct File/Dump Modes & Single-Binary Release"
status: completed
priority: P1
effort: "20m"
dependencies: [4]
---

# Phase 5: CLI Entrypoint, Direct File/Dump Modes & Single-Binary Release

## Overview
Wire up CLI arguments and flags, implement direct file reading and stdout dump modes, compile a self-contained static single binary `treemd`, and copy/link to `linux-utils/treemd`.

## Requirements
- Functional:
  - `treemd`: Launch interactive TUI in current working directory.
  - `treemd <dir>`: Launch interactive TUI searching in specified directory.
  - `treemd <file.md>`: Open directly in full-screen reader.
  - `treemd -p <file.md>` / `treemd --dump <file.md>`: Print ANSI-rendered markdown to stdout (supports unix pipes e.g. `treemd -p README.md | grep ...`).
  - `treemd --version` / `treemd -h`: Standard help and version info.
- Non-functional:
  - Single standalone binary output (`CGO_ENABLED=0 go build -ldflags="-s -w"`).
  - Executable size under 15MB, zero shared library or dynamic runtime requirements.

## Architecture
- `cmd/treemd/main.go`: Flag parsing and execution dispatch.
- Build target: `bin/treemd` and symlink/copy to root `linux-utils/treemd`.

## Related Code Files
- Modify: `go-tools/treemd/cmd/treemd/main.go`
- Create: `go-tools/treemd/Makefile`

## Implementation Steps

### Step 1: Wire CLI Flags & Subcommand Modes
- **Files**: `go-tools/treemd/cmd/treemd/main.go`
- **Action**: Implement flag parsing for `--theme`, `--dump`, `--version`, and positional path arguments.

### Step 2: Build Single Static Binary
- **Files**: `go-tools/treemd/Makefile`, `.gitignore`
- **Action**: Add `/treemd` to root `.gitignore`. Compile single binary using `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=1.0.0" -o ../../treemd ./cmd/treemd`.
- **Verification Command**: `export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH" && ../../treemd --version`
- **Expected Output**: `treemd version 1.0.0`

### Step 3: End-to-End Verification
- **Files**: `linux-utils/treemd`
- **Action**: Run non-interactive headless smoke check: verify `--dump` / `-p` prints rendered ANSI markdown, test `--list`, and inspect binary linkage.
- **Verification Command**: `export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH" && ../../treemd -p ../../plans/2026-09-23-0816-treemd-cli/plan.md | head -n 10 && file ../../treemd`
- **Expected Output**: Formatted markdown lines and `statically linked` (or standalone executable ELF 64-bit).

## Success Criteria
- [ ] Compiled single binary `treemd` has zero external runtime dependencies.
- [ ] Direct file mode, dump mode, and directory mode function seamlessly.
