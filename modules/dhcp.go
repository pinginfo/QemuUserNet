package modules

import (
	"QemuUserNet/entities"
	"QemuUserNet/tools"
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Dhcp represents a DHCP server module.
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

// NewDhcp creates a new Dhcp instance with the provided parameters.
func NewDhcp(subnet string, gateway string, gatewayM string, rangeIp string, dnsIp string, clients *entities.Clients) (*Dhcp, error) {
	// Parse subnet and gateway IP
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return nil, errors.New("Invalid gateway IP")

	}

	// Parse gateway MAC address
	gatewayMAC, err := net.ParseMAC(gatewayM)
	if err != nil {
		return nil, err
	}

	// Generate IP range
	freeIP, err := tools.GenerateIPRange(rangeIp)
	if err != nil {
		return nil, err
	}

	// Parse DNS IP
	dnsIP := net.ParseIP(dnsIp)
	if dnsIP == nil {
		return nil, errors.New("Invalid DNS IP")
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

// Listen processes a DHCP packet, assigns an IP address to the client, and constructs a DHCP response.
func (d *Dhcp) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	etherLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)

	// Check if all required layers are present
	if etherLayer == nil || ipLayer == nil || udpLayer == nil || dhcpLayer == nil {
		return packet.Data(), All, nil, errors.New("Not a dhcp packet")
	}

	// Get an available IP address from the DHCP pool
	clientIP, err := d.getAnIp()

	if err != nil {
		return packet.Data(), All, nil, err
	}

	// Extract Ethernet and DHCP layers
	ether, _ := etherLayer.(*layers.Ethernet)
	dhcp, _ := dhcpLayer.(*layers.DHCPv4)

	// Determine DHCP message type
	var messagetype layers.DHCPMsgType
	for _, option := range dhcp.Options {
		if option.Type == layers.DHCPOptMessageType {
			messageType := layers.DHCPMsgType(option.Data[0])
			switch messageType {
			case layers.DHCPMsgTypeDiscover:
				messagetype = layers.DHCPMsgTypeOffer
			case layers.DHCPMsgTypeRequest:
				messagetype = layers.DHCPMsgTypeAck
			default:
				return packet.Data(), Nobody, nil, errors.New("DHCP message type not supported")
			}
		}
	}

	// Construct DHCP response layers
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

	// Construct DHCP response options
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

	// Serialize the response packet
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err = gopacket.SerializeLayers(buf, opts, responseEther, responseIP, responseUDP, responseDHCP)

	if err != nil {
		return packet.Data(), Nobody, nil, errors.New("Packet serialization error")
	}

	// Retrieve the client based on the DHCP client's MAC address
	client, err := d.clients.GetClientByMac(dhcp.ClientHWAddr.String())
	if err != nil {
		return packet.Data(), Nobody, nil, errors.New("Client not found")
	}

	// Update the client's IP address
	ip := clientIP.String()
	client.VM.Ip = &ip

	return buf.Bytes(), Himself, nil, nil
}

// Quit handles any cleanup operations needed for a client upon disconnection.
// It releases the IP address used by the client back into the DHCP pool.
func (d *Dhcp) Quit(client *entities.Thread) error {
	// Release the used IP address back into the pool of free IP addresses
	for _, ip := range d.usedIP {
		if client.VM.Ip != nil && ip.String() == *client.VM.Ip {
			d.freeIP = append(d.freeIP, ip)
		}
	}
	return nil
}

// returns an available IP address from the DHCP pool.
func (d *Dhcp) getAnIp() (*net.IP, error) {
	var value net.IP
	if len(d.freeIP) > 0 {
		value, d.freeIP = d.freeIP[0], d.freeIP[1:]
		d.usedIP = append(d.usedIP, value)
		return &value, nil
	} else {
		return nil, errors.New("No IP left in the DHCP pool")
	}
}
