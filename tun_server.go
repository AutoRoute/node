package node

import (
	"encoding/json"
	"fmt"
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
	quit        chan bool
	err         chan error
}

func NewTunServer(tun TCPTun, d DataConnection, amt int64) *TunServer {
	return &TunServer{d, tun, amt, make(map[string]types.NodeAddress), make(map[types.NodeAddress]bool), 0, make(chan bool), make(chan error, 1)}
}

func (ts *TunServer) Close() {
	close(ts.quit)
}

func (ts *TunServer) Error() chan error {
	return ts.err
}

func (ts *TunServer) connect(connectingNode types.NodeAddress) {
	ip := fmt.Sprintf("6666::%d", ts.currIP)
	ts.currIP++
	nodeAddr := connectingNode
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
		src := types.NodeAddress(p.Dest)
		if _, ok := ts.connections[src]; !ok {
			ts.connect(src)
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
