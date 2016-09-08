package node

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

type TCPTunServer struct {
	node        NodeConnection
	tun         TCPTun
	amt         int64
	nodes       map[string]types.NodeAddress
	connections map[types.NodeAddress]bool
	currIP      int
	quit        chan bool
	err         chan error
}

// NewTCPTunServer just constructs and returns a TCPTunServer with the given paramters.
// Unlike NewTCPTunClient, it does not start listening.
func NewTCPTunServer(n NodeConnection, tun TCPTun, amt int64) *TCPTunServer {
	return &TCPTunServer{n, tun, amt, make(map[string]types.NodeAddress), make(map[types.NodeAddress]bool), 0, make(chan bool), make(chan error, 1)}
}

func (ts *TCPTunServer) Close() {
	close(ts.quit)
}

func (ts *TCPTunServer) Error() chan error {
	return ts.err
}

// connect takes the NodeAddress of the connecting Node (the Node that sent a request to tunnel),
// adds it to the connection table, and assigns and sends it an IP address
func (ts *TCPTunServer) connect(connectingNode types.NodeAddress) {
	ip := fmt.Sprintf("2001::%d", ts.currIP)
	ts.currIP++
	nodeAddr := connectingNode
	ts.nodes[ip] = nodeAddr
	ts.connections[nodeAddr] = true

	resp := types.TCPTunnelResponse{net.ParseIP(ip)}
	resp_b, _ := resp.MarshalBinary()
	ep := types.Packet{nodeAddr, ts.amt, resp_b}
	err := ts.node.SendPacket(ep)
	if err != nil {
		ts.err <- err
		return
	}
}

// Listen() starts listening on the tun and AutoRoute connection
func (ts *TCPTunServer) Listen() {
	go ts.listenNode()
	go ts.listenTun()
}

// listenNode reads a packet from the node connection and determines whether
// it's a TCP tunnel request or data packet. If it's a request to tunnel, connect
// is called. Otherwise the packet is unwrapped and sent out the tun device.
func (ts *TCPTunServer) listenNode() {
	for p := range ts.node.Packets() {
		// src := types.NodeAddress(p.Dest)
		// if _, ok := ts.connections[src]; !ok {
		if p.Data[1] == 0 {
			var req types.TCPTunnelRequest
			err := req.UnmarshalBinary(p.Data)
			if err != nil {
				ts.err <- err
			}

			if _, ok := ts.connections[req.Source]; ok {
				// If we are getting a request from someone that
				// we're already tunneling with we should figure
				// out how to handle that
				continue
			}

			ts.connect(req.Source)
		} else if p.Data[1] == 2 {
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
		// else don't do anyting with the packet
		// We don't handle response packets
	}
}

// listenTun reads a packet from the tun device, wraps it in a TCP tunnel packet
// and sends it out the node connection
func (ts *TCPTunServer) listenTun() {
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
		err = ts.node.SendPacket(ep)
		if err != nil {
			ts.err <- err
			return
		}
	}
}
