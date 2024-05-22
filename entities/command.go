package entities

type CommandType string

const CreateCommandType CommandType = "create"
const ConnectCommandType CommandType = "connect"
const DisconnectCommandType CommandType = "disconnect"
const InspectCommandType CommandType = "inspect"
const LsCommandType CommandType = "ls"
const PruneCommandType CommandType = "prune"
const RmCommandType CommandType = "rm"

type CommandWrapper struct {
	Type    CommandType
	Command interface{}
}

type CreateCommand struct {
	NetworkName string
	Subnet      string
	GatewayIP   string
	GatewayMAC  string
	RangeIP     string
	DnsIP       string
	DnsMAC      string
}

type ConnectCommand struct {
	NetworkName string
	MacAddress  string
}

type DisconnectCommand struct {
	NetworkName string
	MacAddress  string
}

type InspectCommand struct {
	NetworkNames []string
}

type LsCommand struct{}

type PruneCommand struct{}

type RmCommand struct {
	NetworkName string
}
