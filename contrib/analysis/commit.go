package analysis

import (
	"time"
	"fmt"
	"regexp"
	"strings"
	"path/filepath"
	"context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/shards"
)

type CommitFileInfo struct {
	// relative path like /README.md
	Path string							`json:"path"`
	// what happened for this file; added, deleted, modified ...
	Actions string						`json:"act"`
	// if it is the same with commit Id, leave blank
	// if not, leave with value
	// for example, in git it will always be blank
	//              in p4 it will be file revision #
	Revision string					`json:"rev"`
}

type CommitDetails struct {
	Id string							`json:"id"`
	Timestamp int64					`json:"timestamp"`
	Author string						`json:"author"`
	Title string						`json:"title"`
	Description string				`json:"desc"`
	CommitFiles []*CommitFileInfo	`json:"files"`
	// it can be used to store doc, review, bug links ...
	Fields map[string]string		`json:"fields"`
}

var p4DescInvalidMatcher = regexp.MustCompile(`^Invalid changelist number .*$`)
var p4DescCommitHeaderMatcher = regexp.MustCompile(`^Change (\d+) by (.+)@(.+) on (\d{4})/(\d{2})/(\d{2}) (\d{2}):(\d{2}):(\d{2})$`)
var p4DescMessageMatcher = regexp.MustCompile(`^\t(.*)$`)
var p4DescFileChangeListStartMatcher = regexp.MustCompile(`^Affected files \.\.\.$`)
var p4DescFileChangeMatcher = regexp.MustCompile(`^\.\.\. (.+)#(\d+) (\w+)$`)

func (p *P4Project) GetCommitDetails (commitId string) (*CommitDetails, error) {
	cmd := fmt.Sprintf(
		"P4PORT=%s P4USER=%s P4CLIENT=%s %s describe -s %s",
		p.P4Port, p.P4User, p.P4Client, P4_BIN, commitId,
	)
	printDebugCommand(cmd)

	// status:
	// - -1: error
	// -  0: start
	// -  1: normal
	// -  2: message
	// -  3: file change list
	status := 0
	details := &CommitDetails{}
	details.CommitFiles = make([]*CommitFileInfo, 0)
	var err error
	var parts []string
	Exec2Lines(cmd, func (line string) {
		switch status {
		case -1:
			return
		case 0:
			parts = p4DescInvalidMatcher.FindStringSubmatch(line)
			if parts != nil {
				status = -1
				err = fmt.Errorf("%s", parts[1])
				return
			}
			status = 1
			fallthrough
		case 1:
			parts = p4DescCommitHeaderMatcher.FindStringSubmatch(line)
			if parts == nil {
				status = -1
				err = fmt.Errorf("invalid commit details")
				return
			}
			details.Id = parts[1]
			details.Author = parts[2]
			t, timeErr := time.Parse(
				time.RFC3339,
				fmt.Sprintf(
					"%s-%s-%sT%s:%s:%sZ",
					parts[4], parts[5], parts[6], parts[7], parts[8], parts[9],
				),
			)
			if timeErr != nil {
				status = -1
				err = fmt.Errorf("invalid commit date")
				return
			}
			details.Timestamp = t.Unix()
			status = 2
		case 2:
			parts = p4DescMessageMatcher.FindStringSubmatch(line)
			if parts != nil {
				if details.Title == "" {
					details.Title = parts[1]
				} else {
					details.Description += parts[1] + "\n"
				}
				// TODO: extract more info like doc, review, bug links
				return
			}
			parts = p4DescFileChangeListStartMatcher.FindStringSubmatch(line)
			if parts != nil {
				status = 3
				return
			}
		case 3:
			parts = p4DescFileChangeMatcher.FindStringSubmatch(line)
			if parts != nil {
				mappedLocalPath := p.MapLocalPath(parts[1])
				if mappedLocalPath != "" {
					mappedLocalPath = strings.TrimPrefix(mappedLocalPath, p.BaseDir)
					fileinfo := &CommitFileInfo{mappedLocalPath, parts[3], parts[2]}
					details.CommitFiles = append(details.CommitFiles, fileinfo)
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return details, nil
}

func (p *P4Project) getCommitSummary (commitId string) (*CommitDetails, error) {
	cmd := fmt.Sprintf(
		"P4PORT=%s P4USER=%s P4CLIENT=%s %s describe -s %s",
		p.P4Port, p.P4User, p.P4Client, P4_BIN, commitId,
	)
	printDebugCommand(cmd)

	details := &CommitDetails{}
	finished := false
	var err error
	var parts []string
	Exec2Lines(cmd, func (line string) {
		if finished { return }
		finished = true
		parts = p4DescInvalidMatcher.FindStringSubmatch(line)
		if parts != nil {
			err = fmt.Errorf("%s", parts[1])
			return
		}
		parts = p4DescCommitHeaderMatcher.FindStringSubmatch(line)
		if parts == nil {
			err = fmt.Errorf("invalid commit details")
			return
		}
		details.Id = parts[1]
		details.Author = parts[2]
		t, timeErr := time.Parse(
			time.RFC3339,
			fmt.Sprintf(
				"%s-%s-%sT%s:%s:%sZ",
				parts[4], parts[5], parts[6], parts[7], parts[8], parts[9],
			),
		)
		if timeErr != nil {
			err = fmt.Errorf("invalid commit date")
			return
		}
		details.Timestamp = t.Unix()
	})
	if err != nil {
		return nil, err
	}
	return details, nil
}

var gitLogFatalMatcher = regexp.MustCompile(`^fatal: (.+)$`)
var gitLogCommitIdMatcher = regexp.MustCompile(`^commit ([a-f0-9]+)$`)
var gitLogAuthorMatcher = regexp.MustCompile(`^Author: (.*) <(.+@.+)>$`)
var gitLogDateMatcher = regexp.MustCompile(`^Date:\s+(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) ([+\-]\d{2})(\d{2})$`)
var gitLogMessageMatcher = regexp.MustCompile(`^    (.+)$`)
var gitLogFileChangeMatcher = regexp.MustCompile(`^([ADMT])\s+(.+)$`)
var gitLogFileMoveMatcher = regexp.MustCompile(`^([CR])\s(.+) -> (.+)$`)

func (p *GitProject) GetCommitDetails (commitId string) (*CommitDetails, error) {
	cmd := fmt.Sprintf(
		`%s -C %s log --name-status --date=iso --diff-filter=ACDMRT -1 -U %s`,
		GIT_BIN, p.BaseDir, commitId,
	)
	printDebugCommand(cmd)

	// status:
	// - -1: error
	// -  0: start
	// -  1: normal
	// -  2: message
	// -  3: file change list
	status := 0
	details := &CommitDetails{}
	details.CommitFiles = make([]*CommitFileInfo, 0)
	var err error
	var parts []string
	Exec2Lines(cmd, func (line string) {
		switch status {
		case -1:
			return
		case 0:
			parts = gitLogFatalMatcher.FindStringSubmatch(line)
			if parts != nil {
				status = -1
				err = fmt.Errorf("%s", parts[1])
				return
			}
			parts = gitLogCommitIdMatcher.FindStringSubmatch(line)
			if parts == nil {
				status = -1
				err = fmt.Errorf("invalid commit details")
				return
			}
			details.Id = parts[1]
			status = 1
		case 1:
			parts = gitLogAuthorMatcher.FindStringSubmatch(line)
			if parts != nil {
				details.Author = parts[2]
				return
			}
			parts = gitLogDateMatcher.FindStringSubmatch(line)
			if parts != nil {
				t, timeErr := time.Parse(
					time.RFC3339,
					fmt.Sprintf(
						"%sT%s%s:%s",
						parts[1], parts[2], parts[3], parts[4],
					),
				)
				if timeErr != nil {
					status = -1
					err = fmt.Errorf("invalid commit date")
					return
				}
				details.Timestamp = t.Unix()
				return
			}
			parts = gitLogMessageMatcher.FindStringSubmatch(line)
			if parts != nil {
				status = 2
				details.Title = parts[1]
				return
			}
		case 2:
			parts = gitLogFileChangeMatcher.FindStringSubmatch(line)
			if parts == nil {
				parts = gitLogFileMoveMatcher.FindStringSubmatch(line)
			}
			if parts == nil {
				details.Description += line + "\n"
				return
			}
			status = 3
			fallthrough
		case 3:
			parts = gitLogFileChangeMatcher.FindStringSubmatch(line)
			if parts == nil {
				parts = gitLogFileMoveMatcher.FindStringSubmatch(line)
			}
			if parts == nil {
				return
			}
			fileinfo := &CommitFileInfo{parts[2], parts[1], ""}
			details.CommitFiles = append(details.CommitFiles, fileinfo)
		}
	})
	if err != nil {
		return nil, err
	}
	return details, nil
}

func commitSearch(indexPath string, ctx context.Context, q string, num int) (*zoekt.SearchResult, error) {
	empty, err := IsEmptyFolder(indexPath)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, fmt.Errorf("invalid index path")
	}
	printDebugCommand(fmt.Sprintf("search in '%s'", indexPath))
	searcher, err := shards.NewDirectorySearcher(indexPath)
	if err != nil {
		return nil, err
	}
	defer searcher.Close()
	Q, err := query.Parse(q)
	sOpts := zoekt.SearchOptions{
		MaxWallTime: 10 * time.Second,
	}
	sOpts.SetDefaults()
	// limit doc number in case there are too many
	// ref: web/server.go
	if plan, err := searcher.Search(ctx, Q, &zoekt.SearchOptions{EstimateDocCount: true}); err != nil {
		return nil, err
	} else if numdocs := plan.ShardFilesConsidered; numdocs > 10000 {
		// 10k docs, top 50 -> max match/important = 275/4
		sOpts.ShardMaxMatchCount = num*5 + (5*num)/(numdocs/1000)
		sOpts.ShardMaxImportantMatch = num/20 + num/(numdocs/500)
	} else {
		n := numdocs + num*100
		sOpts.ShardMaxImportantMatch = n
		sOpts.ShardMaxMatchCount = n
		sOpts.TotalMaxMatchCount = n
	}
	sOpts.MaxDocDisplayCount = num
	// ref: api.go
	// sres.Files -> f.LineMatches
	// f.Language, f.Branches, string(f.Checksum), f.Filename, f.Repository
	// f.Version
	// m.LineNumber, m.Line, m.LineFragments -> x
	// x.LineOffset, x.MatchLength
	sres, err :=  searcher.Search(ctx, Q, &sOpts)
	if err != nil {
		return nil, err
	}
	return sres, nil
}

// ref: contrib/analysis/metadata.go
// define folder structure; /path/to/repo/.<type>/.zoekt/index
// TODO: add "r:commit"

func (p *P4Project) SearchCommits(ctx context.Context, query string, num int) (*zoekt.SearchResult, error) {
	path := filepath.Join(p.BaseDir, ".p4", ".zoekt", "index")
	return commitSearch(path, ctx, query, num)
}

func (p *GitProject) SearchCommits(ctx context.Context, query string, num int) (*zoekt.SearchResult, error) {
	path := filepath.Join(p.BaseDir, ".git", ".zoekt", "index")
	return commitSearch(path, ctx, query, num)
}
