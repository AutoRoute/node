package l2

import (
	"log"
	"time"
)

// This is a wrapper that can take a device and limit the bandwidth that can
// go through it.
type deviceWithLatency struct {
	dev FrameReadWriteCloser

	// Maximum number of outgoing bytes per second.
	send_bandwidth int
	// Maximum number of incoming bytes per second.
	receive_bandwidth int
	// How much data we wrote the last time we called WriteFrame().
	last_write_size int
	// How much data we read the last time we called ReadFrame().
	last_read_size int
}

// Basically initializes latency and then delegates to the tapDevice version of
// NewTapDevice. For bandwidth parameters, 0 means no limitation.
// Args:
//	mac: MAC address of the device.
//  dev: A name for the device.
//	send_bandwidth: Maximum number of outgoing bytes per second.
//  receive_bandwidth: Maximum number of incoming bytes per second.
// Returns:
//	The new tap device, error.
func NewTapDeviceWithLatency(mac, dev string, send_bandwidth,
	receive_bandwidth int) (FrameReadWriteCloser, error) {
	tap, err := NewTapDevice(mac, dev)
	if err != nil {
		return nil, err
	}

	wrapped_tap := deviceWithLatency{
		dev:               tap,
		send_bandwidth:    send_bandwidth,
		receive_bandwidth: receive_bandwidth,
	}
	return &wrapped_tap, nil
}

// Reads a frame from the tap device.
// Returns:
// 	The frame read, error.
func (t *deviceWithLatency) ReadFrame() (EthFrame, error) {
	start_time := time.Now()

	frame, err := t.dev.ReadFrame()
	if err != nil {
		return nil, err
	}

	if t.receive_bandwidth == 0 {
		// No limitation.
		return frame, nil
	}

	// Compute how long it took us.
	end_time := time.Now()
	elapsed := end_time.Sub(start_time)

	// Target latency based on our bandwidth. (This is the same formula OpenVPN
	// uses, apparently.)
	target_latency := float32(t.last_read_size) / float32(t.receive_bandwidth)
	t.last_read_size = len(frame)

	// We want to have the exact latency, so wait the rest of the time.
	to_wait := target_latency*float32(time.Second) - float32(elapsed)
	if to_wait < 0 {
		log.Print("Receiving packet took longer than latency!")
	}
	time.Sleep(time.Duration(to_wait))

	return frame, nil
}

// Writes a frame to the tap device.
// Returns:
//	Error.
func (t *deviceWithLatency) WriteFrame(data EthFrame) error {
	if t.send_bandwidth != 0 {
		// Wait the requisite latency.
		target_latency := float32(t.last_write_size) / float32(t.send_bandwidth)
		t.last_write_size = len(data)
		time.Sleep(time.Duration(target_latency * float32(time.Second)))
	}

	return t.dev.WriteFrame(data)
}

// Closes the underlying device.
func (t *deviceWithLatency) Close() error {
	return t.dev.Close()
}
