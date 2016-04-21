package node

import (
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
	data   DataConnection
	tun    TCPTun
	nodes  map[string]types.NodeAddress
	currIP int
}

func NewTunServer(d DataConnection) *TunServer {
	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		log.Fatal(err)
	}
	return &TunServer{d, i, make(map[string]types.NodeAddress), 0}
}

func (ts *TunServer) connect(connectingNode string) {
	ip := fmt.Sprintf("6666::%d", ts.currIP)
	ts.currIP++
	ts.nodes[ip] = types.NodeAddress(connectingNode)
}

func (ts *TunServer) Listen() {
	go func() {
		for packet := range ts.data.Packets() {
			if packet.Data == requestHeader {
				ts.connect(packet.Data)
			} else {
				/* do something with packet */
			}
		}
	}()
}
