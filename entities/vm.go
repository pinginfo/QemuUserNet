package entities

type VM struct {
	ID           string
	Mac          string
	Socket       string
	RemoteSocket string
	LocalSocket  string
	Ip           *string
}
