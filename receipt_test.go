package node

import (
	"testing"
)

type testReceiptConnection struct {
	in  chan PacketReceipt
	out chan PacketReceipt
}

func (c testReceiptConnection) SendReceipt(r PacketReceipt) error {
	c.out <- r
	return nil
}
func (c testReceiptConnection) PacketReceipts() <-chan PacketReceipt {
	return c.in
}

func makePairedReceiptConnections() (ReceiptConnection, ReceiptConnection) {
	one := make(chan PacketReceipt)
	two := make(chan PacketReceipt)
	return testReceiptConnection{one, two}, testReceiptConnection{two, one}
}

type pr struct {
	hash string
	src  NodeAddress
}

func (p pr) ListPackets() []PacketHash { return []PacketHash{PacketHash(p.hash)} }
func (p pr) Source() NodeAddress       { return p.src }
func (p pr) Verify() error             { return nil }

func TestReceiptHandler(t *testing.T) {
	a1, a2 := NodeAddress("1"), NodeAddress("2")
	i1, i2 := make(chan RoutingDecision), make(chan RoutingDecision)
	ri1, ri2 := newReceipt(a1, i1), newReceipt(a2, i2)

	c1, c2 := makePairedReceiptConnections()
	ri1.AddConnection(a2, c1)
	ri2.AddConnection(a1, c2)

	p := testPacket("2")

	i1 <- RoutingDecision{p, a1, a2}
	i2 <- RoutingDecision{p, a1, a2}

	go ri2.SendReceipt(pr{"2", a2})

	<-ri2.PacketHashes()
	<-ri1.PacketHashes()
}
