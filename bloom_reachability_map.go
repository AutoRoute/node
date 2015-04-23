package node

import (
	"fmt"
	"github.com/AutoRoute/bloom"
)

type BloomReachabilityMap struct {
	filters      []*bloom.BloomFilter
	conglomerate *bloom.BloomFilter
}

func NewBloomReachabilityMap() ReachabilityMap {
	fs := make([]*bloom.BloomFilter, 1)
	fs[0] = bloom.New(1000, 4)

	m := BloomReachabilityMap{
		filters:      fs,
		conglomerate: fs[0].Copy(),
	}
	return m
}

func (m BloomReachabilityMap) IsReachable(n NodeAddress) bool {
	entry := []byte(n)
	res := m.conglomerate.Test(entry)
	return res
}

func (m BloomReachabilityMap) AddEntry(n NodeAddress) {
	entry := []byte(n)
	m.filters[0].Add(entry)
	m.conglomerate.Add(entry)
}

func (m BloomReachabilityMap) Increment() {
	newZeroth := make([]*bloom.BloomFilter, 1)
	newZeroth[0] = bloom.New(1000, 4)
	m.filters = append(newZeroth, m.filters...)
}

func (m BloomReachabilityMap) Merge(nr ReachabilityMap) error {
	var err error

	if n, ok := nr.(BloomReachabilityMap); ok {
		if len(m.filters) < len(n.filters) {
			for k, v := range m.filters {
				err = v.Merge(n.filters[k])
				if err != nil {
					return err
				}
			}
			// append the remaining filters
			m.filters = append(m.filters, n.filters[len(n.filters):]...)
		} else {
			for k, v := range n.filters {
				err = m.filters[k].Merge(v)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = fmt.Errorf("Mismatched reachability map types")
		return err
	}
	// reconstruct the conglomerate
	for _, v := range m.filters {
		err = m.conglomerate.Merge(v)
		if err != nil {
			return err
		}
	}
	return err
}

func (m BloomReachabilityMap) Copy() ReachabilityMap {
	newFilters := make([]*bloom.BloomFilter, len(m.filters))

	// copy each filter
	for k, v := range m.filters {
		newFilters[k] = v.Copy()
	}
	newConglomerate := m.conglomerate.Copy()

	mc := BloomReachabilityMap{
		filters:      newFilters,
		conglomerate: newConglomerate,
	}
	return mc
}
