package finder

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cggithub333/deepmd/internal/model"
)

func TestScoutCache_SaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "deepmd-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	files := []model.FileInfo{
		{RelPath: "README.md", Path: filepath.Join(tempDir, "README.md"), Depth: 1, Size: 100},
		{RelPath: "docs/architecture.md", Path: filepath.Join(tempDir, "docs/architecture.md"), Depth: 2, Size: 250},
	}

	// 1. Save cache
	if err := SaveCache(tempDir, files); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	// 2. Load fresh cache
	cached, ok := LoadCache(tempDir, 1*time.Hour)
	if !ok {
		t.Fatal("expected LoadCache to return true for fresh cache")
	}
	if cached.Count != 2 {
		t.Fatalf("expected 2 cached files, got %d", cached.Count)
	}
	if cached.Files[0].RelPath != "README.md" {
		t.Fatalf("expected first file README.md, got %s", cached.Files[0].RelPath)
	}

	// 3. Test expired cache
	_, okExpired := LoadCache(tempDir, -1*time.Second)
	if okExpired {
		t.Fatal("expected LoadCache to return false for expired cache")
	}
}
