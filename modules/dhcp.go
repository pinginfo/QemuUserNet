package modules

import (
	"QemuUserNet/entities"
	"QemuUserNet/tools"
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Dhcp struct {
	gatewayIP  net.IP
	gatewayMAC net.HardwareAddr
	subnetIP   net.IP
	subnetMask net.IPMask
	dnsIP      net.IP
	freeIP     []net.IP
	usedIP     []net.IP
	clients    *entities.Clients
}

func NewDhcp(subnet string, gateway string, gatewayM string, rangeIp string, dnsIp string, clients *entities.Clients) (*Dhcp, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return nil, errors.New("Invalid gateway IP")
	}
	gatewayMAC, err := net.ParseMAC(gatewayM)
	if err != nil {
		return nil, err
	}

	freeIP, err := tools.GenerateIPRange(rangeIp)
	if err != nil {
		return nil, err
	}

	dnsIP := net.ParseIP(dnsIp)
	if dnsIP == nil {
		return nil, errors.New("Invalid dns IP")
	}

	return &Dhcp{
		gatewayIP:  gatewayIP,
		gatewayMAC: gatewayMAC,
		subnetIP:   ipnet.IP,
		subnetMask: ipnet.Mask,
		dnsIP:      dnsIP,
		freeIP:     freeIP,
		usedIP:     []net.IP{},
		clients:    clients,
	}, nil
}

func (d *Dhcp) getAnIp() (*net.IP, error) {
	var value net.IP
	if len(d.freeIP) > 0 {
		value, d.freeIP = d.freeIP[0], d.freeIP[1:]
		d.usedIP = append(d.usedIP, value)
		return &value, nil
	} else {
		return nil, errors.New("No IP left")
	}
}

func (d *Dhcp) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	etherLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)

	if etherLayer == nil || ipLayer == nil || udpLayer == nil || dhcpLayer == nil {
		return packet.Data(), All, nil, errors.New("Not a dhcp packet")
	}
	var messagetype layers.DHCPMsgType

	clientIP, err := d.getAnIp()

	if err != nil {
		return packet.Data(), All, nil, err
	}

	ether, _ := etherLayer.(*layers.Ethernet)
	dhcp, _ := dhcpLayer.(*layers.DHCPv4)

	for _, option := range dhcp.Options {
		if option.Type == layers.DHCPOptMessageType {
			messageType := layers.DHCPMsgType(option.Data[0])
			switch messageType {
			case layers.DHCPMsgTypeDiscover:
				messagetype = layers.DHCPMsgTypeOffer
			case layers.DHCPMsgTypeRequest:
				messagetype = layers.DHCPMsgTypeAck
			default:
				return packet.Data(), Nobody, nil, errors.New("DHCPOptMessageType is not supported")
			}
		}
	}

	responseEther := &layers.Ethernet{
		SrcMAC:       d.gatewayMAC,
		DstMAC:       ether.SrcMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}

	responseIP := &layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    d.gatewayIP,
		DstIP:    net.IPv4bcast,
		Protocol: layers.IPProtocolUDP,
	}

	responseUDP := &layers.UDP{
		SrcPort: layers.UDPPort(67),
		DstPort: layers.UDPPort(68),
	}
	responseUDP.SetNetworkLayerForChecksum(responseIP)

	responseDHCP := &layers.DHCPv4{
		Operation:    layers.DHCPOpReply,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  6,
		Xid:          dhcp.Xid,
		YourClientIP: *clientIP,
		NextServerIP: d.gatewayIP,
		RelayAgentIP: net.IPv4zero,
		ClientHWAddr: dhcp.ClientHWAddr,
	}

	options := []layers.DHCPOption{
		{
			Type:   layers.DHCPOptMessageType,
			Data:   []byte{byte(messagetype)},
			Length: 1,
		},
		{
			Type:   layers.DHCPOptServerID,
			Data:   d.gatewayIP.To4(),
			Length: 4,
		},
		{
			Type:   layers.DHCPOptLeaseTime,
			Data:   []byte{0, 0x98, 0x96, 0x7f},
			Length: 4,
		},
		{
			Type:   layers.DHCPOptRouter,
			Data:   d.gatewayIP.To4(),
			Length: 4,
		},
		{
			Type:   layers.DHCPOptSubnetMask,
			Data:   d.subnetMask,
			Length: 4,
		},
		{
			Type:   layers.DHCPOptDNS,
			Data:   d.dnsIP.To4(),
			Length: 4,
		},
		{
			Type:   layers.DHCPOptEnd,
			Length: 0,
		},
	}

	responseDHCP.Options = options

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err = gopacket.SerializeLayers(buf, opts, responseEther, responseIP, responseUDP, responseDHCP)

	if err != nil {
		return packet.Data(), Nobody, nil, errors.New("Packet serialization error")
	}

	client, err := d.clients.GetClientByMac(dhcp.ClientHWAddr.String())
	if err != nil {
		return packet.Data(), Nobody, nil, errors.New("Client not found")
	}

	ip := clientIP.String()
	client.VM.Ip = &ip

	return buf.Bytes(), Himself, nil, nil
}

func (d *Dhcp) Quit(client *entities.Thread) error {
	var value net.IP
	for _, ip := range d.usedIP {
		if client.VM.Ip != nil && ip.String() == *client.VM.Ip {
			value, d.usedIP = d.usedIP[0], d.usedIP[1:]
			d.freeIP = append(d.freeIP, value)
		}
	}
	return nil
}
