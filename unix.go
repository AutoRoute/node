package node

import (
	"encoding/json"
	"log"
	"net"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

type UnixSocket struct {
	l net.Listener
	d node.DataConnection
}

// Creates a unix socket which all packets are sent to /from.
func NewUnixSocket(path string, d node.DataConnection) (*UnixSocket, error) {
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	u := &UnixSocket{l, d}
	go u.awaitConnection()
	return u, nil
}

func (u *UnixSocket) awaitConnection() {
	for {
		c, err := u.l.Accept()
		if err != nil {
			log.Print(err)
			return
		}
		go u.sendPackets(c)
		go u.receivePackets(c)
	}
}

func (u *UnixSocket) sendPackets(c net.Conn) {
	dec := json.NewDecoder(c)
	for {
		p := types.Packet{}
		err := dec.Decode(&p)
		if err != nil {
			log.Print(err)
			return
		}
		err = u.d.SendPacket(p)
		if err != nil {
			log.Print(err)
		}
	}
}

func (u *UnixSocket) receivePackets(c net.Conn) {
	enc := json.NewEncoder(c)
	for p := range u.d.Packets() {
		err := enc.Encode(p)
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func (u *UnixSocket) Close() {
	u.l.Close()
}
