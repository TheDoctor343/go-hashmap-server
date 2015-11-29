package main

import (
	"net"
	"fmt"
	"bufio"
	"time"
	"math/rand"
	"strconv"
	"reflect"
	"os"
	"strings"
)

type CRUD int

const (
	CREATE CRUD = iota
	READ
)

/**
Function that makes a connection to the DB and reads or writes the given values
*/
func Client(clientID int,ipAdress string, done chan int, values map[string]string, instruction CRUD) map[string]string {
	// connect to this socket
	conn, err := net.Dial("tcp", ipAdress)

	if err != nil {
		fmt.Println("Client: err:", err)
		return nil
	}


	var text string
	results := make(map[string]string)

	for key, value := range values {

		switch instruction {
		case CREATE:
			text = writeRequest(key, value)
			fmt.Fprintf(conn, text + "\n")
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Client:", clientID, "Client Error:", err)
				return nil
			}
			_, success := interpretResponse(message[:len(message) - 1])
			if !success {
				fmt.Println("Key: "+key,"failed to write")
				results[key] = ""
				continue
			} else {
				results[key] = "success"
			}
		case READ:
			text = readRequest(key)
			fmt.Fprintf(conn, text + "\n")
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Client:", clientID, "Client Error:", err)
				return nil
			}
			val, success := interpretResponse(message[:len(message) - 1])
			if !success {
				fmt.Println("Key: "+key,"value not found")
				results[key] = ""
				continue
			} else {
				results[key] = val
			}

		}
	}

	conn.Close()
	done <- 1

	return results

}

/*If the key is longer than 99,999,999 chars this will break but, I don't think that's something to worry about*/
func encodeVals(key string, value string) string {
	length := len(key)
	lenStr := fmt.Sprint(length)
	for len(lenStr) < 8 {
		lenStr = "0" + lenStr
	}

	/*Encoding is: length of key + key + value*/
	return lenStr + key + value
}

func readRequest(key string) string {
	return "R" + encodeVals(key, "")
}

func writeRequest(key string, value string) string {
	return "C" + encodeVals(key, value)
}

func interpretResponse(response string) (string, bool) {
	var responseVal string
	var responseSuccess bool
	if response[0] == 'T' {
		responseVal, responseSuccess = response[1:], true
	} else {
		responseVal, responseSuccess = "", false
	}
	return responseVal, responseSuccess
}

func filDictWithRandomVals(keyStart int, keyEnd int) map[string]string {
	dict := make(map[string]string)
	s1 := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(s1)

	for i:=keyStart; i<keyEnd; i++  {
		key := strconv.Itoa(i)
		val := strconv.Itoa(rng.Int())
		dict[key] = val
	}
	return dict
}

func testServer(clientID int, ipAdress string, done chan int, msgDict map[string]string) {
	done2 := make(chan int, 2)

	Client(clientID, ipAdress, done2, msgDict, CREATE)
	respDict := Client(clientID, ipAdress, done2, msgDict, READ)
	if respDict != nil {
		success := reflect.DeepEqual(msgDict, respDict)
		if success {
			fmt.Printf("Client %d: Success\n", clientID)
		} else {
			fmt.Printf("Client %d: Failure\n", clientID)
		}
	}

	done <- 1
}

var numclients int = 5
var testsPerClient int  = 10

func main() {
	args := os.Args[1:]
	ip := "127.0.0.1:9000"

	if len(args) > 0 {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Test Client: Command Mode")

		for  {
			done := make(chan int, 1) //the channel is a necessary arg for the Client function I built
			request := make(map[string]string)

			fmt.Print("R or W: ")
			c, err := reader.ReadByte()
			reader.ReadString('\n')
			if err != nil {
				continue
			}
			switch c {
			case 'r':
				fallthrough
			case 'R':
				fmt.Print("Key: ")
				key, err := reader.ReadString('\n')
				key = strings.TrimSpace(key)
				if err != nil {
					continue
				}
				request[key] = key
				value := Client(0, ip, done, request, READ)
				fmt.Println("Value: ",value[key])
			case 'w':
				fallthrough
			case 'W':
				fmt.Print("Key: ")
				key, err := reader.ReadString('\n')
				key = strings.TrimSpace(key)
				if err != nil {
					continue
				}
				fmt.Print("Value: ")
				value, err := reader.ReadString('\n')
				value = strings.TrimSpace(value)
				if err != nil {
					continue
				}
				request[key] = value
				Client(0, ip, done, request, CREATE)
			}
			fmt.Println("--------------------")
		}

	} else {

		fmt.Print("Number of Clients: ")
		fmt.Scanf("%d", &numclients)

		fmt.Print("Tests per Client: ")
		fmt.Scanf("%d", &testsPerClient)

		done := make(chan int, numclients)

		fmt.Println("Testing the DB...")

		for i := 0; i < numclients; i++ {
			msgDict := filDictWithRandomVals(i * testsPerClient, (i + 1) * testsPerClient)
			go testServer(i, ip, done, msgDict)
		}

		numDone := 0
		for i := range done {
			numDone += i
			if numDone == numclients {
				fmt.Println("Program Ended")
				return
			}
		}
	}
}

