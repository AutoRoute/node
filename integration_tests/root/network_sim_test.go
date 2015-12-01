// These integration tests require root as they need access to network devices.
package root

import (
	integration "github.com/AutoRoute/node/integration_tests"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"testing"
	"path/filepath"
	"encoding/json"
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
	config := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "AutoRoute", "loopback2", "examples", "sample.json")
	
	cmd := integration.NewWrappedBinary(GetLoopBack2Path(), "--config=" + config)
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
	interfaces := make(map[string] *net.Interface)
	// dev list for creating the binaries
	devs := make(map[string][]string)
	// binaries
	bins := make(map[string]integration.AutoRouteBinary)
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
		
		if !ok_src && !ok_dst{
			i1 := "i" + strconv.Itoa(i) + "-0"
			i++
			i2 := "i" + strconv.Itoa(i) + "-0"
			i++
			tap_interfaces[src][dst] = i1 + ":" + i2
			// now add to the devs map
			devs[src] = append(devs[src],i1)
			devs[dst] = append(devs[dst],i2)
		}
	}
	// populate ifaces map, waits for all devices, and sets the links up
	for _, dsts := range tap_interfaces {
		for _, ifaces := range dsts {
			looptaps := strings.Split(ifaces,":")

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

			out, err := exec.Command("ip", strings.Split("link set dev " + looptaps[0] + " up", " ")...).CombinedOutput()
			if err != nil {
				t.Fatal(err, string(out))
			}
			out, err = exec.Command("ip", strings.Split("link set dev " + looptaps[1] + " up", " ")...).CombinedOutput()
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
	for _, names := range devs {
		// "leaf"
		if len(names) == 1 {
			name := names[0]
			listen_port := integration.GetUnusedPort()
			listen := integration.NewNodeBinary(integration.BinaryOptions{
				Listen:               BuildListenAddress(interfaces[name], listen_port),
				Fake_money:           true,
				Autodiscover:         true,
				Autodiscover_devices: names,
			})
			bins[name] = listen;
			listen.Start()
			defer listen.KillAndPrint(t)
		} else {
			connect := integration.NewNodeBinary(integration.BinaryOptions{
				Fake_money:           true,
				Autodiscover:         true,
				Autodiscover_devices: names,
			})
			for _, name := range names {
				bins[name] = connect;
			}
			connect.Start()
			defer connect.KillAndPrint(t)
		}

	}
	// wait for each pair of interfaces to connect
	for _, dsts := range tap_interfaces {
		for _, ifaces := range dsts {
			looptaps := strings.Split(ifaces,":")
			listen := bins[looptaps[0]]
			connect := bins[looptaps[1]]
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
	}

	
}
