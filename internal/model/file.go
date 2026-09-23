package model

import "time"

// FileInfo represents metadata for a discovered markdown file.
type FileInfo struct {
	Path    string    `json:"path"`
	RelPath string    `json:"rel_path"`
	Depth   int       `json:"depth"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}
