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

func (m BloomReachabilityMap) Merge(n BloomReachabilityMap) error {
	if len(m.Filters) < len(n.Filters) {
		for k, v := range m.Filters {
			err := v.Merge(n.Filters[k])
			if err != nil {
				return err
			}
		}
		// append the remaining Filters
		m.Filters = append(m.Filters, n.Filters[len(n.Filters):]...)
	} else {
		for k, v := range n.Filters {
			err := m.Filters[k].Merge(v)
			if err != nil {
				return err
			}
		}
	}
	// reconstruct the Conglomerate
	for _, v := range m.Filters {
		err := m.Conglomerate.Merge(v)
		if err != nil {
			return err
		}
	}
	return nil
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
