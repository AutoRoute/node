package node

import (
	"code.google.com/p/tuntap"

	"encoding/json"
	"errors"
	"log"
)

type TCP struct {
	data DataConnection
	tun  *tuntap.Interface
	dest NodeAddress
	amt  int64
	quit chan bool
	err  error
}

func NewTCPTunnel(tun *tuntap.Interface, d DataConnection, dest NodeAddress, amt int64) *TCP {
	t := &TCP{d, tun, dest, amt, make(chan bool), nil}
	go t.readtun()
	go t.writetun()
	return t
}

func (t *TCP) Close() {
	t.quit <- true
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
			log.Print(err)
			t.err = err
			return
		}
		if p.Truncated {
			t.err = errors.New("truncated packet?")
			continue
		}
		b, err := json.Marshal(p)
		if err != nil {
			log.Print(err)
			t.err = err
			return
		}
		ep := Packet{t.dest, t.amt, string(b)}
		err = t.data.SendPacket(ep)
		if err != nil {
			log.Print(err)
			t.err = err
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
				log.Print(err)
				t.err = err
				return
			}
			err = t.tun.WritePacket(ep)
			if err != nil {
				log.Print(err)
				t.err = err
				return
			}
		case <-t.quit:
			return
		}
	}
}
