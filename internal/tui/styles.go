package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Border styles
	ActiveBorderStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00D7FF"))
	InactiveBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#585858"))

	// Typography & accents
	HeaderStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFAF")).Padding(0, 1)
	StatusStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A8A")).Padding(0, 1)
	SearchPromptStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5FD7"))
	MatchHighlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAF00")).Underline(true)
	SelectedItemStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#264F78"))
	NormalItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#D0D0D0"))
	DimItemStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	// Resize mode styling
	ResizeBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF5FD7"))
	ResizeHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#FF5FD7")).Padding(0, 1)

	// Danger / Warning
	WarningStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
)

// LayoutDimensions calculates pane dimensions safely.
type LayoutDimensions struct {
	TotalWidth  int
	TotalHeight int
	LeftWidth   int
	LeftHeight  int
	RightWidth  int
	RightHeight int
	TooSmall    bool
}

// CalculateLayout computes dimensions given terminal size with default ratio.
func CalculateLayout(width, height int) LayoutDimensions {
	return CalculateLayoutWithCustomLeft(width, height, 0)
}

// CalculateLayoutWithCustomLeft computes dimensions with a custom left width.
func CalculateLayoutWithCustomLeft(width, height int, customLeft int) LayoutDimensions {
	if width < 60 || height < 14 {
		return LayoutDimensions{
			TotalWidth:  width,
			TotalHeight: height,
			TooSmall:    true,
		}
	}

	leftW := customLeft
	if leftW <= 0 {
		leftW = int(float64(width) * 0.35)
	}

	minLeft := 20
	maxLeft := width - 25
	if maxLeft < minLeft {
		maxLeft = minLeft
	}
	if leftW < minLeft {
		leftW = minLeft
	}
	if leftW > maxLeft {
		leftW = maxLeft
	}

	rightW := width - leftW - 4 // Account for left pane border and right pane border
	if rightW < 15 {
		rightW = 15
	}

	innerH := height - 4 // header + status bar + borders
	if innerH < 5 {
		innerH = 5
	}

	return LayoutDimensions{
		TotalWidth:  width,
		TotalHeight: height,
		LeftWidth:   leftW,
		LeftHeight:  innerH,
		RightWidth:  rightW,
		RightHeight: innerH,
		TooSmall:    false,
	}
}

// TruncateString safely truncates a string with an ellipsis if it exceeds maxWidth.
// It guarantees zero negative slice indexing and handles small/zero maxWidth gracefully.
func TruncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

