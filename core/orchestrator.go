package core

import (
	"github.com/Rahul6700/foodo/shared"
	"github.com/Rahul6700/load-balancer/models"
)

// this is the struct that we use inside the HandleUpload function 
type ClientUploadRequest struct {
	Filename string
	Chunks []struct {
		ChunkID string
		Index int
	}
}

func HandleUpload (c *gin.Conext) {

	// this function basically recieves Filename, ChunkID and Index and decides on which chunk is stored where and informs the raft leader the same
	// each hcunk is stored in replicatation_factor number of datanodes
	const replicatation_factor = 3

	var variable ClientUploadRequest // the variable stores the client request content
	if err := c.BindJSON(&uploadRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request :details": err.Error()})
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
		for i := 0; i < replicatation_factor; i++ {
			server := models.SelectServer() // this selects the least active server connections from the heap
			if server == nil { // heap is empty, so no server added
				c.JSON(404, gin.H{"error" : "No datanodes available"})
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
			chunkID = chunk.chunkID,
			chunkIndex = chunk.Index,
			Locations = chunk.locationsSlice
		})

	}

	

}
