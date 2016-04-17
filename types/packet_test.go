package types

import (
	"crypto/rand"
	"encoding/json"
	"testing"
)

func TestPacketMarshalling(t *testing.T) {
	// Make sure that encoding can handle random binary data.
	b := make([]byte, 256)
	_, err := rand.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	m := Packet{NodeAddress(string(b)), 3, "test"}
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var m2 Packet
	err = json.Unmarshal(b, &m2)
	if err != nil {
		t.Fatal(err)
	}
	if m2 != m {
		t.Fatalf("Different packets? %v != %v", m2, m)
	}
	_ = m.String()
}
