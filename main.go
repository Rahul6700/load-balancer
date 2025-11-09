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

	go core.StartLeaderDiscovery()

	r.POST("/addServer", core.AddServer)
	r.GET("/listServers", core.ListServers)
	r.POST("/deleteServer", core.DeleteServer)
	
	r.POST("/uploadFile", core.HandleUpload)
	r.POST("/heartbeat", core.HandleHeartbeat)
	r.GET("/get-file-locations", core.HandleGetFileLocations)
	
	//r.NoRoute(core.Balancer)

	fmt.Println("server listening on port 8000")
	r.Run(":8000")
}

