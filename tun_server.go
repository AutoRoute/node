package node

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

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
	ip := fmt.Sprintf("2001::%d", ts.currIP)
	ts.currIP++
	nodeAddr := connectingNode
	ts.nodes[ip] = nodeAddr
	ts.connections[nodeAddr] = true

	resp := types.TCPTunnelResponse{net.ParseIP(ip)}
	resp_b, _ := resp.MarshalBinary()
	ep := types.Packet{nodeAddr, ts.amt, resp_b}
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
			var req types.TCPTunnelRequest
			err := req.UnmarshalBinary(p.Data)
			if err != nil {
				ts.err <- err
			}

			ts.connect(src)
		} else {
			/* do something with packet */
			var tcp_data types.TCPTunnelData
			err := tcp_data.UnmarshalBinary(p.Data)
			if err != nil {
				ts.err <- err
			}

			ep := &tuntap.Packet{}
			err = json.Unmarshal(tcp_data.Data, ep)
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
		version := p.Packet[0]
		if (version & 0xF) != 0x6 {
			continue
		}

		ip := net.IP(p.Packet[64:192])
		dest_node := ts.nodes[string(ip)]

		tcp_data := types.TCPTunnelData{b}
		tcp_data_b, _ := tcp_data.MarshalBinary()
		ep := types.Packet{dest_node, ts.amt, tcp_data_b}
		err = ts.data.SendPacket(ep)
		if err != nil {
			ts.err <- err
			return
		}
	}
}
