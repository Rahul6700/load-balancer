package models

import (
	"fmt"
	"container/heap"
	"sync"
)

var lock sync.Mutex // this is the mutex lock we'll use to lock the heap to prevent race conditions when one goroutine is modyfing it

// ServerHeap is the data structure
type ServerHeap []*ServerStruct

// declaring the global heap
var MyHeap ServerHeap

func SelectServer() *ServerStruct {

	lock.Lock() // lock the mutex here
	defer lock.Unlock() // this runs the unlock call once the function ends

	// var to store the server (now using pointer)
	var server *ServerStruct
	// remove the node from the heap
	server = heap.Pop(&MyHeap).(*ServerStruct)
	// increment active value
	//server.Active++ -> we comment out this part, this works for our reverse proxy but not our raft part
	// add it to the heap again to re-heapify so it goes to the right place
	heap.Push(&MyHeap, server)
	return server
}

// new raft exclusive function, the datanode sends a heartbeat to the loadb once every few seconds with its actual current load, which is written 
// to the loadb's heap. this way we have an actual value of how many are active
func UpdateServerLoad(url string, activeWrites int) error {
	lock.Lock()
	defer lock.Unlock()

	for _, server := range MyHeap {
		if server.URL == url {
			// Set the 'Active' count to the "true" value
			server.Active = activeWrites
			
			// Re-sort the heap with the new, correct priority
			heap.Fix(&MyHeap, server.Index)
			return nil
		}
	}
	return fmt.Errorf("server %s not found in heap", url)
}

func DoneWithServer(server *ServerStruct) {
	lock.Lock() 
	defer lock.Unlock()
	// decrement active counter after the server is done
	server.Active--
	heap.Fix(&MyHeap, server.Index)
}

// helper functions
// for heap len
func (h ServerHeap) Len() int {
	return len(h)
}

func (h ServerHeap) Less(i, j int) bool {
	return h[i].Active < h[j].Active
}

func (h ServerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *ServerHeap) Push(x any) {
	n := len(*h)
	server := x.(*ServerStruct)
	server.Index = n
	*h = append(*h, server)
}

func (h *ServerHeap) Pop() any {
	old := *h
	n := len(old)
	server := old[n-1]
	server.Index = -1
	*h = old[0 : n-1]
	return server
}
