package node

import (
	"fmt"
	"net"
	"testing"
)

var port = 10000

func ConnectSSH(sk1, sk2 PrivateKey) (*SSHConnection, *SSHConnection, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	lt, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	l := ListenSSH(lt, sk2)
	if l.Error() != nil {
		return nil, nil, l.Error()
	}

	ct, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	c1, err := EstablishSSH(ct, addr, sk1)
	port += 1
	if err != nil {
		return nil, nil, err
	}

	c2 := <-l.Connections()
	return c1, c2, nil
}

func TestSSHEstablishment(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	c1.Close()
	c2.Close()
}

func TestSSHPaymentTransmission(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	defer c1.Close()
	defer c2.Close()

	h := PaymentHash("foo")
	err = c1.SendPayment(h)
	if err != nil {
		t.Fatal(err)
	}
	r := <-c2.Payments()
	if r != h {
		t.Fatalf("Error in relaying received %v, expected %v", r, h)
	}
}

func TestSSHMapTransmission(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	defer c1.Close()
	defer c2.Close()

	m := NewBloomReachabilityMap()
	m.AddEntry(NodeAddress("1"))
	err = c1.SendMap(m)
	if err != nil {
		t.Fatal(err)
	}
	m2 := <-c2.ReachabilityMaps()
	if !m2.IsReachable(NodeAddress("1")) {
		t.Fatalf("1 not in %v", m2)
	}
}

func TestSSHReceiptTransmission(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	defer c1.Close()
	defer c2.Close()

	m := CreateMerkleReceipt(sk1, []PacketHash{PacketHash("hi")})
	if m.Verify() != nil {
		t.Fatalf("Error verifying generated receipt: %v", m.Verify())
	}
	err = c1.SendReceipt(m)
	if err != nil {
		t.Fatal(err)
	}
	r := <-c2.PacketReceipts()
	if r.Verify() != nil {
		t.Fatalf("Error verifying received receipt: %v", r.Verify())
	}
}

func TestSSHPacketTransmission(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	defer c1.Close()
	defer c2.Close()

	p := Packet{NodeAddress("foo"), 3, "test"}
	err = c1.SendPacket(p)
	if err != nil {
		t.Fatal(err)
	}
	p2 := <-c2.Packets()
	if p2 != p {
		t.Fatalf("Different packets? %v != %v", p2, p)
	}
}

func TestSSHSatisfiesConnection(t *testing.T) {
	_ = Connection(&SSHConnection{})
}

func TestSSHKeysAreCorrect(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c1, c2, err := ConnectSSH(sk1, sk2)
	if err != nil {
		t.Fatalf("Problems establish ssh connection: %v", err)
		return
	}
	if c1.Key().Hash() != sk2.PublicKey().Hash() {
		t.Fatalf("Hashes don't match %x != %x", c1.Key().Hash(), sk2.PublicKey().Hash())
	}
	if c2.Key().Hash() != sk1.PublicKey().Hash() {
		t.Fatalf("Hashes don't match %x != %x", c2.Key().Hash(), sk1.PublicKey().Hash())
	}
	c1.Close()
	c2.Close()
}
