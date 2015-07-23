package node

import (
	"log"
	"testing"
	"time"
)

type testDataConnection struct {
	in  chan Packet
	out chan Packet
}

func (c testDataConnection) SendPacket(m Packet) error {
	c.out <- m
	return nil
}
func (c testDataConnection) Packets() <-chan Packet {
	return c.in
}
func (c testDataConnection) Close() error {
	close(c.in)
	return nil
}

func makePairedDataConnections() (testDataConnection, testDataConnection) {
	one := make(chan Packet)
	two := make(chan Packet)
	return testDataConnection{one, two}, testDataConnection{two, one}
}

type testConnection struct {
	testDataConnection
	testMapConnection
	testReceiptConnection
	testPaymentConnection
	k PublicKey
}

func (t testConnection) Close() error {
	t.testDataConnection.Close()
	t.testMapConnection.Close()
	t.testReceiptConnection.Close()
	t.testPaymentConnection.Close()
	return nil
}

func (t testConnection) Key() PublicKey { return t.k }

func makePairedConnections(k1, k2 PublicKey) (Connection, Connection) {
	d1, d2 := makePairedDataConnections()
	m1, m2 := makePairedMapConnections()
	r1, r2 := makePairedReceiptConnections()
	p1, p2 := makePairedPaymentConnections()
	return testConnection{d1, m1, r1, p1, k1}, testConnection{d2, m2, r2, p2, k2}
}

func testPacket(n NodeAddress) Packet {
	return Packet{n, 3, "test"}
}

type linkable interface {
	GetAddress() PublicKey
	AddConnection(Connection)
}

func link(a, b linkable) {
	c1, c2 := makePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

func TestDirectRouter(t *testing.T) {
	sk1, _ := NewECDSAKey()
	k1 := sk1.PublicKey()
	sk2, _ := NewECDSAKey()
	k2 := sk2.PublicKey()
	r1 := newRouter(k1)
	r2 := newRouter(k2)
	defer r1.Close()
	defer r2.Close()
	a2 := k2.Hash()
	link(r1, r2)

	// Send a test packet over the connection
	p2 := testPacket(a2)
	go func() {
		err := r1.SendPacket(p2)
		if err != nil {
			t.Fatal(err)
		}
	}()

	received := <-r2.Packets()
	if received != p2 {
		t.Fatalf("%q != %q", received, p2)
	}

	// Make sure a bad packet fails
	a3 := NodeAddress("3")
	p3 := testPacket(a3)
	err := r1.SendPacket(p3)
	if err == nil {
		t.Fatal("Expected error got nil")
	}
}

func TestRelayRouter(t *testing.T) {
	sk1, err := NewECDSAKey()
	k1 := sk1.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	sk2, err := NewECDSAKey()
	k2 := sk2.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	sk3, err := NewECDSAKey()
	k3 := sk3.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	a3 := k3.Hash()
	r1 := newRouter(k1)
	r2 := newRouter(k2)
	r3 := newRouter(k3)
	defer r1.Close()
	defer r2.Close()
	defer r3.Close()
	link(r1, r2)
	link(r2, r3)

	// Send a test packet over the connection
	p3 := testPacket(a3)
	clean := make(chan bool)
	go func() {
		defer close(clean)
		tries := time.Tick(10 * time.Millisecond)
		timeout := time.After(time.Second)
		for {
			select {
			case <-tries:
				err := r1.SendPacket(p3)
				if err == nil {
					return
				}
			case <-timeout:
				log.Fatal("Timed out waiting for succesful send")
			}
		}
	}()

	received := <-r3.Packets()
	if received != p3 {
		t.Fatalf("%q != %q", received, p3)
	}
	<-clean
}
