package entities

import (
	"QemuUserNet/tools"
	"errors"
)

// Thread represents a VM instance, including its active status and a done channel for signaling.
type Thread struct {
	VM     VM            // Virtual Machine instance
	Active bool          // Indicates if the VM is active
	Done   chan struct{} // Channel to signal when the VM is stopped
}

// Stop closes the done channel to signal that the VM is stopped.
func (t *Thread) Stop() {
	close(t.Done)
}

// Clients manages a collection of VM threads.
type Clients struct {
	Threads []*Thread
}

// GetClientByID retrieves a VM thread by its ID.
// Returns the thread and an error if the VM is not found.
func (c Clients) GetClientByID(id string) (*Thread, error) {
	for _, client := range c.Threads {
		if client.VM.ID == id {
			return client, nil
		}
	}
	return nil, errors.New("VM not found")
}

// GetClientByMac retrieves a VM thread by its MAC address.
// Returns the thread and an error if the VM is not found.
func (c Clients) GetClientByMac(mac string) (*Thread, error) {
	for _, client := range c.Threads {
		if client.VM.Mac == mac {
			return client, nil
		}
	}
	return nil, errors.New("VM not found")
}

// GetVMs returns a slice of all VMs managed by Clients.
func (c Clients) GetVMs() ([]VM, error) {
	var vm = []VM{}
	for _, client := range c.Threads {
		vm = append(vm, client.VM)
	}
	return vm, nil
}

// GetClientByLocalSocket retrieves a VM thread by its local socket.
// Returns the thread and an error if the VM is not found.
func (c Clients) GetClientByLocalSocket(localSock string) (*Thread, error) {
	for _, client := range c.Threads {
		if client.VM.LocalSocket == localSock {
			return client, nil
		}
	}
	return nil, errors.New("VM not found")
}

// RemoveClient removes a VM thread from the Clients list and closes its local socket.
// Returns the updated Clients and an error if the VM is not found.
func (c Clients) RemoveClient(client *Thread) (Clients, error) {
	index := -1
	for i, c := range c.Threads {
		if client.VM.ID == c.VM.ID {
			index = i
			break
		}
	}

	if index == -1 {
		return c, errors.New("VM not found")
	}

	if client.VM.LocalSock != nil {
		client.VM.LocalSock.Close()
	}

	return Clients{
		Threads: append(c.Threads[:index], c.Threads[index+1:]...),
	}, nil
}

// UpdateIPIFEmpty updates the IP address of a VM if it is currently empty.
// Returns an error if the IP is invalid or if the VM already has an IP.
func (c Clients) UpdateIPIFEmpty(mac string, ip string) error {
	if !tools.IsUsableIP(ip) {
		return errors.New("Ip is invalid")
	}
	for _, client := range c.Threads {
		if client.VM.Mac == mac {
			if client.VM.Ip == nil {
				client.VM.Ip = &ip
			} else {
				return errors.New("The VM already has an Ip")
			}
		}
	}
	return errors.New("VM not found")
}
