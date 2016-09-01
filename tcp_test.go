package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os/user"
	"strings"
	"testing"

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

type testTCPData struct {
	in          chan types.Packet
	out         chan types.Packet
	write_error error
}

func (d testTCPData) SendPacket(p types.Packet) error {
	var req types.TCPTunnelRequest
	err := req.UnmarshalBinary(p.Data)
	if err != nil {
		resp := types.TCPTunnelResponse{net.ParseIP("fe80::")}
		resp_b, _ := resp.MarshalBinary()
		d.out <- types.Packet{"test", 7, resp_b}
		return nil
	} else {
		d.in <- p
		return d.write_error
	}
}
func (d testTCPData) Packets() <-chan types.Packet {
	return d.out
}

/*
func TestTCPTunToData(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p
	p_enc := <-data.in

	// Make sure the output packet is correct
	var p_after tuntap.Packet
	err = json.Unmarshal([]byte(p_enc.Data), &p_after)
	if err != nil {
		t.Fatal(err)
	}
	if p_after.Protocol != p.Protocol ||
		p_after.Truncated != p.Truncated ||
		!bytes.Equal(p_after.Packet, p.Packet) {
		t.Fatalf("%v != %v", p, p_after)
	}

	select {
	case err := <-tcp.Error():
		t.Fatal(err)
	default:
	}
}
*/

func TestTCPTunReadError(t *testing.T) {
	if !isRoot() {
		t.Skip()
	}

	amt := int64(7)
	dest := types.NodeAddress("destination")
	read_error := errors.New("Read Error")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), read_error, nil}
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p

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
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: true, Packet: []byte("test")}
	tun.out <- &p

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
	data := testTCPData{make(chan types.Packet, 1), make(chan types.Packet), write_error}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p

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
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	p_in := types.Packet{dest, amt, b}
	data.out <- p_in
	p_recv := <-tun.in

	// Make sure we receive the correct response
	if p_recv.Protocol != p.Protocol ||
		p_recv.Truncated != p.Truncated ||
		!bytes.Equal(p_recv.Packet, p.Packet) {
		t.Fatalf("%v != %v", p, p_recv)
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
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	p_in := types.Packet{dest, amt, b}
	data.out <- p_in

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
	data := testTCPData{make(chan types.Packet), make(chan types.Packet), nil}

	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		t.Fatal(err)
	}

	tcp := NewTCPTunnel(tun, data, dest, amt, strings.TrimRight(i.Name(), "\x00"))
	defer tcp.Close()

	// Send in a test packet
	p_in := types.Packet{dest, amt, []byte("NOTJSON")}
	data.out <- p_in

	err = <-tcp.Error()
	if err == nil {
		t.Fatalf("%v != nil", err)
	}
}
