package modules

import (
	"QemuUserNet/entities"
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Arp struct {
	clients *entities.Clients
}

func NewArp(clients *entities.Clients) (*Arp, error) {
	return &Arp{clients: clients}, nil
}

func (a *Arp) Listen(packet gopacket.Packet) ([]byte, Receiver, error) {
	etherLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	arpLayer := packet.Layer(layers.LayerTypeARP)

	if etherLayer != nil && arpLayer != nil {
		ethl, _ := etherLayer.(*layers.Ethernet)
		srcMAC := ethl.SrcMAC
		arp, _ := arpLayer.(*layers.ARP)
		ip := net.IP(arp.SourceProtAddress)
		a.clients.UpdateIPIFEmpty(srcMAC.String(), ip.String())
		return packet.Data(), All, errors.New("Job done")
	}

	if etherLayer == nil || ipLayer == nil {
		return packet.Data(), All, errors.New("Job done")
	}

	ethl, _ := etherLayer.(*layers.Ethernet)
	srcMAC := ethl.SrcMAC
	ipl, _ := ipLayer.(*layers.IPv4)
	a.clients.UpdateIPIFEmpty(srcMAC.String(), ipl.SrcIP.String())

	return packet.Data(), All, errors.New("Job done")
}

func (a *Arp) Quit(client *entities.Thread) error {
	return nil
}
