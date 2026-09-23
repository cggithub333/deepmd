---
phase: 1
title: "Go Toolchain Setup & Project Scaffolding"
status: completed
priority: P1
effort: "15m"
dependencies: []
---

# Phase 1: Go Toolchain Setup & Project Scaffolding

## Overview
Install or verify the local Go compiler in `~/.local/go` (requiring zero sudo privileges), initialize the Go module `go-tools/treemd`, setup directory layout, and add foundational types.

## Requirements
- Functional: Ensure `go version` works cleanly in terminal, create Go module at `/home/james/temp/linux-utils/go-tools/treemd`.
- Non-functional: Zero sudo dependency; clean, idiomatic Go project structure (`cmd/treemd`, `internal/finder`, `internal/render`, `internal/tui`).

## Architecture
- Target directory: `go-tools/treemd`
- Module path: `github.com/cggithub333/linux-utils/go-tools/treemd`
- Directory layout:
  - `cmd/treemd/main.go` - Main entrypoint
  - `internal/finder/` - Concurrent file walker and fuzzy filter
  - `internal/render/` - Glamour markdown renderer
  - `internal/tui/` - Bubble Tea dual-pane and pager models

## Related Code Files
- Create: `go-tools/treemd/go.mod`
- Create: `go-tools/treemd/cmd/treemd/main.go`
- Create: `go-tools/treemd/internal/model/file.go`

## Implementation Steps

### Step 1: Ensure Go Toolchain in `~/.local/go`
- **Files**: `~/.local/go`, `~/.local/bin/go`
- **Action**: Download Go 1.23 linux-amd64 tarball, unpack into `~/.local/go`, and create symlinks in `~/.local/bin/go`.
- **Verification Command**: `export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH" && go version`
- **Expected Output**: `go version go1.23.* linux/amd64`

### Step 2: Initialize Go Module & Scaffolding
- **Files**: `go-tools/treemd/go.mod`, `go-tools/treemd/cmd/treemd/main.go`
- **Action**: Run `go mod init github.com/cggithub333/linux-utils/go-tools/treemd`, create basic directory structure, and add minimal entrypoint.
- **Verification Command**: `export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH" && go build -o /dev/null ./cmd/treemd`
- **Expected Output**: Exit code 0

## Success Criteria
- [ ] Go compiler is available in PATH.
- [ ] `go.mod` created with Go 1.23+ directive.
- [ ] Directory layout ready for core implementation.
