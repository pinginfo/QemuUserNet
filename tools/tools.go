package tools

import (
	"crypto/rand"
	"fmt"
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
