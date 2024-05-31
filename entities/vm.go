package entities

import "net"

type VM struct {
	ID           string
	Mac          string
	Socket       string
	RemoteSocket string
	LocalSocket  string
	Ip           *string
	LocalSock    *net.UnixConn
}
