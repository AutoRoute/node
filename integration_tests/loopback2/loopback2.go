package loopback2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AutoRoute/l2"
)

type settings struct {
	Connections []ConnectionType
	Devices     []DeviceType
}

type ConnectionType struct {
	Source      string
	Destination string
	Bandwidth   string
}

type DeviceType struct {
	Name      string
	Bandwidth string
}

// Creates a new tap device with the appropriate parameters.
// Args:
//	name: Name of the device.
//	bandwidth: The max bandwidth for the device.
// Returns:
//	The new tap device.
func makeTap(name string, bandwidth int) l2.FrameReadWriteCloser {
	if bandwidth == 0 {
		// For some reason, some of the integration tests don't like it when we set
		// our own address.
		tap, err := l2.NewTapDevice("", name)
		if err != nil {
			log.Fatal("Error opening tap device:", err)
		}
		return tap
	}

	log.Printf("Setting bandwidth on %s to %d.\n", name, bandwidth)
	tap, err := l2.NewTapDevice("", name)
	if err != nil {
		log.Fatal("Error opening tap device with bandwidth:", err)
	}
	return l2.NewDeviceWithLatency(tap, bandwidth,
		bandwidth)
}

// Waits for a set of devices to go offline.
// Args:
//  devices: The devices that we want to wait for.
func waitForOffline(devices map[string]l2.FrameReadWriteCloser) {
	// Spin until the devices actually appear. We're not really in a hurry here,
	// so we can rate-limit to reduce CPU usage.
	timeout := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		interfaces, err := net.Interfaces()
		if err != nil {
			log.Printf("Warning: Checking interfaces failed: %s\n", err)
			// Skip this check.
			break
		}

		found := false
		for _, iface := range interfaces {
			_, name_found := devices[iface.Name]
			if name_found {
				found = true
			}
		}
		if !found {
			break
		}

		// Check for timeout.
		select {
		case <-timeout:
			log.Fatal("Waiting for devices to go offline took too long.\n")
		default:
			continue
		}
	}
}

// A fake network made out of tap devices.
type TapNetwork struct {
	// Suffix to append to the names of tap devices. This is in case we wish to
	// use multiple networks and want to make sure they don't conflict.
	suffix string
	// Channel we use to notify it when we want to quit.
	quitChannel chan bool
	// Channel we use to notify the main goroutine when we are done initializing.
	readyChannel chan bool
	// This is so we can wait for the internal goroutine to finish cleaning
	// everything up when we want to exit.
	waiter sync.WaitGroup
	// Path to the configuration file we're using.
	config string
}

// Builds a map of all the connections between tap devices based on an input
// configuration file.
// Returns:
//  * A map that maps a device name to submap which itself maps the name of the
//  device paired with that first device to a string containing the actual names
//  of the tap interfaces that will be used separated by a colon.
//  * A map that maps each device to the bandwidth for that device.
//  * An error if appropriate.
func (t *TapNetwork) ReadConfigFile() (map[string]map[string]string,
	map[string]int, error) {
	file, err := ioutil.ReadFile(t.config)
	if err != nil {
		log.Print("Error opening the configuration file:", err)
		return nil, nil, err
	}

	var s settings
	err = json.Unmarshal(file, &s)
	if err != nil {
		log.Print("Error retrieving JSON from configuration file:", err)
		return nil, nil, err
	}

	// various data structures for holding relationships between structures
	// interfaces
	tap_interfaces := make(map[string]map[string]string)
	// Contains bandwidths for each device.
	tap_bandwidths := make(map[string]int)

	// generate connection pairs
	// goes through each pair of connections
	i := 0
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
			i1 := "i" + strconv.Itoa(i) + t.suffix
			i++
			i2 := "i" + strconv.Itoa(i) + t.suffix
			i++
			tap_interfaces[src][dst] = i1 + ":" + i2
		}
	}

	// Set bandwidths.
	for _, device := range s.Devices {
		name := device.Name
		bandwidth, err := strconv.Atoi(device.Bandwidth)
		if err != nil {
			log.Printf("Invalid bandwidth parameter '%s'.\n", device.Bandwidth)
			return nil, nil, err
		}

		_, ok_name := tap_interfaces[name]
		if !ok_name {
			log.Printf("Unknown device '%s'.\n", name)
			return nil, nil, err
		}

		tap_bandwidths[name] = bandwidth
	}

	return tap_interfaces, tap_bandwidths, nil
}

// Actually creates the network.
func (t *TapNetwork) doCreateNetwork() {
	log.Printf("Loopback2: Starting...\n")

	devices := make(map[string]l2.FrameReadWriteCloser)
	if len(t.config) > 0 {
		tap_interfaces, tap_bandwidths, err := t.ReadConfigFile()
		if err != nil {
			log.Fatal("Error reading the configuration file.")
		}

		// Actually make all the tap devices.
		var device_number int64 = 0
		for src, dsts := range tap_interfaces {
			for _, ifaces := range dsts {
				looptaps := strings.Split(ifaces, ":")
				name1 := fmt.Sprintf(looptaps[0])
				device_number += 1
				name2 := fmt.Sprintf(looptaps[1])
				device_number += 1

				bandwidth := tap_bandwidths[src]

				log.Print(fmt.Sprintf("Creating device pair: %s:%s\n", name1, name2))
				lo0 := makeTap(name1, bandwidth)
				lo1 := makeTap(name2, bandwidth)
				devices[name1] = lo0
				devices[name2] = lo1

				go l2.SendFrames(lo0, lo1)
				go l2.SendFrames(lo1, lo0)
			}
		}

	}

	// We're done initializing.
	t.readyChannel <- true
	log.Print("Loopback2: Ready.\n")

	// Wait for someone to tell us to quit.
	<-t.quitChannel
	// Close devices.
	for _, device := range devices {
		device.Close()
	}
	waitForOffline(devices)

	log.Print("Loopback2: Exiting...\n")
	// Tell the caller we're done quitting.
	t.waiter.Done()
}

// Stops a running network simulation.
func (t *TapNetwork) Stop() {
	t.quitChannel <- true
	// Wait for it to clean up.
	log.Printf("Waiting for cleanup...\n")
	t.waiter.Wait()
}

// Creates a network of tap devices that can be used for testing autoroute code.
// Args:
//  config: The location of the config file to use when creating the network. An
//  empty string means use the default network configuration.
//  suffix: Suffix to add to devices created for this network.
// Returns:
//  The TapNetwork object.
func NewTapNetwork(config, suffix string) *TapNetwork {
	network := TapNetwork{
		suffix,
		make(chan bool, 1),
		make(chan bool, 1),
		sync.WaitGroup{},
		config,
	}

	go network.doCreateNetwork()

	network.waiter.Add(1)
	<-network.readyChannel
	return &network
}
