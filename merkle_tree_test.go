package node

import (
	"testing"
)

func TestMerkle(t *testing.T) {
	k, _ := NewECDSAKey()

	p := make([]PacketHash, 3)
	p[0] = "Fo0"
	p[1] = "Fo1"
	p[2] = "Fo2"

	pr := CreateMerkleReceipt(k, p)
	if err := pr.Verify(); err != nil {
		t.Fatal(err)
	}
	if pr.Source() != k.PublicKey().Hash() {
		t.Fatal("Expected %v got %v", pr.Source(), k.PublicKey().Hash())
	}
	for i, _ := range pr.ListPackets() {
		if pr.ListPackets()[i] != p[i] {
			t.Fatal("Expected %v got %v", pr.ListPackets()[i], p[i])
		}
	}
}
