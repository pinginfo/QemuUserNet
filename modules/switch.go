package modules

import (
	"QemuUserNet/entities"
	"QemuUserNet/tools"
	"errors"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Switch represents a network switch that handles packets for a list of clients.
type Switch struct {
	clients *entities.Clients
}

// NewSwitch creates a new Switch instance with the provided clients.
func NewSwitch(clients *entities.Clients) (*Switch, error) {
	return &Switch{clients: clients}, nil
}

// Listen processes a network packet and determines its forwarding action based on the destination MAC address.
// It extracts the Ethernet layer from the packet, checks if the destination MAC address is a broadcast address,
// and retrieves the corresponding client if it is not. The function returns the packet data, a receiver type
// (indicating how to forward the packet), the target client thread (if applicable), and any error encountered.
func (s *Switch) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	// Extract the Ethernet layer from the packet
	etherLayer := packet.Layer(layers.LayerTypeEthernet)
	if etherLayer == nil {
		return packet.Data(), Nobody, nil, errors.New("Ethernet layer is missing from the packet")
	}
	eth, ok := etherLayer.(*layers.Ethernet)
	if !ok {
		return packet.Data(), Nobody, nil, errors.New("Ethernet layer is missing from the packet")
	}

	// Get the destination MAC address from the Ethernet layer
	mac := eth.DstMAC.String()

	// Check if the MAC address is a broadcast address
	isBroadcast, err := tools.IsBroadcastMAC(mac)
	if err != nil {
		return packet.Data(), Nobody, nil, fmt.Errorf("Failed to check if MAC is broadcast: %v", err)
	}

	// If the destination MAC is a broadcast address, return the packet to be sent to all clients
	if isBroadcast {
		return packet.Data(), All, nil, err
	}

	// Find the client associated with the destination MAC address
	client, err := s.clients.GetClientByMac(mac)
	if err != nil {
		return packet.Data(), Nobody, nil, err
	}

	// Return the packet to be sent to the specific client
	return packet.Data(), Explicit, client, nil
}

// Quit handles any necessary cleanup for a client when it disconnects. Currently, it does nothing.
func (s *Switch) Quit(client *entities.Thread) error {
	return nil
}
