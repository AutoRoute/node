package root

import (
	"github.com/AutoRoute/node"
	integration "github.com/AutoRoute/node/integration_tests"

	"fmt"
	"os/exec"
	"testing"
	"time"
)

func SetDevAddr(dev string, addr string) error {
	_, err := exec.Command("ip", "addr", "add", addr, "dev", dev).CombinedOutput()
	return err
}

func TestPing(t *testing.T) {
	err := CheckRoot()
	if err != nil {
		t.Skip(err)
	}

	_, key0, err := node.CreateKey("/tmp/keyfile0")
	if err != nil {
		t.Fatal(err)
	}

	node_addr0 := fmt.Sprintf("%x", key0.PublicKey().Hash())

	// tun0
	ponger := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:      "localhost:9999",
		Fake_money:  true,
		Tcptunserve: true,
		Keyfile:     "/tmp/keyfile0",
	})
	ponger.Start()
	defer ponger.KillAndPrint(t)
	_, err = WaitForDevice("tun0")
	if err != nil {
		t.Fatal(err)
	}
	err = SetDevUp("tun0")
	if err != nil {
		t.Fatal(err)
	}
	err = SetDevAddr("tun0", "fe80::1/64")
	if err != nil {
		t.Fatal(err)
	}

	// tun1
	pinger := integration.NewNodeBinary(integration.BinaryOptions{
		Listen:     "localhost:9998",
		Fake_money: true,
		Connect:    []string{"localhost:9999"},
		Tcptun:     node_addr0,
		Keyfile:    "/tmp/keyfile1",
	})
	pinger.Start()
	defer pinger.KillAndPrint(t)
	_, err = WaitForDevice("tun1")
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	for {
		cmd := exec.Command("ping", "-6", "-I", "tun1", "-c", "1", "fe80::1")
		buf, err := cmd.CombinedOutput()
		//fmt.Println(string(buf))
		if err == nil {
			break
		}
		select {
		case <-timeout:
			t.Fatal(err, string(buf))
		default:
		}
	}
}
