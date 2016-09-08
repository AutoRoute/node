package types

import (
	"bytes"
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

	m := Packet{NodeAddress(string(b)), 3, []byte("test")}
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var m2 Packet
	err = json.Unmarshal(b, &m2)
	if err != nil {
		t.Fatal(err)
	}
	if m2.Dest != m.Dest {
		t.Fatalf("Different dests? %v != %v", m2.Dest, m.Dest)
	}

	if m2.Amt != m.Amt {
		t.Fatalf("Different amounts? %v != %v", m2.Amt, m.Amt)
	}

	if bytes.Compare(m2.Data, m.Data) != 0 {
		t.Fatalf("Different data? %q != %q", m2.Data, m.Data)
	}
	_ = m.String()
}
