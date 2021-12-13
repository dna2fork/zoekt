package main

import (
	"io/ioutil"
	"path/filepath"
	"os"
	"regexp"
	"strings"
	"fmt"
)

func regexpSplit(text string, delimeter string) []string {
    reg := regexp.MustCompile(delimeter)
    indexes := reg.FindAllStringIndex(text, -1)
    laststart := 0
    result := make([]string, len(indexes) + 1)
    for i, element := range indexes {
            result[i] = text[laststart:element[0]]
            laststart = element[1]
    }
    result[len(indexes)] = text[laststart:len(text)]
    return result
}

func main () {
	baseDirStr := os.Args[1]
	baseDir, err := filepath.Abs(baseDirStr)
	if err != nil {
		fmt.Println("(Error: InvalidBaseDir)")
		return
	}
	indexHtmlBytes, err := ioutil.ReadFile(filepath.Join(baseDir, "index.html"))
	indexHtml := string(indexHtmlBytes)
	if err != nil {
		indexHtml = "(Error: InvalidTemplate)"
	}
	indexHtmlParts := regexpSplit(indexHtml, "{{|}}")
	for i, part := range indexHtmlParts {
		if i % 2 == 0 {
			continue
		}
		partialTextBytes, err := ioutil.ReadFile(filepath.Join(baseDir, part))
		partialText := string(partialTextBytes)
		if err != nil {
			partialText = "(Error: InvalidPartial)"
		}
		indexHtmlParts[i] = partialText
	}
	err = ioutil.WriteFile(filepath.Join(baseDir, "generated-search.html.tpl"), []byte(strings.Join(indexHtmlParts, "")), 0644)
}
