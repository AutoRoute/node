package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os/user"
	"strings"
	"testing"
	"time"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

func isRoot() bool {
	user, err := user.Current()
	if err != nil {
		return false
	}
	if user.Username != "root" {
		return false
	}
	return true
}

type testTun struct {
	in          chan *tuntap.Packet
	out         chan *tuntap.Packet
	read_error  error
	write_error error
}

func (t testTun) ReadPacket() (*tuntap.Packet, error) {
	return <-t.out, t.read_error
}
func (t testTun) WritePacket(p *tuntap.Packet) error {
	t.in <- p
	return t.write_error
}

type testNode struct {
	in          chan types.Packet
	out         chan types.Packet
	write_error error
}

func (n testNode) SendPacket(p types.Packet) error {
	n.in <- p
	return n.write_error
}
func (n testNode) Packets() <-chan types.Packet {
	return n.out
}
func (n testNode) GetNodeAddress() types.NodeAddress {
	return "source"
}

func exampleTunPacket() tuntap.Packet {
	p := make([]byte, 192, 192)
	p[0] = 0x6
	addr := []byte(net.ParseIP("2002::"))
	copy(p[64:192], addr)
	tp := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: p}
	return tp
}

func TestTCPTunRequest(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, i.Name())
	defer tcp.Close()

	// Check request from new tunnel
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-tcp.Error():
		t.Fatal(err)
	default:
	}
}

func TestTCPTunResponse(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	// Start new client
	tcp := NewTCPTunClient(node, tun, dest, amt, i.Name())
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	time.Sleep(300 * time.Millisecond)

	// Make sure the device is set up with the correct address
	tun_name := strings.Trim(i.Name(), "\x00")

	ifi, err := net.InterfaceByName(tun_name)
	if err != nil {
		t.Fatal(err)
	}

	ifi_addrs, err := ifi.Addrs()
	if err != nil {
		t.Fatal(err)
	}

	for _, addr := range ifi_addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			t.Fatal(err)
		}
		if ip.String() == "2001::" {
			return
		}
	}
	t.Fatalf("Client did not set up addrss given on interface")
}

func TestTCPTunToData(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	// Start new client
	tcp := NewTCPTunClient(node, tun, dest, amt, i.Name())
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Write to client's tun device
	tp := exampleTunPacket()
	tun.out <- &tp

	// Receive the packet over AutoRoute and make
	// sure it's the correct message type (TCPTunnelData)
	p = <-node.in
	var tcp_data types.TCPTunnelData
	err = tcp_data.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the output packet is correct
	var tp_after tuntap.Packet
	err = json.Unmarshal(tcp_data.Data, &tp_after)
	if err != nil {
		t.Fatal(err)
	}
	if tp_after.Protocol != tp.Protocol ||
		tp_after.Truncated != tp.Truncated ||
		!bytes.Equal(tp_after.Packet, tp.Packet) {
		t.Fatalf("%q != %q", tp, tp_after)
	}

	select {
	case err := <-tcp.Error():
		t.Fatal(err)
	default:
	}
}

func TestTCPTunReadError(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	read_error := errors.New("Read Error")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), read_error, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	tp := exampleTunPacket()
	tun.out <- &tp

	// Make sure the error appears
	err = <-tcp.Error()
	if err != read_error {
		t.Fatalf("%v != %v", read_error, err)
	}
}

func TestTCPTunReadTruncated(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	tp := exampleTunPacket()
	tp.Truncated = true
	tun.out <- &tp

	// Make sure the error appears
	err = <-tcp.Error()
	if err != truncated_error {
		t.Fatalf("%v != %v", truncated_error, err)
	}
}

func TestTCPTunReadWriteFails(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	write_error := errors.New("Write Error")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet, 1), make(chan types.Packet), write_error}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	tp := exampleTunPacket()
	tun.out <- &tp

	// Make sure the error appears
	err = <-tcp.Error()
	if err != write_error {
		t.Fatalf("%v != %v", write_error, err)
	}
}

func TestTCPTunWrite(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	tp := exampleTunPacket()
	b, err := json.Marshal(&tp)
	if err != nil {
		t.Fatal(err)
	}

	tcp_data := types.TCPTunnelData{b}
	tcp_data_b, _ := tcp_data.MarshalBinary()

	p_in := types.Packet{dest, amt, tcp_data_b}
	node.out <- p_in
	tp_recv := <-tun.in

	// Make sure we receive the correct response
	if tp_recv.Protocol != tp.Protocol ||
		tp_recv.Truncated != tp.Truncated ||
		!bytes.Equal(tp_recv.Packet, tp.Packet) {
		t.Fatalf("%v != %v", tp, tp_recv)
	}
}

func TestTCPTunWriteSendError(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	write_error := errors.New("Write Error")
	tun := testTun{make(chan *tuntap.Packet, 1), make(chan *tuntap.Packet), nil, write_error}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	tp := exampleTunPacket()
	b, err := json.Marshal(&tp)
	if err != nil {
		t.Fatal(err)
	}

	tcp_data := types.TCPTunnelData{b}
	tcp_data_b, _ := tcp_data.MarshalBinary()
	p_in := types.Packet{dest, amt, tcp_data_b}
	node.out <- p_in

	err = <-tcp.Error()
	if err != write_error {
		t.Fatalf("%v != %v", err, write_error)
	}
}

func TestTCPTunWriteUnmarshalError(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	node := testNode{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunClient(node, tun, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Check request from new client
	p := <-node.in
	var req types.TCPTunnelRequest
	err = req.UnmarshalBinary(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	// Have server send back response
	resp := types.TCPTunnelResponse{net.ParseIP("2001::")}
	resp_b, _ := resp.MarshalBinary()
	p = types.Packet{"source", amt, resp_b}
	node.out <- p

	// Send in a test packet
	p_in := types.Packet{dest, amt, []byte("NOTJSON")}
	node.out <- p_in

	err = <-tcp.Error()
	if err == nil {
		t.Fatalf("%v != nil", err)
	}
}
