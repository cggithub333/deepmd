package model

import (
	"testing"
	"time"
)

func TestFileInfo(t *testing.T) {
	now := time.Now()
	fi := FileInfo{
		Path:    "/home/user/notes/test.md",
		RelPath: "notes/test.md",
		Depth:   2,
		Size:    1024,
		ModTime: now,
		IsDir:   false,
	}

	if fi.RelPath != "notes/test.md" {
		t.Fatalf("expected notes/test.md, got %s", fi.RelPath)
	}
	if fi.Depth != 2 {
		t.Fatalf("expected depth 2, got %d", fi.Depth)
	}
}
