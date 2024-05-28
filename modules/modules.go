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
	Explicit
)

type Module interface {
	Listen(gopacket.Packet) ([]byte, Receiver, *entities.Thread, error)
	Quit(client *entities.Thread) error
}
