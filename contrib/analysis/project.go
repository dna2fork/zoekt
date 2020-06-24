package analysis

import (
	"os"
	"io"
	"log"
	"fmt"
	"path/filepath"
	"strings"
	"regexp"
	"context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/contrib"
)

var (
	P4_BIN string
	GIT_BIN string
	CTAGS_BIN string
)

func init() {
	P4_BIN = os.Getenv("ZOEKT_P4_BIN")
	GIT_BIN = os.Getenv("ZOEKT_GIT_BIN")
	CTAGS_BIN = os.Getenv("ZOEKT_CTAGS_BIN")
}

// IProject project operator interface
type IProject interface {
	GetBaseDir() string
	Sync() (map[string]string, error) // return filepath to store latest modified file list
	Compile() error // virtually compile project; store metadata into disk: dump commit message, build ast tree ...
	GetProjectType() string // return p4, git, ...
	GetFileTextContents(path, revision string) (string, error)
	GetFileBinaryContents(path, revision string) ([]byte, error)
	GetFileLength(path, revision string) (int64, error)
	GetFileHash(path, revision string) (string, error)
	GetFileBlameInfo(path, revision string, startLine, endLine int) ([]*BlameDetails, error)
	GetFileCommitInfo(path string, offset, N int) ([]string, error) // N = -1 for dumping all
	GetDirContents(path, revision string) ([]string, error)
	GetCommitDetails(commitId string) (*CommitDetails, error)
	SearchCommits(ctx context.Context, query string, num int) (*zoekt.SearchResult, error)
}

type BlameDetails struct {
	Author string					`json:"author"`
	Commit string					`json:"commit"`
}

var _ IProject = &P4Project{}
var _ IProject = &GitProject{}

func NewProject (projectName string, baseDir string) IProject {
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil
	}
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil
	}
	options := make(map[string]string)
	// git project:
	// - .git
	gitGuess := filepath.Join(baseDir, ".git")
	info, err = os.Stat(gitGuess)
	if err == nil {
		if !info.IsDir() {
			return nil
		}
		getGitProjectOptions(baseDir, &options)
		return NewGitProject(projectName, baseDir, options)
	}
	// p4 project:
	// - .p4
	p4Guess := filepath.Join(baseDir, ".p4")
	info, err = os.Stat(p4Guess)
	if err == nil {
		if !info.IsDir() {
			return nil
		}
		getP4ProjectOptions(baseDir, &options)
		return NewP4Project(projectName, baseDir, options)
	}
	// not support yet
	return nil
}

var gitRemoteMatcher = regexp.MustCompile(`^origin\s+(.*)\s+\([a-z]+\)$`)

func getGitProjectOptions(baseDir string, options *map[string]string) {
	cmd := fmt.Sprintf("%s -C %s remote -v", GIT_BIN, baseDir)
	contrib.Exec2Lines(cmd, func (line string) {
		parts := gitRemoteMatcher.FindStringSubmatch(line)
		if parts == nil {
			return
		}
		(*options)["Url"] = parts[1]
	})
}

func getP4ProjectOptions(baseDir string, options *map[string]string) {
	configFilepath := filepath.Join(baseDir, ".p4", "config")
	f, err := os.Open(configFilepath)
	if err != nil {
		return
	}
	defer f.Close()
	// config file max size is 4KB
	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if err != nil {
		return
	}
	for _, keyval := range strings.Split(string(buf[0:n]), "\n") {
		if keyval == "" {
			continue
		}
		parts := strings.SplitN(keyval, "=", 2)
		(*options)[parts[0]] = parts[1]
	}
}

// P4Project //////////////////////////////////////////////////////////////////
type P4Project struct {
	Name string
	BaseDir string
	P4Port, P4User, P4Client string
	P4Details p4Details
}

type p4Details struct {
	Root string
	Owner string
	Views map[string]string
}

func NewP4Project (projectName string, baseDir string, options map[string]string) *P4Project {
	if P4_BIN == "" {
		log.Panic("[E] ! cannot find p4 command")
	}
	// baseDir: absolute path
	port, ok := options["P4PORT"]
	if !ok {
		log.Printf("P/%s: [E] missing P4PORT\n", projectName)
		return nil
	}
	user, ok := options["P4USER"]
	if !ok {
		log.Printf("P/%s: [E] missing P4USER\n", projectName)
		return nil
	}
	client, ok := options["P4CLIENT"]
	if !ok {
		log.Printf("P/%s: [E] missing P4CLIENT\n", projectName)
		return nil
	}
	p := &P4Project{projectName, baseDir, port, user, client, p4Details{}};
	p.getDetails()
	return p
}

func (p *P4Project) GetBaseDir () string {
	return p.BaseDir
}

var p4DetailRootMather = regexp.MustCompile(`^Root:\s+(.+)$`)
var p4DetailOwnerMather = regexp.MustCompile(`^Owner:\s+(.+)$`)
var p4DetailViewMather = regexp.MustCompile(`^View:$`)
// TODO: only support view map like //depot/path/to/... //client/path/to/...
//                 not support      //depot/path/to/file //client/path/to/file
var p4DetailViewLineMather = regexp.MustCompile(`^\s(//.+/)\.{3}\s+(//.+/)\.{3}$`)

func (p *P4Project) getDetails () error {
	cmd := fmt.Sprintf(
		"P4PORT=%s P4USER=%s P4CLIENT=%s %s client -o",
		p.P4Port, p.P4User, p.P4Client, P4_BIN,
	)
	contrib.PrintDebugCommand(cmd)
	viewMapLines := false
	err := contrib.Exec2Lines(cmd, func (line string) {
		if strings.HasPrefix(line, "#") {
			return
		}

		if viewMapLines {
			viewMap := p4DetailViewLineMather.FindStringSubmatch(line)
			if viewMap != nil {
				localPath := strings.TrimPrefix(viewMap[2], fmt.Sprintf("//%s/", p.P4Client))
				if filepath.Separator == '\\' {
					localPath = strings.ReplaceAll(localPath, "/", "\\")
				}
				localPath = fmt.Sprintf("%s%s", p.P4Details.Root, localPath)
				p.P4Details.Views[viewMap[1]] = localPath
			}
			return
		}

		parts := p4DetailRootMather.FindStringSubmatch(line)
		if parts != nil {
			p.P4Details.Root = strings.TrimRight(parts[1], string(filepath.Separator)) + string(filepath.Separator)
			return
		}
		parts = p4DetailOwnerMather.FindStringSubmatch(line)
		if parts != nil {
			p.P4Details.Owner = parts[1]
			return
		}
		parts = p4DetailViewMather.FindStringSubmatch(line)
		if parts != nil {
			viewMapLines = true
			p.P4Details.Views = make(map[string]string)
			return
		}
	})
	return err
}

func (p *P4Project) prepareP4folder () error {
	p4folder := filepath.Join(p.BaseDir, ".p4")
	fileinfo, err := os.Stat(p4folder)
	if os.IsNotExist(err) {
		os.Mkdir(p4folder, 0755)
	} else if err != nil {
		return err
	} else if !fileinfo.IsDir() {
		return fmt.Errorf(".p4 has been used as a normal file not a directory")
	}

	p4config := filepath.Join(p4folder, "config")
	f, err := os.Create(p4config)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("P4PORT=%s\nP4USER=%s\nP4CLIENT=%s\n", p.P4Port, p.P4User, p.P4Client))
	if err != nil {
		return err
	}
	return nil
}

// p4 output e.g. //depot/b#1 - added as /path/to/b
var p4SyncLineMatcher = regexp.MustCompile(`^(.*)#(\d+) - (\w+) as (.*)$`)
// when we manually remove all files in a client
// and then do a force sync, p4 will output delete all files
// and refreshing them ...
var p4SyncLineRefreshMatcher = regexp.MustCompile(`^(.*)#(\d+) - refreshing (.*)$`)

func (p *P4Project) extractSyncPath(line string, updatedList *map[string]string) {
	parts := p4SyncLineMatcher.FindStringSubmatch(line)
	if parts != nil {
		filename := strings.TrimPrefix(parts[4], p.BaseDir)
		(*updatedList)[filename] = parts[3]
		return
	}

	parts = p4SyncLineRefreshMatcher.FindStringSubmatch(line)
	if parts != nil {
		filename := strings.TrimPrefix(parts[3], p.BaseDir)
		(*updatedList)[filename] = "added"
	}
}

func (p *P4Project) clone (updatedList *map[string]string) error {
	cmd := fmt.Sprintf(
		"P4PORT=%s P4USER=%s P4CLIENT=%s %s sync -f",
		p.P4Port, p.P4User, p.P4Client, P4_BIN,
	)
	contrib.PrintDebugCommand(cmd)
	err := contrib.Exec2Lines(cmd, nil)
	doWalk(p.BaseDir, ".p4", updatedList)
	err = p.prepareP4folder()
	return err
}

func (p *P4Project) sync (updatedList *map[string]string) error {
	cmd := fmt.Sprintf(
		"P4PORT=%s P4USER=%s P4CLIENT=%s %s sync",
		p.P4Port, p.P4User, p.P4Client, P4_BIN,
	)
	contrib.PrintDebugCommand(cmd)
	err := contrib.Exec2Lines(cmd, func (line string) {
		p.extractSyncPath(line, updatedList)
	})
	return err
}

func (p *P4Project) Sync () (map[string]string, error) {
	updatedList := make(map[string]string)
	fileinfo, err := os.Stat(p.BaseDir)
	if os.IsNotExist(err) {
		err = p.clone(&updatedList)
		return updatedList, err
	}
	if err != nil {
		return updatedList, err
	}
	if !fileinfo.IsDir() {
		return updatedList, fmt.Errorf("P/%s: [E] cannot clone repo since \"%s\" is not a directory", p.Name)
	}
	err = p.sync(&updatedList)
	return updatedList, err
}

func (p *P4Project) Compile () error {
	return nil
}

func (p *P4Project) GetProjectType () string {
	return "p4"
}

// P4Project.MapViewPath
// - it is a special func for p4 only; to map a local path to server path
// /client/root/path/to/file --> //depot/path/to/file
func (p *P4Project) MapViewPath (path string) string {
	if path == "/" {
		for oneViewPath, oneLocalPath := range p.P4Details.Views {
			path = strings.TrimSuffix(oneViewPath, strings.TrimPrefix(oneLocalPath, p.BaseDir))
			return path
		}
		return ""
	}
	fullPath := filepath.Join(p.BaseDir, path)
	matchedView := ""
	matchedLocal := ""
	maxLen := 0
	for viewPath, localPath := range p.P4Details.Views {
		if strings.HasPrefix(fullPath, localPath) {
			n := len(localPath)
			if n > maxLen {
				matchedView = viewPath
				matchedLocal = localPath
			}
		} else if fullPath + string(filepath.Separator) == localPath {
			return viewPath
		}
	}
	if matchedView == "" {
		return ""
	}
	mappedPath := matchedView + strings.TrimPrefix(fullPath, matchedLocal)
	return mappedPath
}

func (p *P4Project) MapLocalPath (serverPath string) string {
	matchedView := ""
	matchedLocal := ""
	maxLen := 0
	for viewPath, localPath := range p.P4Details.Views {
		if strings.HasPrefix(serverPath, viewPath) {
			n := len(viewPath)
			if n > maxLen {
				matchedView = viewPath
				matchedLocal = localPath
			}
		}
	}
	if matchedLocal == "" {
		return ""
	}
	mappedPath := matchedLocal + strings.TrimPrefix(serverPath, matchedView)
	mappedPath = strings.TrimPrefix(mappedPath, p.BaseDir)
	return mappedPath
}

func (p *P4Project) GetFileTextContents (path, revision string) (string, error) {
	B, err := p.GetFileBinaryContents(path, revision)
	if err != nil {
		return "", err
	}
	T := string(B)
	if strings.Index(T, "\x00") >= 0 {
		return "", fmt.Errorf("binary")
	}
	return T, nil
}

func (p *P4Project) GetFileBinaryContents (path, revision string) ([]byte, error) {
	// P4CONFIG=.p4/config p4 print -q /path/to/file#54
	url := p.MapViewPath(path)
	if url == "" {
		return nil, fmt.Errorf("non-tracked file")
	}
	if revision != "" {
		url += "#" + revision
	}
	cmd := fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s print -q %s",
		p.BaseDir, P4_BIN, url,
	)
	contrib.PrintDebugCommand(cmd)
	var err error
	B := make([]byte, 0)
	L := 0
	contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
		n := 1024 * 1024 * 1
		buf := make([]byte, n)
		for n >= 1024 * 1024 * 1 {
			L += n
			if L > 1024 * 1024 * 10 {
				// max reading size 10 MB
				err = fmt.Errorf("larger than 10 MB")
				return
			}
			n, err = stream.Read(buf)
			if err != nil {
				return
			}
			B = append(B, buf[0:n]...)
		}
		err = nil
	})
	if err != nil {
		return nil, err
	}
	return B, nil
}

func (p *P4Project) GetFileHash (path, revision string) (string, error) {
	// P4CONFIG=.p4/config p4 print -q /path/to/file#54
	var url string
	if revision == "" {
		url = filepath.Join(p.BaseDir, path)
		return contrib.FileHash(url)
	} else {
		url = p.MapViewPath(path)
		if url == "" {
			return "", fmt.Errorf("non-tracked file")
		}
		if revision != "" {
			url += "#" + revision
		}
		cmd := fmt.Sprintf(
			"P4CONFIG=%s/.p4/config %s print -q %s",
			p.BaseDir, P4_BIN, url,
		)
		contrib.PrintDebugCommand(cmd)
		var hash string
		var err error
		contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
			hash, err = contrib.IoHash(stream)
		})
		return hash, err
	}
}

func (p *P4Project) GetFileLength (path, revision string) (int64, error) {
	// P4CONFIG=.p4/config p4 print -q /path/to/file#54
	var url string
	if revision == "" {
		url = filepath.Join(p.BaseDir, path)
		return contrib.FileLen(url)
	} else {
		url = p.MapViewPath(path)
		if url == "" {
			return -1, fmt.Errorf("non-tracked file")
		}
		if revision != "" {
			url += "#" + revision
		}
		cmd := fmt.Sprintf(
			"P4CONFIG=%s/.p4/config %s print -q %s",
			p.BaseDir, P4_BIN, url,
		)
		contrib.PrintDebugCommand(cmd)
		var L int64
		var err error
		contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
			L, err = contrib.IoLen(stream)
		})
		return L, err
	}
}

var p4AnnotateMatcher = regexp.MustCompile(`^(\d+):.*$`)

func (p *P4Project) GetFileBlameInfo (path, revision string, startLine, endLine int) ([]*BlameDetails, error) {
	// P4CONFIG=.p4/config p4 annotate -q /path/to/file#54 (rev)
	// P4CONFIG=.p4/config p4 annotate -I -q /path/to/file#54 (cln)

	// Step 1: get fielog (ChangeNumber-Author map)
	url := p.MapViewPath(path)
	if url == "" {
		return nil, fmt.Errorf("non-tracked file")
	}
	cmd := fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s filelog -s -i %s",
		p.BaseDir, P4_BIN, url,
	)
	contrib.PrintDebugCommand(cmd)
	commits := make(map[string]string, 0)
	contrib.Exec2Lines(cmd, func (line string) {
		parts := p4FilelogRevMatcher.FindStringSubmatch(line)
		if parts != nil {
			commits[parts[2]] = parts[5]
			return
		}
	})

	// Step 2: get annotate
	if revision != "" {
		url += "#" + revision
	}
	cmd = fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s annotate -q -c -I %s",
		p.BaseDir, P4_BIN, url,
	)
	contrib.PrintDebugCommand(cmd)
	blames := make([]*BlameDetails, 0)
	lastCommit := ""
	lineNo := 1
	contrib.Exec2Lines(cmd, func (line string) {
		parts := p4AnnotateMatcher.FindStringSubmatch(line)
		if parts != nil {
			if lineNo < startLine || lineNo > endLine {
				return
			}
			C := parts[1]
			author, ok := commits[C]
			if !ok {
				commitDetails, err := p.getCommitSummary(C)
				if err == nil {
					author = commitDetails.Author
					commits[commitDetails.Id] = commitDetails.Author
				} else {
					author = "(unknown)"
				}
			}
			if lastCommit == C {
				C = "^"
				author = "^"
			} else {
				lastCommit = C
			}
			details := &BlameDetails{C, author}
			blames = append(blames, details)
			lineNo ++
			return
		}
	})
	return blames, nil
}

var p4FilelogRevMatcher = regexp.MustCompile(`^\.\.\. #(\d+) change (\d+) ([a-z]+) on (\d{4}/\d{2}/\d{2} by ([^\s]+)@[^\s]+ .*)$`)
var p4FilelogExtraMatcher = regexp.MustCompile(`^\.\.\. \.\.\. ([a-z]+) from (.+)$`)

func (p *P4Project) GetFileCommitInfo (path string, offset, N int) ([]string, error) {
	// P4CONFIG=.p4/config p4 filelog -s /path/to/file
	/* samples
	... #2 change \d+ integrate on YYYY/MM/DD by who@where (text) 'commit message short'
	... ... copy from //depot/path/to/file#2
	... #1 change \d+ branch on YYYY/MM/DD by who@where (text) 'commit message short'
	... ... branch from //depot/path/to/file#1
	*/
	url := p.MapViewPath(path)
	if url == "" {
		return nil, fmt.Errorf("non-tracked file")
	}
	cmd := fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s filelog -s %s",
		p.BaseDir, P4_BIN, url,
	)
	contrib.PrintDebugCommand(cmd)
	commits := make([]string, 0)
	contrib.Exec2Lines(cmd, func (line string) {
		parts := p4FilelogRevMatcher.FindStringSubmatch(line)
		if parts != nil {
			if offset > 0 {
				offset --
				return
			}
			if N == 0 {
				return
			}
			commits = append(commits, parts[2])
			N --
			return
		}
		parts = p4FilelogExtraMatcher.FindStringSubmatch(line)
		// TODO: deal with extra info
	})
	return commits, nil
}

var p4NoSuchFileMatcher = regexp.MustCompile(`^.* - no such file\(s\)\.$`)
var p4FoundFileMatcher = regexp.MustCompile(`^(.*)#(\d+) - [a-z/]+ change (\d+) .*$`)

func (p *P4Project) GetDirContents (path, revision string) ([]string, error) {
	serverPath := p.MapViewPath(path)
	var suffix string
	if revision != "" {
		suffix = "#" + suffix
	}
	if serverPath == "" {
		return nil, fmt.Errorf("path not found")
	}
	list := make([]string, 0)
	if !strings.HasSuffix(serverPath, string(filepath.Separator)) {
		serverPath = serverPath + string(filepath.Separator)
	}
	cmd := fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s files -e %s*%s",
		p.BaseDir, P4_BIN, serverPath, suffix,
	)
	contrib.PrintDebugCommand(cmd)
	contrib.Exec2Lines(cmd, func (line string) {
		// //depot/path/to/file#4 - delete change 1234 (text)
		// //depot/path/to/file#4 - branch change 1234 (text)
		// //depot/path/to/file#4 - move/add change 1234 (text)
		// //depot/path/to/file#4 - add change 1234 (text)
		parts := p4NoSuchFileMatcher.FindStringSubmatch(line)
		if parts != nil {
			return
		}
		parts = p4FoundFileMatcher.FindStringSubmatch(line)
		checkLocal := p.MapLocalPath(parts[1])
		if checkLocal != "" {
			list = append(list, filepath.Base(parts[1]))
		}
	})
	cmd = fmt.Sprintf(
		"P4CONFIG=%s/.p4/config %s dirs -C %s*%s",
		p.BaseDir, P4_BIN, serverPath, suffix,
	)
	contrib.PrintDebugCommand(cmd)
	contrib.Exec2Lines(cmd, func (line string) {
		// //depot/path/to/dir
		parts := p4NoSuchFileMatcher.FindStringSubmatch(line)
		if parts != nil {
			return
		}
		list = append(list, filepath.Base(line) + "/")
	})
	return list, nil
}

// GitProject /////////////////////////////////////////////////////////////////
type GitProject struct {
	Name string
	BaseDir string
	Url, Branch string
}

func NewGitProject (projectName string, baseDir string, options map[string]string) *GitProject {
	if GIT_BIN == "" {
		log.Panic("[E] ! cannot find git command")
	}
	// baseDir: absolute path
	url, ok := options["Url"]
	if !ok {
		log.Printf("P/%s: [E] missing Url\n", projectName)
		return nil
	}
	branch, ok := options["Branch"]
	if !ok {
		branch = ""
	}
	p := &GitProject{projectName, baseDir, url, branch};
	info, err := os.Stat(baseDir)
	if err == nil {
		if info.IsDir() {
			p.getCurrentBranch()
		} else {
			log.Printf("P/%s: [E] %s is a file\n", projectName, baseDir)
			return nil
		}
	} else {
		log.Printf("P/%s: [W] missing Branch; using default\n", projectName)
	}
	return p
}

func (p *GitProject) GetBaseDir () string {
	return p.BaseDir
}

func (p *GitProject) getCurrentBranch () (string, error) {
	cmd := fmt.Sprintf("%s -C %s branch", GIT_BIN, p.BaseDir)
	contrib.PrintDebugCommand(cmd)
	err := contrib.Exec2Lines(cmd, func (line string) {
		if strings.HasPrefix(line, "* ") {
			p.Branch = strings.Fields(line)[1]
		}
	})
	return p.Branch, err
}

func (p *GitProject) clone (updatedList *map[string]string) error {
	cmd := ""
	if p.Branch == "" {
		cmd = fmt.Sprintf(
			"%s clone %s %s",
			GIT_BIN, p.Url, p.BaseDir,
		)
		contrib.PrintDebugCommand(cmd)
		err := contrib.Exec2Lines(cmd, nil)
		if err != nil {
			return err
		}
		p.getCurrentBranch()
	} else {
		cmd = fmt.Sprintf(
			"%s clone %s -b %s %s",
			GIT_BIN, p.Url, p.Branch, p.BaseDir,
		)
		contrib.PrintDebugCommand(cmd)
		err := contrib.Exec2Lines(cmd, nil)
		if err != nil {
			return err
		}
	}
	doWalk(p.BaseDir, ".git", updatedList)
	return nil
}

var gitSyncLineMatcher = regexp.MustCompile(`^diff --git a([/].*) b([/].*)$`)

func (p *GitProject) extractSyncPath(line string, updatedList *map[string]string) {
	parts := gitSyncLineMatcher.FindStringSubmatch(line)
	if parts == nil {
		return
	}
	a := parts[1]
	b := parts[2]
	if a == b {
		(*updatedList)[b] = "modified"
	} else {
		// move a to b
		(*updatedList)[a] = "deleted"
		(*updatedList)[b] = "added"
	}
}

func (p *GitProject) sync (updatedList *map[string]string) error {
	cmd := fmt.Sprintf(
		"%s -C %s fetch --all",
		GIT_BIN, p.BaseDir,
	)
	contrib.PrintDebugCommand(cmd)
	contrib.Exec2Lines(cmd, nil)
	if p.Branch == "" {
		p.getCurrentBranch()
	}

	cmd = fmt.Sprintf(
		"%s -C %s diff HEAD \"origin/%s\"",
		GIT_BIN, p.BaseDir, p.Branch,
	)
	contrib.PrintDebugCommand(cmd)
	err := contrib.Exec2Lines(cmd, func (line string) {
		p.extractSyncPath(line, updatedList)
	})
	for path, val := range *updatedList {
		if val != "modified" {
			continue
		}
		_, err := os.Stat(filepath.Join(p.BaseDir, path))
		if os.IsNotExist(err) {
			(*updatedList)[path] = "added"
		}
	}

	cmd = fmt.Sprintf(
		"%s -C %s reset --hard \"origin/%s\"",
		GIT_BIN, p.BaseDir, p.Branch,
	)
	contrib.PrintDebugCommand(cmd)
	err = contrib.Exec2Lines(cmd, nil)
	for path, val := range *updatedList {
		if val != "modified" {
			continue
		}
		_, err := os.Stat(filepath.Join(p.BaseDir, path))
		if os.IsNotExist(err) {
			(*updatedList)[path] = "deleted"
		}
	}
	return err
}

func (p *GitProject) Sync () (map[string]string, error) {
	updatedList := make(map[string]string)
	fileinfo, err := os.Stat(p.BaseDir)
	if os.IsNotExist(err) {
		err = p.clone(&updatedList)
		return updatedList, err
	}
	if err != nil {
		return updatedList, err
	}
	if !fileinfo.IsDir() {
		return updatedList, fmt.Errorf("P/%s: [E] cannot clone repo since \"%s\" is not a directory", p.Name)
	}
	err = p.sync(&updatedList)
	return updatedList, err
}

func (p *GitProject) Compile () error {
	return nil
}

func (p *GitProject) GetProjectType () string {
	return "git"
}

func (p *GitProject) GetFileTextContents (path, revision string) (string, error) {
	B, err := p.GetFileBinaryContents(path, revision)
	if err != nil {
		return "", err
	}
	T := string(B)
	if strings.Index(T, "\x00") >= 0 {
		return "", fmt.Errorf("binary")
	}
	return T, nil
}

func (p *GitProject) GetFileBinaryContents (path, revision string) ([]byte, error) {
	url := fmt.Sprintf("%s:%s", revision, strings.TrimLeft(path, string(filepath.Separator)))
	cmd := fmt.Sprintf("%s -C %s show %s", GIT_BIN, p.BaseDir, url)
	contrib.PrintDebugCommand(cmd)
	var err error
	B := make([]byte, 0)
	L := 0
	contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
		n := 1024 * 1024 * 1
		buf := make([]byte, n)
		for n >= 1024 * 1024 * 1 {
			L += n
			if L > 1024 * 1024 * 10 {
				// max reading size 10 MB
				err = fmt.Errorf("larger than 10 MB")
				return
			}
			n, err = stream.Read(buf)
			if err != nil {
				return
			}
			B = append(B, buf[0:n]...)
		}
		err = nil
	})
	if err != nil {
		return nil, err
	}
	return B, nil
}

func (p *GitProject) GetFileHash (path, revision string) (string, error) {
	var url string
	if revision == "" {
		url = filepath.Join(p.BaseDir, path)
		return contrib.FileHash(url)
	} else {
		url = fmt.Sprintf("%s:%s", revision, strings.TrimLeft(path, string(filepath.Separator)))
		cmd := fmt.Sprintf("%s -C %s show %s", GIT_BIN, p.BaseDir, url)
		contrib.PrintDebugCommand(cmd)
		var hash string
		var err error
		contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
			hash, err = contrib.IoHash(stream)
		})
		return hash, err
	}
}

func (p *GitProject) GetFileLength (path, revision string) (int64, error) {
	var url string
	if revision == "" {
		url = filepath.Join(p.BaseDir, path)
		return contrib.FileLen(url)
	} else {
		url = fmt.Sprintf("%s:%s", revision, strings.TrimLeft(path, string(filepath.Separator)))
		cmd := fmt.Sprintf("%s -C %s show %s", GIT_BIN, p.BaseDir, url)
		contrib.PrintDebugCommand(cmd)
		var L int64
		var err error
		contrib.Exec2Bytes(cmd, func (stream io.ReadCloser) {
			L, err = contrib.IoLen(stream)
		})
		return L, err
	}
}

var gitBlameLineMatcher = regexp.MustCompile(`^\^?([a-f0-9]+) \(<(.*@.*)>\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [+\-]\d{4})\s+\d+\) +.*$`)

func (p *GitProject) GetFileBlameInfo (path, revision string, startLine, endLine int) ([]*BlameDetails, error) {
	var Lrange string
	if endLine < 0 {
		Lrange = ""
	} else {
		Lrange = fmt.Sprintf("-L %d,%d", startLine, endLine)
	}
	cmd := fmt.Sprintf(
		"%s -C %s blame -e -l %s %s -- %s",
		GIT_BIN, p.BaseDir, Lrange, revision, filepath.Join(p.BaseDir, path),
	)
	contrib.PrintDebugCommand(cmd)
	blames := make([]*BlameDetails, 0)
	lastCommit := ""
	contrib.Exec2Lines(cmd, func (line string) {
		parts := gitBlameLineMatcher.FindStringSubmatch(line)
		var email string
		var commit string
		if parts == nil {
			email = "(unknown)"
			commit = "(unknown)"
		} else {
			email = parts[2]
			commit = parts[1]
		}
		if commit == lastCommit {
			// to shorten blame emails for lines
			email = "^"
			commit = "^"
		} else {
			lastCommit = commit
		}
		details := &BlameDetails{email, commit}
		blames = append(blames, details)
	})
	return blames, nil
}

func (p *GitProject) GetFileCommitInfo (path string, offset, N int) ([]string, error) {
	cmd := fmt.Sprintf(
		"%s -C %s log --pretty=format:%%H -- %s",
		GIT_BIN, p.BaseDir, filepath.Join(p.BaseDir, path),
	)
	contrib.PrintDebugCommand(cmd)
	commits := make([]string, 0)
	contrib.Exec2Lines(cmd, func (line string) {
		if line == "" {
			return
		}
		if offset > 0 {
			offset --
			return
		}
		if N == 0 {
			// if N = -1, dump all commit hashes
			return
		}
		commits = append(commits, line)
		N --
	})
	return commits, nil
}

func (p *GitProject) GetDirContents (path, revision string) ([]string, error) {
	path = filepath.Join(p.BaseDir, path)
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		path += string(filepath.Separator)
	}
	list := make([]string, 0)
	if p.Branch == "" {
		p.getCurrentBranch()
	}
	if revision == "" {
		revision = p.Branch
	}
	cmd := fmt.Sprintf(
		"%s -C %s ls-tree --name-only %s -- %s",
		GIT_BIN, p.BaseDir, revision, path,
	)
	contrib.PrintDebugCommand(cmd)
	contrib.Exec2Lines(cmd, func (line string) {
		if line == "" {
			return
		}
		fullPath := filepath.Join(p.BaseDir, line)
		info, err := os.Stat(fullPath)
		if err == nil {
			if info.IsDir() {
				list = append(list, line + "/")
				return
			}
		}
		list = append(list, line)
	})
	return list, nil
}

func doWalk (baseDir string, ignoredDir string, updatedList *map[string]string) error {
	return filepath.Walk(baseDir, func (path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("D/%s: [analysis.doWalk/W] cannot get file list ...\n", baseDir)
			return err
		}
		if info.IsDir() {
			if info.Name() == ignoredDir {
				return filepath.SkipDir
			}
		} else {
			(*updatedList)[strings.TrimPrefix(path, baseDir)] = "added"
		}
		return nil
	})
}

