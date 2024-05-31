package modules

import (
	"QemuUserNet/entities"
	"QemuUserNet/tools"
	"errors"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Switch struct {
	clients *entities.Clients
}

func NewSwitch(clients *entities.Clients) (*Arp, error) {
	return &Arp{clients: clients}, nil
}

func (s *Switch) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	etherLayer := packet.Layer(layers.LayerTypeEthernet)

	if etherLayer == nil {
		return packet.Data(), Nobody, nil, errors.New("Ether layer not found")
	}

	eth, ok := etherLayer.(*layers.Ethernet)
	if !ok {
		return packet.Data(), Nobody, nil, errors.New("Ether layer not found")
	}
	mac := eth.SrcMAC.String()

	isBroadcast, err := tools.IsBroadcastMAC(mac)
	if err != nil {
		return packet.Data(), Nobody, nil, errors.New("Ether layer not found")
	}

	if isBroadcast {
		return packet.Data(), All, nil, err
	}

	client, err := s.clients.GetClientByMac(mac)

	if err != nil {
		return packet.Data(), Nobody, nil, err
	}

	return packet.Data(), Explicit, client, nil
}

func (s *Switch) Quit(client *entities.Thread) error {
	return nil
}
