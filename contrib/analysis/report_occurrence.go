package analysis

import (
	"fmt"
	"os"
	"time"
	"regexp"
	"strings"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"github.com/google/zoekt/contrib"
)

type OccurrenceReport struct {
	GenTime int64             `json:"gentime"`
	// 0 = ready, 1 = generating
	Status  int               `json:"status"`
	Groups  []*OccurrenceGroup `json:"groups"`
}

type OccurrenceGroup struct {
	Name  string   `json:"-"`
	Items []string `json:"items"`
	Count int      `json:"count"`
	Cases []string `json:"cases"`
	/* e.g. {
			Items: ["Klass1.Rename", "Klass2.Rename"],
			Count: 3,
			Cases: ["test1.java#L32", "test2.cpp#L89"]
			     // 1 in L32,         2 in L89
		}
	*/
}

func getOccurrenceReport(p IProject, name string) (*OccurrenceReport, error) {
	metaBaseDir := p.GetMetadataDir()
	reportBaseDir := filepath.Join(metaBaseDir, ".zoekt", "report", "occurrence")
	reportPath := filepath.Join(reportBaseDir, fmt.Sprintf("%s.json", name))
	_, err := os.Stat(reportPath)
	if err != nil { return nil, err }
	b, err := ioutil.ReadFile(reportPath)
	if err != nil { return nil, err }
	report := &OccurrenceReport{}
	err = json.Unmarshal(b, &report)
	if err != nil { return nil, err }
	return report, nil
}

func (p *P4Project) GetOccurrenceReport(name string) (*OccurrenceReport, error) {
	return getOccurrenceReport(p, name)
}

func (p *GitProject) GetOccurrenceReport(name string) (*OccurrenceReport, error) {
	return getOccurrenceReport(p, name)
}

func stringContains(list []string, a string) bool {
	for _, one := range list {
		if one == a { return true }
	}
	return false
}

func getScopeGroupName(scope string) string {
	parts := strings.Split(scope, ".")
	n := len(parts)
	return parts[n - 1]
}

var stops = regexp.MustCompile("[~!@#$%^&*()_+\\-={}|:\"<>?\\[\\]\\\\;',./ \\t\\r\\n\\`]+")

func processOneFileOccurrence(path, baseDir string, groupMap map[string]*OccurrenceGroup) error {
	if yes, _ := contrib.IsBinaryFile(path); yes {
		// skip binary file
		return nil
	}
	lineNo := 1
	contrib.File2Lines(path, func (line string) {
		wordMap := make(map[string]bool)
		for _, word := range stops.Split(line, -1) {
			if word == "" { continue }
			group, ok := groupMap[word]
			if !ok { continue }
			group.Count ++
			if _, ok = wordMap[word]; ok { continue }
			wordMap[word] = true
			relativePath := strings.TrimPrefix(path, baseDir)
			group.Cases = append(group.Cases, fmt.Sprintf("%s#L%d", relativePath, lineNo))
		}
		lineNo ++
	})
	return nil
}

func genOccurrenceReport(p IProject, name string, items []string) error {
	// items = scope1.scope2.....name
	var report OccurrenceReport
	// TODO read prev report if exists
	report.Status = 1
	groupMap := make(map[string]*OccurrenceGroup)

	for _, one := range items {
		groupName := getScopeGroupName(one)
		group, ok := groupMap[groupName]
		if !ok {
			group = &OccurrenceGroup{}
			group.Name = groupName
			group.Items = make([]string, 0)
			group.Cases = make([]string, 0)
			groupMap[groupName] = group
		}
		if !stringContains(group.Items, one) {
			group.Items = append(group.Items, one)
		}
	}

	baseDir := p.GetBaseDir()
	ignoredDirs := []string {".git", ".p4", ".zoekt"}
	filepath.Walk(baseDir, func (path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if stringContains(ignoredDirs, info.Name()) {
				return filepath.SkipDir
			}
		} else {
			e := processOneFileOccurrence(path, baseDir, groupMap)
			if e != nil {
				fmt.Println("occurrenceReport | cannot process:", name, path, e)
			}
		}
		return nil
	})

	report.Status = 0
	report.GenTime = time.Now().Unix()
	report.Groups = make([]*OccurrenceGroup, 0)
	for _, group := range groupMap {
		report.Groups = append(report.Groups, group)
	}

	metaBaseDir := p.GetMetadataDir()
	reportBaseDir := filepath.Join(metaBaseDir, ".zoekt", "report", "occurrence")
	contrib.PrepareDirectory(reportBaseDir)
	reportPath := filepath.Join(reportBaseDir, fmt.Sprintf("%s.json", name))
	rpf, err := os.Create(reportPath)
	if err != nil { return err }
	defer rpf.Close()
	b, err := json.Marshal(report)
	if err != nil { return err }
	_, err = rpf.Write(b)
	if err != nil { return err }
	return nil
}

func (p *P4Project) GenOccurrenceReport(name string, items []string) error {
	return genOccurrenceReport(p, name, items)
}

func (p *GitProject) GenOccurrenceReport(name string, items []string) error {
	return genOccurrenceReport(p, name, items)
}

