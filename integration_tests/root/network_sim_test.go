// These integration tests require root as they need access to network devices.
package root

import (
	integration "github.com/AutoRoute/node/integration_tests"
	"github.com/AutoRoute/node/types"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

type settings struct {
	Connections []ConnectionType
}

type ConnectionType struct {
	Source      string
	Destination string
}

func TestNetwork(t *testing.T) {
	WarnRoot(t)

	// config file
	config := "network_sim_network.json"

	cmd := integration.NewWrappedBinary(GetLoopBack2Path(), "--config="+config)
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	file, err := ioutil.ReadFile(config)
	if err != nil {
		t.Fatal("Error opening the configuration file:", err)
	}

	var s settings
	tap_interfaces := make(map[string]map[string]string)

	err = json.Unmarshal(file, &s)
	if err != nil {
		t.Fatal("Error retrieving JSON from configuration file:", err)
	}
	i := 0

	defer cmd.KillAndPrint(t)

	// various data structures for holding relationships between structures
	// interfaces
	interfaces := make(map[string]*net.Interface)
	// dev list for creating the binaries
	devs := make(map[string][]string)
	// binaries
	bins := make(map[string]integration.AutoRouteBinary)
	// socket names
	sockets := make(map[string]string)
	// generate connection pairs
	// goes through each pair of connections
	for _, v := range s.Connections {
		src := v.Source
		dst := v.Destination

		_, ok_src := tap_interfaces[src]
		_, ok_dst := tap_interfaces[dst][src]
		// source-destination mapping is not in map at all
		if !ok_src && !ok_dst {
			tap_interfaces[src] = make(map[string]string)
		}

		_, ok_src = tap_interfaces[src][dst]

		if !ok_src && !ok_dst {
			i1 := "i" + strconv.Itoa(i) + "-0"
			i++
			i2 := "i" + strconv.Itoa(i) + "-0"
			i++
			tap_interfaces[src][dst] = i1 + ":" + i2
			// now add to the devs map
			devs[src] = append(devs[src], i1)
			devs[dst] = append(devs[dst], i2)
		}
	}
	// populate ifaces map, waits for all devices, and sets the links up
	for _, dsts := range tap_interfaces {
		for _, ifaces := range dsts {
			looptaps := strings.Split(ifaces, ":")

			dev1, err := WaitForDevice(looptaps[0])
			if err != nil {
				t.Fatal(err)
			}
			dev2, err := WaitForDevice(looptaps[1])
			if err != nil {
				t.Fatal(err)
			}
			interfaces[looptaps[0]] = dev1
			interfaces[looptaps[1]] = dev2

			out, err := exec.Command("ip", strings.Split("link set dev "+looptaps[0]+" up", " ")...).CombinedOutput()
			if err != nil {
				t.Fatal(err, string(out))
			}
			out, err = exec.Command("ip", strings.Split("link set dev "+looptaps[1]+" up", " ")...).CombinedOutput()
			if err != nil {
				t.Fatal(err, string(out))
			}

			err = WaitForListen(BuildListenAddress(dev1, integration.GetUnusedPort()))
			if err != nil {
				t.Fatal(err)
			}
			err = WaitForListen(BuildListenAddress(dev2, integration.GetUnusedPort()))
			if err != nil {
				t.Fatal(err)
			}

		}
	}
	i = 0
	for _, names := range devs {
		// "leaf"
		socket := "/tmp/unix" + strconv.Itoa(i)
		if len(names) == 1 {
			name := names[0]
			listen_port := integration.GetUnusedPort()
			listen := integration.NewNodeBinary(integration.BinaryOptions{
				Listen:               BuildListenAddress(interfaces[name], listen_port),
				Fake_money:           true,
				Autodiscover:         true,
				Autodiscover_devices: names,
				Unix:                 socket,
			})
			bins[name] = listen
			sockets[name] = socket
			listen.Start()
			defer listen.KillAndPrint(t)
		} else {
			listen_port := integration.GetUnusedPort()
			connect := integration.NewNodeBinary(integration.BinaryOptions{
				Listen:               fmt.Sprintf("[::]:%d", listen_port),
				Fake_money:           true,
				Autodiscover:         true,
				Autodiscover_devices: names,
				Unix:                 socket,
			})
			for _, name := range names {
				bins[name] = connect
				sockets[name] = socket
			}
			connect.Start()
			defer connect.KillAndPrint(t)
		}
		i++
	}
	// wait for each pair of interfaces to connect
	for _, dsts := range tap_interfaces {
		for _, ifaces := range dsts {
			looptaps := strings.Split(ifaces, ":")
			listen := bins[looptaps[0]]
			connect := bins[looptaps[1]]

			listen_id, err := integration.WaitForID(listen)
			if err != nil {
				t.Fatal(err)
			}

			connect_id, err := integration.WaitForID(connect)
			if err != nil {
				t.Fatal(err)
			}
			err = integration.WaitForConnection(connect, listen_id)
			if err != nil {
				err = integration.WaitForConnection(listen, connect_id)
				if err != nil {
					t.Fatal(err)
				}
			}
			raw_id, err := hex.DecodeString(connect_id)
			p := types.Packet{types.NodeAddress(string(raw_id)), 10, "data"}
			// create unix sockets for both nodes
			c, err := integration.WaitForSocket(sockets[looptaps[0]])
			if err != nil {
				t.Fatal(err)
			}
			c2, err := integration.WaitForSocket(sockets[looptaps[1]])
			if err != nil {
				t.Fatal(err)
			}

			w := json.NewEncoder(c)
			err = w.Encode(p)
			if err != nil {
				t.Fatal(err)
			}
			// verify packet transmissions for each pair of nodes
			err = integration.WaitForPacketsReceived(listen, listen_id)
			if err != nil {
				t.Fatal(err)
			}
			err = integration.WaitForPacketsSent(listen, connect_id)
			if err != nil {
				t.Fatal(err)
			}
			err = integration.WaitForPacketsReceived(connect, listen_id)
			if err != nil {
				t.Fatal(err)
			}
			err = integration.WaitForPacketsSent(connect, connect_id)
			if err != nil {
				t.Fatal(err)
			}
			// send a packet between each pair of nodes
			packets := make(chan types.Packet)
			go integration.WaitForPacket(c2, t, packets)
			select {
			case <-time.After(10 * time.Second):
				t.Fatal("Never received packet")
			case p2 := <-packets:
				if p != p2 {
					t.Fatal("Packets %v != %v", p, p2)
				}
			}

		}
	}

}
