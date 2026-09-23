package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cggithub333/deepmd/internal/config"
	"github.com/cggithub333/deepmd/internal/finder"
	"github.com/cggithub333/deepmd/internal/model"
	"github.com/cggithub333/deepmd/internal/render"
)

// AppState denotes the current active screen mode.
type AppState int

const (
	StateDualPane AppState = iota
	StateReader
)

// FocusedPane denotes which pane has keyboard focus in DualPane mode.
type FocusedPane int

const (
	FocusList FocusedPane = iota
	FocusPreview
)

type previewRenderedMsg struct {
	path    string
	content string
	gen     int64
	err     error
}

type rescoutCompletedMsg struct {
	files []model.FileInfo
	err   error
}

// App is the root Bubble Tea application model.
type App struct {
	Files           []model.FileInfo
	Root            string
	State           AppState
	Focus           FocusedPane
	Rescouting      bool
	LeaderMode      bool
	DraggingDivider bool
	SplitRatio      float64
	CustomLeftWidth int
	Notification    string
	Config          config.Config
	List            ListPane
	Preview         PreviewPane
	Reader          ReaderPane
	Renderer        *render.Renderer
	Layout          LayoutDimensions
	PreviewGen      int64
	CurrentPath     string
}

// NewApp initializes the deepmd TUI model with default/persisted config.
func NewApp(files []model.FileInfo, root string) App {
	return NewAppWithConfig(files, root, config.LoadConfig())
}

// NewAppWithConfig initializes the deepmd TUI model with a specific configuration.
func NewAppWithConfig(files []model.FileInfo, root string, cfg config.Config) App {
	ratio := 0.60
	layout := CalculateLayoutWithRatio(80, 24, ratio)
	list := NewListPane(files)
	list.SetDimensions(layout.LeftWidth, layout.LeftHeight)
	preview := NewPreviewPane(layout.RightWidth-2, layout.RightHeight)
	preview.SetMouseWheelConfig(cfg.MouseWheelEnabled, cfg.MouseWheelDelta)
	reader := NewReaderPane(80, 24)
	reader.SetMouseWheelConfig(cfg.MouseWheelEnabled, cfg.MouseWheelDelta)

	theme := cfg.Theme
	if theme == "" {
		theme = "dark"
	}

	return App{
		Files:           files,
		Root:            root,
		State:           StateDualPane,
		Focus:           FocusList,
		Rescouting:      false,
		LeaderMode:      false,
		DraggingDivider: false,
		SplitRatio:      ratio,
		CustomLeftWidth: layout.LeftWidth,
		Notification:    "",
		Config:          cfg,
		List:            list,
		Preview:         preview,
		Reader:          reader,
		Renderer:        render.NewRenderer(theme),
		Layout:          layout,
		PreviewGen:      0,
	}
}

// Init triggers initial preview render and background cache refresh.
func (a App) Init() tea.Cmd {
	var cmds []tea.Cmd
	selected := a.List.SelectedFile()
	if selected != nil {
		a.CurrentPath = selected.Path
		cmds = append(cmds, a.dispatchPreview(selected.Path))
	}
	cmds = append(cmds, a.triggerRescout())
	return tea.Batch(cmds...)
}

func (a *App) dispatchPreview(path string) tea.Cmd {
	a.PreviewGen++
	gen := a.PreviewGen
	width := a.Layout.RightWidth - 4
	if width < 20 {
		width = 40
	}

	return func() tea.Msg {
		rendered, err := a.Renderer.RenderPreview(path, width, 500)
		return previewRenderedMsg{
			path:    path,
			content: rendered,
			gen:     gen,
			err:     err,
		}
	}
}

func (a *App) triggerRescout() tea.Cmd {
	root := a.Root
	return func() tea.Msg {
		opts := finder.DefaultWalkOptions(root)
		files, err := finder.WalkMarkdownFiles(opts)
		if err == nil {
			_ = finder.SaveCache(root, files)
		}
		return rescoutCompletedMsg{files: files, err: err}
	}
}

func (a *App) updateLayoutWithWidth(leftWidth int) {
	if a.Layout.TotalWidth > 0 {
		a.SplitRatio = float64(leftWidth) / float64(a.Layout.TotalWidth)
	}
	a.Layout = CalculateLayoutWithRatio(a.Layout.TotalWidth, a.Layout.TotalHeight, a.SplitRatio)
	a.CustomLeftWidth = a.Layout.LeftWidth
	if !a.Layout.TooSmall {
		a.List.SetDimensions(a.Layout.LeftWidth, a.Layout.LeftHeight)
		a.Preview.SetDimensions(a.Layout.RightWidth, a.Layout.RightHeight)
		a.Reader.SetDimensions(a.Layout.TotalWidth, a.Layout.TotalHeight)
	}
}

// AdjustSplit adjusts the width of the left pane by delta and dispatches preview re-render.
func (a *App) AdjustSplit(delta int) tea.Cmd {
	current := a.Layout.LeftWidth
	if current <= 0 {
		current = int(float64(a.Layout.TotalWidth) * a.SplitRatio)
	}
	a.updateLayoutWithWidth(current + delta)

	selected := a.List.SelectedFile()
	if selected != nil {
		return a.dispatchPreview(selected.Path)
	}
	return nil
}

// Update coordinates messages across list, preview, and reader panes.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if a.SplitRatio <= 0.05 || a.SplitRatio >= 0.95 {
			a.SplitRatio = 0.60
		}
		a.Layout = CalculateLayoutWithRatio(msg.Width, msg.Height, a.SplitRatio)
		a.CustomLeftWidth = a.Layout.LeftWidth
		if !a.Layout.TooSmall {
			a.List.SetDimensions(a.Layout.LeftWidth, a.Layout.LeftHeight)
			a.Preview.SetDimensions(a.Layout.RightWidth, a.Layout.RightHeight)
			a.Reader.SetDimensions(a.Layout.TotalWidth, a.Layout.TotalHeight)

			selected := a.List.SelectedFile()
			if selected != nil {
				cmds = append(cmds, a.dispatchPreview(selected.Path))
			}
		}
		return a, tea.Batch(cmds...)

	case rescoutCompletedMsg:
		a.Rescouting = false
		if msg.err == nil && len(msg.files) > 0 {
			a.Files = msg.files
			a.List.AllFiles = msg.files
			a.List.applyFilter(a.List.TextInput.Value())
		}
		return a, nil

	case previewRenderedMsg:
		if msg.gen == a.PreviewGen {
			if msg.err != nil {
				a.Preview.SetContent(fmt.Sprintf("Error rendering preview:\n%v", msg.err))
			} else {
				a.Preview.SetContent(msg.content)
			}
		}
		return a, nil

	case tea.MouseMsg:
		mouseEv := tea.MouseEvent(msg)

		// 1. Wheel scroll handling: strictly disabled in DualPane mode!
		// Trackpad 2-finger inertia or wheel clicks will NEVER scroll or jump through the file list.
		if mouseEv.IsWheel() {
			if a.State == StateDualPane {
				return a, nil
			}
			if a.State == StateReader && a.Config.MouseWheelEnabled {
				var cmd tea.Cmd
				a.Reader, cmd = a.Reader.Update(msg)
				return a, cmd
			}
			return a, nil
		}

		// 2. Mouse dragging of the divider between List and Preview
		if a.State == StateDualPane {
			dividerX := a.Layout.LeftWidth + 2

			// Release mouse button: stop dragging and refresh preview width
			if mouseEv.Action == tea.MouseActionRelease || (mouseEv.Button == tea.MouseButtonNone && a.DraggingDivider) {
				if a.DraggingDivider {
					a.DraggingDivider = false
					selected := a.List.SelectedFile()
					if selected != nil {
						return a, a.dispatchPreview(selected.Path)
					}
					return a, nil
				}
			}

			// Press mouse button: check if divider dragging or tab/pane click focus
			if mouseEv.Action == tea.MouseActionPress && mouseEv.Button == tea.MouseButtonLeft {
				if mouseEv.X >= dividerX-2 && mouseEv.X <= dividerX+2 {
					a.DraggingDivider = true
					return a, nil
				}

				// Click on Left (File Explorer) Pane: switch focus to List
				if mouseEv.X < dividerX-2 {
					if a.Focus != FocusList {
						a.Focus = FocusList
						a.List.TextInput.Focus()
						if a.Preview.Searching {
							a.Preview.StopSearch()
						}
					}

					// Click directly on a file item in the list
					if mouseEv.Y >= 4 && mouseEv.Y < 4+a.Layout.LeftHeight-3 {
						clickedIdx := a.List.Offset + (mouseEv.Y - 4)
						maxItems := len(a.List.Filtered)
						if a.List.Mode == SearchModeContent {
							maxItems = len(a.List.ContentMatches)
						}
						if clickedIdx >= 0 && clickedIdx < maxItems {
							oldSel := a.List.SelectedFile()
							a.List.Cursor = clickedIdx
							newSel := a.List.SelectedFile()
							if newSel != nil && (oldSel == nil || oldSel.Path != newSel.Path) {
								a.CurrentPath = newSel.Path
								return a, a.dispatchPreview(newSel.Path)
							}
						}
					}
					return a, nil
				}

				// Click on Right (Preview) Pane: switch focus to Preview
				if mouseEv.X > dividerX+2 {
					if a.Focus != FocusPreview {
						a.Focus = FocusPreview
						a.List.TextInput.Blur()
					}
					return a, nil
				}
			}

			// Motion while dragging: slide divider left <-> right
			if a.DraggingDivider {
				newLeft := mouseEv.X - 2
				a.updateLayoutWithWidth(newLeft)
				return a, nil
			}
		}

		if a.State == StateReader {
			var cmd tea.Cmd
			a.Reader, cmd = a.Reader.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		// Clear previous notification on any new keypress (except when generating a new one)
		if a.Notification != "" && msg.String() != "ctrl+c" && msg.String() != "ctrl+a" {
			a.Notification = ""
		}

		// Leader Key: Alt+Q toggles Leader Mode
		if msg.String() == "alt+q" || msg.String() == "alt+Q" {
			if a.State == StateDualPane {
				a.LeaderMode = !a.LeaderMode
				return a, nil
			}
		}

		// Active Leader Mode controls
		if a.State == StateDualPane && a.LeaderMode {
			switch msg.String() {
			// Leader + left / h: switch view to File Explorer
			case "left", "h":
				a.Focus = FocusList
				a.List.TextInput.Focus()
				a.LeaderMode = false
				return a, nil

			// Leader + right / l: switch view to Preview Box
			case "right", "l":
				a.Focus = FocusPreview
				a.List.TextInput.Blur()
				a.LeaderMode = false
				return a, nil

			// Leader + Alt+Left / Alt+h: resize divider left (narrows list, widens preview)
			case "alt+left", "alt+h":
				return a, a.AdjustSplit(-2)

			// Leader + Alt+Right / Alt+l: resize divider right (widens list, narrows preview)
			case "alt+right", "alt+l":
				return a, a.AdjustSplit(2)

			// Esc / Enter / q: exit LeaderMode
			case "esc", "enter", "q":
				a.LeaderMode = false
				return a, nil
			}
		}

		// Direct resize shortcuts anytime in DualPane
		if a.State == StateDualPane {
			switch msg.String() {
			case "alt+left", "alt+h":
				return a, a.AdjustSplit(-2)
			case "alt+right", "alt+l":
				return a, a.AdjustSplit(2)
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case "ctrl+p", "ctrl+y":
			// Copy filepath to clipboard
			var path string
			if a.State == StateDualPane {
				selected := a.List.SelectedFile()
				if selected != nil {
					path = selected.Path
				}
			} else if a.State == StateReader && a.Reader.File != nil {
				path = a.Reader.File.Path
			}
			if path != "" {
				_ = CopyToClipboard(path)
				a.Notification = fmt.Sprintf(" 󰅍 Copied path: %s ", filepath.Base(path))
			}
			return a, nil

		case "ctrl+a":
			// Ctrl+A copies raw markdown file content (for both file explorer focus and preview focus)
			var targetPath string
			if a.State == StateDualPane {
				selected := a.List.SelectedFile()
				if selected != nil {
					targetPath = selected.Path
				}
			} else if a.State == StateReader && a.Reader.File != nil {
				targetPath = a.Reader.File.Path
			}
			if targetPath != "" {
				data, err := os.ReadFile(targetPath)
				if err == nil {
					_ = CopyToClipboard(string(data))
					a.Notification = fmt.Sprintf(" 󰅍 Copied content: %s (%d bytes) ", filepath.Base(targetPath), len(data))
				}
			}
			return a, nil

		case "ctrl+q":
			return a, tea.Quit

		case "ctrl+r":
			if a.State == StateDualPane {
				a.Rescouting = true
				return a, a.triggerRescout()
			}

		case "ctrl+f":
			if a.State == StateDualPane {
				if a.Preview.Searching {
					a.Preview.StopSearch()
					a.Focus = FocusList
					a.List.TextInput.Focus()
				} else {
					a.Focus = FocusPreview
					a.List.TextInput.Blur()
					a.Preview.StartSearch()
				}
				return a, nil
			}

		case "q":
			if a.State == StateReader {
				a.State = StateDualPane
				return a, nil
			}
			if a.State == StateDualPane && a.Focus == FocusPreview && !a.Preview.Searching {
				a.Focus = FocusList
				a.List.TextInput.Focus()
				return a, nil
			}
			// Do NOT quit on q!

		case "esc":
			if a.State == StateReader {
				a.State = StateDualPane
				return a, nil
			}
			if a.State == StateDualPane && a.Focus == FocusPreview {
				if a.Preview.Searching {
					a.Preview.StopSearch()
					return a, nil
				}
				a.Focus = FocusList
				a.List.TextInput.Focus()
				return a, nil
			}

		case "backspace":
			if a.State == StateReader {
				a.State = StateDualPane
				return a, nil
			}
			// In DualPane mode, do not intercept backspace so TextInput / SearchInput can delete characters

		case "enter":
			if a.State == StateDualPane {
				// Typing exit / :exit / quit / :quit / :q in search input exits
				query := strings.TrimSpace(strings.ToLower(a.List.TextInput.Value()))
				if query == "exit" || query == ":exit" || query == "quit" || query == ":quit" || query == ":q" {
					return a, tea.Quit
				}

				if a.Focus == FocusPreview && a.Preview.Searching {
					previewQuery := strings.TrimSpace(strings.ToLower(a.Preview.SearchInput.Value()))
					if previewQuery == "exit" || previewQuery == ":exit" || previewQuery == "quit" || previewQuery == ":quit" || previewQuery == ":q" {
						return a, tea.Quit
					}
					a.Preview.NextMatch()
					return a, nil
				}

				selected := a.List.SelectedFile()
				if selected != nil {
					width := a.Layout.TotalWidth - 4
					if width < 20 {
						width = 80
					}
					fullContent, err := a.Renderer.RenderFile(selected.Path, width)
					if err != nil {
						fullContent = fmt.Sprintf("Error loading file:\n%v", err)
					}
					a.Reader.LoadFile(selected, fullContent)
					a.State = StateReader
					return a, nil
				}
			}
		}
	}

	if a.State == StateReader {
		var cmd tea.Cmd
		a.Reader, cmd = a.Reader.Update(msg)
		return a, cmd
	}

	// Dual pane: when Preview is focused
	if a.Focus == FocusPreview {
		if a.Preview.Searching {
			var prevCmd tea.Cmd
			a.Preview, prevCmd = a.Preview.Update(msg)
			return a, prevCmd
		}

		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "/" {
			a.Preview.StartSearch()
			return a, nil
		}

		var prevCmd tea.Cmd
		a.Preview, prevCmd = a.Preview.Update(msg)
		return a, prevCmd
	}

	// Dual pane: when List is focused, route to List
	oldSelected := a.List.SelectedFile()
	var listCmd tea.Cmd
	a.List, listCmd = a.List.Update(msg)
	cmds = append(cmds, listCmd)

	newSelected := a.List.SelectedFile()
	if newSelected != nil && (oldSelected == nil || oldSelected.Path != newSelected.Path) {
		a.CurrentPath = newSelected.Path
		cmds = append(cmds, a.dispatchPreview(newSelected.Path))
	}

	return a, tea.Batch(cmds...)
}

// View renders either the full-screen reader or the dual-pane layout.
func (a App) View() string {
	if a.Layout.TooSmall {
		return WarningStyle.Render("\n    Terminal window too small (min 60x14). Please enlarge.")
	}

	if a.State == StateReader {
		return a.Reader.View()
	}

	var sb strings.Builder

	// Top Title Bar
	title := HeaderStyle.Render("   deepmd — Markdown Explorer & Reader ")
	infoBadge := StatusStyle.Render(fmt.Sprintf("[%d files]", len(a.Files)))
	topHeader := lipgloss.JoinHorizontal(lipgloss.Top, title, infoBadge)

	if a.Notification != "" {
		notifBadge := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FFAF")).
			Padding(0, 1).
			Render(a.Notification)
		topHeader = lipgloss.JoinHorizontal(lipgloss.Top, topHeader, notifBadge)
	}

	if a.LeaderMode {
		leaderBadge := ResizeHeaderStyle.Render(" 󰩨 LEADER: [←/→] Switch View │ [Alt+←/→] Resize │ [Esc] Exit ")
		topHeader = lipgloss.JoinHorizontal(lipgloss.Top, topHeader, leaderBadge)
	}
	sb.WriteString(topHeader)
	sb.WriteString("\n")

	// Main Dual-Pane Section with dynamic focus borders
	var leftBorder, rightBorder lipgloss.Style
	if a.LeaderMode {
		leftBorder = ResizeBorderStyle
		rightBorder = ResizeBorderStyle
	} else if a.Focus == FocusPreview {
		leftBorder = InactiveBorderStyle
		rightBorder = ActiveBorderStyle
	} else {
		leftBorder = ActiveBorderStyle
		rightBorder = InactiveBorderStyle
	}

	leftView := leftBorder.Width(a.Layout.LeftWidth).Height(a.Layout.LeftHeight).Render(a.List.View())
	rightView := rightBorder.Width(a.Layout.RightWidth).Height(a.Layout.RightHeight).Render(a.Preview.View())

	mainPanes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	sb.WriteString(mainPanes)
	sb.WriteString("\n")

	// Bottom Status / Help Bar
	var helpText string
	if a.LeaderMode {
		helpText = StatusStyle.Render(" [←/→] Switch Pane │ [Alt+←/→] Resize Divider │ [Enter/Esc] Done ")
	} else if a.Focus == FocusPreview {
		if a.Preview.Searching {
			helpText = StatusStyle.Render(" [Tab/Enter] Next match │ [Shift+Tab] Prev match │ [Esc] Close search ")
		} else {
			helpText = StatusStyle.Render(" [Ctrl+F or /] Grep Preview │ [Tab/n] Next match │ [Ctrl+A] Copy Content │ [j/k/d/u] Scroll │ [Esc] Back to List │ [Ctrl+C] Exit ")
		}
	} else {
		rescoutNotice := ""
		if a.Rescouting {
			rescoutNotice = " [Rescouting...] │"
		}
		helpText = StatusStyle.Render(fmt.Sprintf(" [Click Pane / Leader+←/→] Switch Focus │ [Drag Divider / Leader+Alt+←/→] Resize │ [Ctrl+P] Copy Path │ [Ctrl+A] Copy Content │ [Ctrl+F] Grep %s│ [Ctrl+C / exit] Exit ", rescoutNotice))
	}
	sb.WriteString(helpText)

	return sb.String()
}
