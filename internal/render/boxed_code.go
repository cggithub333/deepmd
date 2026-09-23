package render

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/wrap"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// BoxedCodeRenderer renders fenced and indented code blocks with background styling.
type BoxedCodeRenderer struct {
	width int
	style string
}

// NewBoxedCodeRenderer creates a new BoxedCodeRenderer.
func NewBoxedCodeRenderer(width int, style string) *BoxedCodeRenderer {
	if width < 20 {
		width = 80
	}
	return &BoxedCodeRenderer{
		width: width,
		style: style,
	}
}

// RegisterFuncs implements renderer.NodeRenderer.
func (b *BoxedCodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, b.renderCodeBlock)
	reg.Register(ast.KindCodeBlock, b.renderCodeBlock)
}

func (b *BoxedCodeRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	var code string
	var lang string

	switch n := node.(type) {
	case *ast.FencedCodeBlock:
		lang = string(n.Language(source))
		l := n.Lines().Len()
		var sb strings.Builder
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			sb.Write(line.Value(source))
		}
		code = sb.String()
	case *ast.CodeBlock:
		l := n.Lines().Len()
		var sb strings.Builder
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			sb.Write(line.Value(source))
		}
		code = sb.String()
	default:
		return ast.WalkContinue, nil
	}

	boxed := RenderBoxedCode(code, lang, b.width, b.style)
	_, err := w.WriteString(boxed)
	if err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// RenderBoxedCode formats source code inside a borderless card with a solid gray background,
// language badge, and Chroma syntax highlighting.
func RenderBoxedCode(code, lang string, width int, styleName string) string {
	if width < 30 {
		width = 80
	}

	// Layout margin and box sizing
	margin := "  "
	if width < 50 && width >= 35 {
		margin = " "
	} else if width < 35 {
		margin = ""
	}

	boxW := width - (len(margin) * 2)
	if boxW < 24 {
		boxW = 24
	}
	innerW := boxW - 4
	if innerW < 12 {
		innerW = 12
	}

	cleanLang := strings.TrimSpace(strings.ToLower(lang))
	_, badgeColor := getLanguageColors(cleanLang)
	badgeStyle := lipgloss.NewStyle().Foreground(badgeColor).Bold(true)

	// Clean, modern gray background (dark mode: #262626 / RGB(38,38,38), light mode: #F0F0F0 / RGB(240,240,240))
	bgEscape := "\x1b[48;2;38;38;38m"
	if strings.Contains(strings.ToLower(styleName), "light") {
		bgEscape = "\x1b[48;2;240;240;240m"
	}
	const bgReset = "\x1b[0m"

	// Expand tabs to 4 spaces for consistent terminal alignment and trim trailing newlines
	code = strings.TrimRight(code, "\r\n")
	code = strings.ReplaceAll(code, "\t", "    ")

	// Chroma syntax highlighting
	lexer := lexers.Get(cleanLang)
	if lexer == nil && cleanLang != "" {
		lexer = lexers.Match(cleanLang)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	chromaTheme := "dracula"
	if strings.Contains(strings.ToLower(styleName), "light") {
		chromaTheme = "github"
	} else if strings.Contains(strings.ToLower(styleName), "monokai") {
		chromaTheme = "monokai"
	} else if strings.Contains(strings.ToLower(styleName), "tokyo") {
		chromaTheme = "nord"
	}

	themeStyle := styles.Get(chromaTheme)
	if themeStyle == nil {
		themeStyle = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, code)
	if err == nil {
		_ = formatter.Format(&buf, themeStyle, iterator)
	} else {
		buf.WriteString(code)
	}
	highlighted := buf.String()

	rawLines := strings.Split(highlighted, "\n")
	for len(rawLines) > 0 && strings.TrimSpace(xansi.Strip(rawLines[len(rawLines)-1])) == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	displayLang := cleanLang

	var sb strings.Builder
	sb.WriteString("\n") // separation before code block

	// Top line with language badge if language is specified
	if displayLang != "" && displayLang != "code" {
		badgeText := badgeStyle.Render(displayLang) + bgEscape
		pad := strings.Repeat(" ", max(0, innerW-lipgloss.Width(displayLang)))
		topLine := bgEscape + "  " + badgeText + pad + "  " + bgReset
		sb.WriteString(margin + topLine + "\n")
	} else {
		// Subtle top padding line
		topLine := bgEscape + strings.Repeat(" ", boxW) + bgReset
		sb.WriteString(margin + topLine + "\n")
	}

	// Content lines: code lines rendered on gray background
	if len(rawLines) == 0 {
		emptyLine := bgEscape + strings.Repeat(" ", boxW) + bgReset
		sb.WriteString(margin + emptyLine + "\n")
	} else {
		for _, line := range rawLines {
			w := lipgloss.Width(line)
			if w <= innerW {
				styledLine := strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgEscape)
				pad := strings.Repeat(" ", innerW-w)
				sb.WriteString(margin + bgEscape + "  " + styledLine + pad + "  " + bgReset + "\n")
			} else {
				// Wrap line gracefully inside innerW
				wrapped := wrap.String(line, innerW)
				sublines := strings.Split(wrapped, "\n")
				for _, sub := range sublines {
					subStyled := strings.ReplaceAll(sub, "\x1b[0m", "\x1b[0m"+bgEscape)
					subW := lipgloss.Width(sub)
					subPad := strings.Repeat(" ", max(0, innerW-subW))
					sb.WriteString(margin + bgEscape + "  " + subStyled + subPad + "  " + bgReset + "\n")
				}
			}
		}
	}

	// Bottom padding line on gray background
	bottomLine := bgEscape + strings.Repeat(" ", boxW) + bgReset
	sb.WriteString(margin + bottomLine + "\n")

	return sb.String()
}

// getLanguageColors maps a markdown code fence language to a distinctive theme color.
func getLanguageColors(lang string) (borderColor lipgloss.Color, badgeColor lipgloss.Color) {
	switch lang {
	case "go", "golang":
		return lipgloss.Color("#00ADD8"), lipgloss.Color("#00E5FF")
	case "py", "python":
		return lipgloss.Color("#3776AB"), lipgloss.Color("#FFD438")
	case "js", "javascript":
		return lipgloss.Color("#F7DF1E"), lipgloss.Color("#FFE600")
	case "ts", "typescript":
		return lipgloss.Color("#3178C6"), lipgloss.Color("#61AFEF")
	case "rs", "rust":
		return lipgloss.Color("#DEA584"), lipgloss.Color("#FF6E4A")
	case "c", "cpp", "c++", "cxx", "h", "hpp":
		return lipgloss.Color("#5E97D0"), lipgloss.Color("#82B1FF")
	case "java", "kt", "kotlin", "scala":
		return lipgloss.Color("#ED8B00"), lipgloss.Color("#FFA726")
	case "sh", "bash", "zsh", "shell":
		return lipgloss.Color("#23D18B"), lipgloss.Color("#50FA7B")
	case "json", "yaml", "yml", "toml":
		return lipgloss.Color("#FF5FD2"), lipgloss.Color("#FF79C6")
	case "html", "xml", "svg":
		return lipgloss.Color("#E34F26"), lipgloss.Color("#FF7043")
	case "css", "scss", "sass":
		return lipgloss.Color("#2965F1"), lipgloss.Color("#40C4FF")
	case "sql", "psql", "mysql":
		return lipgloss.Color("#F5821F"), lipgloss.Color("#FFB74D")
	case "rb", "ruby":
		return lipgloss.Color("#CC342D"), lipgloss.Color("#FF5252")
	case "php":
		return lipgloss.Color("#777BB4"), lipgloss.Color("#B388FF")
	case "docker", "dockerfile":
		return lipgloss.Color("#2496ED"), lipgloss.Color("#40C4FF")
	case "md", "markdown":
		return lipgloss.Color("#5C94E5"), lipgloss.Color("#82B1FF")
	case "diff", "patch":
		return lipgloss.Color("#50FA7B"), lipgloss.Color("#69FF94")
	default:
		return lipgloss.Color("#7D56F4"), lipgloss.Color("#BD93F9")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
