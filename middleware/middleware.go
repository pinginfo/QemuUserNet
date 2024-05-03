package middleware

import (
	"QemuUserNet/entities"
	"QemuUserNet/network"
	"errors"
	"fmt"
	"strings"
)

type Middleware struct {
	networks []*network.Network
}

func (s *Middleware) Init() error {
	s.networks = []*network.Network{}

	return nil
}

func (s *Middleware) getNetwork(nameNetwork string) (*network.Network, error) {
	for _, net := range s.networks {
		if nameNetwork == net.Name {
			return net, nil
		}
	}
	return nil, errors.New("Network not found")
}

func (s *Middleware) Create(cmd entities.CreateCommand) ([]byte, error) {
	_, err := s.getNetwork(cmd.NetworkName)
	if err == nil {
		r := []string{"This name is already in use"}
		return []byte(strings.Join(r, "\n")), nil
	}
	s.networks = append(
		s.networks,
		&network.Network{
			Name: cmd.NetworkName,
			MTU:  1500 + 14})
	r := []string{cmd.NetworkName}
	return []byte(strings.Join(r, "\n")), nil
}

func (s *Middleware) Connect(cmd entities.ConnectCommand) ([]byte, error) {
	net, err := s.getNetwork(cmd.NetworkName)
	if err != nil {
		return []byte(err.Error()), nil
	}
	_, err = net.GetVM(cmd.MacAddress)
	if err == nil {
		return []byte("This mac address is already in use"), nil
	}
	vm, err := net.AddVM(cmd.MacAddress)
	if err != nil {
		return []byte(err.Error()), nil
	}
	return []byte("-netdev dgram,id=" + vm.Socket + ",remote.type=unix,remote.path=" + vm.RemoteSocket + ",local.type=unix,local.path=" + vm.LocalSocket + " -device virtio-net,netdev=" + vm.Socket + ",mac=" + vm.Mac), nil
}

func (s *Middleware) Disconnect(cmd entities.DisconnectCommand) ([]byte, error) {
	net, err := s.getNetwork(cmd.NetworkName)
	if err != nil {
		return []byte(err.Error()), nil
	}
	net.RemoveVM(cmd.MacAddress)
	return []byte(cmd.MacAddress), nil
}

func (s *Middleware) Inspect(cmd entities.InspectCommand) ([]byte, error) {
	r := []string{"Mac Addres		Ip		Socket"}
	for _, network := range s.networks {
		for _, selectedNetwork := range cmd.NetworkNames {
			if network.Name == selectedNetwork {
				r = append(r, "---------------------------------------------------------------------------------------------")
				vms, _ := network.GetVMs()
				for _, vm := range vms {
					if vm.Ip == nil {
						r = append(r, vm.Mac+"	"+"None	"+"	"+vm.Socket)
					} else {
						r = append(r, vm.Mac+"	"+*vm.Ip+"	"+vm.Socket)
					}
				}
			}
		}
	}
	return []byte(strings.Join(r, "\n")), nil
}

func (s *Middleware) Ls(cmd entities.LsCommand) ([]byte, error) {
	r := []string{"NAME", "----"}
	for _, network := range s.networks {
		r = append(r, network.Name)
	}

	return []byte(strings.Join(r, "\n")), nil
}

func (s *Middleware) Prune(cmd entities.PruneCommand) ([]byte, error) {
	r := []string{"Not implemented"}
	return []byte(strings.Join(r, "\n")), nil
}

func (s *Middleware) Rm(cmd entities.RmCommand) ([]byte, error) {
	var updatedList []*network.Network
	var r = []string{}
	removed := false
	for _, net := range s.networks {
		if cmd.NetworkName != net.Name {
			updatedList = append(updatedList, net)
		} else {
			removed = true
		}
	}
	s.networks = updatedList
	if removed {
		r = []string{cmd.NetworkName}
	} else {
		str := fmt.Sprintf("Error: unable to find network with name %s: network not found", cmd.NetworkName)
		r = []string{str}
	}
	return []byte(strings.Join(r, "\n")), nil
}
