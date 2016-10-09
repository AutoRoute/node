package integration_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Represents a binary execution of the autoroute binary.
type AutoRouteBinary struct {
	WrappedBinary
	port int
}

// Represents various options which can be passed to the binary.
type BinaryOptions struct {
	Listen              string
	FakeMoney           bool
	Connect             []string
	Unix                string
	Autodiscover        bool
	AutodiscoverDevices []string
	Tcptun              string
	Tcptunserve         bool
	Keyfile             string
	BTCHost             string
	BTCUser             string
	BTCPass             string
	RouteLogPath        string
}

// Transforms a BinaryOptions into a valid AutoRoute command line.
func ProduceCommandLine(b BinaryOptions) []string {
	args := make([]string, 0)
	if len(b.Listen) > 0 {
		args = append(args, "--listen="+b.Listen)
	}
	if b.FakeMoney {
		args = append(args, "--fake_money")
	}
	if len(b.Connect) > 0 {
		args = append(args, "--connect="+strings.Join(b.Connect, ","))
	}
	if b.Autodiscover {
		args = append(args, "--auto=true")
	}
	if len(b.AutodiscoverDevices) > 0 {
		args = append(args, "--devs="+strings.Join(b.AutodiscoverDevices, ","))
	}
	if len(b.Tcptun) > 0 {
		args = append(args, "--tcp_tun="+b.Tcptun)
	}
	if b.Tcptunserve {
		args = append(args, "--tcp_tun_serve")
	}
	if len(b.Keyfile) > 0 {
		args = append(args, "--key_file="+b.Keyfile)
	}
	if len(b.BTCHost) > 0 {
		args = append(args, "--btc_host="+b.BTCHost)
	}
	if len(b.BTCUser) > 0 {
		args = append(args, "--btc_user="+b.BTCUser)
	}
	if len(b.BTCPass) > 0 {
		args = append(args, "--btc_pass="+b.BTCPass)
	}
	if len(b.Unix) > 0 {
		args = append(args, "--unix="+b.Unix)
	}
	if len(b.RouteLogPath) > 0 {
		args = append(args, "--route_log_path="+b.RouteLogPath)
	}
	return args
}

// Produces a AutoRoute Binary which can be run. Must call Start in order to
// make it start running.
// Args:
//  b: Options for the binary.
// Returns:
//  The new autoroute binary.
func NewNodeBinary(b BinaryOptions) AutoRouteBinary {
	port := GetUnusedPort()
	args := ProduceCommandLine(b)
	args = append(args, "--status=[::1]:"+fmt.Sprint(port))

	path := GetAutoRoutePath()
	binary := NewWrappedBinary(path, args...)

	return AutoRouteBinary{binary, port}
}

type statusStruct struct {
	Connections      map[string]int
	Packets_sent     map[string]int
	Packets_dropped  int
	Packets_received map[string]int
	Id               string
}

func (b AutoRouteBinary) FetchStatus() (*statusStruct, error) {
	resp, err := http.Get(fmt.Sprintf("http://[::1]:%d/debug/vars", b.port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	status := &statusStruct{}
	err = json.Unmarshal(body, status)
	return status, err
}

// Returns a list of connections
func (b AutoRouteBinary) GetConnections() ([]string, error) {
	status, err := b.FetchStatus()
	if err != nil {
		return nil, err
	}
	c := make([]string, 0)
	for k, _ := range status.Connections {
		c = append(c, k)
	}
	return c, nil

}

// Returns a map of address -> packets received
func (b AutoRouteBinary) GetPacketsReceived() (map[string]int, error) {
	status, err := b.FetchStatus()
	if err != nil {
		return nil, err
	}
	return status.Packets_received, err

}

// Returns a map of address -> packets received
func (b AutoRouteBinary) GetPacketsSent() (map[string]int, error) {
	status, err := b.FetchStatus()
	if err != nil {
		return nil, err
	}
	return status.Packets_sent, err

}

// Returns the hex encoded network ID of the binary.
func (b AutoRouteBinary) GetID() (string, error) {
	status, err := b.FetchStatus()
	if err != nil {
		return "", err
	}
	return status.Id, nil

}
