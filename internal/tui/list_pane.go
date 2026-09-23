package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cggithub333/deepmd/internal/finder"
	"github.com/cggithub333/deepmd/internal/model"
)

// SearchMode defines whether we are searching file paths or file content.
type SearchMode int

const (
	SearchModePath SearchMode = iota
	SearchModeContent
)

// ListPane manages the left-side file list and search filter.
type ListPane struct {
	TextInput      textinput.Model
	AllFiles       []model.FileInfo
	Filtered       []finder.MatchResult
	ContentMatches []finder.ContentMatch
	Cursor         int
	Offset         int
	Width          int
	Height         int
	Mode           SearchMode
}

// NewListPane creates a new ListPane component.
func NewListPane(files []model.FileInfo) ListPane {
	ti := textinput.New()
	ti.Placeholder = "Type to search files... (Ctrl+F to grep preview)"
	ti.Prompt = " "
	ti.PromptStyle = SearchPromptStyle
	ti.Focus()

	lp := ListPane{
		TextInput: ti,
		AllFiles:  files,
		Cursor:    0,
		Offset:    0,
		Mode:      SearchModePath,
	}
	lp.applyFilter("")
	return lp
}

// SetDimensions sets the dimensions of the ListPane and adapts input width & placeholder.
func (lp *ListPane) SetDimensions(width, height int) {
	lp.Width = width
	lp.Height = height
	lp.TextInput.Width = max(10, width-6)
	lp.updatePlaceholder()
}

func (lp *ListPane) updatePlaceholder() {
	if lp.Mode == SearchModePath {
		if lp.Width < 35 {
			lp.TextInput.Placeholder = "Search..."
		} else if lp.Width < 55 {
			lp.TextInput.Placeholder = "Search files..."
		} else {
			lp.TextInput.Placeholder = "Type to search files... (Ctrl+F for grep)"
		}
	} else {
		if lp.Width < 35 {
			lp.TextInput.Placeholder = "Content..."
		} else if lp.Width < 55 {
			lp.TextInput.Placeholder = "Search content..."
		} else {
			lp.TextInput.Placeholder = "Search text inside markdown files..."
		}
	}
}

func (lp *ListPane) applyFilter(query string) {
	if lp.Mode == SearchModePath {
		lp.Filtered = finder.FuzzyFilter(lp.AllFiles, query)
		if lp.Cursor >= len(lp.Filtered) {
			lp.Cursor = max(0, len(lp.Filtered)-1)
		}
	} else {
		matches, _ := finder.SearchContent(lp.AllFiles, query, 50)
		lp.ContentMatches = matches
		if lp.Cursor >= len(lp.ContentMatches) {
			lp.Cursor = max(0, len(lp.ContentMatches)-1)
		}
	}
}

// SelectedFile returns the currently highlighted file info.
func (lp *ListPane) SelectedFile() *model.FileInfo {
	if lp.Mode == SearchModePath {
		if len(lp.Filtered) == 0 || lp.Cursor < 0 || lp.Cursor >= len(lp.Filtered) {
			return nil
		}
		return &lp.Filtered[lp.Cursor].File
	}

	if len(lp.ContentMatches) == 0 || lp.Cursor < 0 || lp.Cursor >= len(lp.ContentMatches) {
		return nil
	}
	return &lp.ContentMatches[lp.Cursor].File
}

// ToggleMode toggles between path fuzzy search and full-text content search.
func (lp *ListPane) ToggleMode() {
	if lp.Mode == SearchModePath {
		lp.Mode = SearchModeContent
		lp.TextInput.Prompt = " "
	} else {
		lp.Mode = SearchModePath
		lp.TextInput.Prompt = " "
	}
	lp.updatePlaceholder()
	lp.Cursor = 0
	lp.Offset = 0
	lp.applyFilter(lp.TextInput.Value())
}

// Update handles keyboard messages for the search box and list navigation.
func (lp *ListPane) Update(msg tea.Msg) (ListPane, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "ctrl+p":
			if lp.Cursor > 0 {
				lp.Cursor--
				if lp.Cursor < lp.Offset {
					lp.Offset = lp.Cursor
				}
			}
			return *lp, nil

		case "down", "ctrl+n":
			maxItems := len(lp.Filtered)
			if lp.Mode == SearchModeContent {
				maxItems = len(lp.ContentMatches)
			}
			if lp.Cursor < maxItems-1 {
				lp.Cursor++
				visibleRows := lp.Height - 3
				if visibleRows <= 0 {
					visibleRows = 5
				}
				if lp.Cursor >= lp.Offset+visibleRows {
					lp.Offset = lp.Cursor - visibleRows + 1
				}
			}
			return *lp, nil
		}
	}

	prevVal := lp.TextInput.Value()
	lp.TextInput, cmd = lp.TextInput.Update(msg)
	if lp.TextInput.Value() != prevVal {
		lp.Cursor = 0
		lp.Offset = 0
		lp.applyFilter(lp.TextInput.Value())
	}

	return *lp, cmd
}

// View renders the list pane into a string.
func (lp *ListPane) View() string {
	var sb strings.Builder

	// Input bar
	sb.WriteString(lp.TextInput.View())
	sb.WriteString("\n")

	// Separator line
	divider := strings.Repeat("─", max(10, lp.Width-2))
	sb.WriteString(DimItemStyle.Render(divider))
	sb.WriteString("\n")

	visibleRows := lp.Height - 3
	if visibleRows <= 0 {
		visibleRows = 5
	}

	if lp.Mode == SearchModePath {
		if len(lp.Filtered) == 0 {
			sb.WriteString(DimItemStyle.Render("  No matching markdown files"))
			return sb.String()
		}

		end := lp.Offset + visibleRows
		if end > len(lp.Filtered) {
			end = len(lp.Filtered)
		}

		for i := lp.Offset; i < end; i++ {
			item := lp.Filtered[i]
			line := item.File.RelPath
			if lp.Width-6 > 0 {
				line = SmartTruncatePath(line, lp.Width-6)
			}

			if i == lp.Cursor {
				sb.WriteString(SelectedItemStyle.Render(fmt.Sprintf("  %s ", line)))
			} else {
				sb.WriteString(NormalItemStyle.Render(fmt.Sprintf("   %s", line)))
			}
			sb.WriteString("\n")
		}
	} else {
		if len(lp.ContentMatches) == 0 {
			sb.WriteString(DimItemStyle.Render("  No content matches found"))
			return sb.String()
		}

		end := lp.Offset + visibleRows
		if end > len(lp.ContentMatches) {
			end = len(lp.ContentMatches)
		}

		for i := lp.Offset; i < end; i++ {
			m := lp.ContentMatches[i]
			label := fmt.Sprintf("%s:%d", SmartTruncatePath(m.File.RelPath, max(12, lp.Width/2)), m.LineNum)
			snippet := strings.TrimSpace(m.LineText)
			maxSnippetW := lp.Width - len(label) - 6
			if maxSnippetW > 0 {
				snippet = TruncateString(snippet, maxSnippetW)
			} else if lp.Width > 0 {
				snippet = ""
			}

			fullLine := fmt.Sprintf("%s - %s", label, snippet)
			if lp.Width-6 > 0 {
				fullLine = TruncateString(fullLine, lp.Width-6)
			}
			if i == lp.Cursor {
				sb.WriteString(SelectedItemStyle.Render(fmt.Sprintf("  %s ", fullLine)))
			} else {
				sb.WriteString(NormalItemStyle.Render(fmt.Sprintf("   %s", fullLine)))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
