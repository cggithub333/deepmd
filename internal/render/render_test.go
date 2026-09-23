package render

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

func TestRender(t *testing.T) {
	r := NewRenderer("dark")
	out, err := r.Render("# Hello World\n\nThis is **bold** text.\n\n```go\nfmt.Println()\n```", 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Hello") || !strings.Contains(out, "World") {
		t.Fatalf("expected output to contain Hello and World, got: %s", out)
	}
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "DEEPMD") {
			t.Logf("LINE IS: %q", l)
		}
	}
	if !strings.Contains(out, "Println") {
		t.Fatalf("expected output to contain code block")
	}
	t.Logf("RENDERED OUTPUT:\n%s\n", out)
}

func TestRenderPreview_Truncation(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "long.md")

	var sb strings.Builder
	for i := 1; i <= 30; i++ {
		sb.WriteString("Line of text in document\n")
	}
	os.WriteFile(f1, []byte(sb.String()), 0644)

	r := NewRenderer("dark")
	out, err := r.RenderPreview(f1, 80, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Truncated preview") {
		t.Fatalf("expected truncation banner, got: %s", out)
	}
}

func TestRender_InvalidFile(t *testing.T) {
	r := NewRenderer("dark")
	_, err := r.RenderPreview("does_not_exist.md", 80, 5)
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	_, err = r.RenderFile("does_not_exist.md", 80)
	if err == nil {
		t.Fatal("expected error for non-existent file in RenderFile, got nil")
	}
}

func TestRenderBoxedCode_Alignment(t *testing.T) {
	boxed := RenderBoxedCode("func main() {\n    println(\"hello world\")\n}", "go", 60, "dark")
	rawLines := strings.Split(boxed, "\n")
	var lines []string
	for _, l := range rawLines {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (top, content, bottom), got %d", len(lines))
	}

	widths := make([]int, len(lines))
	for i, line := range lines {
		widths[i] = lipgloss.Width(line)
		t.Logf("Line %d width=%d text=%q", i, widths[i], line)
	}

	for i := 1; i < len(widths); i++ {
		if widths[i] != widths[0] {
			t.Errorf("line %d width %d != line 0 width %d", i, widths[i], widths[0])
		}
	}
}

func TestRenderBoxedCode_MultiLanguage(t *testing.T) {
	langs := []string{"python", "bash", "json", "typescript", "rust", ""}
	for _, lang := range langs {
		boxed := RenderBoxedCode("echo 'hello'", lang, 60, "dark")
		// Must not contain border characters
		if strings.Contains(boxed, "╭") || strings.Contains(boxed, "│") || strings.Contains(boxed, "╰") {
			t.Errorf("expected no box borders for lang %q", lang)
		}
		// Must contain gray background ANSI escape
		if !strings.Contains(boxed, "\x1b[48;2;38;38;38m") {
			t.Errorf("expected gray background escape in box for lang %q", lang)
		}
		if lang != "" {
			if !strings.Contains(boxed, lang) {
				t.Errorf("expected badge %q in box for lang %q", lang, lang)
			}
		}
	}
}

func TestRenderBoxedCode_Wrapping(t *testing.T) {
	longCode := "this_is_an_extremely_long_line_of_code_that_definitely_exceeds_the_inner_width_limit_of_a_narrow_terminal_window_and_must_wrap_cleanly()"
	boxed := RenderBoxedCode(longCode, "python", 50, "dark")

	rawLines := strings.Split(boxed, "\n")
	var lines []string
	for _, l := range rawLines {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}

	if len(lines) <= 3 {
		t.Fatalf("expected wrapped code block to have more than 3 lines, got %d", len(lines))
	}

	widths := make([]int, len(lines))
	for i, line := range lines {
		widths[i] = lipgloss.Width(line)
	}

	for i := 1; i < len(widths); i++ {
		if widths[i] != widths[0] {
			t.Errorf("wrapped line %d width %d != line 0 width %d", i, widths[i], widths[0])
		}
	}
}

func TestRenderFile_SampleFile(t *testing.T) {
	sampleMd := `# Sample
` + "```go\npackage main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n```" + `
` + "```python\nprint(\"py\")\n```" + `
` + "```bash\necho \"sh\"\n```" + `
` + "```json\n{\"k\": \"v\"}\n```"

	tmpDir := t.TempDir()
	samplePath := filepath.Join(tmpDir, "sample.md")
	if err := os.WriteFile(samplePath, []byte(sampleMd), 0644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}

	r := NewRenderer("dark")
	out, err := r.RenderFile(samplePath, 80)
	if err != nil {
		t.Fatalf("unexpected error rendering sample file: %v", err)
	}

	for _, badge := range []string{"go", "python", "bash", "json"} {
		if !strings.Contains(out, badge) {
			t.Errorf("expected rendered sample to contain badge %q", badge)
		}
	}

	t.Logf("SAMPLE FILE RENDERED:\n%s\n", out)
}

func TestChromaFormatterBehavior(t *testing.T) {
	lexer := lexers.Get("go")
	st := styles.Get("dracula")
	fmttr := formatters.Get("terminal256")
	it, _ := lexer.Tokenise(nil, "package main")
	var buf bytes.Buffer
	_ = fmttr.Format(&buf, st, it)
	t.Logf("Chroma formatted: %q", buf.String())
}
