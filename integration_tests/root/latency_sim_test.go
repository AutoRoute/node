// These integration tests require root as they need access to network devices.
package root

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	integration "github.com/AutoRoute/node/integration_tests"
	"github.com/AutoRoute/node/types"
)

// Takes a list of the tap devices on our system and returns the endpoints for
// nodes A and B.
// Args:
//  t: Testing interface.
//  tap_interfaces: The map of tap interfaces returned by SetupNetwork().
//  sockets: The map of sockets returned by SetupNetwork().
//  bins: The map of binaries returned by SetupNetwork().
// Returns:
//  Connections that we can use to read and write data to and from A and B,
//  as well as the raw node addresses of A and B.
func getNetworkEndpoints(t *testing.T,
	tap_interfaces map[string]map[string]string,
	sockets map[string]string,
	bins map[string]integration.AutoRouteBinary) (
	net.Conn, net.Conn, types.NodeAddress, types.NodeAddress) {

	// Try connecting to the A and B nodes in our system.
	a_index := 0
	tap_pair_a_c, ok_a_c := tap_interfaces["A"]["C"]
	if !ok_a_c {
		// Try the other way around if they're not in there.
		a_index = 1
		tap_pair_a_c, ok_a_c = tap_interfaces["C"]["A"]
		if !ok_a_c {
			// Now we have an issue.
			t.Fatal("Unable to find link between nodes A and C!")
		}
	}

	b_index := 0
	tap_pair_b_f, ok_b_f := tap_interfaces["B"]["F"]
	if !ok_b_f {
		b_index = 1
		tap_pair_b_f, ok_b_f = tap_interfaces["F"]["B"]
		if !ok_b_f {
			t.Fatal("Unable to find link between nodes B and F!")
		}
	}

	// Separate the correct interfaces.
	looptaps_a_c := strings.Split(tap_pair_a_c, ":")
	looptaps_b_f := strings.Split(tap_pair_b_f, ":")
	tap_a := looptaps_a_c[a_index]
	tap_b := looptaps_b_f[b_index]

	// Make sure the sockets are up and running.
	connection_a, err := integration.WaitForSocket(sockets[tap_a])
	if err != nil {
		t.Fatalf("Failed to connect to A: %s\n", err)
	}
	connection_b, err := integration.WaitForSocket(sockets[tap_b])
	if err != nil {
		t.Fatalf("Failed to connect to B: %s\n", err)
	}

	// Get node addresses.
	str_address, err := bins[tap_a].GetID()
	if err != nil {
		t.Fatalf("Getting node A address failed: %s\n", err)
	}
	var address_a types.NodeAddress
	err = address_a.UnmarshalText([]byte(str_address))
	if err != nil {
		t.Fatalf("Decoding node A address failed: %s\n", err)
	}
	str_address, err = bins[tap_b].GetID()
	if err != nil {
		t.Fatalf("Getting node B address failed: %s\n", err)
	}
	var address_b types.NodeAddress
	err = address_b.UnmarshalText([]byte(str_address))
	if err != nil {
		t.Fatalf("Decoding node B address failed: %s\n", err)
	}

	return connection_a, connection_b, address_a, address_b
}

// Common function to set things up for all the tests in this file.
// Args:
//  t: Testing interface we are using.
//  tap_interfaces: Interfaces from SetupNetwork().
//  bins: Binaries from SetupNetwork().
//  sockets: Sockets from SetupNetwork().
// Returns:
//  Socket connections to the A and B nodes which constitute the source and
//  destination for all these tests, respectively, as well as the raw node
//  addresses for A and B.
func doAllSetup(t *testing.T, tap_interfaces map[string]map[string]string,
	bins map[string]integration.AutoRouteBinary,
	sockets map[string]string) (
	net.Conn, net.Conn, types.NodeAddress, types.NodeAddress) {

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

			// We're not actually going to bother with explicit checks for sending
			// packets back and forth. We have so many interfaces and enough latency
			// that those sorts of tests take FOREVER. And then everything times out
			// and it's just generally a mess, so it's really better to just skip
			// them.
		}
	}

	// Set up connections to devices A and B.
	return getNetworkEndpoints(t, tap_interfaces, sockets, bins)
}

// Make sure nodes A and B can talk to each other at all.
func TestReadAndWrite(t *testing.T) {
	err := CheckRoot()
	if err != nil {
		t.Skip(err)
	}

	tap_interfaces, bins,
		sockets, taps := SetupNetwork(t, "latency_sim_network.json", true)

	// Clean up everything when we're done.
	defer taps.Stop()
	for _, bin := range bins {
		defer bin.KillAndPrint(t)
	}

	conn_a, conn_b, _, addr_b := doAllSetup(t, tap_interfaces, bins, sockets)

	// Write a simple message and read it out.
	test_message := "Damn, Daniel!"
	integration.SendPacket(conn_a, t, addr_b, test_message)
	packets := make(chan types.Packet)
	go integration.WaitForPacket(conn_b, t, packets)
	packet := <-packets
	t.Logf("Got message: %s\n", packet.Data)
	if packet.Data != test_message {
		t.Fatalf("Expected message '%s', got '%s'.\n", test_message, packet.Data)
	}
}

func makeSpaceString(length int) string {
	str := ""
	for i := 0; i < length; i++ {
		str += " "
	}
	return str
}

// Make sure the routing algorithm makes optimal choices for sending packets.
func TestRoutingAlgorithm(t *testing.T) {
	err := CheckRoot()
	if err != nil {
		t.Skip(err)
	}

	tap_interfaces, bins,
		sockets, taps := SetupNetwork(t, "latency_sim_network.json", false)

	// Clean up everything when we're done.
	defer taps.Stop()
	for _, bin := range bins {
		defer bin.KillAndPrint(t)
	}

	conn_a, conn_b, _, addr_b := doAllSetup(t, tap_interfaces, bins, sockets)

	// Send some data, and it should transmit it through multiple paths.
	packet_number := 0
	packet_padding := 30
	send_packets := 50
	wait_for := send_packets - 5
	go func() {
		for i := 0; i < send_packets; i++ {
			packet := makeSpaceString(packet_padding)
			packet += fmt.Sprintf("%d", packet_number)
			integration.SendPacket(conn_a, t, addr_b, packet)
			packet_number++
		}
	}()

	start_time := time.Now()
	// Wait for all the packet to make it to the other end.
	got_packets := 0
	packet_chan := make(chan types.Packet)
	go integration.WaitForPackets(conn_b, t, packet_chan, wait_for)
	// Make it close to account for any dropped packets.
	for got_packets < wait_for {
		<-packet_chan
		got_packets++
	}
	transmission_time := float64(time.Since(start_time)) / float64(time.Second)
	t.Logf("Packet transmission took %f seconds.\n", transmission_time)

	// Do a sanity check on the amount of time it took.
	overhead := 169
	min_bandwidth := 5000
	packet_size := overhead + packet_padding + 3
	if transmission_time/1.5 >
		float64(packet_size*wait_for)/
			float64(min_bandwidth) {
		t.Fatalf("Packet transmission took too long! (%fs)\n", transmission_time)
	}
}
