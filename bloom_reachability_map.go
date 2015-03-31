package node

import (
	"fmt"
	"github.com/AutoRoute/bloom"
)

type BloomReachabilityMap struct {
	f *bloom.BloomFilter
}

func NewBloomReachabilityMap() ReachabilityMap {
	m := BloomReachabilityMap{f: bloom.New(1000, 4)}
	return m
}

func (m BloomReachabilityMap) IsReachable(n NodeAddress) bool {
	entry := []byte(n)
	res := m.f.Test(entry)
	return res
}

func (m BloomReachabilityMap) AddEntry(n NodeAddress) {
	entry := []byte(n)
	m.f.Add(entry)
}

// TODO: Figure out how to increment
func (m BloomReachabilityMap) Increment() {
	return
}

func (m BloomReachabilityMap) Merge(nr ReachabilityMap) error {
	var err error

	if n, ok := nr.(BloomReachabilityMap); ok {
		err = m.f.Merge(n.f)
	} else {
		err = fmt.Errorf("Mismatched reachability map types")
	}

	return err
}

// TODO: Figure out how to copy
func (m BloomReachabilityMap) Copy() ReachabilityMap {
	return m
}
