package node

import (
  "fmt"
	"net"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

var request_header string = "hello"
var begin_ip_block net.IP = net.ParseIP("6666::0")
var end_ip_block net.IP = net.PasrseIP("6666::6666")

type TunServer struct {
  data        DataConnection
	dev         tuntap.Interface
	nodes       map[net.IP]types.NodeAddress
	connections map[types.NodeAddress]*TCP
  curr_ip     int
}

func NewTunServer(d DataConnection) *TunServer {
	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		log.Fatal(err)
	}
	return &TunServer{d, i, make(map[net.TIP]types.NodeAddress), make(map[net.IP]*TCP, 0)}
}

func (ts *TunServer) Connect(tcptun string, amt int64) {
  dest := ""
  _, err = fmt.Sscanf(*tcptun, "%x", &dest)
  if err != nil {
    log.Fatal(err)
  }
  t := NewTCPTunnel(ts.dev, ts.data, types.NodeAddress(dest), amt)
  ip := net.ParseIP(net.Sprintf("6666::%d", ts.curr_ip++))
  nodes[ip] = types.NodeAddress(dest)
  connections[types.NodeAddress] = t
}

func (ts *TunServer) Listen() {
  go func() {
    for packet := range data.Packets() {
      if packet.Data == request_header {
        Connect(packet.Data, 10000)
      } else {
        /* do something with packet */
      }
    }
  }()
}
