package finder

import (
	"sort"

	"github.com/cggithub333/deepmd/internal/model"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

type MatchResult struct {
	File      model.FileInfo
	Score     int
	Positions []int
}

func FuzzyFilter(files []model.FileInfo, pattern string) []MatchResult {
	if pattern == "" {
		results := make([]MatchResult, len(files))
		for i, f := range files {
			results[i] = MatchResult{File: f, Score: 0}
		}
		return results
	}

	var results []MatchResult
	patternRunes := []rune(pattern)
	slab := util.MakeSlab(100, 2048)

	for _, file := range files {
		chars := util.ToChars([]byte(file.RelPath))

		res, pos := algo.FuzzyMatchV2(false, true, true, &chars, patternRunes, true, slab)
		if res.Score > 0 {
			var positions []int
			if pos != nil {
				positions = make([]int, len(*pos))
				for i, p := range *pos {
					positions[i] = int(p)
				}
			}
			results = append(results, MatchResult{
				File:      file,
				Score:     res.Score,
				Positions: positions,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
