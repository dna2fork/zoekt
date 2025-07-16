package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"strings"
	"sync"
	"github.com/sourcegraph/zoekt/contrib"
)

type DiffMap struct {
	A1, B1, A2, B2 int
}

type DiffRange struct {
	Maps []*DiffMap
}

func NewDiffMap(a1, b1, a2, b2 int) *DiffMap {
	return &DiffMap{a1, b1, a2, b2}
}

func (m *DiffMap) Add (x, y int) bool {
	if m.B1 >= x && m.A1 - 1 <= x && m.B2 >= y && m.A2 - 1 <= y {
		if m.B1 == x { m.B1 ++ }
		if m.A1 - 1 == x { m.A1 -- }
		if m.B2 == y { m.B2 ++ }
		if m.A2 - 1 == y { m.A2 -- }
		return true
	}
	return false
}

func (m *DiffMap) Swap () {
	a := m.A1
	b := m.B1
	m.A1 = m.A2
	m.B1 = m.B2
	m.A2 = a
	m.B2 = b
}

func (m *DiffMap) Clone () *DiffMap {
	newone := NewDiffMap(m.A1, m.B1, m.A2, m.B2)
	return newone
}

func (m *DiffMap) ToString () string {
	return fmt.Sprintf("(%d-%d, %d-%d)", m.A1, m.B1, m.A2, m.B2)
}

func NewDiffRange() *DiffRange {
	d := &DiffRange{}
	d.Maps = make([]*DiffMap, 0)
	return d
}

func (d *DiffRange) Swap () {
	for _, m := range d.Maps {
		m.Swap()
	}
}

func (d *DiffRange) Add (x, y int) {
	// assume x, y monotonically increasing
	// we only check if last element can hold x, y
	// otherwise, add new DiffMap
	n := len(d.Maps)
	if n > 0 {
		last := d.Maps[n - 1]
		if !last.Add(x, y) {
			m := NewDiffMap(x, x + 1, y, y + 1)
			d.Maps = append(d.Maps, m)
		}
	} else {
		m := NewDiffMap(x, x + 1, y, y + 1)
		d.Maps = append(d.Maps, m)
	}
}

func (d *DiffRange) Rank () int {
	n := 0
	for _, m := range d.Maps {
		n += m.B1 - m.A1
	}
	return n
}

func (d *DiffRange) Prior (another *DiffRange) int {
	return d.Rank() - another.Rank()
}

func (d *DiffRange) Equal (another *DiffRange) bool {
	n1 := len(d.Maps)
	n2 := len(another.Maps)
	if n1 != n2 { return false }
	if d.Prior(another) != 0 { return false }
	for i, m1 := range d.Maps {
		m2 := another.Maps[i]
		if m1.A1 != m2.A1 || m1.A2 != m2.A2 || m1.B1 != m2.B1 || m1.B2 != m2.B2 {
			return false
		}
	}
	return true
}

func (d *DiffRange) Clone () *DiffRange {
	newone := NewDiffRange()
	for _, m := range d.Maps {
		newone.Maps = append(newone.Maps, m.Clone())
	}
	return newone
}

func (d *DiffRange) Dump () {
	fmt.Print("range[")
	for _, m := range d.Maps {
		fmt.Print(m.ToString())
	}
	fmt.Println("]")
}

type Diff struct {
}

func (x *Diff) cloneDiffRangeSet (set []*DiffRange) []*DiffRange {
	if set == nil { return nil }
	n := len(set)
	cloned := make([]*DiffRange, n)
	for i, r := range set {
		cloned[i] = r.Clone()
	}
	return cloned
}

func (x *Diff) mergeDiffRangeSet (setA, setB []*DiffRange) []*DiffRange {
	if setA == nil && setB == nil { return nil }
	if setA == nil { return x.cloneDiffRangeSet(setB) }
	if setB == nil { return x.cloneDiffRangeSet(setA) }
	newone := x.cloneDiffRangeSet(setA)
	for _, r := range setB {
		exist := false
		for _, s := range newone {
			if s.Equal(r) {
				exist = true
				break
			}
		}
		if exist { break }
		newone = append(newone, r.Clone())
	}
	return newone
}

func (x *Diff) addMap(set []*DiffRange, a, b int) []*DiffRange {
	if set == nil {
		newone := make([]*DiffRange, 1)
		newone[0] = NewDiffRange()
		newone[0].Add(a, b)
		return newone
	}
	for _, r := range set {
		r.Add(a, b)
	}
	return set
}

func (x *Diff) Act (rA, rB []string) []*DiffRange {
	lenA := len(rA)
	lenB := len(rB)
	curA := rA
	curB := rB
	reverse := false
	if lenB < lenA {
		curA = rB
		curB = rA
		lenA = len(curA)
		lenB = len(curB)
		reverse = true
	}
	stackPrev := make([][]*DiffRange, lenA + 1)
	for j, chB := range curB {
		stack := make([][]*DiffRange, lenA + 1)
		for i, chA := range curA {
			// fmt.Println("--->", i, j)
			if chA == chB {
				stack[i+1] = x.cloneDiffRangeSet(stackPrev[i])
				stack[i+1] = x.addMap(stack[i+1], i, j)
				// fmt.Println("==")
				// DumpDiffRangeSet(stackPrev[i])
				// DumpDiffRangeSet(stack[i+1])
			} else {
				pathA := stack[i]
				pathB := stackPrev[i+1]
				ra := 0
				rb := 0
				if pathA != nil { ra = pathA[0].Rank() }
				if pathB != nil { rb = pathB[0].Rank() }
				// fmt.Println("!=", ra, rb)
				// DumpDiffRangeSet(pathA)
				// DumpDiffRangeSet(pathB)
				if ra > rb {
					stack[i+1] = x.cloneDiffRangeSet(pathA)
				} else if ra < rb {
					stack[i+1] = x.cloneDiffRangeSet(pathB)
				} else {
					stack[i+1] = x.mergeDiffRangeSet(pathA, pathB)
				}
				// DumpDiffRangeSet(stack[i+1])
			}
		}
		stackPrev = stack
	}
	result := stackPrev[lenA]
	if result == nil { return result }

	if reverse {
		for _, d := range result {
			d.Swap()
		}
	}

	return result
}

func (x *Diff) TrackLine (diff *DiffRange, linesA, linesB []string, line int) (int, int, int) {
	// return: bestLine, possibleStartLine, possibleEndLine
	st := 0
	ed := 0
	bestL := 0
	n := len(diff.Maps)
	lenA := len(linesA)
	lenB := len(linesB)
	if line >= lenA {
		return -1, -1, -1
	}
	for i, m := range diff.Maps {
		if m.A1 <= line && m.B1 > line {
			L := m.A2 + line - m.A1
			fmt.Println(line, "-->", L, linesA[line])
			return L, L, L+1
		}
		if line < m.A1 {
			if i == 0 {
				st = 0
				ed = m.A2
				break
			} else {
				prev := diff.Maps[i-1]
				st = prev.B2
				ed = m.A2
				break
			}
		}
	}
	if ed == 0 {
		last := diff.Maps[n-1]
		st = last.B2
		ed = lenB
	}

	if ed - st == 1 {
		bestL = st
		fmt.Println(line, linesA[line], "-->", st, linesB[st])
	} else if ed == st {
		bestL = -1
		fmt.Println(line, linesA[line], "[deleted]")
	} else {
		originRunes := []rune(linesA[line])
		targetRunes := []rune(linesB[st])
		bestL = st
		bestLscore := LCS(originRunes, targetRunes)
		for curL := st+1; curL < ed; curL++ {
			targetRunes = []rune(linesB[curL])
			score := LCS(originRunes, targetRunes)
			if score > bestLscore {
				bestL = curL
				bestLscore = score
			}
		}
		fmt.Println(line, linesA[line], "-->", bestL, linesB[bestL])
	}
	return bestL, st, ed
}

func LCS(rA, rB []rune) int {
	lenA := len(rA)
	lenB := len(rB)
	if lenA == 0 || lenB == 0 { return 0 }
	curA := rA
	curB := rB
	stackPrev := make([]int, lenA+1)
	for _, chB := range curB {
		stack := make([]int, lenA+1)
		for i, chA := range curA {
			if chA == chB {
				stack[i+1] = stackPrev[i] + 1
			} else {
				ra := stack[i]
				rb := stackPrev[i+1]
				if ra > rb {
					stack[i+1] = ra
				} else {
					stack[i+1] = rb
				}
			}
		}
		stackPrev = stack
	}
	return stackPrev[lenA]
}

func DumpDiffRangeSet (set []*DiffRange) {
	if set == nil {
		fmt.Println("(empty)")
		return
	}
	fmt.Println("{")
	for _, r := range set {
		fmt.Print("   ")
		r.Dump()
	}
	fmt.Println("}")
}

const MAX_TRACKPOINT_ID = 1000000

type TrackPoint struct {
	Id     int            `json:"id"`
	// { "rev": line }
	Origin map[string]int `json:"o"`
	Line   int            `json:"line"`
}

type TrackFileMeta struct {
	NextId    int    `json:"nid"`
	Timestamp int64  `json:"mtime"`
	Rev       string `json:"rev"`
}

type TrackFile struct {
	Project IProject
	Path    string
	Meta    TrackFileMeta
	Points  []TrackPoint
	mutex   *sync.Mutex
}

// metadata
// /project/(.git|.p4)/.zoekt/track/path... ._ -> timestmp, points

func NewTrackFile(p IProject, path string) *TrackFile {
	tf := &TrackFile{p, path, TrackFileMeta{1, 0, ""}, nil, &sync.Mutex{}}
	return tf
}

func (f *TrackFile) Load() error {
	metaBaseDir := f.Project.GetMetadataDir()
	track := filepath.Join(metaBaseDir, ".zoekt", "track", f.Path + "._")
	header := true

	var loadErr error
	err := contrib.File2Lines(track, func (line string) {
		if loadErr != nil {
			return
		}
		if line == "" {
			return
		}
		if strings.HasPrefix(line, "#") {
			return
		}
		b := []byte(line)
		if header {
			header = false
			// load basic info
			loadErr = json.Unmarshal(b, &f.Meta)
			if loadErr != nil {
				return
			}
			return
		}
		if f.Points == nil {
			f.Points = make([]TrackPoint, 0)
		}
		// load track points
		tp := TrackPoint{}
		loadErr = json.Unmarshal(b, &tp)
		if loadErr != nil {
			return
		}
		f.Points = append(f.Points, tp)
	})
	if loadErr == nil {
		loadErr = err
	}
	return loadErr
}

func (f *TrackFile) GetById(id int) *TrackPoint {
	if f.Points == nil { return nil }
	for _, tp := range f.Points {
		if tp.Id == id {
			return &tp
		}
	}
	return nil
}

func (f *TrackFile) GetByLine(line int) *TrackPoint {
	if f.Points == nil { return nil }
	for _, tp := range f.Points {
		if tp.Line == line {
			return &tp
		}
	}
	return nil
}

func (f *TrackFile) Add(line int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.Meta.NextId >= MAX_TRACKPOINT_ID {
		return -1, fmt.Errorf("reach max id")
	}
	id := f.Meta.NextId
	c, err := f.Project.GetFileCommitInfo(f.Path, 0, 1)
	if err != nil {
		return -1, fmt.Errorf("cannot get file commit")
	}
	if len(c) == 0 {
		return -1, fmt.Errorf("file commit empty")
	}
	commit := c[0]
	// TODO: Sync and GetByLine to check if exists?
	cur := f.GetByLine(line)
	if cur != nil {
		cur.Origin[commit] = line
		return cur.Id, nil
	}
	tp := TrackPoint{id, nil, line}
	tp.Origin = make(map[string]int)
	tp.Origin[commit] = line
	// TODO: check line range
	if f.Points == nil {
		f.Points = make([]TrackPoint, 0)
	}
	f.Points = append(f.Points, tp)
	f.Meta.NextId ++
	// TODO: how to prevent data loss when crash
	return id, nil
}

func (f *TrackFile) Del(id int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.Points == nil { return nil }
	index := -1
	for i, tp := range f.Points {
		if tp.Id == id {
			index = i
			break
		}
	}
	if index < 0 { return nil }
	r := f.Points[0:index]
	r = append(r, f.Points[index+1:]...)
	f.Points = r
	return nil
}

func (f *TrackFile) Save() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	metaBaseDir := f.Project.GetMetadataDir()
	track := filepath.Join(metaBaseDir, ".zoekt", "track", f.Path + "._")
	dirname := filepath.Dir(track)
	contrib.PrepareDirectory(dirname)

	trf, err := os.Create(track)
	if err != nil {
		return err
	}
	defer trf.Close()
	b, err := json.Marshal(f.Meta)
	if err != nil { return err }
	_, err = trf.Write(b)
	if err != nil { return err }
	_, err = trf.WriteString("\n")
	if err != nil { return err }
	if f.Points != nil {
		for _, p := range f.Points {
			b, err = json.Marshal(p)
			if err != nil { return err }
			_, err = trf.Write(b)
			if err != nil { return err }
			_, err = trf.WriteString("\n")
			if err != nil { return err }
		}
	}
	return nil
}

func (f *TrackFile) TrackLines () error {
	if f.Points == nil { return nil }
	commits, err := f.Project.GetFileCommitInfo(f.Path, 0, 1)
	if err != nil { return err }
	if len(commits) < 1 { return fmt.Errorf("file commit empty") }
	latest := commits[0]
	// data is synced
	if latest == f.Meta.Rev { return nil }

	srcText, err := f.Project.GetFileTextContents(f.Path, f.Meta.Rev)
	if err != nil { return err }
	dstText, err := f.Project.GetFileTextContents(f.Path, latest)
	if err != nil { return err }
	d := &Diff{}
	srcLines := strings.Split(srcText, "\n")
	dstLines := strings.Split(dstText, "\n")
	ret := d.Act(srcLines, dstLines)
	if ret == nil { return fmt.Errorf("cannot diff") }
	if len(ret) == 0 { return fmt.Errorf("totally different files") }

	for i, tp := range f.Points {
		// skip disconnected track point
		if tp.Line <= 0 { continue }
		L, Lst, Led := d.TrackLine(ret[0], srcLines, dstLines, tp.Line)
		if Lst < 0 && Led < 0 {
			// line disconnected
			f.Points[i].Line = -1
		} else {
			f.Points[i].Line = L
		}
	}

	f.Meta.Rev = latest
	return nil
}

func (f *TrackFile) SyncMetadata () error {
	commits, err := f.Project.GetFileCommitInfo(f.Path, 0, 1)
	if err != nil { return nil }
	if len(commits) == 0 { return fmt.Errorf("file commit empty") }
	latest := commits[0]
	f.Meta.Rev = latest
	return nil
}

func (f *TrackFile) Sync() error {
	if strings.HasSuffix(f.Path, "/") {
		f.Path = string([]rune(f.Path)[0:len(f.Path)-1])
	}
	baseDir := f.Project.GetBaseDir()
	source := filepath.Join(baseDir, f.Path)
	metaBaseDir := f.Project.GetMetadataDir()
	track := filepath.Join(metaBaseDir, ".zoekt", "track", f.Path + "._")
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("not suport track directory")
	}
	updatedTs := info.ModTime().Unix()

	dirname := filepath.Dir(track)
	contrib.PrepareDirectory(dirname)
	info, err = os.Stat(track)

	if os.IsNotExist(err) {
		// init track
		f.Meta.Timestamp = 0
		err = f.SyncMetadata()
		if err != nil { return err }
	} else if err == nil {
		// load data and compare with current timestamp
		err = f.Load()
		if err != nil { return err }
	} else {
		return err
	}
	if updatedTs != f.Meta.Timestamp {
		f.Meta.Timestamp = updatedTs
		err = f.TrackLines()
		if err != nil { return err }
		err = f.Save()
		if err != nil { return err }
	}
	return nil
}
