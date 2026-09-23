package render

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/glamour"
)

// Renderer manages markdown rendering using Charm's Glamour library and boxed code blocks.
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

type extractedCodeBlock struct {
	placeholder string
	code        string
	lang        string
}

var placeholderRegex = regexp.MustCompile(`DEEPMD_BLOCK_TOKEN_(\d+)`)

// Render converts markdown to ANSI terminal text with boxed code blocks.
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

	// Step 1: Extract code blocks and substitute with placeholders
	modifiedMarkdown, blocks := extractFencedCodeBlocks(content)

	// Step 2: Render remaining markdown with Glamour
	tr, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(st),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("glamour init: %w", err)
	}

	rendered, err := tr.Render(modifiedMarkdown)
	if err != nil {
		return "", fmt.Errorf("glamour render: %w", err)
	}

	if len(blocks) == 0 {
		return rendered, nil
	}

	// Step 3: Substitute placeholders with boxed code blocks
	renderedLines := strings.Split(rendered, "\n")
	var finalSb strings.Builder

	for _, line := range renderedLines {
		plainLine := xansi.Strip(line)
		matches := placeholderRegex.FindStringSubmatch(plainLine)
		if len(matches) == 2 {
			var idx int
			_, scanErr := fmt.Sscanf(matches[1], "%d", &idx)
			if scanErr == nil && idx >= 0 && idx < len(blocks) {
				cb := blocks[idx]
				boxed := RenderBoxedCode(cb.code, cb.lang, width, st)
				// Trim extra surrounding blank lines from boxed output if needed
				finalSb.WriteString(boxed)
				continue
			}
		}
		finalSb.WriteString(line)
		finalSb.WriteString("\n")
	}

	return finalSb.String(), nil
}

// extractFencedCodeBlocks scans markdown for fenced code blocks (``` or ~~~)
// and replaces them with safe unique placeholders.
func extractFencedCodeBlocks(content string) (string, []extractedCodeBlock) {
	lines := strings.Split(content, "\n")
	var outLines []string
	var blocks []extractedCodeBlock

	inBlock := false
	fenceChar := byte(0)
	fenceLen := 0
	var curCode strings.Builder
	var curLang string

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		if !inBlock {
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				fc := trimmed[0]
				fl := 0
				for fl < len(trimmed) && trimmed[fl] == fc {
					fl++
				}
				if fl >= 3 {
					inBlock = true
					fenceChar = fc
					fenceLen = fl
					curLang = strings.TrimSpace(trimmed[fl:])
					curCode.Reset()
					continue
				}
			}
			outLines = append(outLines, line)
		} else {
			// Check for closing fence
			if strings.HasPrefix(trimmed, strings.Repeat(string(fenceChar), fenceLen)) {
				rest := strings.TrimSpace(trimmed[fenceLen:])
				if rest == "" {
					// End of code block
					inBlock = false
					idx := len(blocks)
					ph := fmt.Sprintf("DEEPMD_BLOCK_TOKEN_%d", idx)
					blocks = append(blocks, extractedCodeBlock{
						placeholder: ph,
						code:        curCode.String(),
						lang:        curLang,
					})
					// Insert blank line before and after placeholder to ensure it forms a standalone paragraph in Glamour
					outLines = append(outLines, "")
					outLines = append(outLines, ph)
					outLines = append(outLines, "")
					continue
				}
			}
			curCode.WriteString(line)
			curCode.WriteString("\n")
		}
	}

	// If block remained unclosed at EOF, close it gracefully
	if inBlock {
		idx := len(blocks)
		ph := fmt.Sprintf("DEEPMD_BLOCK_TOKEN_%d", idx)
		blocks = append(blocks, extractedCodeBlock{
			placeholder: ph,
			code:        curCode.String(),
			lang:        curLang,
		})
		outLines = append(outLines, "")
		outLines = append(outLines, ph)
		outLines = append(outLines, "")
	}

	return strings.Join(outLines, "\n"), blocks
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
