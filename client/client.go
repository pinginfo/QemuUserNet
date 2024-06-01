// Package client provides functions to send various network commands
// to a server and handle the responses.
package client

import (
	"QemuUserNet/entities"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// send establishes a TCP connection to the given IP and port,
// and sends the provided data. It returns the connection and any error encountered.
func send(ip string, port int, data []byte) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Println("Socket dial error: ", err.Error())
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Println("Socket dial write error: ", err.Error())
	}
	return conn, nil
}

// listen reads the response from the server on the given connection
// and prints it if it is not "nil". Closes the connection after reading.
func listen(conn net.Conn) error {
	buffer := make([]byte, 2048)
	l, err := conn.Read(buffer)
	if err != nil {
		log.Println("Socket dial read error: ", err.Error())
	}
	if string(buffer[:l]) != "nil" {
		fmt.Println(string(buffer[:l]))
	}
	return conn.Close()
}

// Create sends a create network command to the server with the specified parameters.
func Create(ip string, port int, nameNetwork string, subnet string, gatewayIP string, gatewayMAC string, rangeIP string, dnsIP string, dnsMAC string, disconnectOnPowerOff bool) error {
	cmd := entities.CreateCommand{nameNetwork, subnet, gatewayIP, gatewayMAC, rangeIP, dnsIP, dnsMAC, disconnectOnPowerOff}
	wrapper := entities.CommandWrapper{"create", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Println("Json marshal error: ", err.Error())
	}

	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Connect sends a connect VM command to the server with the specified parameters.
func Connect(ip string, port int, nameNetwork string, vmId string) error {
	cmd := entities.ConnectCommand{nameNetwork, vmId}
	wrapper := entities.CommandWrapper{"connect", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}

	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Disconnect sends a disconnect VM command to the server with the specified parameters.
func Disconnect(ip string, port int, nameNetwork string, vmId string) error {
	cmd := entities.DisconnectCommand{nameNetwork, vmId}
	wrapper := entities.CommandWrapper{"disconnect", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}
	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Inspect sends an inspect network command to the server with the specified network names.
func Inspect(ip string, port int, names []string) error {
	cmd := entities.InspectCommand{names}
	wrapper := entities.CommandWrapper{"inspect", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}
	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Ls sends a list networks command to the server to retrieve all network names.
func Ls(ip string, port int) error {
	cmd := entities.LsCommand{}
	wrapper := entities.CommandWrapper{"ls", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}
	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Prune sends a prune command to the server to remove unused resources.
func Prune(ip string, port int) error {
	cmd := entities.PruneCommand{}
	wrapper := entities.CommandWrapper{"prune", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}
	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}

// Rm sends a remove network command to the server with the specified network name.
func Rm(ip string, port int, name string) error {
	cmd := entities.RmCommand{name}
	wrapper := entities.CommandWrapper{"rm", cmd}

	data, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Json marshal error: ", err.Error())
	}
	conn, err := send(ip, port, data)
	if err != nil {
		return err
	}
	return listen(conn)
}
