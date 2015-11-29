package main

/*
This program demonstrates how go channels work.
 */

import (
	"fmt"
	"Wesley/go-hashmap-server/concurrentmap"
	"time"
	"strconv"
	"math/rand"
)

func pipe(receive <-chan int, send chan<- int, done chan bool) {
	//just send back what was received
	i := 0
	for {
		msg := <-receive
		send <- msg
		i++
		if i >= numRequests*numIterations {
			done <- true
			break
		}
	}
}

const numRequests int = 5
const numIterations int = 3

func makeRequest(receive <-chan int, send chan<- int, i int) {
	for k := 1; k <= numIterations; k++ {
		send <- i
		msg := <-receive
		fmt.Println("Sent:", i, "Recieved:", msg, "Message:",k)
	}
}

func ChannelExample() {
	c1 := make(chan int)
	c2 := make(chan  int)
	done := make(chan bool)

	go pipe(c1, c2, done)

	for i := 1; i<= numRequests; i++  {
		go makeRequest(c2, c1, i)
	}

	<-done
}

func mapExample() {
	cMap := concurrentmap.ConcurrentMap1{}
	cMap.ConstructMap(111)
	fmt.Println("Writing some values...")

	cMap.Put("1", "cat")
	cMap.Put("2", "dog")
	cMap.Put("3", "zebra")

	fmt.Println("Reading...")

	v, e := cMap.Get("1")
	fmt.Println("Key: 1,","Value:",v,e)
	v, e = cMap.Get("2")
	fmt.Println("Key: 2,","Value:",v,e)
	v, e = cMap.Get("3")
	fmt.Println("Key: 3,","Value:",v,e)

	cMap.DestructMap()
	<-time.After(time.Second*2)

}
const numberOfTests int = 2000000
func mapReadTestSeries(cMap concurrentmap.ConcurrentMap) int  {

	s1 := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(s1)
	notFound := 0

	for i:=0; i<numberOfTests ;i++  {
		key := strconv.Itoa(rng.Intn(numberOfTests))
		val, e := cMap.Get(key)
		if val != key {
			fmt.Println("Key/Val Mismatch:",key,val)
		}
		if !e {
			notFound++
		}
	}
	return notFound
}

func mapReadTestConcurrent(cMap concurrentmap.ConcurrentMap, numRThreads int, done chan int64) {

	for i := 0; i<numRThreads; i++ {
		go readConcurrent(cMap, i, numRThreads, done)
	}
}

func readConcurrent(cMap concurrentmap.ConcurrentMap, threadID int, numReadThreads int, done chan int64) {
	start := time.Now().UnixNano()

	s1 := rand.NewSource(time.Now().UnixNano()*int64(threadID))
	rng := rand.New(s1)
	notFound := 0

	for i:=0; i<numberOfTests/numReadThreads ;i++  {
		key := strconv.Itoa(rng.Intn(numberOfTests))
		_, e := cMap.Get(key)
		if !e {
			notFound++
		}
	}

	end := time.Now().UnixNano()

	done <- end-start
	return

}

func buildMapSeries(size int) concurrentmap.ConcurrentMap {
	cMap := concurrentmap.ConcurrentMap1{}
	cMap.ConstructMap(size)

	for i := 0; i<numberOfTests; i++  {
		s := strconv.Itoa(i)
		cMap.Put(s, s)
	}
	return &cMap
}

func buildMapConcurrent(size int, threads int, done chan int64) concurrentmap.ConcurrentMap {

	cMap := concurrentmap.ConcurrentMap1{}
	cMap.ConstructMap(size)

	for i := 0; i<threads; i++ {
		go writeConcurrent(&cMap,threads,i, done)
	}

	return &cMap
}

func writeConcurrent(cMap concurrentmap.ConcurrentMap, threads int, threadID int, done chan int64) {
	start := time.Now().UnixNano()
	numWrites := numberOfTests/threads

	for i := numWrites*threadID; i<numWrites*(threadID+1); i++ {
		s := strconv.Itoa(i)
		cMap.Put(s, s)
	}
	end := time.Now().UnixNano()

	done <- end-start
	return
}

func mapTestSeries(size int) int64 {
	start := time.Now().UnixNano()

	sMap := buildMapSeries(size)
	//notFound := mapReadTestSeries(sMap)
	mapReadTestSeries(sMap)

	end := time.Now().UnixNano()

	fmt.Println("Series Test:",float64((end-start))/1000000,"ms")

	//fmt.Println("#notFound:",notFound)
	sMap.DestructMap()
	<-time.After(time.Second*2)

	return end-start
}

func mapTestConcurrent(size int, threads int) int64 {
	//channel for getting times
	done := make(chan int64, 2*threads)

	var funcTimes int64 = 0
	start := time.Now().UnixNano()

	cMap := buildMapConcurrent(size, threads, done)
	mapReadTestConcurrent(cMap, threads, done)

	i := 0
	for t := range done  {
		funcTimes += t
		i++
		if (i == 2*threads) { //means I've read all of the values that will be read
			close(done)
		}
	}

	end := time.Now().UnixNano()

	fmt.Println("Concurrent Test:",float64((end-start))/1000000,"ms")
	//fmt.Println("Concurrent Test additive", (funcTimes)/1000, "us")

	<-time.After(time.Second*3)
	cMap.DestructMap()
	<-time.After(time.Second*2)

	return end-start
}

func mapTestDefault() int64 {
	start := time.Now().UnixNano()
	mp := make(map[string]string)
	for i:=0; i<numberOfTests; i++ {
		s := strconv.Itoa(i)
		mp[s] = s
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(s1)
	notFound := 0

	for i:=0; i<numberOfTests ;i++  {
		key := strconv.Itoa(rng.Intn(numberOfTests))
		val, e := mp[key]
		if val != key {
			fmt.Println("Key/Val Mismatch:",key,val)
		}
		if !e {
			notFound++
		}
	}
	end := time.Now().UnixNano()

	fmt.Println("Default Map Test:",float64((end-start))/1000000,"ms")

	return end-start
}


func mapTestSeries2(size int) int64 {
	start := time.Now().UnixNano()

	sMap := buildMapSeries2(size)
	//notFound := mapReadTestSeries(sMap)
	mapReadTestSeries(sMap)

	end := time.Now().UnixNano()

	fmt.Println("V2: Series Test:",float64((end-start))/1000000,"ms")

	//fmt.Println("#notFound:",notFound)
	sMap.DestructMap()
	<-time.After(time.Second*2)

	return end-start
}

func mapTestConcurrent2(size int, threads int) int64 {
	//channel for getting times
	done := make(chan int64, 2*threads)

	var funcTimes int64 = 0
	start := time.Now().UnixNano()

	cMap := buildMapConcurrent2(size, threads, done)
	mapReadTestConcurrent(cMap, threads, done)

	i := 0
	for t := range done  {
		funcTimes += t
		i++
		if (i == 2*threads) { //means I've read all of the values that will be read
			close(done)
		}
	}

	end := time.Now().UnixNano()

	fmt.Println("V2: Concurrent Test:",float64((end-start))/1000000,"ms")
	//fmt.Println("Concurrent Test additive", (funcTimes)/1000, "us")

	<-time.After(time.Second*3)
	cMap.DestructMap()
	<-time.After(time.Second*2)

	return end-start
}

func buildMapSeries2(size int) concurrentmap.ConcurrentMap {
	cMap := concurrentmap.ConcurrentMap2{}
	cMap.ConstructMap(size)

	for i := 0; i<numberOfTests; i++ {
		s := strconv.Itoa(i)
		cMap.Put(s, s)
	}
	return &cMap
}

func buildMapConcurrent2(size int, threads int, done chan int64) concurrentmap.ConcurrentMap {

	cMap := concurrentmap.ConcurrentMap2{}
	cMap.ConstructMap(size)

	for i := 0; i<threads; i++ {
		go writeConcurrent(&cMap,threads,i, done)
	}

	return &cMap
}
