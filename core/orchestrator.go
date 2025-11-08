package core

import (
	"github.com/Rahul6700/Foodo/shared"
	"github.com/Rahul6700/load-balancer/models"
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
)

// this is the struct that we use inside the HandleUpload function 
type ClientUploadRequest struct {
	Filename string `json:"filename"`
	Chunks []struct {
		ChunkID string `json:"chunkID"`
		Index int `json:"index"`
	} `json:"chunks"`
}

// var currentLeaderAddress string // the is the url:port of the current raft cluster leader

func HandleUpload (c *gin.Context) {

	// this function basically recieves Filename, ChunkID and Index and decides on which chunk is stored where and informs the raft leader the same
	// each hcunk is stored in replicatation_factor number of datanodes
	const replication_factor = 3

	var variable ClientUploadRequest// the variable stores the client request content
	if err := c.BindJSON(&variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request:" + err.Error()})
		return
	}

	uploadPlan := make(map[string][]string) // plan for the client -> chunkID - [url1, url2, url3] map
	// shared.chunkStruct is the struct imported from the foodo project, it basically looks like this,
//	type ChunkStruct struct {
//	ChunkID string
//	ChunkIndex int
//	Locations []string
//} 
	raftChunks := []shared.ChunkStruct{} // plan for the Namenode -> the official data that we'll send and will be stored in the NameNode
	// the namenode will use this metadata to map chunks to datanodes

	// now we iterate through every node
	for _, chunk := range variable.Chunks {
		// for every chunk we make a new 'locations' array which contains the addresses of the datanodes in which its running
		var locationsSlice []string
		// now we run the server selection for each chunk 'replicatation_factor' number of times
		for i := 0; i < replication_factor; i++ {
			server := models.SelectServer() // this selects the least active server connections from the heap
			if server == nil { // heap is empty, so no server added
				c.JSON(404, gin.H{"error" : "No datanodes available"})
				return
			}
			locationsSlice = append(locationsSlice, server.URL)
		}
		uploadPlan[chunk.ChunkID] = locationsSlice // we push the chunkID as the key and its locationSLice as the value, do this for every chunkID
		// this is the raftChunks slice that is sent to the namenode, using this only the namenode will map
		// this is how it looks,
// 		raftChunks = [
//     ChunkStruct{
//         ChunkID:    "123",
//         ChunkIndex: 0,
//         Locations:  ["http://dn1:9000", "http://dn3:9000", "http://dn2:9000"]
//     },
//     
//     ChunkStruct{
//         ChunkID:    "789",
//         ChunkIndex: 1,
//         Locations:  ["http://dn3:9000", "http://dn1:9000", "http://dn2:9000"]
//     }
// ]

	raftChunks = append(raftChunks, shared.ChunkStruct{
    ChunkID:    chunk.ChunkID,
    ChunkIndex: chunk.Index,
    Locations:  locationsSlice,
	})

	}

	// creates a raftCommand that will be sent to the namenode
	// it contains the operation (adding file to DN), filename and the raftChunks (slice of chunk metadata)
	raftCommand := shared.RaftCommand {
		Operation: "REGISTER_FILE",
		Filename: variable.Filename,
		Chunks: raftChunks,
	}
	
	// now we have raft Command which is a struct of metadata
	// we should sent this metadata to the namenode now, but this is not the right format
	// we'll convert out go struct into json byte slice and send that over the network (use the "marshal()" function for this)

	JSONByteSlice, err := json.Marshal(raftCommand) // JSONByteSlice is the JSON byte slice of our raftCommand struct
	if err != nil {
		c.JSON(500, gin.H{"error" : "error converting raftCommand to JSONByteSlice in the loadb"})
		return
	}

	leaderAddr := GetCurrentLeader() // func in discovery.go
	if leaderAddr == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error" : "No NameNode Leader found"})
		return
	}

	// now that we have found the leader NameNode we send the JSONByteSlice to it, so it can store the metadata
	// we'll use a HTTP post req for this, we'll construct the ReqURL for the req
	ReqURL := leaderAddr + "/raft/propose" // so it'll look something like http://127.0.0.1:8080/raft/propose
	resp, err := http.Post(ReqURL, "application/json", bytes.NewBuffer(JSONByteSlice)) // sending the post req to ReqURl
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(500, gin.H{"error" : "couldnt send the loadb plan to the namenode thru http post"})
		return
	}

	resp.Body.Close()

	// if everything is successful, the plan is sent
	c.JSON(201, gin.H{"success" : true, "upload_plan" : uploadPlan})


}
