// These integration tests require root as they need access to network devices.
package root

import (
	integration "github.com/AutoRoute/node/integration_tests"
	"github.com/AutoRoute/node/integration_tests/loopback2"

	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func GetLinkLocalAddr(dev net.Interface) (net.IP, error) {
	dev_addrs, err := dev.Addrs()
	if err != nil {
		return nil, err
	}

	for _, dev_addr := range dev_addrs {
		addr, _, err := net.ParseCIDR(dev_addr.String())
		if err != nil {
			return nil, err
		}

		if addr.IsLinkLocalUnicast() {
			return addr, nil
		}
	}
	return nil, errors.New("Unable to find Link local address")
}

func BuildListenAddress(i *net.Interface, port int) string {
	ll, err := GetLinkLocalAddr(*i)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("[%s%%%s]:%d", ll, i.Name, port)
}

func WaitForDevice(s string) (*net.Interface, error) {
	timeout := time.After(1 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-timeout:
			return nil, errors.New(fmt.Sprintf("Error waiting for %s to be reachable", s))
		default:
		}
		d, err := net.InterfaceByName(s)
		if err == nil {
			return d, nil
		}
	}
	panic("Unreachable")
}

func WaitForListen(s string) error {
	timeout := time.After(2 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-timeout:
			return errors.New(fmt.Sprintf("Error waiting for %s to be reachable", s))
		default:
		}
		_, err := net.Listen("tcp6", s)
		if err == nil {
			return nil
		}
	}
	panic("Unreachable")
}

func SetDevUp(dev string) error {
	cmd := fmt.Sprintf("link set dev %s up", dev)
	_, err := exec.Command("ip", strings.Split(cmd, " ")...).CombinedOutput()
	return err
}

func TestConnection(t *testing.T) {
	// set -e
	err := CheckRoot()
	if err != nil {
		t.Skip(err)
	}
	// loopback2
	network := loopback2.NewTapNetwork("", "")
	defer network.Stop()
	listen_dev, err := WaitForDevice("looptap0-0")
	if err != nil {
		t.Fatal(err)
	}
	connect_dev, err := WaitForDevice("looptap0-1")
	if err != nil {
		t.Fatal(err)
	}

	err = SetDevUp("looptap0-0")
	if err != nil {
		t.Fatal(err)
	}
	err = SetDevUp("looptap0-1")
	if err != nil {
		t.Fatal(err)
	}
	// starts listening on the address
	err = WaitForListen(BuildListenAddress(listen_dev, integration.GetUnusedPort()))
	if err != nil {
		t.Fatal(err)
	}
	err = WaitForListen(BuildListenAddress(connect_dev, integration.GetUnusedPort()))
	if err != nil {
		t.Fatal(err)
	}

	listen_port := integration.GetUnusedPort()
	// autoroute -fake_money -listen "[%ip0%looptap0-0]:31337" -auto=true -devs='looptap0-0'
	listen := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:               BuildListenAddress(listen_dev, listen_port),
		Fake_money:           true,
		Autodiscover:         true,
		Autodiscover_devices: []string{listen_dev.Name},
	}, true)
	listen.Start()
	defer listen.KillAndPrint(t)
	// autoroute -fake_money -listen "[%ip1%looptap0-1]:31337" -auto=true -devs='looptap0-1'
	connect_port := integration.GetUnusedPort()
	connect := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:               BuildListenAddress(connect_dev, connect_port),
		Fake_money:           true,
		Autodiscover:         true,
		Autodiscover_devices: []string{connect_dev.Name},
	}, true)
	connect.Start()
	defer connect.KillAndPrint(t)

	_, err = integration.WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}
	// get connect ID once it's generated
	connect_id, err := integration.WaitForID(connect)
	if err != nil {
		t.Fatal(err)
	}
	// connect it to the listener
	err = integration.WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}
}
