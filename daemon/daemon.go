// Package daemon provides functions to initialize and handle network commands as a daemon server.
package daemon

import (
	"QemuUserNet/entities"
	"QemuUserNet/middleware"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var myMiddleware middleware.Middleware

// InitDaemon initializes the daemon server with the specified IP interface and port.
func InitDaemon(ipInterface string, port int) {
	myMiddleware = middleware.Middleware{}
	err := myMiddleware.Init()
	if err != nil {
		log.Println("WARNING: Error initializing middleware: ", err.Error())
	}

	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ipInterface, port))
	if err != nil {
		log.Panic("Socket error: ", err.Error())
		os.Exit(0)
	}
	defer socket.Close()

	log.Println("Middleware listing " + ipInterface + ":" + strconv.Itoa(port))

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Println("WARNING: Socket accept error: ", err.Error())
			return
		}
		go handle(conn)
	}
}

// response sends a response to the client with the provided result data and error.
func response(conn net.Conn, result []byte, e error) {
	if e != nil {
		log.Println("WARNING: error during response: ", e.Error())
		conn.Close()
		return
	}
	conn.Write(result)
	conn.Close()
	return
}

// deserialiseCommand deserializes the command from the wrapper interface and returns the command and any error encountered.
func deserialiseCommand[T any](wrapper interface{}, finalCmd T) (*T, error) {
	cmd := wrapper.(map[string]interface{})
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cmdBytes, &finalCmd)
	if err != nil {
		return nil, err
	}
	return &finalCmd, nil
}

// handle handles incoming connections by reading commands, executing them, and sending responses.
func handle(conn net.Conn) {
	var wrapper entities.CommandWrapper

	buffer := make([]byte, 2048)
	l, err := conn.Read(buffer)
	if err != nil {
		log.Println("WARNING: Socket read error: ", err.Error())
		return
	}

	err = json.Unmarshal(buffer[:l], &wrapper)

	switch wrapper.Type {
	case entities.CreateCommandType:
		var cmd entities.CreateCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : Create : ", *command)
		r, err := myMiddleware.Create(*command)
		response(conn, r, err)

	case entities.ConnectCommandType:
		var cmd entities.ConnectCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : Connect : ", *command)
		r, err := myMiddleware.Connect(*command)
		response(conn, r, err)

	case entities.DisconnectCommandType:
		var cmd entities.DisconnectCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : Disconnect : ", *command)
		r, err := myMiddleware.Disconnect(*command)
		response(conn, r, err)

	case entities.InspectCommandType:
		var cmd entities.InspectCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : Inspect : ", *command)
		r, err := myMiddleware.Inspect(*command)
		response(conn, r, err)

	case entities.LsCommandType:
		var cmd entities.LsCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : ls : ", *command)
		r, err := myMiddleware.Ls(*command)
		response(conn, r, err)

	case entities.PruneCommandType:
		var cmd entities.PruneCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : Prune : ", *command)
		r, err := myMiddleware.Prune(*command)
		response(conn, r, err)

	case entities.RmCommandType:
		var cmd entities.RmCommand
		command, err := deserialiseCommand(wrapper.Command, cmd)
		if err != nil {
			log.Println("WARNING: deserialiseCommand error")
		}
		log.Println("INFO: daemon received : rm : ", *command)
		r, err := myMiddleware.Rm(*command)
		response(conn, r, err)

	default:
		log.Println("WARNING: Unknow command")
	}
}
