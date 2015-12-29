package node

import (
	"encoding/json"
	"testing"

	"github.com/AutoRoute/node/types"
)

func TestMerkle(t *testing.T) {
	k, _ := NewECDSAKey()

	p := make([]types.PacketHash, 3)
	p[0] = "Fo0"
	p[1] = "Fo1"
	p[2] = "Fo2"

	pr := CreateMerkleReceipt(k, p)
	if err := pr.Verify(); err != nil {
		t.Fatal(err)
	}
	if pr.Source() != k.PublicKey().Hash() {
		t.Fatalf("Expected %v got %v", pr.Source(), k.PublicKey().Hash())
	}
	if len(pr.ListPackets()) != len(p) {
		t.Fatalf("Expected %v got %v", pr.ListPackets(), p)
	}
	for _, i := range pr.ListPackets() {
		found := false
		for _, j := range p {
			if j == i {
				found = true
			}
		}
		if !found {
			t.Fatalf("Expected %v in %v", i, p)
		}
	}
}

func TestMerkleMarshalling(t *testing.T) {
	sk1, _ := NewECDSAKey()
	m := CreateMerkleReceipt(sk1, []types.PacketHash{types.PacketHash("hi")})
	if m.Verify() != nil {
		t.Fatalf("Error verifying generated receipt: %v", m.Verify())
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var m2 PacketReceipt
	err = json.Unmarshal(b, &m2)
	if err != nil {
		t.Fatal(err)
	}
}
