// Package middleware provides functionalities to manage network creation,
// connection, disconnection, inspection, and removal within a virtualized
// environment. It acts as a middleware layer between user commands and
// the underlying network operations.
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

// Middleware struct holds a slice of network pointers representing the
// managed networks.
type Middleware struct {
	networks []*network.Network
}

// Init initializes the Middleware by creating an empty slice for networks.
func (s *Middleware) Init() error {
	s.networks = []*network.Network{}

	return nil
}

// Create initializes and adds a new network to the Middleware. It takes a
// CreateCommand object, creates necessary network modules (DHCP, DNS, ARP,
// and Switch), and appends the network to the Middleware's networks slice.
// Returns the network name and any error encountered during creation.
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
	ar, err := modules.NewAddressResolution(clients)
	if err != nil {
		return []byte(err.Error()), nil
	}
	vswitch, err := modules.NewSwitch(clients)
	if err != nil {
		return []byte(err.Error()), nil
	}

	s.networks = append(
		s.networks,
		&network.Network{
			Name:                 cmd.NetworkName,
			MTU:                  1500 + 14,
			Modules:              []modules.Module{ar, dhcp, dns, vswitch},
			Clients:              clients,
			DisconnectOnPowerOff: cmd.DisconnectOnPowerOff})
	r := []string{cmd.NetworkName}
	return []byte(strings.Join(r, "\n")), nil
}

// Connect attaches a virtual machine (VM) to the specified network. It takes
// a ConnectCommand object, adds the VM to the network, and returns the network
// command required for the VM to join the network, along with any error encountered.
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

// Disconnect removes a VM from the specified network. It takes a DisconnectCommand
// object, removes the VM from the network, and returns the VM ID along with any
// error encountered.
func (s *Middleware) Disconnect(cmd entities.DisconnectCommand) ([]byte, error) {
	net, err := s.getNetwork(cmd.NetworkName)
	if err != nil {
		return []byte(err.Error()), nil
	}
	net.RemoveVM(cmd.VmID)
	return []byte(cmd.VmID), nil
}

// Inspect provides detailed information about VMs in specified networks. It takes
// an InspectCommand object, retrieves information about each VM in the networks,
// and returns the details as a formatted byte slice along with any error encountered.
func (s *Middleware) Inspect(cmd entities.InspectCommand) ([]byte, error) {
	r := []string{"ID	Mac Address		Ip		Socket"}
	for _, network := range s.networks {
		for _, selectedNetwork := range cmd.NetworkNames {
			if network.Name == selectedNetwork {
				r = append(r, "-"+selectedNetwork+"-------------------------------------------------------------------------------------------")
				vms, _ := network.Clients.GetVMs()
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

// Ls lists the names of all networks managed by the Middleware. It takes an
// LsCommand object and returns the network names as a formatted byte slice along
// with any error encountered.
func (s *Middleware) Ls(cmd entities.LsCommand) ([]byte, error) {
	r := []string{"NAME", "----"}
	for _, network := range s.networks {
		r = append(r, network.Name)
	}

	return []byte(strings.Join(r, "\n")), nil
}

// Prune is a placeholder method for pruning unused resources. Currently, it is not
// implemented and returns a "Not implemented" message.
func (s *Middleware) Prune(cmd entities.PruneCommand) ([]byte, error) {
	r := []string{"Not implemented"}
	return []byte(strings.Join(r, "\n")), nil
}

// Rm removes a network from the Middleware. It takes an RmCommand object, stops
// the specified network, and removes it from the Middleware's networks slice.
// Returns the network name if successful or an error message if the network is not found.
func (s *Middleware) Rm(cmd entities.RmCommand) ([]byte, error) {
	var updatedList []*network.Network
	var r = []string{}

	removed := false
	for _, net := range s.networks {
		if cmd.NetworkName != net.Name {
			updatedList = append(updatedList, net)
		} else {
			err := net.Stop()
			if err != nil {
				str := fmt.Sprintf("Error: cannot remove the network : %s", err.Error())
				r = []string{str}
				return []byte(strings.Join(r, "\n")), nil
			}
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

// getNetwork searches for a network by name and returns the corresponding
// network object and an error if the network is not found.
func (s *Middleware) getNetwork(nameNetwork string) (*network.Network, error) {
	for _, net := range s.networks {
		if nameNetwork == net.Name {
			return net, nil
		}
	}
	return nil, errors.New("Network not found")
}
