package node

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

func TestConnection(t *testing.T) {
	key1, _ := NewKey()
	key2, _ := NewKey()

	n1 := NewServer(key1, internal.FakeMoney{}, nil)
	defer n1.Close()
	n2 := NewServer(key2, internal.FakeMoney{}, nil)
	defer n2.Close()

	err := n1.Listen("[::1]:16543")
	if err != nil {
		t.Fatalf("Error listening %v", err)
	}
	err = n2.Connect("[::1]:16543")
	if err != nil {
		t.Fatalf("Error connecting %v", err)
	}
}

func WaitForReachable(n Node, addr types.NodeAddress) error {
	tick := time.Tick(100 * time.Millisecond)
	timeout := time.After(1 * time.Second)
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
	key1, _ := NewKey()
	key2, _ := NewKey()

	n1 := NewServer(key1, internal.FakeMoney{}, nil)
	defer n1.Close()
	n2 := NewServer(key2, internal.FakeMoney{}, nil)
	defer n2.Close()
	err := n1.Listen("[::1]:16544")
	if err != nil {
		t.Fatalf("Error listening %v", err)
	}
	err = n2.Connect("[::1]:16544")
	if err != nil {
		t.Fatalf("Error connecting %v", err)
	}

	// Wait for n1 to know about n2
	err = WaitForReachable(n1.Node(), key2.k.PublicKey().Hash())
	if err != nil {
		t.Fatalf("Error waiting for information %v", err)
	}

	for i := 0; i < 10; i++ {
		p2 := types.Packet{key2.k.PublicKey().Hash(), 3, fmt.Sprintf("test%d", i)}
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
}

func benchmarkDataTransmission(size int, b *testing.B) {
	key1, _ := NewKey()
	key2, _ := NewKey()

	n1 := NewServer(key1, internal.FakeMoney{}, nil)
	defer n1.Close()
	n2 := NewServer(key2, internal.FakeMoney{}, nil)
	defer n2.Close()
	err := n1.Listen(fmt.Sprintf("[::1]:16%03d", (size+b.N)%127+1))
	if err != nil {
		b.Fatalf("Error listening %v", err)
	}
	err = n2.Connect(fmt.Sprintf("[::1]:16%03d", (size+b.N)%127+1))
	if err != nil {
		b.Fatalf("Error connecting %v", err)
	}

	// Wait for n1 to know about n2
	err = WaitForReachable(n1.Node(), key2.k.PublicKey().Hash())
	if err != nil {
		b.Fatalf("Error waiting for information %v", err)
	}

	p2 := types.Packet{key2.k.PublicKey().Hash(), 3, strings.Repeat("a", size)}
	done := make(chan bool)

	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			<-n2.Node().Packets()
		}
		done <- true
	}()
	for i := 0; i < b.N; i++ {
		err = n1.Node().SendPacket(p2)
		if err != nil {
			b.Fatalf("Error sending packet: %v", err)
		}
	}
	<-done
}

func BenchmarkDataTransmission1(b *testing.B) {
	benchmarkDataTransmission(1, b)
}

func BenchmarkDataTransmission1k(b *testing.B) {
	benchmarkDataTransmission(1024, b)
}
