package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

var requestHeader = "hello"
var beginIPBlock = net.ParseIP("6666::0")
var endIPBlock = net.ParseIP("6666::6666")

type TunServer struct {
	data        DataConnection
	tun         TCPTun
	amt         int64
	nodes       map[string]types.NodeAddress
	connections map[types.NodeAddress]bool
	currIP      int
	err         chan error
}

func NewTunServer(d DataConnection, amt int64) *TunServer {
	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		log.Fatal(err)
	}
	return &TunServer{d, i, amt, make(map[string]types.NodeAddress), make(map[types.NodeAddress]bool), 0, make(chan error, 1)}
}

func (ts *TunServer) Error() chan error {
	return ts.err
}

func (ts *TunServer) connect(connectingNode string) {
	ip := fmt.Sprintf("6666::%d", ts.currIP)
	ts.currIP++
	nodeAddr := types.NodeAddress(connectingNode)
	ts.nodes[ip] = nodeAddr
	ts.connections[nodeAddr] = true
	ep := types.Packet{nodeAddr, ts.amt, ip}
	err := ts.data.SendPacket(ep)
	if err != nil {
		ts.err <- err
		return
	}
}

func (ts *TunServer) Listen() *TunServer {
	go ts.listenNode()
	go ts.listenTun()
	return ts
}

func (ts *TunServer) listenNode() {
	for p := range ts.data.Packets() {
        src := types.NodeAddress([]byte(p.Data)[12:16])
		if _, ok := ts.connections[src]; !ok {
			ts.connect(p.Data)
		} else {
			/* do something with packet */
			ep := &tuntap.Packet{}
			err := json.Unmarshal([]byte(p.Data), ep)
			if err != nil {
				ts.err <- err
				return
			}
			err = ts.tun.WritePacket(ep)
			if err != nil {
				ts.err <- err
				return
			}
			err = ts.tun.WritePacket(ep)
			if err != nil {
				ts.err <- err
				return
			}
		}
	}
}

func (ts *TunServer) listenTun() {
	for {
		p, err := ts.tun.ReadPacket()
		if err != nil {
			ts.err <- err
			return
		}
		if p.Truncated {
			ts.err <- truncated_error
			return
		}
		b, err := json.Marshal(p)
		if err != nil {
			ts.err <- err
			return
		}
		ip := string(p.Packet[0:6])
		dest_node := ts.nodes[ip]
		ep := types.Packet{dest_node, ts.amt, string(b)}
		err = ts.data.SendPacket(ep)
		if err != nil {
			ts.err <- err
			return
		}
	}
}
