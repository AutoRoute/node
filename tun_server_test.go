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

func requestPacket(source types.NodeAddress) types.Packet {
	req := types.TCPTunnelRequest{source}
	req_b, _ := req.MarshalBinary()
	ep := types.Packet{"destination", 7, req_b}
	return ep
}

func TestReceiveRequest(t *testing.T) {
	amt := int64(7)
	source := types.NodeAddress("source")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	node.out <- requestPacket(source)

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-node.in
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
	source := types.NodeAddress("source")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	node.out <- requestPacket(source)

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-node.in
	err := resp.UnmarshalBinary(p_resp.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IP we are assigned is the first one the tun_server
	// hands out
	if !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
		t.Fatalf("%v != %v", resp.IP, "2001::")
	}

	// Create test node packet
	p := exampleTunPacket()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	// Send test node packet
	tcp_data := types.TCPTunnelData{b}
	tcp_data_b, _ := tcp_data.MarshalBinary()
	ep := types.Packet{"destination", 7, tcp_data_b}
	node.out <- ep

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
	source := types.NodeAddress("source")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send test request packet
	node.out <- requestPacket(source)

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_resp := <-node.in
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
	p := exampleTunPacket()
	tun.out <- &p

	// Receive the packet from the node connection
	ep := <-node.in
	var tcp_data types.TCPTunnelData
	err = tcp_data.UnmarshalBinary(ep.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we got it from the node connection
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
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), read_error, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := exampleTunPacket()
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != read_error {
		t.Fatalf("%v != %v", read_error, err)
	}
}

func TestTunServerReadTruncated(t *testing.T) {
	amt := int64(7)
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
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
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet, 1), make(chan types.Packet), write_error}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := exampleTunPacket()
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != write_error {
		t.Fatalf("%v != %v", write_error, err)
	}
}

func TestTunServerWriteSendError(t *testing.T) {
	amt := int64(7)
	source := types.NodeAddress("source")
	write_error := errors.New("Write Error")
	tun := testTun{make(chan *tuntap.Packet, 1), make(chan *tuntap.Packet), nil, write_error}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Start handshake
	node.out <- requestPacket(source)

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_back := <-node.in
	err := resp.UnmarshalBinary(p_back.Data)
	if err != nil {
		t.Fatal(err)
	}

	if p_back.Dest != source || p_back.Amt != amt || !bytes.Equal(resp.IP, net.ParseIP("2001::")) {
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
	p_in := types.Packet{source, amt, tcp_data_b}
	node.out <- p_in

	err = <-tunserver.Error()
	if err != write_error {
		t.Fatalf("%v != %v", err, write_error)
	}
}

func TestTunServerWriteUnmarshalError(t *testing.T) {
	amt := int64(7)
	source := types.NodeAddress("source")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTCPTunServer(node, tun, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Start handshake
	node.out <- requestPacket(source)

	// See if we get a response
	var resp types.TCPTunnelResponse
	p_back := <-node.in
	err := resp.UnmarshalBinary(p_back.Data)
	if err != nil {
		t.Fatal(err)
	}

	if p_back.Dest != source || p_back.Amt != amt || !bytes.Equal(resp.IP, net.ParseIP("2001::0")) {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	tcp_data := types.TCPTunnelData{[]byte("NOTJSON")}
	tcp_data_b, _ := tcp_data.MarshalBinary()
	p_in := types.Packet{source, amt, tcp_data_b}
	node.out <- p_in

	err = <-tunserver.Error()
	if err == nil {
		t.Fatalf("%v != nil", err)
	}
}
