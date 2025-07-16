package index

import (
	"strings"
)

func contribEvalBuildRepoListPattern(qStr string) ([]string, error) {
	// TODO: validate qStr (length, split(',') length) if too long return error
	if len(qStr) == 0 {
		return nil, nil
	}
	patterns := strings.Split(qStr, ",")
	return patterns, nil
}

func contribEvalMatchRepoListPattern(name string, patternList []string) bool {
	if (patternList == nil) {
		return true
	}
	if (len(patternList) == 0) {
		return true
	}
	for _, a := range patternList {
		if strings.Contains(name, a) {
			return true
		}
	}
	return false
}
