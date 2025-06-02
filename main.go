package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/Rahul6700/load-balancer/core"
	"github.com/Rahul6700/load-balancer/models"
	"container/heap"
)

func main() {

	heap.Init(&models.MyHeap) //min-heap

	r := gin.Default()

	r.POST("/addServer", core.AddServer)
	r.GET("/listServers", core.ListServers)

	fmt.Println("server listening on port 8000")
	r.Run(":8000")
}

