package core

import (
	"github.com/Rahul6700/load-balancer/models"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
)

// helper func which takes the dest server URL and creates a reverse proxy to the selected server
func ProxyRequest(c* gin.Context, targetURL string) {
	
	server, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(500, gin.H{"error" : "targetURL is invalid in proxyReq func"})
		return
	}
	//parsing the URL splits it into 4 parts
	// 1. scheme -> http, https
	// 2. host -> rahul.com, localhost:8000
	// 3. path -> /api/users
	// 4. query -> ?page=2

	// we need to take the req's scheme, host and all and reconstruct a new req to the server from the LB

	//======

	// this method -> takes care of establishing a TCP connection between the balancer and the selected server (call this connection B).
	// TCP connection A is the one between the client and the balancer for this particular req.
	// this function sets up a proxy to forward the req's from connection A to B and acts a bridge between the 2
	// this function abstracts all the working there
	proxy := httputil.NewSingleHostReverseProxy(server)


	//now we make the req to send to the server from the LB (basically rebuilding the req from client to LB)
	c.Request.URL.Scheme = server.Scheme
	c.Request.URL.Host = server.Host
	c.Request.Host = server.Host


	// in the created proxy, it sends the c.Request to the server, waits for response and 'writes' the contents of
	//the response (c.Writer) to the client.
	proxy.ServeHTTP(c.Writer, c.Request)

}

//the main function for the balancing
func Balancer(c* gin.Context) {
	
	//select a server to use
	server := models.SelectServer()

	if server.URL == "" {
		c.JSON(503, gin.H{"error" : "no servers available to serve the request"})
		return
	}
	//this decrements the Active field for the selected server after its done
	defer models.DoneWithServer(&server)

	ProxyRequest(c, server.URL)

}


