package node

type srmentry struct {
	address  NodeAddress
	distance uint8
}

type SimpleReachabilityMap map[NodeAddress]srmentry

func NewSimpleReachabilityMap() ReachabilityMap {
	m := make(map[NodeAddress]srmentry)
	return SimpleReachabilityMap(m)
}

func (m SimpleReachabilityMap) IsReachable(n NodeAddress) bool {
	_, ok := m[n]
	return ok
}

func (m SimpleReachabilityMap) AddEntry(n NodeAddress) {
	m[n] = srmentry{n, 0}
}

func (m SimpleReachabilityMap) Increment() {
	for k, v := range m {
		v.distance += 1
		m[k] = v
	}
}

func (m SimpleReachabilityMap) Merge(nr ReachabilityMap) {
	n := nr.(SimpleReachabilityMap)
	for k, v := range n {
		m[k] = v
	}
}
