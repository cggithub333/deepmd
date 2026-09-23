package finder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cggithub333/deepmd/internal/model"
	"github.com/charlievieth/fastwalk"
	ignore "github.com/sabhiram/go-gitignore"
)

type WalkOptions struct {
	Root          string
	MaxDepth      int
	IncludeHidden bool
	Extensions    []string
}

// DefaultWalkOptions returns standard configuration for markdown searching.
func DefaultWalkOptions(root string) WalkOptions {
	if root == "" {
		root = "."
	}
	return WalkOptions{
		Root:          root,
		MaxDepth:      5,
		IncludeHidden: false,
		Extensions:    []string{".md", ".markdown", ".mdown", ".mdx"},
	}
}

func WalkMarkdownFiles(opts WalkOptions) ([]model.FileInfo, error) {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = 5
	}
	if len(opts.Extensions) == 0 {
		opts.Extensions = []string{".md", ".markdown", ".mdown", ".mdx"}
	}

	extMap := make(map[string]bool)
	for _, ext := range opts.Extensions {
		extMap[strings.ToLower(ext)] = true
	}

	var mu sync.Mutex
	var files []model.FileInfo

	ign, _ := ignore.CompileIgnoreFile(filepath.Join(opts.Root, ".gitignore"))

	err := fastwalk.Walk(nil, opts.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		relPath, err := filepath.Rel(opts.Root, path)
		if err != nil {
			return nil
		}
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator)) + 1
		if depth > opts.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			switch name {
			case ".git", "node_modules", ".cache", ".vscode", "vendor", ".next", "target", "dist", "build", "out", ".turbo", ".gradle", ".idea", "storybook-static", ".pnpm", "coverage", "tmp":
				return filepath.SkipDir
			}
			if !opts.IncludeHidden && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		if ign != nil && ign.MatchesPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && d.Type().IsRegular() {
			name := d.Name()
			if !opts.IncludeHidden && strings.HasPrefix(name, ".") {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(name))
			if extMap[ext] {
				info, err := d.Info()
				if err != nil {
					return nil
				}
				absPath, err := filepath.Abs(path)
				if err != nil {
					absPath = path
				}
				mu.Lock()
				files = append(files, model.FileInfo{
					Path:    absPath,
					RelPath: relPath,
					Depth:   depth,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
				})
				mu.Unlock()
			}
		}

		return nil
	})

	return files, err
}
