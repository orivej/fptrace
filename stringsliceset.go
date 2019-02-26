package main

import (
	"flag"
	"strings"
)

type StringSliceSet struct {
	Slice []string
	Has   map[string]bool
}

func NewStringSliceSet() StringSliceSet {
	return StringSliceSet{
		Slice: []string{},
		Has:   map[string]bool{},
	}
}

func (ss *StringSliceSet) Add(x string) {
	if !ss.Has[x] {
		ss.Slice = append(ss.Slice, x)
		ss.Has[x] = true
	}
}

func (ss *StringSliceSet) String() string {
	return strings.Join(ss.Slice, ",")
}

func (ss *StringSliceSet) Set(x string) error {
	ss.Add(x)
	return nil
}

func StringSliceSetFlag(name, usage string) *StringSliceSet {
	ss := NewStringSliceSet()
	flag.Var(&ss, name, usage)
	return &ss
}
