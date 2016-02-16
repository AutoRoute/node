package internal

import (
	"testing"

	"github.com/AutoRoute/node/types"
)

func TestReceiptHandler(t *testing.T) {
	pk2, _ := NewECDSAKey()
	a1, a2 := types.NodeAddress("1"), pk2.PublicKey().Hash()
	i1, i2 := make(chan routingDecision), make(chan routingDecision)
	ri1, ri2 := newReceipt(a1, i1), newReceipt(a2, i2)
	defer ri1.Close()
	defer ri2.Close()

	c1, c2 := makePairedReceiptConnections()
	ri1.AddConnection(a2, c1)
	ri2.AddConnection(a1, c2)

	p := testPacket(a2)

	i1 <- newRoutingDecision(p, a1, a2)
	i2 <- newRoutingDecision(p, a1, a2)

	go ri2.SendReceipt(CreateMerkleReceipt(pk2, []types.PacketHash{p.Hash()}))

	<-ri2.PacketHashes()
	<-ri1.PacketHashes()
}
