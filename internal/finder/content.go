package finder

import (
	"bufio"
	"os"
	"strings"

	"github.com/cggithub333/deepmd/internal/model"
)

type ContentMatch struct {
	File     model.FileInfo
	LineNum  int
	LineText string
	Score    int // optional scoring
}

func SearchContent(files []model.FileInfo, query string, maxResults int) ([]ContentMatch, error) {
	var matches []ContentMatch
	if query == "" || maxResults <= 0 {
		return matches, nil
	}

	lowerQuery := strings.ToLower(query)

	for _, file := range files {
		if len(matches) >= maxResults {
			break
		}

		if file.Size > 2*1024*1024 {
			continue // Skip files larger than 2MB
		}

		f, err := os.Open(file.Path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 1
		for scanner.Scan() {
			lineText := scanner.Text()
			if strings.Contains(strings.ToLower(lineText), lowerQuery) {
				matches = append(matches, ContentMatch{
					File:     file,
					LineNum:  lineNum,
					LineText: lineText,
				})
				if len(matches) >= maxResults {
					break
				}
			}
			lineNum++
		}
		f.Close()
	}

	return matches, nil
}
