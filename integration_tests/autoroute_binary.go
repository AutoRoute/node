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
	Listen               string
	Fake_money           bool
	Connect              []string
	Unix                 string
	Autodiscover         bool
	Autodiscover_devices []string
	BTCHost              string
	BTCUser              string
	BTCPass              string
}

// Transforms a BinaryOptions into a valid AutoRoute command line.
func ProduceCommandLine(b BinaryOptions) []string {
	args := make([]string, 0)
	args = append(args, "--listen="+b.Listen)
	if b.Fake_money {
		args = append(args, "--fake_money")
	}
	if len(b.Connect) > 0 {
		args = append(args, "--connect="+strings.Join(b.Connect, ","))
	}
	if b.Autodiscover {
		args = append(args, "--auto=true")
	}
	if len(b.Autodiscover_devices) > 0 {
		args = append(args, "--devs="+strings.Join(b.Autodiscover_devices, ","))
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
	return args
}

// Produces a AutoRoute Binary which can be run. Must call Start in order to
// make it start running.
func NewNodeBinary(b BinaryOptions) AutoRouteBinary {
	port := GetUnusedPort()
	args := ProduceCommandLine(b)
	args = append(args, "--status=[::1]:"+fmt.Sprint(port))
	binary := NewWrappedBinary(GetAutoRoutePath(), args...)
	return AutoRouteBinary{binary, port}
}

type statusStruct struct {
	Connections map[string]int
	Id          string
}

func (b AutoRouteBinary) fetchStatus() (*statusStruct, error) {
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
	status, err := b.fetchStatus()
	if err != nil {
		return nil, err
	}
	c := make([]string, 0)
	for k, _ := range status.Connections {
		c = append(c, k)
	}
	return c, nil

}

// Returns the hex encoded network ID of the binary.
func (b AutoRouteBinary) GetID() (string, error) {
	status, err := b.fetchStatus()
	if err != nil {
		return "", err
	}
	return status.Id, nil

}
