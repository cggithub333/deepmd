package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ansiStripRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiStripRegex.ReplaceAllString(s, "")
}

// PreviewPane manages the right-side live Glamour preview with in-preview grep.
type PreviewPane struct {
	Viewport     viewport.Model
	Width        int
	Height       int
	Loaded       bool
	RawContent   string
	Lines        []string
	Searching    bool
	SearchInput  textinput.Model
	MatchLines   []int
	CurrentMatch int
}

// NewPreviewPane creates a new PreviewPane component.
func NewPreviewPane(width, height int) PreviewPane {
	vp := viewport.New(width, height)
	vp.SetContent(DimItemStyle.Render("Select a file to preview..."))

	ti := textinput.New()
	ti.Placeholder = "grep text..."
	ti.Prompt = " "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5FD7")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ti.CharLimit = 40

	return PreviewPane{
		Viewport:     vp,
		Width:        width,
		Height:       height,
		Loaded:       false,
		SearchInput:  ti,
		Searching:    false,
		CurrentMatch: 0,
	}
}

// SetContent updates the viewport content and resets scroll to top.
func (pp *PreviewPane) SetContent(content string) {
	pp.RawContent = content
	pp.Lines = strings.Split(content, "\n")
	pp.Viewport.SetContent(content)
	pp.Viewport.GotoTop()
	pp.Loaded = true
	if pp.Searching && pp.SearchInput.Value() != "" {
		pp.applySearch(pp.SearchInput.Value())
	}
}

// SetDimensions updates the viewport dimensions.
func (pp *PreviewPane) SetDimensions(width, height int) {
	pp.Width = width
	pp.Height = height
	vpHeight := height
	if pp.Searching {
		vpHeight = height - 2
		if vpHeight < 3 {
			vpHeight = 3
		}
	}
	pp.Viewport.Width = width
	pp.Viewport.Height = vpHeight
}

// SetMouseWheelConfig configures mouse wheel behavior on the viewport.
func (pp *PreviewPane) SetMouseWheelConfig(enabled bool, delta int) {
	pp.Viewport.MouseWheelEnabled = enabled
	pp.Viewport.MouseWheelDelta = delta
}

// StartSearch activates the in-preview grep search box.
func (pp *PreviewPane) StartSearch() {
	pp.Searching = true
	pp.SetDimensions(pp.Width, pp.Height)
	pp.SearchInput.Focus()
	if pp.SearchInput.Value() != "" {
		pp.applySearch(pp.SearchInput.Value())
	}
}

// StopSearch deactivates the search box and restores full preview height.
func (pp *PreviewPane) StopSearch() {
	pp.Searching = false
	pp.SearchInput.Blur()
	pp.SetDimensions(pp.Width, pp.Height)
}

func (pp *PreviewPane) applySearch(query string) {
	pp.MatchLines = nil
	pp.CurrentMatch = 0
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return
	}
	lowerQuery := strings.ToLower(trimmed)
	for idx, line := range pp.Lines {
		clean := stripANSI(line)
		if strings.Contains(strings.ToLower(clean), lowerQuery) {
			pp.MatchLines = append(pp.MatchLines, idx)
		}
	}
	if len(pp.MatchLines) > 0 {
		pp.scrollToCurrentMatch()
	}
}

func (pp *PreviewPane) scrollToCurrentMatch() {
	if len(pp.MatchLines) == 0 {
		return
	}
	target := pp.MatchLines[pp.CurrentMatch]
	offset := target - (pp.Viewport.Height / 3)
	if offset < 0 {
		offset = 0
	}
	pp.Viewport.SetYOffset(offset)
}

// NextMatch navigates to the next search occurrence.
func (pp *PreviewPane) NextMatch() {
	if len(pp.MatchLines) == 0 {
		return
	}
	pp.CurrentMatch = (pp.CurrentMatch + 1) % len(pp.MatchLines)
	pp.scrollToCurrentMatch()
}

// PrevMatch navigates to the previous search occurrence.
func (pp *PreviewPane) PrevMatch() {
	if len(pp.MatchLines) == 0 {
		return
	}
	pp.CurrentMatch = (pp.CurrentMatch - 1 + len(pp.MatchLines)) % len(pp.MatchLines)
	pp.scrollToCurrentMatch()
}

// Update passes messages to the underlying viewport or search input.
func (pp *PreviewPane) Update(msg tea.Msg) (PreviewPane, tea.Cmd) {
	var cmd tea.Cmd

	if pp.Searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				pp.StopSearch()
				return *pp, nil
			case "enter", "n":
				pp.NextMatch()
				return *pp, nil
			case "N", "shift+enter":
				pp.PrevMatch()
				return *pp, nil
			}
			oldVal := pp.SearchInput.Value()
			pp.SearchInput, cmd = pp.SearchInput.Update(msg)
			if pp.SearchInput.Value() != oldVal {
				pp.applySearch(pp.SearchInput.Value())
			}
			return *pp, cmd
		}
	}

	pp.Viewport, cmd = pp.Viewport.Update(msg)
	return *pp, cmd
}

// View renders the preview pane with an optional top-right search box.
func (pp *PreviewPane) View() string {
	if !pp.Searching {
		return pp.Viewport.View()
	}

	// Build small search box on top right
	var matchBadge string
	if len(pp.MatchLines) > 0 {
		matchBadge = fmt.Sprintf("[%d/%d] [n/N]", pp.CurrentMatch+1, len(pp.MatchLines))
	} else if strings.TrimSpace(pp.SearchInput.Value()) != "" {
		matchBadge = "[0 matches]"
	} else {
		matchBadge = "[type to grep]"
	}

	boxText := fmt.Sprintf("  grep: %s %s │ [Esc] ", pp.SearchInput.View(), matchBadge)
	searchBox := lipgloss.NewStyle().
		Background(lipgloss.Color("#2E3440")).
		Foreground(lipgloss.Color("#ECEFF4")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF5FD7")).
		Bold(true).
		Render(boxText)

	topRow := lipgloss.NewStyle().Width(pp.Width).Align(lipgloss.Right).Render(searchBox)
	return topRow + "\n" + pp.Viewport.View()
}
