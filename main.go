package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/Rahul6700/load-balancer/core"
)

func main() {

	heap.Init(ServerList) //min-heap

	r := gin.Default()

	r.POST("/addServer", core.AddServer)
	r.GET("/ListServers", core.ListServers)

	fmt.Println("server listening on port 8000")
	r.Run(":8000")
}

