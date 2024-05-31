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
	"github.com/google/uuid"
)

type Network struct {
	Name                 string
	MTU                  int
	Clients              *entities.Clients
	Modules              []modules.Module
	DisconnectOnPowerOff bool
}

func (n *Network) send(sockPath string, data []byte) error {
	sock, err := net.DialUnix("unixgram", nil, &net.UnixAddr{sockPath, "unixgram"})
	if err != nil {
		if n.DisconnectOnPowerOff {
			client, err := n.Clients.GetClientByLocalSocket(sockPath)
			if err != nil {
				log.Println("WARNING: error when deleting the VM: ", err.Error())
			}
			err = n.RemoveVM(client.VM.ID)
			if err != nil {
				log.Println("WARNING: error when deleting the VM: ", err.Error())
			}
			return nil
		}

		log.Println("WARNING: error during creation of socket: ", err.Error())
		return err
	}
	defer sock.Close()

	_, err = sock.Write(data)
	return nil
}

func (n *Network) listen(thread *entities.Thread) error {
	log.Println("INFO: Thread started : " + thread.VM.ID)
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
			var client *entities.Thread
			for _, module := range n.Modules {
				request, receiver, client, err = module.Listen(packet)
				if err != nil {
					continue
				}
				break
			}

			switch receiver {
			case modules.Nobody:
				continue
			case modules.Explicit:
				n.send(client.VM.LocalSocket, request)
				continue
			case modules.Himself:
				n.send(thread.VM.LocalSocket, request)
				continue
			case modules.All:
				for _, x := range n.Clients.Threads {
					n.send(x.VM.LocalSocket, request)
				}
			default:
				for _, x := range n.Clients.Threads {
					if x.VM == thread.VM {
						continue
					}
					n.send(x.VM.LocalSocket, request)
				}
			}
		}
	}
}

func (n *Network) GetVMs() ([]entities.VM, error) {
	return n.Clients.GetVMs()
}

func (n *Network) AddVM(id string) (*entities.VM, error) {
	if _, err := n.Clients.GetClientByID(id); err == nil {
		return nil, errors.New("This ID is already used")
	}

	uuid := uuid.New().String()
	var localSock = "/tmp/QemuUserNet_" + uuid + ".local"
	var remoteSock = "/tmp/QemuUserNet_" + uuid + ".remote"

	mac, err := n.getNewMac()

	if err != nil {
		log.Println("WARNING: error during mac generation: ", err.Error())
	}

	vm := entities.VM{ID: id, Mac: mac, Socket: uuid, LocalSocket: localSock, RemoteSocket: remoteSock, Ip: nil}
	thread := &entities.Thread{VM: vm, Active: false, Done: make(chan struct{})}
	n.Clients.Threads = append(n.Clients.Threads, thread)

	go n.listen(thread)

	return &vm, nil
}

func (n *Network) Start(id string) error {
	handler, err := n.Clients.GetClientByID(id)
	if err != nil {
		return err
	}
	go n.listen(handler)
	return nil
}

func (n *Network) Stop(id string) error {
	handler, err := n.Clients.GetClientByID(id)
	if err != nil {
		return err
	}
	handler.Stop()
	for _, module := range n.Modules {
		module.Quit(handler)
	}
	return nil
}

func (n *Network) StartAllThreads() error {
	fmt.Println("StartAllThreads")
	for _, handler := range n.Clients.Threads {
		go n.listen(handler)
	}
	return nil
}

func (n *Network) stopAllThreads() error {
	for _, handler := range n.Clients.Threads {
		handler.Stop()
	}
	return nil
}

func (n *Network) RemoveVM(id string) error {
	handler, err := n.Clients.GetClientByID(id)
	if err != nil {
		return err
	}
	handler.Stop()
	for _, module := range n.Modules {
		module.Quit(handler)
	}
	*n.Clients, err = n.Clients.RemoveClient(handler)
	return err
}

func (n *Network) getNewMac() (string, error) {
	var (
		mac string
		err error
	)
	for {
		mac, err = tools.GenerateMACAddress()

		if err != nil {
			return "", err
		}

		if _, err = n.Clients.GetClientByMac(mac); err != nil {
			break
		}
	}
	return mac, nil
}
