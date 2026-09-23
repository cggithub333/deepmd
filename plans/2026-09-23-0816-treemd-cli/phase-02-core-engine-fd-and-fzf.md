---
phase: 2
title: "Pure Go File Discovery ('fd'), Fuzzy Matching ('fzf') & Content Search"
status: completed
priority: P1
effort: "35m"
dependencies: [1]
---

# Phase 2: Pure Go File Discovery ('fd'), Fuzzy Matching ('fzf') & Content Search

## Overview
Implement high-performance multi-threaded file walking (replacing the external `fd` CLI with depth limit 5), pure Go fuzzy filtering (using `junegunn/fzf/src/algo`, replacing the external `fzf` CLI), and in-file content search (grep inside markdown files).

## Requirements
- Functional:
  - Fast recursive scan for markdown files (`.md`, `.markdown`, `.mdown`, `.mdx`) starting from the calling path.
  - **Max depth constraint**: Strict maximum depth of 5 levels (`maxDepth = 5`) from the root calling path.
  - Ignore hidden files/dirs (`.git`, `.cache`, `.vscode`), `node_modules`, `vendor`, and `.gitignore` entries.
  - Compute fuzzy matching score and highlight index ranges for query terms.
  - **Content Search Engine**: Scan file contents for query matches (full-text search inside markdown files), returning matching line snippets.
- Non-functional:
  - Zero external subprocess execution (`exec.Command("fd")` and `exec.Command("fzf")` are strictly prohibited).
  - Traversal speed exceeding 5,000 files/sec on local SSD.
  - Memory-efficient file caching for instant preview and search.

## Architecture
- `internal/finder/walker.go`: Uses `github.com/charlievieth/fastwalk` with depth tracking (`currentDepth <= 5`) and `sabhiram/go-gitignore` rule matcher.
- `internal/finder/fuzzy.go`: Integrates `github.com/junegunn/fzf/src/algo` (`algo.FuzzyMatchV2`) to compute score, start index, and end index for ranked path matching.
- `internal/finder/content.go`: Fast multi-goroutine content scanner matching query terms inside files with snippet extraction.

## Related Code Files
- Create: `go-tools/treemd/internal/finder/walker.go`
- Create: `go-tools/treemd/internal/finder/fuzzy.go`
- Create: `go-tools/treemd/internal/finder/content.go`
- Create: `go-tools/treemd/internal/finder/finder_test.go`

## Implementation Steps

### Step 1: Implement Concurrent File Walker with MaxDepth=5
- **Files**: `go-tools/treemd/internal/finder/walker.go`
- **Action**: Implement `WalkMarkdownFiles(root string, opts WalkOptions) ([]FileInfo, error)` where `opts.MaxDepth = 5` default, pruning directories when depth exceeds 5.
- **Verification Command**: `go test -v ./internal/finder -run TestWalkerMaxDepth`
- **Expected Output**: PASS (files beyond depth 5 are excluded).

### Step 2: Implement Pure Go FZF Fuzzy Scorer & Content Search
- **Files**: `go-tools/treemd/internal/finder/fuzzy.go`, `go-tools/treemd/internal/finder/content.go`
- **Action**: Implement `FuzzyFilter` using `junegunn/fzf/src/algo` and `SearchContent(files []FileInfo, query string) []ContentMatch` utilizing streaming `bufio.Scanner` and skipping files larger than 2MB to prevent memory exhaustion.
- **Verification Command**: `go test -v ./internal/finder -run TestFuzzyAndContent`
- **Expected Output**: PASS with verified fuzzy ranking and content snippet extraction.

## Success Criteria
- [ ] Directory walker discovers markdown files quickly from calling path with max depth = 5.
- [ ] Fuzzy finder produces accurate fzf-compatible rankings in pure Go.
- [ ] Content search locates text inside markdown files without external tools.
