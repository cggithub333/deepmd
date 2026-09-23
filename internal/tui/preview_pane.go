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
		vpHeight = height - 1
		if vpHeight < 2 {
			vpHeight = 2
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

// StopSearch deactivates the search box and restores full preview height and raw unhighlighted content.
func (pp *PreviewPane) StopSearch() {
	pp.Searching = false
	pp.SearchInput.Blur()
	pp.MatchLines = nil
	pp.Viewport.SetContent(pp.RawContent)
	pp.SetDimensions(pp.Width, pp.Height)
}

func (pp *PreviewPane) applySearch(query string) {
	pp.MatchLines = nil
	pp.CurrentMatch = 0
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		pp.Viewport.SetContent(pp.RawContent)
		return
	}
	lowerQuery := strings.ToLower(trimmed)
	for idx, line := range pp.Lines {
		clean := stripANSI(line)
		if strings.Contains(strings.ToLower(clean), lowerQuery) {
			pp.MatchLines = append(pp.MatchLines, idx)
		}
	}

	pp.Viewport.SetContent(pp.renderHighlights(trimmed))
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

// NextMatch navigates to the next search occurrence and updates highlight states.
func (pp *PreviewPane) NextMatch() {
	if len(pp.MatchLines) == 0 {
		return
	}
	pp.CurrentMatch = (pp.CurrentMatch + 1) % len(pp.MatchLines)
	pp.Viewport.SetContent(pp.renderHighlights(pp.SearchInput.Value()))
	pp.scrollToCurrentMatch()
}

// PrevMatch navigates to the previous search occurrence and updates highlight states.
func (pp *PreviewPane) PrevMatch() {
	if len(pp.MatchLines) == 0 {
		return
	}
	pp.CurrentMatch = (pp.CurrentMatch - 1 + len(pp.MatchLines)) % len(pp.MatchLines)
	pp.Viewport.SetContent(pp.renderHighlights(pp.SearchInput.Value()))
	pp.scrollToCurrentMatch()
}

// renderHighlights produces content where matching query occurrences are highlighted.
// Active match on the target line is styled in bright magenta, other matches in gold/yellow.
func (pp *PreviewPane) renderHighlights(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" || len(pp.Lines) == 0 {
		return pp.RawContent
	}

	quoted := regexp.QuoteMeta(trimmed)
	re, err := regexp.Compile(`(?i)(\x1b\[[0-9;]*[a-zA-Z])|(` + quoted + `)`)
	if err != nil {
		return pp.RawContent
	}

	var sb strings.Builder
	normalHL := "\x1b[48;5;220m\x1b[38;5;16m\x1b[1m" // Gold/yellow bg, dark text, bold
	activeHL := "\x1b[48;5;201m\x1b[97m\x1b[1m"       // Bright magenta bg, white text, bold
	resetHL := "\x1b[0m"

	targetLine := -1
	if len(pp.MatchLines) > 0 && pp.CurrentMatch >= 0 && pp.CurrentMatch < len(pp.MatchLines) {
		targetLine = pp.MatchLines[pp.CurrentMatch]
	}

	for idx, line := range pp.Lines {
		if idx > 0 {
			sb.WriteString("\n")
		}

		clean := stripANSI(line)
		if !strings.Contains(strings.ToLower(clean), strings.ToLower(trimmed)) {
			sb.WriteString(line)
			continue
		}

		isTargetLine := (idx == targetLine)
		hlCode := normalHL
		if isTargetLine {
			hlCode = activeHL
		}

		replaced := re.ReplaceAllStringFunc(line, func(match string) string {
			if strings.HasPrefix(match, "\x1b[") {
				return match
			}
			return hlCode + match + resetHL
		})

		sb.WriteString(replaced)
	}

	return sb.String()
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
			case "tab", "enter":
				pp.NextMatch()
				return *pp, nil
			case "shift+tab", "shift+enter":
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "n":
			if len(pp.MatchLines) > 0 {
				pp.NextMatch()
				return *pp, nil
			}
		case "shift+tab", "N":
			if len(pp.MatchLines) > 0 {
				pp.PrevMatch()
				return *pp, nil
			}
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

	topRow := pp.buildSearchPill()
	return topRow + "\n" + pp.Viewport.View()
}

func (pp *PreviewPane) buildSearchPill() string {
	var matchBadge string
	if len(pp.MatchLines) > 0 {
		if pp.Width >= 48 {
			matchBadge = fmt.Sprintf("[%d/%d] [Tab: next]", pp.CurrentMatch+1, len(pp.MatchLines))
		} else {
			matchBadge = fmt.Sprintf("[%d/%d]", pp.CurrentMatch+1, len(pp.MatchLines))
		}
	} else if strings.TrimSpace(pp.SearchInput.Value()) != "" {
		if pp.Width >= 40 {
			matchBadge = "[0 matches]"
		} else {
			matchBadge = "[0]"
		}
	} else {
		if pp.Width >= 40 {
			matchBadge = "[grep]"
		} else {
			matchBadge = ""
		}
	}

	var boxText string
	if pp.Width >= 55 {
		boxText = fmt.Sprintf("  grep: %s %s │ [Esc] ", pp.SearchInput.View(), matchBadge)
	} else if pp.Width >= 28 {
		boxText = fmt.Sprintf("  grep: %s %s ", pp.SearchInput.View(), matchBadge)
	} else {
		boxText = fmt.Sprintf(" grep:%s", pp.SearchInput.View())
	}

	if pp.Width > 4 && len(stripANSI(boxText)) > pp.Width {
		boxText = TruncateString(boxText, pp.Width-2)
	}

	searchPill := lipgloss.NewStyle().
		Background(lipgloss.Color("#2E3440")).
		Foreground(lipgloss.Color("#ECEFF4")).
		Bold(true).
		Padding(0, 1).
		Render(boxText)

	topRow := lipgloss.NewStyle().Width(pp.Width).Align(lipgloss.Right).Render(searchPill)
	topLines := strings.Split(topRow, "\n")
	if len(topLines) > 0 {
		return topLines[0]
	}
	return topRow
}
