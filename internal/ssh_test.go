package internal

import (
	"fmt"
	"net"
	"testing"

	"github.com/AutoRoute/node/types"
)

var port = 10000

func ConnectSSH(sk1, sk2 PrivateKey) (*SSHConnection, *SSHConnection, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	lt, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	m := func() SSHMetaData {
		return SSHMetaData{Payment_Address: "fake2"}
	}
	l := ListenSSH(lt, sk2, m)
	if l.Error() != nil {
		return nil, nil, l.Error()
	}

	ct, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	c1, err := EstablishSSH(ct, addr, sk1, SSHMetaData{Payment_Address: "fake1"})
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
	m.AddEntry(types.NodeAddress("1"))
	err = c1.SendMap(m)
	if err != nil {
		t.Fatal(err)
	}
	m2 := <-c2.ReachabilityMaps()
	if !m2.IsReachable(types.NodeAddress("1")) {
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

	m := CreateMerkleReceipt(sk1, []types.PacketHash{types.PacketHash("hi")})
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

	p := types.Packet{types.NodeAddress("foo"), 3, "test"}
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
