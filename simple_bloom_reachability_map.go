package node

import (
	"bloom"
)

type SimpleBloomReachabilityMap struct {
	f *bloom.BloomFilter
}

func NewSimpleBloomReachabilityMap() ReachabilityMap {
	m := SimpleBloomReachabilityMap{f: bloom.New(1000, 4)}
	return m
}

func (m SimpleBloomReachabilityMap) IsReachable(n NodeAddress) bool {
	entry := []byte(n)
	res := m.f.Test(entry)
	return res
}

func (m SimpleBloomReachabilityMap) AddEntry(n NodeAddress) {
	entry := []byte(n)
	m.f.Add(entry)
}

// TODO: Figure out how to increment
func (m SimpleBloomReachabilityMap) Increment() {
	return
}

func (m SimpleBloomReachabilityMap) Merge(nr ReachabilityMap) {
	if n, ok := nr.(SimpleBloomReachabilityMap); ok {
		m.f.Merge(n.f)
	} else {
		// TODO: Error goes here.
	}
}
