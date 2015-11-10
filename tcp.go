package node

import (
	"github.com/AutoRoute/tuntap"

	"encoding/json"
	"errors"
)

type TCP struct {
	data DataConnection
	tun  TCPTun
	dest NodeAddress
	amt  int64
	quit chan bool
	err  chan error
}

type TCPTun interface {
	ReadPacket() (*tuntap.Packet, error)
	WritePacket(p *tuntap.Packet) error
}

var truncated_error error = errors.New("truncated packet")

func NewTCPTunnel(tun TCPTun, d DataConnection, dest NodeAddress, amt int64) *TCP {
	t := &TCP{d, tun, dest, amt, make(chan bool), make(chan error, 1)}
	go t.readtun()
	go t.writetun()
	return t
}

func (t *TCP) Close() {
	close(t.quit)
}

func (t *TCP) Error() chan error {
	return t.err
}

func (t *TCP) readtun() {
	for {
		select {
		case <-t.quit:
			return
		default:
		}
		p, err := t.tun.ReadPacket()
		if err != nil {
			t.err <- err
			return
		}
		if p.Truncated {
			t.err <- truncated_error
			return
		}
		b, err := json.Marshal(p)
		if err != nil {
			t.err <- err
			return
		}
		ep := Packet{t.dest, t.amt, string(b)}
		err = t.data.SendPacket(ep)
		if err != nil {
			t.err <- err
			return
		}
	}
}

func (t *TCP) writetun() {
	for {
		select {
		case p := <-t.data.Packets():
			ep := &tuntap.Packet{}
			err := json.Unmarshal([]byte(p.Data), ep)
			if err != nil {
				t.err <- err
				return
			}
			err = t.tun.WritePacket(ep)
			if err != nil {
				t.err <- err
				return
			}
		case <-t.quit:
			return
		}
	}
}
