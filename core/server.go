package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var ServerArray []ServerArray;

func AddServer (c* gin.Context) {

	data := c.BindJSON()
	
	//validate whether json is empty or not
	if !data.url || !data.port {
		c.JSON(404, gin.H{"error" : "missing URL or port number"})
		return
	}

	//verify port num
	if !(data.port >= 1 && data.port <= 65535) {
		c.JSON(403, gin.H{"error" : "invalid port number"})
		return
	}
	
	//validate url
	// since its a http load balancer, out server url needs to be a 'http' or 'https' one, so we convert to that form if its not already
	if !strings.HasPrefix(data.url, "http://") && !strings.HasPrefix(data.url, "https://") {
		myURL := "http://" + data.url 
	} else {
		myURL := data.url
	}
	//go's built in url parser
	parsedURL, err := url.ParseRequestURI(myURL)
	if err != nil {
		c.JSON(403, gin.H{"error" : "enter a valid URL"})
		return
	}
}



