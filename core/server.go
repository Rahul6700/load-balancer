package core

import (
	"github.com/gin-gonic/gin"
	//"net/http"
	"net/url"
	"strings"
	"github.com/Rahul6700/load-balancer/models"
)

var ServerArray []models.ServerInput;

// function to add to array
func AddServer (c* gin.Context) {

	var myURL string
	var data models.ServerInput
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"error" : "error binding json in AddServer func"})
		return
	}
	
	//validate whether json is empty or not
	if data.URL == "" || data.Port == 0 {
		c.JSON(404, gin.H{"error" : "missing URL or port number"})
		return
	}

	//verify port num
	if !(data.Port >= 1 && data.Port <= 65535) {
		c.JSON(403, gin.H{"error" : "invalid port number"})
		return
	}
	
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

	ServerArray = append(ServerArray, models.ServerInput{
		URL : myURL,
		Port : data.Port,
	})

	c.JSON(200, gin.H{"success" : "ip and port addedd successfully"})

}



