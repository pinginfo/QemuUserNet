package modules

import (
	"QemuUserNet/entities"
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Dns struct handles DNS resolution and ARP responses specifically for the DNS server's IP.
type Dns struct {
	ip      net.IP
	mac     net.HardwareAddr
	dict    map[string]net.IP
	clients *entities.Clients
}

// NewDns creates a new Dns instance with the given DNS IP, MAC addresses and the provided clients.
func NewDns(ip string, mac string, clients *entities.Clients) (*Dns, error) {
	nip := net.ParseIP(ip)
	if nip == nil {
		return nil, errors.New("Invalid IP")
	}
	nmac, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}
	return &Dns{nip, nmac, make(map[string]net.IP), clients}, nil
}

// Listen processes incoming packets and responds to ARP and DNS requests.
func (d *Dns) Listen(packet gopacket.Packet) ([]byte, Receiver, *entities.Thread, error) {
	// Check if the packet is an ARP request directed to the DNS server
	t, r, err := d.respondToArpRequest(packet)
	if err == nil {
		return t, r, nil, err
	}

	// If not an ARP request, check if it is a DNS request
	t, r, err = d.respondToDnsRequest(packet)
	return t, r, nil, err
}

// respondToDnsRequest handles DNS requests and builds appropriate responses.
func (d *Dns) respondToDnsRequest(packet gopacket.Packet) ([]byte, Receiver, error) {
	dnsLayer := packet.Layer(layers.LayerTypeDNS)

	if dnsLayer == nil {
		return packet.Data(), All, errors.New("Not a dns packet")
	}

	dnsPacket, _ := dnsLayer.(*layers.DNS)

	// This packet is a response, not a query so we can skip it
	if dnsPacket.QR {
		return packet.Data(), Nobody, errors.New("This is a dns response")
	}

	// Build DNS response
	responseDNS := &layers.DNS{
		ID:           dnsPacket.ID,
		QR:           true,
		OpCode:       dnsPacket.OpCode,
		AA:           true,
		RA:           true,
		Questions:    dnsPacket.Questions,
		ResponseCode: layers.DNSResponseCodeNoErr,
	}

	// Build DNS answers for each question
	for _, question := range dnsPacket.Questions {
		answer := d.buildDNSAnswer(question)
		if answer != nil {
			responseDNS.Answers = append(responseDNS.Answers, *answer)
		}
	}

	// If there are answers, construct the full DNS response packet
	if len(responseDNS.Answers) > 0 {
		responseEther := &layers.Ethernet{
			SrcMAC:       packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet).DstMAC,
			DstMAC:       packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet).SrcMAC,
			EthernetType: layers.EthernetTypeIPv4,
		}

		responseIP := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).DstIP,
			DstIP:    packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).SrcIP,
			Protocol: layers.IPProtocolUDP,
		}

		responseUDP := &layers.UDP{
			SrcPort: layers.UDPPort(53),
			DstPort: packet.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort,
		}
		responseUDP.SetNetworkLayerForChecksum(responseIP)

		// Serialize layers into a single packet
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
		err := gopacket.SerializeLayers(buf, opts, responseEther, responseIP, responseUDP, responseDNS)

		if err != nil {
			return packet.Data(), Nobody, errors.New("Packet serialization error")
		}
		return buf.Bytes(), Himself, nil
	}

	return packet.Data(), Nobody, errors.New("No answers")
}

// respondToArpRequest handles ARP requests specifically for the DNS server's IP and builds appropriate responses.
func (d *Dns) respondToArpRequest(packet gopacket.Packet) ([]byte, Receiver, error) {
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer == nil {
		return nil, All, errors.New("Not a arp request")
	}

	arp, _ := arpLayer.(*layers.ARP)
	if arp.Operation != layers.ARPRequest {
		return packet.Data(), All, errors.New("Not a arp request")
	}

	// Check if the ARP request is for the DNS server's IP
	if net.IP(arp.DstProtAddress).Equal(d.ip) {
		// Build ARP reply
		responseARP := &layers.ARP{
			AddrType:          layers.LinkTypeEthernet,
			Protocol:          layers.EthernetTypeIPv4,
			HwAddressSize:     6,
			ProtAddressSize:   4,
			Operation:         layers.ARPReply,
			SourceHwAddress:   d.mac,
			SourceProtAddress: arp.DstProtAddress,
			DstHwAddress:      arp.SourceHwAddress,
			DstProtAddress:    arp.SourceProtAddress,
		}

		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		eth, _ := ethLayer.(*layers.Ethernet)
		responseEthernet := &layers.Ethernet{
			SrcMAC:       d.mac,
			DstMAC:       eth.SrcMAC,
			EthernetType: layers.EthernetTypeARP,
		}

		// Serialize layers into a single packet
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
		err := gopacket.SerializeLayers(buf, opts, responseEthernet, responseARP)

		if err != nil {
			return packet.Data(), Nobody, errors.New("Packet serialization error")
		}

		return buf.Bytes(), Himself, nil
	}

	return packet.Data(), All, errors.New("ARP request not for dns")
}

// buildDNSAnswer constructs a DNS answer for a given DNS question.
func (d *Dns) buildDNSAnswer(question layers.DNSQuestion) *layers.DNSResourceRecord {
	// Retrieve client information based on the question name
	client, err := d.clients.GetClientByID(string(question.Name))
	if err != nil {
		return nil
	}
	if client.VM.Ip == nil {
		return nil
	}

	ip := net.ParseIP(*client.VM.Ip)
	if ip == nil {
		return nil
	}

	// Create DNS answer based on question type
	answer := &layers.DNSResourceRecord{
		Name:  question.Name,
		Type:  question.Type,
		Class: question.Class,
		TTL:   300,
	}

	switch question.Type {
	case layers.DNSTypeA:
		answer.IP = ip
	case layers.DNSTypeAAAA:
		// TODO Send a dummy ipv6, otherwise the client system will wait for it and slow the system down.
		answer.IP = net.IP{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	default:
		return nil
	}

	return answer
}

// Quit handles any necessary cleanup for a client when it disconnects. Currently, it does nothing.
func (d *Dns) Quit(client *entities.Thread) error {
	return nil
}
