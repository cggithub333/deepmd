package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cggithub333/deepmd/internal/model"
)

// ReaderPane provides a full-screen scrollable reader for a markdown file.
type ReaderPane struct {
	Viewport viewport.Model
	File     *model.FileInfo
	Content  string
	Width    int
	Height   int
}

// NewReaderPane creates an empty full-screen reader.
func NewReaderPane(width, height int) ReaderPane {
	vp := viewport.New(width, height-2)
	return ReaderPane{
		Viewport: vp,
		Width:    width,
		Height:   height,
	}
}

// LoadFile sets the document content and file metadata.
func (rp *ReaderPane) LoadFile(file *model.FileInfo, renderedContent string) {
	rp.File = file
	rp.Content = renderedContent
	rp.Viewport.SetContent(renderedContent)
	rp.Viewport.GotoTop()
}

// SetDimensions updates the full-screen reader viewport dimensions.
func (rp *ReaderPane) SetDimensions(width, height int) {
	rp.Width = width
	rp.Height = height
	rp.Viewport.Width = width
	rp.Viewport.Height = height - 2
}

// SetMouseWheelConfig configures mouse wheel behavior on the reader viewport.
func (rp *ReaderPane) SetMouseWheelConfig(enabled bool, delta int) {
	rp.Viewport.MouseWheelEnabled = enabled
	rp.Viewport.MouseWheelDelta = delta
}

// Update handles vim-like scrolling and in-file navigation.
func (rp *ReaderPane) Update(msg tea.Msg) (ReaderPane, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			rp.Viewport.LineDown(1)
			return *rp, nil
		case "k", "up":
			rp.Viewport.LineUp(1)
			return *rp, nil
		case "d", "ctrl+d":
			rp.Viewport.HalfViewDown()
			return *rp, nil
		case "u", "ctrl+u":
			rp.Viewport.HalfViewUp()
			return *rp, nil
		case "g":
			rp.Viewport.GotoTop()
			return *rp, nil
		case "G":
			rp.Viewport.GotoBottom()
			return *rp, nil
		}
	}

	rp.Viewport, cmd = rp.Viewport.Update(msg)
	return *rp, cmd
}

// View renders the reader with header bar and scroll percentage.
func (rp *ReaderPane) View() string {
	if rp.File == nil {
		return "No document loaded"
	}

	var sb strings.Builder

	// Top header
	pct := int(rp.Viewport.ScrollPercent() * 100)
	headerText := fmt.Sprintf("   %s  (%d%%)", rp.File.RelPath, pct)
	sb.WriteString(HeaderStyle.Render(headerText))
	sb.WriteString("\n")

	// Viewport content
	sb.WriteString(rp.Viewport.View())
	sb.WriteString("\n")

	// Bottom footer
	footer := StatusStyle.Render(" [j/k/d/u] Scroll │ [g/G] Top/Bottom │ [Esc/Backspace] Back to files │ [q] Quit ")
	sb.WriteString(footer)

	return sb.String()
}
