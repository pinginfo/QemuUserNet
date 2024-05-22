package network

import (
	"QemuUserNet/entities"
	"QemuUserNet/modules"
	"QemuUserNet/tools"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Thread struct {
	VM     entities.VM
	Active bool
	Done   chan struct{}
}

func (t *Thread) Stop() {
	close(t.Done)
}

type Network struct {
	Name     string
	MTU      int
	Handlers []*Thread
	Modules  []modules.Module
}

func (n *Network) send(sockPath string, data []byte) error {
	sock, err := net.DialUnix("unixgram", nil, &net.UnixAddr{sockPath, "unixgram"})
	if err != nil {
		log.Println("WARNING: error during creation of socket: ", err.Error())
		return err
	}
	defer sock.Close()

	_, err = sock.Write(data)
	return nil
}

func (n *Network) listen(thread *Thread) error {
	log.Println("INFO: Thread started : " + thread.VM.Socket)
	if _, err := os.Stat(thread.VM.RemoteSocket); err == nil {
		os.Remove(thread.VM.RemoteSocket)
	}

	recv, err := net.ListenUnixgram("unixgram", &net.UnixAddr{thread.VM.RemoteSocket, "unixgram"})
	if err != nil {
		log.Println("WARNING: error during creation of socket: ", err.Error())
		return err
	}

	defer recv.Close()

	for {
		select {
		case <-thread.Done:
			log.Println("INFO: Thread stopped: " + thread.VM.Socket)
			return nil
		default:
			data := make([]byte, n.MTU)
			_, _, err := recv.ReadFromUnix(data)

			if err != nil {
				log.Println("WARNING: error during reading: ", err.Error())
			}
			packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

			var receiver modules.Receiver
			var request []byte
			for _, module := range n.Modules {
				request, receiver, err = module.Listen(packet)
				if err != nil {
					continue
				}
				break
			}

			if receiver == modules.Nobody {
				continue
			}
			if receiver == modules.Himself {
				n.send(thread.VM.LocalSocket, request)
			} else {
				for _, x := range n.Handlers {
					if receiver != modules.All && x.VM == thread.VM {
						continue
					}
					n.send(x.VM.LocalSocket, request)
				}
			}
		}
	}
}

func (n *Network) GetVMs() ([]entities.VM, error) {
	var vm = []entities.VM{}
	for _, handler := range n.Handlers {
		vm = append(vm, handler.VM)
	}
	return vm, nil
}

func (n *Network) GetVM(mac string) (*Thread, error) {
	for _, handler := range n.Handlers {
		if handler.VM.Mac == mac {
			return handler, nil
		}
	}
	return nil, errors.New("VM not found")
}

func (n *Network) AddVM(mac string) (*entities.VM, error) {
	uuid, err := tools.GenerateUUID()
	if err != nil {
		log.Println("WARNING: error during uuid generation: ", err.Error())
	}
	var localSock = "/tmp/QemuUserNet_" + uuid + ".local"
	var remoteSock = "/tmp/QemuUserNet_" + uuid + ".remote"

	vm := entities.VM{Mac: mac, Socket: uuid, LocalSocket: localSock, RemoteSocket: remoteSock, Ip: nil}
	thread := &Thread{VM: vm, Active: false, Done: make(chan struct{})}
	n.Handlers = append(n.Handlers, thread)

	go n.listen(thread)

	return &vm, nil
}

func (n *Network) Start(mac string) error {
	handler, err := n.GetVM(mac)
	if err != nil {
		return err
	}
	go n.listen(handler)
	return nil
}

func (n *Network) Stop(mac string) error {
	handler, err := n.GetVM(mac)
	if err != nil {
		return err
	}
	handler.Stop()
	return nil
}

func (n *Network) StartAllThreads() error {
	fmt.Println("StartAllThreads")
	for _, handler := range n.Handlers {
		go n.listen(handler)
	}
	return nil
}

func (n *Network) stopAllThreads() error {
	for _, handler := range n.Handlers {
		handler.Stop()
	}
	return nil
}

func (n *Network) RemoveVM(mac string) error {
	return errors.New("Not Implemented")
}
