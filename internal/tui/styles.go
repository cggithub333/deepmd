package tui

import (
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

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

// CalculateLayout computes dimensions given terminal size with default ratio (0.60).
func CalculateLayout(width, height int) LayoutDimensions {
	return CalculateLayoutWithRatio(width, height, 0.60)
}

// CalculateLayoutWithRatio computes dimensions with a relative split ratio (0.0 to 1.0).
func CalculateLayoutWithRatio(width, height int, ratio float64) LayoutDimensions {
	if width < 60 || height < 14 {
		return LayoutDimensions{
			TotalWidth:  width,
			TotalHeight: height,
			TooSmall:    true,
		}
	}

	if ratio <= 0.05 || ratio >= 0.95 {
		ratio = 0.60
	}

	leftW := int(float64(width) * ratio)
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

// CalculateLayoutWithCustomLeft computes dimensions with an explicit custom left width.
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
		return CalculateLayoutWithRatio(width, height, 0.60)
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

	rightW := width - leftW - 4
	if rightW < 15 {
		rightW = 15
	}

	innerH := height - 4
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

// SmartTruncatePath intelligently truncates file paths for narrow columns.
// It prioritizes displaying the base filename over deeply nested directory prefixes.
func SmartTruncatePath(path string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(path) <= maxWidth {
		return path
	}
	if maxWidth <= 3 {
		return path[:maxWidth]
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// If no parent directory (e.g. "README.md" or "."), standard truncate
	if dir == "." || dir == "/" || dir == "" {
		return TruncateString(base, maxWidth)
	}

	// If base filename alone is already >= maxWidth
	if len(base) >= maxWidth {
		return TruncateString(base, maxWidth)
	}

	// If there's room for ".../" + base
	prefix := ".../"
	needed := len(prefix) + len(base)
	if maxWidth < needed {
		// Just show the base name (which is <= maxWidth)
		return base
	}

	// There is room for ".../" + some directory component + "/" + base
	availForDir := maxWidth - needed
	// If availForDir is small, just return ".../" + base
	if availForDir <= 3 {
		return prefix + base
	}

	// Try to include the immediate parent dir: ".../parent/base"
	parent := filepath.Base(dir)
	if len(parent)+1 <= availForDir {
		return prefix + parent + "/" + base
	}

	return prefix + base
}

