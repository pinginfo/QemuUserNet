// Package entities defines various types and structures used for managing
// virtual machines (VMs) and network commands in a virtualized environment.
package entities

// CommandType represents the type of command issued.
type CommandType string

// Enumeration of command types for network operations.
const (
	CreateCommandType     CommandType = "create"
	ConnectCommandType    CommandType = "connect"
	DisconnectCommandType CommandType = "disconnect"
	InspectCommandType    CommandType = "inspect"
	LsCommandType         CommandType = "ls"
	PruneCommandType      CommandType = "prune"
	RmCommandType         CommandType = "rm"
)

// CommandWrapper wraps a command with its type for processing.
type CommandWrapper struct {
	Type    CommandType // Type of the command
	Command interface{} // Actual command
}

// CreateCommand defines the structure for the 'create' command,
// including network configuration details.
type CreateCommand struct {
	NetworkName          string // Name of the network
	Subnet               string // Subnet address
	GatewayIP            string // Gateway IP address
	GatewayMAC           string // Gateway MAC address
	RangeIP              string // Range of IP addresses
	DnsIP                string // DNS server IP address
	DnsMAC               string // DNS server MAC address
	DisconnectOnPowerOff bool   // Flag to disconnect on power off
}

// ConnectCommand defines the structure for the 'connect' command,
// specifying the network name and VM ID.
type ConnectCommand struct {
	NetworkName string // Name of the network
	VmID        string // ID of the VM
}

// DisconnectCommand defines the structure for the 'disconnect' command,
// specifying the network name and VM ID.
type DisconnectCommand struct {
	NetworkName string // Name of the network
	VmID        string // ID of the VM
}

// InspectCommand defines the structure for the 'inspect' command,
// containing a list of network names to inspect.
type InspectCommand struct {
	NetworkNames []string // List of network names to inspect
}

// LsCommand defines the structure for the 'ls' command,
// used to list all networks.
type LsCommand struct{}

// PruneCommand defines the structure for the 'prune' command,
// used to prune unused resources.
type PruneCommand struct{}

// RmCommand defines the structure for the 'rm' command,
// specifying the network name to remove.
type RmCommand struct {
	NetworkName string // Name of the network to remove
}
