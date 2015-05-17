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

func TestReceiptHandler(t *testing.T) {
	pk2, _ := NewECDSAKey()
	a1, a2 := NodeAddress("1"), pk2.PublicKey().Hash()
	i1, i2 := make(chan RoutingDecision), make(chan RoutingDecision)
	ri1, ri2 := newReceipt(a1, i1), newReceipt(a2, i2)

	c1, c2 := makePairedReceiptConnections()
	ri1.AddConnection(a2, c1)
	ri2.AddConnection(a1, c2)

	p := testPacket(a2)

	i1 <- NewRoutingDecision(p, a1, a2)
	i2 <- NewRoutingDecision(p, a1, a2)

	go ri2.SendReceipt(CreateMerkleReceipt(pk2, []PacketHash{p.Hash()}))

	<-ri2.PacketHashes()
	<-ri1.PacketHashes()
}
