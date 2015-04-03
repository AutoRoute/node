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

type ra chan PacketHash

func (r ra) Receipt(h PacketHash) {
	r <- h
}

type pr struct {
	hash string
	src  NodeAddress
}

func (p pr) ListPackets() []PacketHash { return []PacketHash{PacketHash(p.hash)} }
func (p pr) Source() NodeAddress       { return p.src }
func (p pr) Verify() error             { return nil }

func TestReceiptHandler(t *testing.T) {
	a1 := NodeAddress("1")
	a2 := NodeAddress("2")

	c1, c2 := makePairedReceiptConnections()
	ra1 := ra(make(chan PacketHash))
	ra2 := ra(make(chan PacketHash))
	ri1 := newReceiptImpl(a1, ra1)
	ri1.AddConnection(a2, c1)
	ri2 := newReceiptImpl(a2, ra2)
	ri2.AddConnection(a1, c2)

	p := testPacket("2")

	ri1.AddSentPacket(p, a1, a2)
	ri2.AddSentPacket(p, a1, a2)

	go ri2.SendReceipt(pr{"2", a2})

	<-ra2
	<-ra1
}
