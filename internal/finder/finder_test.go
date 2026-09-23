package finder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cggithub333/deepmd/internal/model"
)

func setupTestDir(t *testing.T) string {
	dir := t.TempDir()
	
	// Create some markdown files and deep directories
	files := []string{
		"root.md",
		".git/secret.md",
		"node_modules/pkg/README.md",
		"dir1/file1.md",
		"dir1/dir2/file2.md",
		"dir1/dir2/dir3/file3.md",
		"dir1/dir2/dir3/dir4/file4.md",
		"dir1/dir2/dir3/dir4/dir5/file5.md",
		"dir1/dir2/dir3/dir4/dir5/dir6/file6.md",
	}

	for _, f := range files {
		path := filepath.Join(dir, f)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte("test content line 1\nsecond line query_match here\n"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// oversized file
	largePath := filepath.Join(dir, "large.md")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Truncate(3 * 1024 * 1024) // 3MB
	f.Close()

	// gitignore
	ignorePath := filepath.Join(dir, ".gitignore")
	os.WriteFile(ignorePath, []byte("ignored_dir/\n"), 0644)
	
	os.MkdirAll(filepath.Join(dir, "ignored_dir"), 0755)
	os.WriteFile(filepath.Join(dir, "ignored_dir", "ignore.md"), []byte("should be ignored"), 0644)

	return dir
}

func TestWalker(t *testing.T) {
	dir := setupTestDir(t)

	opts := WalkOptions{
		Root:     dir,
		MaxDepth: 5,
	}

	files, err := WalkMarkdownFiles(opts)
	if err != nil {
		t.Fatal(err)
	}

	// Check that .git and node_modules are ignored
	// Check that depth > 5 is ignored (file6.md)
	// Check that ignored_dir is ignored
	for _, f := range files {
		if strings.Contains(f.Path, ".git") {
			t.Errorf("expected .git to be ignored, got %s", f.Path)
		}
		if strings.Contains(f.Path, "node_modules") {
			t.Errorf("expected node_modules to be ignored, got %s", f.Path)
		}
		if strings.Contains(f.Path, "file5.md") {
			t.Errorf("expected file5.md (depth 6) to be ignored, got %s", f.Path)
		}
		if strings.Contains(f.Path, "ignored_dir") {
			t.Errorf("expected ignored_dir to be ignored, got %s", f.Path)
		}
	}

	foundFile4 := false
	for _, f := range files {
		if strings.Contains(f.Path, "file4.md") {
			foundFile4 = true
			break
		}
	}
	if !foundFile4 {
		t.Errorf("expected file5.md (depth 5) to be found")
	}
}

func TestFuzzyFilter(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "docs/readme.md"},
		{RelPath: "src/main.go"},
		{RelPath: "docs/setup_guide.md"},
	}

	results := FuzzyFilter(files, "docmd")
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}

	if results[0].File.RelPath != "docs/readme.md" && results[1].File.RelPath != "docs/readme.md" {
		t.Errorf("expected docs/readme.md to be matched")
	}

	// check if positions are populated
	if len(results[0].Positions) == 0 {
		t.Errorf("expected positions to be populated")
	}
	
	// empty pattern should return all
	emptyRes := FuzzyFilter(files, "")
	if len(emptyRes) != len(files) {
		t.Errorf("expected %d for empty pattern, got %d", len(files), len(emptyRes))
	}
}

func TestSearchContent(t *testing.T) {
	dir := setupTestDir(t)

	opts := WalkOptions{
		Root:     dir,
		MaxDepth: 5,
	}

	files, err := WalkMarkdownFiles(opts)
	if err != nil {
		t.Fatal(err)
	}

	// one of the files is large.md which is 3MB and should be ignored
	// The rest have "second line query_match here" on line 2.
	
	matches, err := SearchContent(files, "query_match", 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) == 0 {
		t.Fatalf("expected matches, got 0")
	}

	for _, m := range matches {
		if strings.Contains(m.File.Path, "large.md") {
			t.Errorf("expected large.md to be skipped")
		}
		if m.LineNum != 2 {
			t.Errorf("expected line num 2, got %d", m.LineNum)
		}
	}
}
