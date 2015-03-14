package node

// Takes care of maintaining maps and insures that we know which interfaces are reachable where.
type ReceiptHandler interface {
	AddConnection(NodeAddress, ReceiptConnection)
	AddSentPacket(p Packet, src, next NodeAddress)
}

type packetRecord struct {
	destination NodeAddress
	src         NodeAddress
	next        NodeAddress
	hash        PacketHash
}

type recieptImpl struct {
	connections map[NodeAddress]ReceiptConnection
	packets     map[PacketHash]packetRecord
}

func (r *receiptImpl) AddConnection(id NodeAddress, c ReceiptConnection) {
	r.connections[id] = c
}

func (r *receiptImpl) AddSentPacket(p Packet, src, next NodeAddress) {
	r.packets[p.Hash()] = packetRecord{p.Destination(), next, src, p.Hash()}
}
