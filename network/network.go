// Package network provides network management for virtual machines in QemuUserNet.
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

// Network represents a virtual network.
type Network struct {
	Name                 string
	MTU                  int
	Clients              *entities.Clients
	Modules              []modules.Module
	DisconnectOnPowerOff bool
}

// AddVM adds a new virtual machine to the network.
func (n *Network) AddVM(id string) (*entities.VM, error) {
	// Check if the ID is already used
	if _, err := n.Clients.GetClientByID(id); err == nil {
		return nil, errors.New("This ID is already used")
	}

	// Generate unique socket identifiers
	uuid := uuid.New().String()
	var localSock = "/tmp/QemuUserNet_" + uuid + ".local"
	var remoteSock = "/tmp/QemuUserNet_" + uuid + ".remote"

	// Generate a new MAC address
	mac, err := n.getNewMac()
	if err != nil {
		log.Println("WARNING: error during mac generation: ", err.Error())
	}

	// Create a new VM and its associated thread
	vm := entities.VM{ID: id, Mac: mac, Socket: uuid, LocalSocket: localSock, RemoteSocket: remoteSock, Ip: nil, LocalSock: nil}
	thread := &entities.Thread{VM: vm, Active: false, Done: make(chan struct{})}
	n.Clients.Threads = append(n.Clients.Threads, thread)

	// Start the listener in a new goroutine
	go func() {
		if err := n.listen(thread); err != nil {
			log.Printf("ERROR: failed to start listener for VM %s: %v", id, err)
		}
	}()

	return &vm, nil
}

// RemoveVM removes a virtual machine from the network by its ID.
func (n *Network) RemoveVM(id string) error {
	client, err := n.Clients.GetClientByID(id)
	if err != nil {
		return err
	}
	return n.stopThread(client)
}

// Stop stops all running threads in the network.
func (n *Network) Stop() error {
	var stopErrors []error
	for _, client := range n.Clients.Threads {
		err := n.stopThread(client)
		if err != nil {
			stopErrors = append(stopErrors, err)
		}
	}
	if len(stopErrors) > 0 {
		return fmt.Errorf("WARNONG: failed to stop some threads: %v", stopErrors)
	}
	return nil
}

// listen starts listening for packets on the VM's remote socket.
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

	defer func() {
		recv.Close()
		os.Remove(thread.VM.RemoteSocket)
	}()

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
			case modules.Explicit:
				err = n.send(client, request)
				if err != nil {
					log.Println(err.Error())
				}
			case modules.Himself:
				err = n.send(thread, request)
				if err != nil {
					log.Println(err.Error())
				}
			case modules.All:
				for _, x := range n.Clients.Threads {
					err = n.send(x, request)
					if err != nil {
						log.Println(err.Error())
					}
				}
			default:
				for _, x := range n.Clients.Threads {
					if x.VM != thread.VM {
						err = n.send(x, request)
						if err != nil {
							log.Println(err.Error())
						}
					}
				}
			}
		}
	}
}

// send sends data to the specified client's local socket.
func (n *Network) send(client *entities.Thread, data []byte) error {
	if client.VM.LocalSock == nil {
		sock, err := net.DialUnix("unixgram", nil, &net.UnixAddr{client.VM.LocalSocket, "unixgram"})
		if err != nil {
			return fmt.Errorf("WARNING: error during creation of socket: %s", err.Error())
		}
		client.VM.LocalSock = sock
		log.Println("INFO: Opened LocalSocket for ", client.VM.ID)
	}
	length, err := client.VM.LocalSock.Write(data)
	if err != nil {
		if n.DisconnectOnPowerOff {
			return n.stopThread(client)
		}
		return fmt.Errorf("WARNING: error during writing : %s", err.Error())
	}
	if length != len(data) {
		return errors.New("Package not send completely")
	}

	return nil
}

// stopThread stops the specified client's thread and cleans up resources.
func (n *Network) stopThread(client *entities.Thread) error {
	var err error
	client.Stop()
	for _, module := range n.Modules {
		module.Quit(client)
	}
	*n.Clients, err = n.Clients.RemoveClient(client)
	return err
}

// getNewMac generates a new unique MAC address for a VM.
func (n *Network) getNewMac() (string, error) {
	for {
		mac, err := tools.GenerateMACAddress()
		if err != nil {
			return "", err
		}
		if _, err = n.Clients.GetClientByMac(mac); err != nil {
			return mac, nil
		}
	}
}
