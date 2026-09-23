package finder

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cggithub333/deepmd/internal/model"
)

// ScoutCache represents persisted cache metadata and discovered files for a directory.
type ScoutCache struct {
	Root      string           `json:"root"`
	ScoutedAt time.Time        `json:"scouted_at"`
	Count     int              `json:"count"`
	Files     []model.FileInfo `json:"files"`
}

// GetCacheDir returns the ~/.deepmd/cache directory path, creating it if necessary.
func GetCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(home, ".deepmd", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}
	return cacheDir, nil
}

// CachePathForRoot returns the unique cache file path for a given root directory.
func CachePathForRoot(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	hasher := sha256.New()
	hasher.Write([]byte(absRoot))
	hashStr := hex.EncodeToString(hasher.Sum(nil))[:16]

	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	baseName := filepath.Base(absRoot)
	if baseName == "/" || baseName == "." {
		baseName = "root"
	}
	fileName := fmt.Sprintf("%s-%s.json", baseName, hashStr)
	return filepath.Join(cacheDir, fileName), nil
}

// LoadCache loads cached file results if the cache exists and is newer than maxAge.
func LoadCache(root string, maxAge time.Duration) (*ScoutCache, bool) {
	cachePath, err := CachePathForRoot(root)
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cache ScoutCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}

	if maxAge <= 0 || time.Since(cache.ScoutedAt) > maxAge {
		return &cache, false // Cache expired or zero maxAge
	}

	return &cache, true
}

// SaveCache writes discovered files to ~/.deepmd/cache/<hash>.json.
func SaveCache(root string, files []model.FileInfo) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}

	cachePath, err := CachePathForRoot(root)
	if err != nil {
		return err
	}

	cache := ScoutCache{
		Root:      absRoot,
		ScoutedAt: time.Now(),
		Count:     len(files),
		Files:     files,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return err
	}

	// Also write human-readable markdown scout file in ~/.deepmd/<date-time>/md-scout.md
	home, err := os.UserHomeDir()
	if err == nil {
		timestampDir := time.Now().Format("2006-01-02-150405")
		scoutDir := filepath.Join(home, ".deepmd", timestampDir)
		_ = os.MkdirAll(scoutDir, 0755)
		scoutMdPath := filepath.Join(scoutDir, "md-scout.md")

		var sb strings.Builder
		sb.WriteString("# deepmd Markdown Scout List \n\n")
		sb.WriteString(fmt.Sprintf("- **Root Directory**: `%s`\n", absRoot))
		sb.WriteString(fmt.Sprintf("- **Scouted At**: `%s`\n", cache.ScoutedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- **Total Files Discovered**: `%d`\n\n", len(files)))
		sb.WriteString("## Discovered Files\n\n")
		for _, f := range files {
			sb.WriteString(fmt.Sprintf("- `%s` (depth: %d, size: %d bytes)\n", f.RelPath, f.Depth, f.Size))
		}
		_ = os.WriteFile(scoutMdPath, []byte(sb.String()), 0644)

		// Create/update ~/.deepmd/latest-scout.md symlink
		latestSymlink := filepath.Join(home, ".deepmd", "latest-scout.md")
		_ = os.Remove(latestSymlink)
		_ = os.Symlink(scoutMdPath, latestSymlink)
	}

	return nil
}
