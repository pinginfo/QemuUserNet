package entities

import "net"

// VM represents a virtual machine with network attributes.
type VM struct {
	ID           string        // ID of the VM
	Mac          string        // MAC address of the VM
	Socket       string        // Network socket
	RemoteSocket string        // Remote network socket
	LocalSocket  string        // Local network socket
	Ip           *string       // IP address of the VM
	LocalSock    *net.UnixConn // Local Unix connection socket
}
