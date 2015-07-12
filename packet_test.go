package node

import (
	"encoding/json"
	"testing"
)

func TestPacketMarshalling(t *testing.T) {
	m := Packet{NodeAddress("foo"), 3, "test"}
	b, err := json.Marshal(m)
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
}
