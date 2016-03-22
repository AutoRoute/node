// These integration tests require root as they need access to network devices.
package root

import (
	integration "github.com/AutoRoute/node/integration_tests"
	"github.com/AutoRoute/node/integration_tests/loopback2"
	"github.com/AutoRoute/node/types"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Reads the devices from the config file, starts autoroute binaries for each of
// them, and waits for everything to connect to each-other.
// Args:
//  t: T instance to use for logging.
//  config: The path to the config file to use for setting up the network.
//  race: Whether to use autoroute binaries with race checking enabled.
// Returns:
//  * Map that contains the names of each device mapped with the devices it
//  connects to, each one of which is mapped to the pair of the actual tap
//  interfaces that are being used.
//  * Map that contains the autoroute binaries for each device.
//  * Map that contains the socket for each interface.
//  * Loopback2 TapNetwork.
func SetupNetwork(t *testing.T, config string, race bool) (
	map[string]map[string]string,
	map[string]integration.AutoRouteBinary,
	map[string]string,
	*loopback2.TapNetwork) {
	network := loopback2.NewTapNetwork(config, "-0")

	tap_interfaces, _, err := network.ReadConfigFile()
	if err != nil {
		t.Fatal("Error opening the configuration file.")
	}

	// various data structures for holding relationships between structures
	// interfaces
	interfaces := make(map[string]*net.Interface)
	// dev list for creating the binaries
	devs := make(map[string][]string)
	// binaries
	bins := make(map[string]integration.AutoRouteBinary)
	// socket names
	sockets := make(map[string]string)

	// populate ifaces map, waits for all devices, and sets the links up
	for src, dsts := range tap_interfaces {
		for dst, ifaces := range dsts {
			looptaps := strings.Split(ifaces, ":")

			// Add device names to the devs map for later use.
			devs[src] = append(devs[src], looptaps[0])
			devs[dst] = append(devs[dst], looptaps[1])

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
	i := 0
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
			}, race)
			bins[name] = listen
			sockets[name] = socket
			listen.Start()
		} else {
			listen_port := integration.GetUnusedPort()
			connect := integration.NewNodeBinary(integration.BinaryOptions{
				Listen:               fmt.Sprintf("[::]:%d", listen_port),
				Fake_money:           true,
				Autodiscover:         true,
				Autodiscover_devices: names,
				Unix:                 socket,
			}, race)
			for _, name := range names {
				bins[name] = connect
				sockets[name] = socket
			}
			connect.Start()
		}
		i++
	}

	return tap_interfaces, bins, sockets, network
}

func TestNetwork(t *testing.T) {
	err := CheckRoot()
	if err != nil {
		t.Skip(err)
	}

	tap_interfaces, bins,
		sockets, taps := SetupNetwork(t, "network_sim_network.json", true)

	// Clean up everything when we're done.
	defer taps.Stop()
	for _, bin := range bins {
		defer bin.KillAndPrint(t)
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
			// Send a packet between each pair of nodes.
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
