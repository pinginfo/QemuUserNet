package client

import (
	"QemuUserNet/entities"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

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
