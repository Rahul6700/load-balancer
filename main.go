package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AddServer (c* gin.Context) {
	data := c.BindJSON()
	
	if !data.url || !data.port {
		c.JSON(404, gin.H{"error" : "missing URL or port number"})
		return
	}

	if !(data.port >= 1 && data.port <= 65535) {
		c.JSON(403, gin.H{"error" : "invalid port number"})
		return
	}
	
	if !strings.HasPrefix(data.url, "http://") && !strings.HasPrefix(data.url, "https://") {
		myURL := "http://" + data.url 
	} else {
		myURL := data.url
	}

	parsedURL, err := url.ParseRequestURI(myURL)
	if err != nil {
		c.JSON(403, gin.H{"error" : "enter a valid URL"})
		return
	}

}

func main () {

	var arr string = []

	r := gin.Default()

	r.POST("/addServer", func (c* gin.Context) {
		link := c.BindJSON()
		if link.url == ""
		arr.append(link.url)
	})
}
