package node

import (
  "fmt"
	"net"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

type TunServer struct {
  data        DataConnection
	dev         tuntap.Interface
	nodes       map[net.IP]types.NodeAddress
	connections map[types.NodeAddress]*TCP
}

func NewTunServer(d DataConnection) *TunServer {
	i, err := tuntap.Open("tun%d", tuntap.DevTun)
	if err != nil {
		log.Fatal(err)
	}
	return &TunServer{d, i, make(map[net.TIP]types.NodeAddress), make(map[net.IP]*TCP)}
}

func (ts *TunServer) Connect(tcptun string, amt int64) {
  dest := ""
  _, err = fmt.Sscanf(*tcptun, "%x", &dest)
  if err != nil {
    log.Fatal(err)
  }
  t := NewTCPTunnel(ts.dev, ts.data, types.NodeAddress(dest), amt)
  connections[types.NodeAddress] = t
}

func (ts *TunServer) Listen() {
  go func() {
    for packet := range data.Packets() {
      Connect(packet.Data, 10000)
    }
  }()
}
