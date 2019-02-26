package main

type IntSliceSet struct {
	Slice []int
	Has   map[int]bool
}

func NewIntSliceSet() IntSliceSet {
	return IntSliceSet{
		Slice: []int{},
		Has:   map[int]bool{},
	}
}

func (ss *IntSliceSet) Add(x int) {
	if !ss.Has[x] {
		ss.Slice = append(ss.Slice, x)
		ss.Has[x] = true
	}
}
