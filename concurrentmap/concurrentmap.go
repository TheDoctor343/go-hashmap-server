package concurrentmap
import (
	"hash/fnv"
)

type ConcurrentMap interface {
	Put(key string, value string)
	Get(key string) (string, bool)
	ConstructMap(numTables int)
	DestructMap()
}

/*
Limitations of my current design:
-keys are limited 127 chars in length (values can be arbitrary length)
 */

/*Concurrent Map*/
type ConcurrentMap1 struct {
	tables []tableCom
	numTables int
	done chan bool
}

/*ConcurrentMap function for adding and replacing values*/
func (CMap *ConcurrentMap1) Put(key string, value string)  {
	//find what table this key should be in
	index := findTable(key, CMap.numTables)
	//construct the 'message' to be sent to the goroutine managing that table
	msg := "W"+encodeValue(key, value)
	//send the message
	CMap.tables[index].request <- msg
}

/*ConcurrentMap function for getting values and membership testing*/
func (CMap *ConcurrentMap1) Get(key string) (string, bool)  {
	//find what table this key should be in
	index := findTable(key, CMap.numTables)
	//construct the 'message' to be sent to the goroutine managing that table
	msg := "R"+encodeValue(key, "")
	//send the message
	CMap.tables[index].request <- msg
	//get the response
	resp := <-CMap.tables[index].response

	//check the result
	if resp[0] == 'T' { //The key exists
		return resp[1:], true
	} else { //The key was not found
		return "", false
	}
}

/*
Struct to hold the channel pointers that will be used to communicate
with pieces of the 'database'
*/
type tableCom struct {
	request chan string
	response chan string
}

/*Function to initialize the 'Database'*/
func (CMap *ConcurrentMap1) ConstructMap(numTables int) {
	CMap.done = make(chan bool)

	tables := make([]tableCom, numTables)

	for index, _ := range tables {
		//create channels for this part of the data table
		tables[index] = tableCom{
			request: make(chan string),
			response: make(chan string)}
		//create a routine to handle this new table
		go tableManager(&tables[index], CMap.done)
	}

	CMap.tables = tables
	CMap.numTables = numTables
}

/*
A function do stop all of the goroutines when the map is
done being used.
 */
func (CMap *ConcurrentMap1) DestructMap()  {
	for _, val := range CMap.tables  {
		close(val.request) //signals that there are no more requests coming; shuts down routines
	}
}

/*
A function to manage a single piece of the 'Database'
This functions waits on its request channel for requests,
and then sends out a response once it receives a request
*/
func tableManager(communicator *tableCom, done chan bool)  {
	storage := make(map[string]string)
	var i int = 0 //keep track of # of requests
	for task := range communicator.request {
		i++
		switch task[0] {
		case 'R':
			//read the value from storage and send it
			value, present := processRead(storage, task[1:])
			if present {
				communicator.response <- "T" + value //indicate a value was found
			} else {
				communicator.response <- "F"
			}
		case 'W':
			processWrite(storage, task[1:])
		}
	}
	//Once loop exits the channel should be closed
	//fmt.Println("On Exit:  Size:",len(storage)," #Requests:",i)
}

func processRead(storage map[string]string, msg string) (string, bool) {
	key, _ := decodeValues(msg)
	val, prs := storage[key]
	return val, prs
}

func processWrite(storage map[string]string, msg string) {
	key, value := decodeValues(msg)
	storage[key] = value
}

/*Parses the key and value from a string*/
func decodeValues(msg string) (string, string) {
	key_end := msg[0]
	key := msg[1:key_end+1]
	value := msg[key_end+1:len(msg)]

	return key, value
}

func encodeValue(key string, value string) string {
	key_len := len(key)
	return string(key_len)+key+value
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

/*Struct to find which table a key-value pair should be in*/
func findTable(key string, numTables int) int {
	return int(hash(key) % uint32(numTables))
}