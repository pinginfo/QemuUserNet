package modules

import (
	"QemuUserNet/entities"
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Arp handles ARP requests and updates the IP addresses of clients if they are not already set.
type Arp struct {
	clients *entities.Clients
}

// NewArp creates a new Arp instance with the provided clients.
func NewArp(clients *entities.Clients) (*Arp, error) {
	return &Arp{clients: clients}, nil
}

// Listen processes a network packet to update the client's IP address if it's an ARP packet or an IPv4 packet.
// It extracts the Ethernet, ARP, and IPv4 layers from the packet, and if an ARP packet is found, it updates the client's IP address.
// If only an IPv4 packet is found, it also updates the client's IP address.
// It only updates the IP address if the client's IP is currently empty.
// Returns an error indicating "Job done" to signify the packet was processed.
func (a *Arp) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	// Extract Ethernet, IPv4, and ARP layers from the packet
	etherLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	arpLayer := packet.Layer(layers.LayerTypeARP)

	// Check if both Ethernet and ARP layers are present
	if etherLayer != nil && arpLayer != nil {
		ethl, ok := etherLayer.(*layers.Ethernet)
		if !ok {
			return nil, Nobody, nil, errors.New("Failed to assert Ethernet layer")
		}
		arp, ok := arpLayer.(*layers.ARP)
		if !ok {
			return nil, Nobody, nil, errors.New("Failed to assert ARP layer")
		}
		srcMAC := ethl.SrcMAC
		ip := net.IP(arp.SourceProtAddress)

		// Update the client's IP address if it is empty
		a.clients.UpdateIPIFEmpty(srcMAC.String(), ip.String())
		return packet.Data(), All, nil, errors.New("Job done")
	}

	// Check if either Ethernet or IPv4 layers are missing
	if etherLayer == nil || ipLayer == nil {
		return packet.Data(), All, nil, errors.New("Required layers not found")
	}

	// Extract source MAC address and source IP address
	ethl, ok := etherLayer.(*layers.Ethernet)
	if !ok {
		return nil, Nobody, nil, errors.New("Failed to assert Ethernet layer")
	}
	srcMAC := ethl.SrcMAC
	ipl, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, Nobody, nil, errors.New("Failed to assert IPv4 layer")
	}

	// Update the client's IP address if it is empty
	a.clients.UpdateIPIFEmpty(srcMAC.String(), ipl.SrcIP.String())
	return packet.Data(), All, nil, errors.New("Job done")
}

// Quit handles any necessary cleanup for a client when it disconnects. Currently, it does nothing.
func (a *Arp) Quit(client *entities.Thread) error {
	return nil
}
