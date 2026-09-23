# BUG-01: Runtime Panic Slice Bounds Out Of Range `[:-9]`

## 💡 Feynman Mental Model & Intuitive Analogy
- 👤 Richard Feynman Insight
  - > *"If you cannot explain something in simple terms, you don't truly understand it."* — Richard P. Feynman
- 🏭 Physical Analogy: The Fixed-Length Cookie Cutter
  - Imagine a machine cutter programmed to always snip 9 inches off the end of a board before trimming
  - When given a brand-new board that is 0 inches long (or not measured yet)
  - The mechanical blade tries to reach behind the starting mark into negative space (`-9 inches`)
  - The gears jam and the entire factory line shuts down instantly
- 💡 Plain-English Aha!
  - `ListPane` starts with `Width: 0` before the terminal reports its size
  - When cutting string `line[:lp.Width-9]`, Go calculates `line[:-9]` which crashes immediately
- ⚙️ Step-by-Step Mechanics
  - 1. User launches `deepmd` in any terminal
  - 2. Bubble Tea starts runtime and triggers initial `View()` render frame
  - 3. `ListPane.Width` is uninitialized (`0`) because `tea.WindowSizeMsg` has not arrived yet
  - 4. Loop iterates over files: `len(line) > 0 - 6` (`true`)
  - 5. Expression evaluates `line[:0-9]` ➔ `line[:-9]` ➔ Go panics with `slice bounds out of range`
- ⚠️ Why Naive Fixes Fail
  - Only adding `if lp.Width > 0` leaves narrow terminals (<9 cols) vulnerable to negative bounds
  - Only fixing line 182 leaves content search or other components unaligned
- ⚖️ Brutally Honest Trade-off
  - Defensive string slicing helper requires slight refactor
  - But eliminates 100% of negative slice boundary panics across all UI states

## 🐛 Problem & Symptoms
- 💥 Immediate Runtime Crash
  - `panic: runtime error: slice bounds out of range [:-9]`
  - Application exits abruptly, restoring terminal state
- 📍 Failure Location
  - `internal/tui/list_pane.go:182` in `ListPane.View()`
  - Called from `internal/tui/app.go:204` in `App.View()`
- 🔍 Trigger Condition
  - Occurs on any repository launch where initial frame renders before `tea.WindowSizeMsg`
  - Reproducible whenever `ListPane.Width == 0` or `ListPane.Width < 9`

## 🎯 Impact Analysis
- 🚫 Zero Usability on Launch
  - Users cannot enter the interactive TUI
  - Immediate failure on first run in any directory
- 📉 Developer Experience
  - High friction; CLI feels fragile despite passing unit tests
  - Unit tests mocked `WindowSizeMsg` and never tested initial uninitialized frame render

## 🛠️ Proposed Solution Options
- 🥇 Option A: Comprehensive Defensive Helper & Default Size (Recommended)
  - 1. Safe Slicing: Create `truncateString(s string, maxWidth int)` with `maxWidth <= 3` edge cases
  - 2. Initial Dimensions: In `NewApp()`, invoke `list.SetSize(layout.LeftWidth-2, layout.LeftHeight)`
  - 3. Guard: Return full string or empty safely if `lp.Width < 10`
  - 4. Unit Test: Add regression test asserting `ListPane.View()` does not panic with `Width: 0`
  - ✅ Pros: Bulletproof against zero/negative/tiny terminal dimensions, zero chance of crash
  - ⚠️ Cons: 15 additional lines of defensive code
- 🥈 Option B: Inline Slicing Clamp Only
  - Simple inline fix: `if lp.Width > 9 { maxLen := max(0, lp.Width-9); line = line[:maxLen] + "..." }`
  - ✅ Pros: Quickest 3-line change
  - ⚠️ Cons: Does not initialize `ListPane.Width` at startup; displays squished/truncated dots on first frame
- 🥉 Option C: Defer Initial Frame Render
  - In `App.View()`, return "Loading..." if `a.Layout.Width == 0 || a.List.Width == 0`
  - ✅ Pros: Completely prevents rendering uninitialized layout
  - ⚠️ Cons: May cause a split-second flash of loading text on startup
