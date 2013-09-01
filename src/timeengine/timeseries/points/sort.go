package points

import (
	"sort"
)

type ptSorter struct {
	P []*P
}

func (ps *ptSorter) Len() int {
	return len(ps.P)
}

func (ps *ptSorter) Less(i, j int) bool {
	return ps.P[i].t < ps.P[j].t
}

func (ps *ptSorter) Swap(i, j int) {
	a := ps.P[i]
	ps.P[i] = ps.P[j]
	ps.P[j] = a
}

func sortPtsByDate(pts []*P) {
	sort.Sort(&ptSorter{pts})
}
