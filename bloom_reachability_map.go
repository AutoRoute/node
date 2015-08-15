package node

import (
	"github.com/AutoRoute/bloom"
)

type BloomReachabilityMap struct {
	Filters      []*bloom.BloomFilter
	Conglomerate *bloom.BloomFilter
}

func NewBloomReachabilityMap() BloomReachabilityMap {
	fs := make([]*bloom.BloomFilter, 1)
	fs[0] = bloom.New(1000, 4)

	m := BloomReachabilityMap{
		Filters:      fs,
		Conglomerate: fs[0].Copy(),
	}
	return m
}

func (m BloomReachabilityMap) IsReachable(n NodeAddress) bool {
	entry := []byte(n)
	res := m.Conglomerate.Test(entry)
	return res
}

func (m BloomReachabilityMap) AddEntry(n NodeAddress) {
	entry := []byte(n)
	m.Filters[0].Add(entry)
	m.Conglomerate.Add(entry)
}

func (m BloomReachabilityMap) Increment() {
	newZeroth := make([]*bloom.BloomFilter, 1)
	newZeroth[0] = bloom.New(1000, 4)
	m.Filters = append(newZeroth, m.Filters...)
}

func (m BloomReachabilityMap) Merge(n BloomReachabilityMap) {
	if len(m.Filters) < len(n.Filters) {
		for k, v := range m.Filters {
			v.Merge(n.Filters[k])
		}
		// append the remaining Filters
		m.Filters = append(m.Filters, n.Filters[len(n.Filters):]...)
	} else {
		for k, v := range n.Filters {
			m.Filters[k].Merge(v)
		}
	}
	// reconstruct the Conglomerate
	for _, v := range m.Filters {
		m.Conglomerate.Merge(v)
	}
}

func (m BloomReachabilityMap) Copy() BloomReachabilityMap {
	newFilters := make([]*bloom.BloomFilter, len(m.Filters))

	// copy each filter
	for k, v := range m.Filters {
		newFilters[k] = v.Copy()
	}
	newConglomerate := m.Conglomerate.Copy()

	mc := BloomReachabilityMap{
		Filters:      newFilters,
		Conglomerate: newConglomerate,
	}
	return mc
}
