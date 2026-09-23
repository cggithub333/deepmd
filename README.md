# deepmd 

> Fast, Single-Binary Markdown Explorer & Reader CLI powered by Charmbracelet Glamour, Bubble Tea, and pure Go implementations of `fd` and `fzf`.

---

## 󰈙 Features

- **Pure Go File Discovery**: Automatically scans markdown files (`.md`, `.markdown`, `.mdown`, `.mdx`) starting from the calling path up to a max depth of 5 (configurable via `--depth`). Fully respects `.gitignore` rules and ignores common vendor/cache directories (`.git`, `node_modules`, `.cache`, `vendor`, `.vscode`). Zero external `fd` dependency.
- **Pure Go Fuzzy Finder**: Sub-millisecond interactive fuzzy path search powered by Junegunn's `fzf` algorithm (`junegunn/fzf/src/algo`). Zero external `fzf` dependency.
- **Instant Dual-Pane Preview**: Glamour terminal markdown preview renders on the right pane in real time as you navigate through files on the left pane. Asynchronous rendering with debounce tokens ensures zero UI stutter and drops stale renders.
- **Full-Screen Reader View**: Press `Enter` to switch into full-screen interactive reader mode (Bubble Tea viewport pager) with vim navigation (`j`/`k`, `d`/`u`, `g`/`G`).
- **In-File Content Grep**:
  - In explorer mode: toggle `Ctrl+f` to search inside file contents.
  - In reader mode: press `/` to search within the open document with `n`/`N` jumping.
- **Direct Terminal Dump Mode**: Use `-p` or `--dump` to render and dump styled ANSI markdown directly to stdout (compatible with pipes and standard pagers like `less -R`).
- **Border-Free Shaded Code Blocks**: Markdown code blocks (` ```...``` `) are displayed inside a modern, borderless card with a solid dark gray background, language badge, and vibrant Chroma syntax highlighting with auto-wrapping for long lines.
- **Single-Binary Zero Dependencies**: Statically compiled binary (`CGO_ENABLED=0`) with zero shared library or external runtime requirements.

---

## 󰐥 Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/cggithub333/deepmd.git
cd deepmd
make install    # Builds and installs to ~/.local/bin/deepmd

# Or build standalone binary directly
make build
./bin/deepmd
```

### CLI Usage

```bash
# 1. Interactive Explorer (Dual-Pane TUI) in current directory
deepmd

# 2. Interactive Explorer under a specific directory with custom depth
deepmd ./docs --depth 3

# 3. Direct Full-Screen Reader on a single file
deepmd README.md

# 4. Dump ANSI rendered markdown to stdout / terminal pager
deepmd -p README.md | less -R

# 5. List all discovered markdown files
deepmd -l

# 6. Enable smooth mouse wheel scrolling (or configure in ~/.deepmd/config.json)
deepmd --enable-mouse --mouse-delta 1
```

---

## 󰒓 Configuration (`~/.deepmd/config.json`)

`deepmd` automatically creates and persists user preferences in `~/.deepmd/config.json`:

```json
{
  "mouse_wheel_enabled": false,
  "mouse_wheel_delta": 1,
  "theme": "dark"
}
```

- **`mouse_wheel_enabled`** (`false` by default): Disables terminal mouse wheel events to prevent terminal emulators from injecting erratic synthetic arrow key bursts.
- **`mouse_wheel_delta`** (`1` by default): Lines scrolled per notch for ultra-smooth, slow scrolling when mouse wheel is enabled.
- **`theme`** (`"dark"` by default): Standard Glamour styling (`dark`, `light`, `dracula`, `notty`, `auto`).

### CLI Overrides
- `--mouse` / `--enable-mouse`: Enable mouse wheel scrolling (automatically saved to config).
- `--no-mouse` / `--disable-mouse`: Disable mouse wheel scrolling (default).
- `--mouse-delta <N>` / `--mouse-speed <N>`: Adjust scroll speed to N lines per notch.


---

## 󰌌 Keybindings

### Dual-Pane Explorer Mode
| Key / Gesture | Action |
| :--- | :--- |
| **Mouse Click on File Explorer** | Focus the file explorer pane; click any file item to directly select and preview |
| **Mouse Click on Preview Box** | Focus the preview pane for scrolling or searching |
| **Mouse Click & Drag Divider** | Press on the middle border between List and Preview to slide the split left <-> right |
| `Alt+q` / `Alt+Q` | Tmux-like leader chord to toggle **Leader Mode** (`←`/`→` to switch view, `Alt+←`/`Alt+→` to resize, `Enter`/`Esc` to exit) |
| `Leader` then `←` / `h` | Switch active focus to **File Explorer** |
| `Leader` then `→` / `l` | Switch active focus to **Preview Box** |
| `Leader` then `Alt+←` / `Alt+h` | Nudge divider left (shrinks file list, widens preview box) |
| `Leader` then `Alt+→` / `Alt+l` | Nudge divider right (widens file list, narrows preview box) |
| `Ctrl+p` / `Ctrl+y` | **Copy Full Absolute Filepath** to system clipboard (when focused on file explorer) |
| `Ctrl+a` | **Copy Full Content** to system clipboard (works for both file explorer and preview focus) |
| `Ctrl+f` / `/` | Open / close in-preview grep search pill on the top right of Preview box |
| `Tab` / `Enter` | **Traverse to next match** in preview grep (active match in magenta, other matches in gold/yellow) |
| `Shift+Tab` / `Shift+Enter` | **Traverse to previous match** in preview grep |
| `Backspace` | Edit search input (deletes characters without switching pane focus) |
| `n` / `N` | Next / previous match navigation (when preview is focused without active search input) |
| `↑` / `k` | Move cursor up in file list (or scroll up when Preview focused) |
| `↓` / `j` | Move cursor down in file list (or scroll down when Preview focused) |
| `d` / `u` | Scroll half-page down / up when Preview focused |
| `Enter` | Open selected file in full-screen reader (or exit if search query is `exit` / `quit` / `:q`) |
| `Ctrl+r` | Rescan directory and refresh ScoutCache (`~/.deepmd/<timestamp>/md-scout.md`) |
| `Esc` | Close preview search box or return focus to file list |
| `Ctrl+c` / `Ctrl+q` / `exit` | **Exit deepmd immediately** (`q` does not exit) |

> [!NOTE]
> **Touchpad & Mouse Wheel Scrolling**: Wheel and 2-finger trackpad gestures smoothly scroll the **Preview Box** (and Full-Screen Reader). Meanwhile, scrolling over the **File Explorer** list remains strictly silenced to prevent accidental inertia jumps across files.


### Full-Screen Reader Mode
| Key | Action |
| :--- | :--- |
| `j` / `↓` | Scroll down 1 line |
| `k` / `↑` | Scroll up 1 line |
| `d` / `Ctrl+d` | Half page down |
| `u` / `Ctrl+u` | Half page up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `/` | Search inside document |
| `n` | Next search match |
| `N` | Previous search match |
| `Esc` / `Backspace` / `q` | Return to dual-pane explorer (preserves cursor position) |

---

## 󰒋 Architecture

```
deepmd/
├── cmd/deepmd/main.go          # CLI entrypoint, flags (-p, -l, -depth, -theme, -mouse)
├── internal/
│   ├── config/config.go        # Config persistence (~/.deepmd/config.json)
│   ├── model/file.go           # FileInfo data structure
│   ├── finder/
│   │   ├── walker.go           # Pure Go walker (fastwalk + go-gitignore, max depth 5)
│   │   ├── fuzzy.go            # Pure Go FZF algo (junegunn/fzf/src/algo)
│   │   ├── content.go          # Streaming in-file content search (bufio.Scanner, 2MB ceiling)
│   │   └── cache.go            # ScoutCache with human-readable md & json indexes
│   ├── render/
│   │   └── glamour.go          # Glamour markdown renderer with truncation guards
│   └── tui/
│       ├── styles.go           # Lip Gloss themes, layout split calculations, & size guards
│       ├── list_pane.go        # Fast file fuzzy finder pane
│       ├── preview_pane.go     # Live preview pane with in-preview grep search
│       ├── reader_pane.go      # Full-screen pager with vim keys & / search
│       └── app.go              # Root tea.Model with mouse divider drag, Alt+Q resize
└── Makefile                    # CGO_ENABLED=0 static build pipeline
```
