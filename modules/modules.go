// Package modules provides interfaces and types for creating and managing network modules
// that process and forward network packets. This package includes definitions for
// packet receivers and the Module interface which standardizes how network packets
// are listened to and processed, as well as how resources are cleaned up when a client
// disconnects.
package modules

import (
	"QemuUserNet/entities"

	"github.com/google/gopacket"
)

// Receiver is an enumeration representing the intended recipient(s) of a packet.
type Receiver int64

// Enumeration values for Receiver, representing different forwarding behaviors.
const (
	Undefined Receiver = iota // Undefined receiver state
	Himself                   // Send the packet back to the sender
	Others                    // Send the packet to all clients except the sender
	All                       // Send the packet to all clients
	Nobody                    // Do not send the packet to anyone
	Explicit                  // Send the packet to a specific client
)

// Module is an interface that defines methods for processing and cleaning up network packets.
type Module interface {
	// Listen processes a network packet and determines how it should be forwarded.
	// Returns the modified packet data, the receiver type, the specific client thread (if applicable),
	// and any error encountered. If the error is nil, it indicates that the packet was successfully processed,
	// and further processing of the packet is unnecessary.
	Listen(gopacket.Packet) ([]byte, Receiver, *entities.Thread, error)

	// Quit handles any necessary cleanup for a client when it disconnects.
	// Returns any error encountered during the cleanup process.
	Quit(*entities.Thread) error
}
