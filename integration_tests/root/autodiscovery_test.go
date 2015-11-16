// These integration tests require root as they need access to network devices.
package root

import (
	"github.com/AutoRoute/node"
	integration "github.com/AutoRoute/node/integration_tests"

	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func BuildListenAddress(i *net.Interface, port int) string {
	ll, err := node.GetLinkLocalAddr(*i)
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

func TestConnection(t *testing.T) {
	// set -e
	WarnRoot(t)
	// loopback2
	cmd := integration.NewWrappedBinary(GetLoopBack2Path())
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.KillAndPrint(t)
	listen_dev, err := WaitForDevice("looptap0-0")
	if err != nil {
		t.Fatal(err)
	}
	connect_dev, err := WaitForDevice("looptap0-1")
	if err != nil {
		t.Fatal(err)
	}
	//  ip link set dev looptap0-0 up
	out, err := exec.Command("ip", strings.Split("link set dev looptap0-0 up", " ")...).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	// ip link set dev looptap 0-1 up
	out, err = exec.Command("ip", strings.Split("link set dev looptap0-1 up", " ")...).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
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
	})
	listen.Start()
	defer listen.KillAndPrint(t)
    // autoroute -fake_money -listen "[%ip1%looptap0-1]:31337" -auto=true -devs='looptap0-1'
	connect_port := integration.GetUnusedPort()
	connect := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:               BuildListenAddress(connect_dev, connect_port),
		Fake_money:           true,
		Autodiscover:         true,
		Autodiscover_devices: []string{connect_dev.Name},
	})
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
