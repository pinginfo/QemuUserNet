package entities

import (
	"QemuUserNet/tools"
	"errors"
)

type Thread struct {
	VM     VM
	Active bool
	Done   chan struct{}
}

func (t *Thread) Stop() {
	close(t.Done)
}

type Clients struct {
	Threads []*Thread
}

func (c Clients) GetClientByID(id string) (*Thread, error) {
	for _, client := range c.Threads {
		if client.VM.ID == id {
			return client, nil
		}
	}
	return nil, errors.New("VM not found")
}

func (c Clients) GetClientByMac(mac string) (*Thread, error) {
	for _, client := range c.Threads {
		if client.VM.Mac == mac {
			return client, nil
		}
	}
	return nil, errors.New("VM not found")
}

func (c Clients) GetVMs() ([]VM, error) {
	var vm = []VM{}
	for _, client := range c.Threads {
		vm = append(vm, client.VM)
	}
	return vm, nil
}

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

	return Clients{
		Threads: append(c.Threads[:index], c.Threads[index+1:]...),
	}, nil
}

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
