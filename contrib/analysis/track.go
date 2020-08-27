package analysis

import (
	"fmt"
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
