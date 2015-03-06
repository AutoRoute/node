package node

type SimpleReachabilityMap map[NodeAddress]bool

func (m SimpleReachabilityMap) IsReachable(n NodeAddress) bool {
	return m[n]
}
