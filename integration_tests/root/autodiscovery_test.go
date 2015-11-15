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

func WaitForDevice(s string) error {
	timeout := time.After(1 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-timeout:
			return errors.New(fmt.Sprintf("Error waiting for %s to be reachable", s))
		default:
		}
		_, err := net.InterfaceByName("looptap0-0")
		if err == nil {
			return nil
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
	WarnRoot(t)

	cmd := integration.NewWrappedBinary(GetLoopBack2Path())
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.KillAndPrint(t)

	err = WaitForDevice("looptap0-0")
	if err != nil {
		t.Fatal(err)
	}
	err = WaitForDevice("looptap0-1")
	if err != nil {
		t.Fatal(err)
	}

	listen_dev, err := net.InterfaceByName("looptap0-0")
	if err != nil {
		t.Fatal(err)
	}
	connect_dev, err := net.InterfaceByName("looptap0-1")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("ip", strings.Split("link set dev looptap0-0 up", " ")...).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	out, err = exec.Command("ip", strings.Split("link set dev looptap0-1 up", " ")...).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}

	err = WaitForListen(BuildListenAddress(listen_dev, integration.GetUnusedPort()))
	if err != nil {
		t.Fatal(err)
	}
	err = WaitForListen(BuildListenAddress(connect_dev, integration.GetUnusedPort()))
	if err != nil {
		t.Fatal(err)
	}

	listen_port := integration.GetUnusedPort()
	listen := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:               BuildListenAddress(listen_dev, listen_port),
		Fake_money:           true,
		Autodiscover:         true,
		Autodiscover_devices: []string{listen_dev.Name},
	})
	listen.Start()
	defer listen.KillAndPrint(t)

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

	connect_id, err := integration.WaitForID(connect)
	if err != nil {
		t.Fatal(err)
	}

	err = integration.WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}
}
