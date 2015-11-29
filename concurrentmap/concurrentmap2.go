package concurrentmap

//import "fmt"

/*
This implementation is less complicated and I thought it would be a lot faster;
In actuality, it is only about twice as fast as ConcurrentMap1; which is still pretty good
-> In the best conditions it is about 72% faster than the default map implementation
*/

import "sync"

type ConcurrentMap2 struct {
	tables []subTable
	numTables int
}

type subTable struct {
	table map[string]string
	mutex *sync.Mutex
}

func (CMap *ConcurrentMap2) ConstructMap(numTables int) {
	CMap.numTables = numTables
	CMap.tables = make([]subTable, numTables)
	for i:=0; i<numTables;i++  {
		CMap.tables[i].table = make(map[string]string)
		CMap.tables[i].mutex = &sync.Mutex{}
	}
}

func (CMap *ConcurrentMap2) DestructMap()  {
	//Method implemented but does not need to do anything
//	var table subTable
//	for _, table = range CMap.tables  {
//		fmt.Println("Table Size:",len(table.table))
//	}
}

func (CMap *ConcurrentMap2) Put(key string, value string) {
	index := findTable(key, CMap.numTables)
	table := CMap.tables[index]

	table.mutex.Lock()
	table.table[key] = value

	table.mutex.Unlock()
}

func (CMap *ConcurrentMap2) Get(key string) (string, bool)  {
	index := findTable(key, CMap.numTables)
	table := CMap.tables[index]

	table.mutex.Lock()
	v, e := table.table[key]
	table.mutex.Unlock()

	return v, e
}