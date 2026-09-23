package render

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

// Renderer manages markdown rendering using Charm's Glamour library.
type Renderer struct {
	Style string
}

// NewRenderer creates a new Markdown renderer with standard style.
func NewRenderer(style string) *Renderer {
	if style == "" {
		style = "dark"
	}
	return &Renderer{Style: style}
}

// Render converts markdown to ANSI terminal text.
func (r *Renderer) Render(content string, width int, customStyle ...string) (string, error) {
	st := r.Style
	if len(customStyle) > 0 && customStyle[0] != "" {
		st = customStyle[0]
	}
	if st == "" {
		st = "dark"
	}
	if width < 20 {
		width = 80
	}
	tr, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(st),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("glamour init: %w", err)
	}
	return tr.Render(content)
}

// RenderPreview reads up to maxBytes or maxLines and renders preview with truncation notice.
func (r *Renderer) RenderPreview(path string, width int, maxLines int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if maxLines <= 0 {
		maxLines = 500
	}

	reader := bufio.NewReader(f)
	var content []byte
	lines := 0
	truncated := false
	bytesRead := 0
	const maxBytes = 64 * 1024

	for {
		line, err := reader.ReadBytes('\n')
		bytesRead += len(line)
		content = append(content, line...)
		lines++

		if lines >= maxLines || bytesRead >= maxBytes {
			truncated = true
			break
		}
		if err != nil {
			break
		}
	}

	rendered, err := r.Render(string(content), width)
	if err != nil {
		return "", err
	}

	if truncated {
		rendered += "\n\x1b[33m[Truncated preview — press Enter to view full document]\x1b[0m\n"
	}

	return rendered, nil
}

// RenderFile reads and renders the full document up to 10MB limit.
func (r *Renderer) RenderFile(path string, width int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("file size exceeds 10MB limit")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return r.Render(string(data), width)
}
