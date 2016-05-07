package internal

import (
  "encoding/base64"
  "hash/fnv"

	"github.com/AutoRoute/bloom"
	"github.com/AutoRoute/node/types"
)

type BloomReachabilityMap struct {
	Filters                []*bloom.BloomFilter
	// Allows us to keep track of filters between nodes.
	filter_hashes          map[string]bool
	Conglomerate           *bloom.BloomFilter
}

// Generates a unique hash for a particular filter.
// Args:
//  filter: The filter to hash.
// Returns:
//  The FNV hash of the filter.
func hashFilter(filter *bloom.BloomFilter) string {
  hasher := fnv.New64()
  filter.WriteTo(hasher)
  return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func NewBloomReachabilityMap() *BloomReachabilityMap {
	fs := make([]*bloom.BloomFilter, 1)
	fs[0] = bloom.New(1000, 4)

	m := BloomReachabilityMap{
		Filters:      fs,
		filter_hashes: make(map[string]bool),
		Conglomerate: fs[0].Copy(),
	}

  // Hash our initial filter to begin with.
  initial_hash := hashFilter(fs[0])
  m.filter_hashes[initial_hash] = true
	return &m
}

func (m *BloomReachabilityMap) IsReachable(n types.NodeAddress) bool {
	entry := []byte(n)
	res := m.Conglomerate.Test(entry)
	return res
}

func (m *BloomReachabilityMap) AddEntry(n types.NodeAddress) {
	entry := []byte(n)
	m.Filters[0].Add(entry)
	m.Conglomerate.Add(entry)
}

func (m *BloomReachabilityMap) Increment() {
	newZeroth := make([]*bloom.BloomFilter, 1)
	newZeroth[0] = bloom.New(1000, 4)
	m.Filters = append(newZeroth, m.Filters...)
}

// Merges two reachability maps.
// Args:
//  n: The map to merge with this one.
// Returns:
//  True if the map was modified, false if it wasn't. Practically, it will only
//  return false if it is being asked to merge a map whose filters are a subset
//  of this one's filters.
func (m *BloomReachabilityMap) Merge(n *BloomReachabilityMap) bool {
  modified := false

	if len(m.Filters) < len(n.Filters) {
    modified = true
		for k, v := range m.Filters {
		  _, found := m.filter_hashes[hashFilter(n.Filters[k])]
      if (found) {
        // This filter is not new.
        continue
      }

		  old_hash := hashFilter(v)
			v.Merge(n.Filters[k])
		  new_hash := hashFilter(v)
		  if old_hash != new_hash {
        delete(m.filter_hashes, old_hash)
        m.filter_hashes[new_hash] = true
      }
		}
		// append the remaining Filters
		for _, filter := range n.Filters[len(m.Filters):] {
		  m.filter_hashes[hashFilter(filter)] = true
		}
		m.Filters = append(m.Filters, n.Filters[len(m.Filters):]...)
	} else {
		for k, v := range n.Filters {
      // Check for an identical filter.
      _, found := m.filter_hashes[hashFilter(v)]
      if (found) {
        // This filter is not new.
        continue
      }

      old_hash := hashFilter(m.Filters[k])
			m.Filters[k].Merge(v)
			new_hash := hashFilter(m.Filters[k])
			if old_hash != new_hash {
        delete(m.filter_hashes, old_hash)
        m.filter_hashes[new_hash] = true
        modified = true
      }
		}
	}

	if !modified {
	  // We didn't add any new filters.
	  return false
	}

	// reconstruct the Conglomerate
	for _, v := range m.Filters {
		m.Conglomerate.Merge(v)
	}

	return true
}

func (m *BloomReachabilityMap) Copy() *BloomReachabilityMap {
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
	return &mc
}
