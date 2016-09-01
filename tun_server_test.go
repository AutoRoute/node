package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"testing"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

type testTCPTun struct {
	in          chan *tuntap.Packet
	out         chan *tuntap.Packet
	read_error  error
	write_error error
}

func (t testTCPTun) ReadPacket() (*tuntap.Packet, error) {
	return <-t.out, t.read_error
}
func (t testTCPTun) WritePacket(p *tuntap.Packet) error {
	t.in <- p
	return t.write_error
}

type testTCPTunData struct {
	in          chan types.Packet
	out         chan types.Packet
	write_error error
}

func (d testTCPTunData) SendPacket(p types.Packet) error {
	d.in <- p
	return d.write_error
}
func (d testTCPTunData) Packets() <-chan types.Packet {
	return d.out
}

func readTunPacket() tuntap.Packet {
	p := make([]byte, 192, 192)
	p[0] = 0x6
	addr := []byte(net.ParseIP("2002::"))
	copy(p[64:192], addr)
	ep := tuntap.Packet{0, false, p}
	return ep
}

func requestPacket() types.Packet {
	req := types.TCPTunnelRequest{}
	req_b, _ := req.MarshalBinary()
	ep := types.Packet{"destination", 7, req_b}
	return ep
}

func TestTunRequest(t *testing.T) {
	amt := int64(7)
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	data.out <- requestPacket()

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-data.in
	err := resp.UnmarshalBinary(p_resp.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IP we are assigned is the first one the tun_server
	// hands out
	if !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
		t.Fatalf("%v != %v", resp.IP, "2001::")
	}

	select {
	case err := <-tunserver.Error():
		t.Fatal(err)
	default:
	}
}

func TestListenNodeData(t *testing.T) {
	amt := int64(7)
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	data.out <- requestPacket()

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-data.in
	err := resp.UnmarshalBinary(p_resp.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IP we are assigned is the first one the tun_server
	// hands out
	if !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
		t.Fatalf("%v != %v", resp.IP, "2001::")
	}

	// Create test data packet
	p := readTunPacket()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	// Send test data packet
	tcp_data := types.TCPTunnelData{b}
	tcp_data_b, _ := tcp_data.MarshalBinary()
	ep := types.Packet{"destination", 7, tcp_data_b}
	data.out <- ep

	// Make sure we got it on the other end
	p_after := <-tun.in

	if p_after.Protocol != p.Protocol ||
		p_after.Truncated != p.Truncated ||
		!bytes.Equal(p_after.Packet, p.Packet) {
		t.Fatalf("%q @= %q", p, p_after)
	}

	select {
	case err := <-tunserver.Error():
		t.Fatal(err)
	default:
	}
}

func TestListenTun(t *testing.T) {
	amt := int64(7)
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	data.out <- requestPacket()

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-data.in
	err := resp.UnmarshalBinary(p_resp.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IP we are assigned is the first one the tun_server
	// hands out
	if !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
		t.Fatalf("%v != %v", resp.IP, "2001::")
	}

	// Send test tun packet
	p := readTunPacket()
	tun.out <- &p

	// Receive the packet from the data connection
	ep := <-data.in
	var tcp_data types.TCPTunnelData
	err = tcp_data.UnmarshalBinary(ep.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we got it from the data connection
	p_after := &tuntap.Packet{}
	err = json.Unmarshal(tcp_data.Data, p_after)
	if err != nil {
		t.Fatal(err)
	}

	if p_after.Protocol != p.Protocol ||
		p_after.Truncated != p.Truncated ||
		!bytes.Equal(p_after.Packet, p.Packet) {
		t.Fatalf("%q @= %q", p, p_after)
	}

	select {
	case err := <-tunserver.Error():
		t.Fatal(err)
	default:
	}
}

func TestTunServerReadError(t *testing.T) {
	amt := int64(7)
	read_error := errors.New("Read Error")
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), read_error, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := readTunPacket()
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != read_error {
		t.Fatalf("%v != %v", read_error, err)
	}
}

func TestTunServerReadTruncated(t *testing.T) {
	amt := int64(7)
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: true, Packet: []byte("test")}
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != truncated_error {
		t.Fatalf("%v != %v", truncated_error, err)
	}
}

func TestTunServerReadWriteFails(t *testing.T) {
	amt := int64(7)
	write_error := errors.New("Write Error")
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet, 1), make(chan types.Packet), write_error}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := readTunPacket()
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != write_error {
		t.Fatalf("%v != %v", write_error, err)
	}
}

func TestTunServerWriteSendError(t *testing.T) {
	amt := int64(7)
	dest := types.NodeAddress("destination")
	write_error := errors.New("Write Error")
	tun := testTCPTun{make(chan *tuntap.Packet, 1), make(chan *tuntap.Packet), nil, write_error}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Start handshake
	data.out <- requestPacket()

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_back := <-data.in
	err := resp.UnmarshalBinary(p_back.Data)
	if err != nil {
		t.Fatal(err)
	}

	if p_back.Dest != dest || p_back.Amt != amt || !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}

	tcp_data := types.TCPTunnelData{b}
	tcp_data_b, _ := tcp_data.MarshalBinary()
	p_in := types.Packet{dest, amt, tcp_data_b}
	data.out <- p_in

	err = <-tunserver.Error()
	if err != write_error {
		t.Fatalf("%v != %v", err, write_error)
	}
}

func TestTunServerWriteUnmarshalError(t *testing.T) {
	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Start handshake
	data.out <- requestPacket()

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_back := <-data.in
	err := resp.UnmarshalBinary(p_back.Data)
	if err != nil {
		t.Fatal(err)
	}

	if p_back.Dest != dest || p_back.Amt != amt || !bytes.Equal(resp.IP, net.ParseIP("2001::0")) {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	p_in := types.Packet{dest, amt, []byte("NOTJSON")}
	data.out <- p_in

	err = <-tunserver.Error()
	if err == nil {
		t.Fatalf("%v != nil", err)
	}
}
