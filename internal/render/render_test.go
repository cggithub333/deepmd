package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(out, "Println") {
		t.Fatalf("expected output to contain code block, got: %s", out)
	}
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
