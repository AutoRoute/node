package node

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestTunToData(t *testing.T) {
	amt := int64(7)
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p
	p_enc := <-data.in

	// Make sure the output is correct
	var p_after tuntap.Packet
	err := json.Unmarshal([]byte(p_enc.Data), &p_after)
	if err != nil {
		t.Fatal(err)
	}
	if p_after.Protocol != p.Protocol ||
		p_after.Truncated != p.Truncated ||
		!bytes.Equal(p_after.Packet, p.Packet) {
		t.Fatalf("%v != %v", p, p_after)
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
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
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
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p

	// Make sure the error appears
	err := <-tunserver.Error()
	if err != write_error {
		t.Fatalf("%v != %v", write_error, err)
	}
}

func TestTunServerWrite(t *testing.T) {
	amt := int64(7)
	dest := types.NodeAddress("destination")
	tun := testTCPTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testTCPTunData{make(chan types.Packet), make(chan types.Packet), nil}
	tunserver := NewTunServer(tun, data, amt)
	tunserver.Listen()
	defer tunserver.Close()

	// Start handshake
	p_in := types.Packet{dest, amt, "hello"}
	data.out <- p_in
	p_back := <-data.in

	if p_back.Dest != dest || p_back.Amt != amt || p_back.Data != "6666::0" {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}

	p_in = types.Packet{dest, amt, string(b)}

	data.out <- p_in
	p_recv := <-tun.in

	// Make sure we receive the correct response
	if p_recv.Protocol != p.Protocol ||
		p_recv.Truncated != p.Truncated ||
		!bytes.Equal(p_recv.Packet, p.Packet) {
		t.Fatalf("%v != %v", p, p_recv)
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
	p_in := types.Packet{dest, amt, "hello"}
	data.out <- p_in
	p_back := <-data.in

	if p_back.Dest != dest || p_back.Amt != amt || p_back.Data != "6666::0" {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}

	p_in = types.Packet{dest, amt, string(b)}
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
	p_in := types.Packet{dest, amt, "hello"}
	data.out <- p_in
	p_back := <-data.in

	if p_back.Dest != dest || p_back.Amt != amt || p_back.Data != "6666::0" {
		t.Fatalf("Incorrect handshake packet received")
	}

	// Send in a test packet
	p_in = types.Packet{dest, amt, string("NOTJSON")}
	data.out <- p_in

	err := <-tunserver.Error()
	if err == nil {
		t.Fatalf("%v != nil", err)
	}
}
