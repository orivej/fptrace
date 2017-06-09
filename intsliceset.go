package main

type IntSliceSet struct {
	Slice []int
	Has   map[int]bool
}

func NewIntSliceSet() *IntSliceSet {
	return &IntSliceSet{
		Slice: []int{},
		Has:   map[int]bool{},
	}
}

func (iss *IntSliceSet) Add(x int) {
	if !iss.Has[x] {
		iss.Slice = append(iss.Slice, x)
		iss.Has[x] = true
	}
}
