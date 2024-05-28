package tools

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
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

func GenerateIPRange(rangeStr string) ([]net.IP, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, errors.New("Invalid format")
	}

	startIP := parts[0]
	endSuffix := parts[1]

	start := net.ParseIP(startIP)
	if start == nil {
		return nil, errors.New("Invalid format")
	}

	startParts := strings.Split(startIP, ".")
	if len(startParts) != 4 {
		return nil, errors.New("Invalid format")
	}

	baseIP := strings.Join(startParts[:3], ".") + "."
	startNum, err := strconv.Atoi(startParts[3])
	if err != nil {
		return nil, errors.New("Invalid format")
	}

	endNum, err := strconv.Atoi(endSuffix)
	if err != nil {
		return nil, errors.New("Invalid format")
	}

	if startNum > endNum {
		return nil, fmt.Errorf("start IP address must be less than or equal to end IP address")
	}

	var ips []net.IP
	for i := startNum; i <= endNum; i++ {
		ips = append(ips, net.ParseIP(fmt.Sprintf("%s%d", baseIP, i)))
	}

	return ips, nil
}
