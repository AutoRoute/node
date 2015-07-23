package node

import (
	"testing"
)

func TestDirect(t *testing.T) {
	key1, _ := NewECDSAKey()
	key2, _ := NewECDSAKey()

	n1 := NewServer(key1)
	n2 := NewServer(key2)

	err := n1.Listen("127.0.0.1:16543")
	if err != nil {
		t.Fatalf("Error listening %v", err)
	}
	err = n2.Connect("127.0.0.1:16543")
	if err != nil {
		t.Fatalf("Error connecting %v", err)
	}
}
