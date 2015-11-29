package main

import (
	"net"
	"fmt"
	"bufio"
	"Wesley/concurrentmap"
	"os"
	"strconv"
	"errors"
)

type Server struct  {
	dict concurrentmap.ConcurrentMap
	sizeMap int
	port string
}

func (server *Server) handleConnection(conn net.Conn)  {
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		//my error handling isn't great right now
		if err != nil {
			fmt.Println("Server: Conn Terminated:",err)
			conn.Close()
			return
		}
		// return response
		message = message[:len(message)-1]
		response, err := server.execInstruction(message)
		// output message received
		fmt.Print("Message:  ", string(message)," Response: ",response, "\n")
		// send new string back to client
		conn.Write([]byte(response + "\n"))
	}
}

func (server *Server) execInstruction(message string) (string, error)  {
	switch message[0]{
	case 'U': //"update"
		fallthrough
	case 'C': //"create"
		key, val, err := parseVals(message[1:])
		if err != nil {
			return "", err
		}
		server.dict.Put(key, val)
		return "T", nil
	case 'R': //"read"
		key, _, err := parseVals(message[1:])
		if err != nil {
			return "", err
		}
		val, success := server.dict.Get(key)
		if !success {
			return "F",nil
		} else {
			return "T"+val, nil
		}
	default:
		return "", errors.New("Invalid Command")
	}
}

func (server *Server) create(mess string) (string, error) {
	return "", nil
}

/*
So my format for sending key-value pairs is:
C00000004AkeyAvalue
-the 8 digit number encodes the length of the key
-the rest of the string is the value
 */
func parseVals(mess string) (string, string, error)  {
	num64, err := strconv.ParseInt(mess[0:8],10,32)
	if err != nil {
		//failure
		return "","",err
	}
	num := int(num64)
	key := mess[8:num+8]
	val := mess[num+8:]

	return key, val, nil
}

func InitServer(args []string) *Server {
	server := Server{}
	if len(args) > 0 {
		server.port = args[0]
	} else {
		server.port = "9000"
	}
	if len(args) > 1 {
		size, err := strconv.ParseInt(args[1],10,32)
		if err != nil {
			server.sizeMap = 11
		} else {
			server.sizeMap = int(size)
		}
	} else {
		server.sizeMap = 11
	}
	server.dict = &concurrentmap.ConcurrentMap2{} //or use Implementation 1
	server.dict.ConstructMap(server.sizeMap)

	return &server
}

func main() {
	fmt.Println("Launching server...")

	args := os.Args[1:]
	server := InitServer(args)

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":"+server.port)

	if err != nil {
		fmt.Println("Server failed to start:",err)
		return
	}

	// run loop forever (or until ctrl-c)
	for {
		// accept connection on port
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Server: Error:",err)
			continue
		}
		go server.handleConnection(conn)
	}
}
