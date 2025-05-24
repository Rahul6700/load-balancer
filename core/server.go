package core

import (
	"github.com/gin-gonic/gin"
	//"net/http"
	"net/url"
	"strings"
	"github.com/Rahul6700/load-balancer/models"
)

var serverList = &serverHeap{}

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
		URL = myURL,
		Active = 0
	}

	heap.Push(ServerList, server)

	c.JSON(200, gin.H{"success" : "ip addedd successfully"})

}

// func to list all the config'd servers
func ListServers (c* gin.Context) {
	c.JSON(200, ServerArray)
}


