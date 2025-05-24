package models

import (
	"container/heap"
)

type ServerHeap []*models.ServerStruct

func SelectServer() *models.ServerStruct {
	if ServerPool.Len() == 0 {
		return nil
	}
	server := heap.Pop(ServerPool).(*models.ServerStruct)
	server.Active++
	heap.Push(ServerPool, server)
	return server
}

func DoneWithServer(server *models.ServerStruct) {
	server.Active--
	heap.Fix(ServerPool, server.Index)
}

func (h ServerHeap) Len() int           { return len(h) }
func (h ServerHeap) Less(i, j int) bool { return h[i].Active < h[j].Active }
func (h ServerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *ServerHeap) Push(x any) {
	n := len(*h)
	server := x.(*models.ServerStruct)
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
