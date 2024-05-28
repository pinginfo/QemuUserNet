package tools

import (
	"crypto/rand"
	"fmt"
	"net"
)

func GenerateMACAddress() (string, error) {
	mac := make([]byte, 6)

	if _, err := rand.Read(mac); err != nil {
		return "", err
	}

	mac[0] = (mac[0] | 2) & 0xfe

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
}

func CraftQemuNetworkCommand(socket string, socketRemote string, socketLocal string, mac string) []byte {
	return []byte("-netdev dgram,id=" +
		socket +
		",remote.type=unix,remote.path=" +
		socketRemote +
		",local.type=unix,local.path=" +
		socketLocal +
		" -device virtio-net,netdev=" +
		socket +
		",mac=" +
		mac)
}

func IsUsableIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if ip.IsLoopback() {
		return false
	}

	if ip.IsMulticast() {
		return false
	}

	if ip.IsLinkLocalUnicast() {
		return false
	}

	if ip.IsLinkLocalMulticast() {
		return false
	}

	if ip.IsUnspecified() {
		return false
	}

	if ip.To4() != nil {
		ip = ip.To4()

		if ip.Equal(net.IPv4bcast) {
			return false
		}

		if ip[0] == 169 && ip[1] == 254 {
			return false
		}
	}

	return true
}
