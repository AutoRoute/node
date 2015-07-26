package node

import (
	"errors"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
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

func WaitForReachable(n *Node, addr NodeAddress) error {
	tick := time.Tick(100 * time.Millisecond)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-tick:
			if n.IsReachable(addr) {
				return nil
			}
		case <-timeout:
			return errors.New("Timed out waiting for node to be reachable")
		}
	}

}

func TestDataTransmission(t *testing.T) {
	key1, _ := NewECDSAKey()
	key2, _ := NewECDSAKey()

	n1 := NewServer(key1)
	n2 := NewServer(key2)
	err := n1.Listen("127.0.0.1:16544")
	if err != nil {
		t.Fatalf("Error listening %v", err)
	}
	err = n2.Connect("127.0.0.1:16544")
	if err != nil {
		t.Fatalf("Error connecting %v", err)
	}

	// Wait for n1 to know about n2
	err = WaitForReachable(n1.Node(), key2.PublicKey().Hash())
	if err != nil {
		t.Fatalf("Error waiting for information %v", err)
	}

	p2 := testPacket(key2.PublicKey().Hash())
	err = n1.Node().SendPacket(p2)
	if err != nil {
		t.Fatalf("Error sending packet: %v", err)
	}

	timeout := time.After(10 * time.Second)
	for found := false; found != true; {
		select {
		case <-n2.Node().Packets():
			found = true
			break
		case <-timeout:
			t.Fatal("Timeout waiting for packets")
		}
	}

}
