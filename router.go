package node

// The brains of everything, takes in connections and wires them togethor.
type Router interface {
	AddConnection(Connection)
}

type routerImpl struct {
	// A map of public key hashes to connections
	connections map[string]Connection
}

func (r *routerImpl) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	go r.handleReachability(c)
	go r.handleData(c)
	go r.handleReceipts(c)
}

func (r *routerImpl) handleReachability(m MapConnection) {

}

func (r *routerImpl) handleData(p DataConnection) {
	for packet := range p.Packets() {
		for _, c := range r.connections {
			if c.Key().Hash() == packet.Destination() {
				c.SendPacket(packet)
			}
		}
	}
}

func (r *routerImpl) handleReceipts(p ReceiptConnection) {

}
