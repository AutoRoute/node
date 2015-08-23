package node

import (
	"code.google.com/p/tuntap"

	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type testTun struct {
	in          chan *tuntap.Packet
	out         chan *tuntap.Packet
	read_error  error
	write_error error
}

func (t testTun) Close() error { return nil }
func (t testTun) Name() string { return "dummy" }

func (t testTun) ReadPacket() (*tuntap.Packet, error) {
	return <-t.out, t.read_error
}
func (t testTun) WritePacket(p *tuntap.Packet) error {
	t.in <- p
	return t.write_error
}

type testData struct {
	in          chan Packet
	out         chan Packet
	write_error error
}

func (d testData) SendPacket(p Packet) error {
	d.in <- p
	return d.write_error
}
func (d testData) Packets() <-chan Packet {
	return d.out
}

func TestTCPTunToData(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p
	p_enc := <-data.in

	// Make sure the output packet is correct
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
	case err := <-tcp.Error():
		t.Fatal(err)
	default:
	}
}

func TestTCPTunReadError(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	read_error := errors.New("Read Error")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), read_error, nil}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p

	// Make sure the error appears
	err := <-tcp.Error()
	if err != read_error {
		t.Fatalf("%v != %v", read_error, err)
	}
}

func TestTCPTunReadTruncated(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: true, Packet: []byte("test")}
	tun.out <- &p

	// Make sure the error appears
	err := <-tcp.Error()
	if err != truncated_error {
		t.Fatalf("%v != %v", truncated_error, err)
	}
}

func TestTCPTunReadWriteFails(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	write_error := errors.New("Write Error")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testData{make(chan Packet, 1), make(chan Packet), write_error}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	tun.out <- &p

	// Make sure the error appears
	err := <-tcp.Error()
	if err != write_error {
		t.Fatalf("%v != %v", write_error, err)
	}
}

func TestTCPTunWrite(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	p_in := Packet{dest, amt, string(b)}
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
	amt := int64(7)
	dest := NodeAddress("destination")
	write_error := errors.New("Write Error")
	tun := testTun{make(chan *tuntap.Packet, 1), make(chan *tuntap.Packet), nil, write_error}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p := tuntap.Packet{Protocol: 0x8000, Truncated: false, Packet: []byte("test")}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	p_in := Packet{dest, amt, string(b)}
	data.out <- p_in

	err = <-tcp.Error()
	if err != write_error {
		t.Fatalf("%v != %v", err, write_error)
	}
}

func TestTCPTunWriteUnmarshalError(t *testing.T) {
	amt := int64(7)
	dest := NodeAddress("destination")
	tun := testTun{make(chan *tuntap.Packet), make(chan *tuntap.Packet), nil, nil}
	data := testData{make(chan Packet), make(chan Packet), nil}
	tcp := NewTCPTunnel(tun, data, dest, amt)
	defer tcp.Close()

	// Send in a test packet
	p_in := Packet{dest, amt, string("NOTJSON")}
	data.out <- p_in

	err := <-tcp.Error()
	if err == nil {
		t.Fatalf("v != nil", err, nil)
	}
}
