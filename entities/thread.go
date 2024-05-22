package entities

import "errors"

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
