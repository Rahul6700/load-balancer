package core

import (
	"github.com/gin-gonic/gin"
	"fmt"
	//"net/http"
	"net/url"
	"strings"
	"github.com/Rahul6700/load-balancer/models"
	"container/heap"
)

var NameNodeSlice []string // has the url:port for all the namenode servers

func AddNameNode (c* gin.Context) {
	var URL string
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"error" : "error binding json in AddNameNode"})
	}
	NameNodeSlice = append(NameNodeSlice, data.URL)
	c.JSON(200, gin.H{"success" : fmt.Sprintf("successfully added %s",data.URL)})
	return
}


// function to add to array
func AddServer (c* gin.Context) {

	var myURL string
	var data models.ServerStruct
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"error" : "error binding json in AddServer func"})
		return
	}
	
	//validate whether json is empty or not
	if data.URL == "" {
		c.JSON(404, gin.H{"error" : "missing URL"})
		return
	}

	//verify port num
	// if !(data.Port >= 1 && data.Port <= 65535) {
	// 	c.JSON(403, gin.H{"error" : "invalid port number"})
	// 	return
	// }
	
	//validate url
	// since its a http load balancer, out server url needs to be a 'http' or 'https' one, so we convert to that form if its not already
	if !strings.HasPrefix(data.URL, "http://") && !strings.HasPrefix(data.URL, "https://") {
		myURL = "http://" + data.URL 
	} else {
		myURL = data.URL
	}
	//go's built in url parser
	_, err = url.ParseRequestURI(myURL)
	if err != nil {
		c.JSON(403, gin.H{"error" : "enter a valid URL"})
		return
	}

	// ServerArray = append(ServerArray, models.ServerStruct{
	// 	URL : myURL,
	// 	Active : 0,
	// })
	//

	server := &models.ServerStruct{
		URL:  myURL,
		Active:  0,
	}

	heap.Push(&models.MyHeap, server)

	c.JSON(200, gin.H{"success" : "ip addedd successfully"})

}


// Helper function to check if a URL/IP exists in the heap
func serverExists(h *models.ServerHeap, url string) (*models.ServerStruct, int, bool) {
	for i, s := range *h {
		if s.URL == url {
			return s, i, true
		}
	}
	return nil, -1, false
}

func DeleteServer(c *gin.Context) {

	var server models.ServerStruct
	myURL := server.URL
	var err error

	if err := c.BindJSON(&server); err != nil {
		c.JSON(400, gin.H{"error": "failed to parse JSON in deleteServer"})
		return
	}

	//validate whether json is empty or not
	if server.URL == "" {
		c.JSON(404, gin.H{"error" : "missing URL"})
		return
	}

	//validate url
	// since its a http load balancer, out server url needs to be a 'http' or 'https' one, so we convert to that form if its not already
	if !strings.HasPrefix(server.URL, "http://") && !strings.HasPrefix(server.URL, "https://") {
		myURL = "http://" + server.URL 
	} else {
		myURL = server.URL
	}
	//go's built in url parser
	_, err = url.ParseRequestURI(myURL)
	if err != nil {
		c.JSON(403, gin.H{"error" : "enter a valid URL"})
		return
	}

	//checking if the server exists in the heap
	target, index, found := serverExists(&models.MyHeap, myURL)
	if !found {
		c.JSON(404, gin.H{"error": "the server does not exist in the heap"})
		return
	}

	// Remove the server from the heap
	heap.Remove(&models.MyHeap, index)

	c.JSON(200, gin.H{"success": fmt.Sprintf("removed the server %s", target.URL)})
}


// func to list all the config'd servers
func ListServers (c* gin.Context) {
	c.JSON(200, &models.MyHeap)
}


