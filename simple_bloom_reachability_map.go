package node

import (
	"bloom"
	"fmt"
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

func (m SimpleBloomReachabilityMap) Merge(nr ReachabilityMap) error {
	var err error

	if n, ok := nr.(SimpleBloomReachabilityMap); ok {
		err = m.f.Merge(n.f)
	} else {
		err = fmt.Errorf("Mismatched reachability map types")
	}

	return err
}
