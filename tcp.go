package node

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	//"strings"
	"time"

	"github.com/AutoRoute/tuntap"

	"github.com/AutoRoute/node/types"
)

// NodeConnection is an AutoRoute connection between two AutoRoute nodes.
// It implements a DataConnection (can send and receive AutoRoute packets)
// but also has a NodeAddress.
type NodeConnection interface {
	DataConnection
	GetNodeAddress() types.NodeAddress
}

type TCPTunClient struct {
	node NodeConnection
	tun  TCPTun
	dest types.NodeAddress
	amt  int64
	quit chan bool
	err  chan error
}

type TCPTun interface {
	ReadPacket() (*tuntap.Packet, error)
	WritePacket(p *tuntap.Packet) error
}

var truncated_error error = errors.New("truncated packet")

// SetDevAddr takes a network interface name and an IP address in string form
// and sets the device up and assigns it an IP address.
func SetDevAddr(dev string, addr string) error {
	_, err := exec.Command("ip", "link", "set", "dev", dev, "up").CombinedOutput()
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	_, err = exec.Command("ip", "addr", "add", addr, "dev", dev).CombinedOutput()
	return err
}

// NewTCPTunnel creates and starts a TCP tunneling client. After creating a TCP object, it starts
// the handshake to the server (exit node) specified by dest and returns a pointer to the TCP
// object. Then it starts listening on the tun and node connection.
func NewTCPTunClient(n NodeConnection, tun TCPTun, dest types.NodeAddress, amt int64, tun_name string) *TCPTunClient {
	t := &TCPTunClient{n, tun, dest, amt, make(chan bool), make(chan error, 1)}
	go t.handshake(tun_name)
	return t
}

func (t *TCPTunClient) Close() {
	close(t.quit)
}

func (t *TCPTunClient) Error() chan error {
	return t.err
}

// Sends a request to tunnel with the server. Receives it's assigned IP address and
// sets the tun device up with that address. That begins listening on the tun and
// node connection.
func (t *TCPTunClient) handshake(tun_name string) {
	req := types.TCPTunnelRequest{t.node.GetNodeAddress()}
	req_b, _ := req.MarshalBinary()
	ep := types.Packet{t.dest, t.amt, req_b}
	go func() {
		err := t.node.SendPacket(ep)
		if err != nil {
			t.err <- err
		}
	}()

	var resp types.TCPTunnelResponse
	p := <-t.node.Packets()
	err := resp.UnmarshalBinary(p.Data)
	if err != nil {
		log.Fatal(err)
	}

	//tun_name = strings.Trim(tun_name, "\x00")
	err = SetDevAddr(tun_name, resp.IP.String())
	if err != nil {
		log.Fatal(err)
	}
	go t.readtun()
	go t.writetun()
}

// Reads from the tun device, wraps the node in a TCP tunneling packet and
// sends it out the node connection.
func (t *TCPTunClient) readtun() {
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

		tcp_data := types.TCPTunnelData{b}
		tcp_data_b, _ := tcp_data.MarshalBinary()
		ep := types.Packet{t.dest, t.amt, tcp_data_b}
		err = t.node.SendPacket(ep)
		if err != nil {
			t.err <- err
			return
		}
	}
}

// Reads from the node connection, unwraps the node from the TCP tunneling packet
// and sends it out the tun device.
func (t *TCPTunClient) writetun() {
	for {
		select {
		case p := <-t.node.Packets():
			var tcp_data types.TCPTunnelData
			err := tcp_data.UnmarshalBinary(p.Data)
			if err != nil {
				t.err <- err
				return
			}

			ep := &tuntap.Packet{}
			err = json.Unmarshal(tcp_data.Data, ep)
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
