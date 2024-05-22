package middleware

import (
	"QemuUserNet/entities"
	"QemuUserNet/modules"
	"QemuUserNet/network"
	"QemuUserNet/tools"
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

	clients := &entities.Clients{}

	dhcp, err := modules.NewDhcp(cmd.Subnet, cmd.GatewayIP, cmd.GatewayMAC, cmd.RangeIP, cmd.DnsIP, clients)
	if err != nil {
		return []byte(err.Error()), nil
	}
	dns, err := modules.NewDns(cmd.DnsIP, cmd.DnsMAC, clients)
	if err != nil {
		return []byte(err.Error()), nil
	}

	s.networks = append(
		s.networks,
		&network.Network{
			Name:    cmd.NetworkName,
			MTU:     1500 + 14,
			Modules: []modules.Module{dhcp, dns},
			Clients: clients,
		})
	r := []string{cmd.NetworkName}
	return []byte(strings.Join(r, "\n")), nil
}

func (s *Middleware) Connect(cmd entities.ConnectCommand) ([]byte, error) {
	net, err := s.getNetwork(cmd.NetworkName)
	if err != nil {
		return []byte(err.Error()), nil
	}
	vm, err := net.AddVM(cmd.VmID)
	if err != nil {
		return []byte(err.Error()), nil
	}
	return tools.CraftQemuNetworkCommand(vm.Socket, vm.RemoteSocket, vm.LocalSocket, vm.Mac), nil
}

func (s *Middleware) Disconnect(cmd entities.DisconnectCommand) ([]byte, error) {
	net, err := s.getNetwork(cmd.NetworkName)
	if err != nil {
		return []byte(err.Error()), nil
	}
	net.RemoveVM(cmd.VmID)
	return []byte(cmd.VmID), nil
}

func (s *Middleware) Inspect(cmd entities.InspectCommand) ([]byte, error) {
	r := []string{"ID	Mac Address		Ip		Socket"}
	for _, network := range s.networks {
		for _, selectedNetwork := range cmd.NetworkNames {
			if network.Name == selectedNetwork {
				r = append(r, "---------------------------------------------------------------------------------------------")
				vms, _ := network.GetVMs()
				for _, vm := range vms {
					if vm.Ip == nil {
						r = append(r, vm.ID+"	"+vm.Mac+"	"+"None	"+"	"+vm.Socket)
					} else {
						r = append(r, vm.ID+"	"+vm.Mac+"	"+*vm.Ip+"	"+vm.Socket)
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
