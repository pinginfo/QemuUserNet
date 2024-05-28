package modules

import (
	"QemuUserNet/entities"

	"github.com/google/gopacket"
)

type Receiver int64

const (
	Undefined Receiver = iota
	Himself
	Others
	All
	Nobody
)

type Module interface {
	Listen(gopacket.Packet) ([]byte, Receiver, error)
	Quit(client *entities.Thread) error
}
