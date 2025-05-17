package core

import (
	"fmt"
	"github.com/Rahul6700/load-balancer/models"
)

func SelectServer () *models.ServerInput {

	//var to store the chosen server
	var ChosenServer models.ServerInput;
	
	ChosenServer = ServerArray[0] //chosing the first by default
	
	var temp int
	temp = 0
	
	// picking the server with the lowest 'Active' value
	for i, server := range ServerArray {
		if(server.Active > ChosenServer.Active) {
			ChosenServer = ServerArray[i]
			temp = i
		}
	}

	// incrementing the Active count for the chosen array
	ServerArray[temp].Active++

	return ChosenServer

}

func ToServer (c* gin.Context) {
	
	//selecting the target server
	target := SelectServer()

	


}
