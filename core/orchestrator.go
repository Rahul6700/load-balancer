package core

import (
	"io"
	"log"
	//"github.com/Rahul6700/Foodo/shared"
	"github.com/Rahul6700/load-balancer/models"
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
)

// this is the struct that we use inside the HandleUpload function 
type ClientUploadRequest struct {
	Filename string `json:"filename"`
	Chunks []struct {
		ChunkID string `json:"chunk_id"`
		Index int `json:"index"`
	} `json:"chunks"`
}

// this is the truct for the RaftCommmand
type RaftCommand struct {
	Operation string `json:"operation"`
	Filename string `json:"filename"`
	Chunks []ChunkStruct `json:"chunks"`
}

// this is the helper struct
type ChunkStruct struct {
	ChunkID string `json:"chunk_id"`
	ChunkIndex int `json:"chunk_index"`
	Locations []string `json:"locations"`
}

// HeartbeatPayload is used by the DN's to send heartbeat's to the LB
type HeartbeatPayload struct {
	NodeID       string `json:"node_id"`       // DN's full url -> "http://192.168.1.15:9001"
	ActiveWrites int    `json:"active_writes"` // load tracked by the atomic counter
}

// var currentLeaderAddress string // the is the url:port of the current raft cluster leader

// HandleHeartbeat is called by the datanode's once every few seconds
// the DN's send the LB info about how many current active writes they have, the loadb updates its active count in the heap accordingly
func HandleHeartbeat(c *gin.Context) {
	//temp struct to store the incoming data from the DN
	var heartbeat struct {
		NodeID       string `json:"node_id"` // This is the Datanode's URL
		ActiveWrites int    `json:"active_writes"`
	}

	if err := c.BindJSON(&heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid heartbeat"})
		return
	}

	//we now call the UpdateServerLoad func in heap.go which takes the nodeID and the active count and updates the heap and reheapify's
	err := models.UpdateServerLoad(heartbeat.NodeID, heartbeat.ActiveWrites)
	if err != nil {
		// this error happens if the DN sent a heartbeat but the DN is not in the heap
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	// on success
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func HandleUpload (c *gin.Context) {
	const replication_factor = 3
	
    // LOG THE RAW REQUEST BODY FIRST
    bodyBytes, _ := io.ReadAll(c.Request.Body)
    log.Printf("=== RAW REQUEST BODY ===")
    log.Printf("%s", string(bodyBytes))
    log.Printf("========================")
    
    // Restore the body so BindJSON can read it
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var variable ClientUploadRequest
	if err := c.BindJSON(&variable); err != nil {
		log.Printf("HandleUpload: ERROR: Failed to bind JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request:" + err.Error()})
		return
	}

	// DEBUG: Log what we received from client
	log.Printf("=== CLIENT REQUEST DEBUG ===")
	log.Printf("Filename: %s", variable.Filename)
	log.Printf("Number of chunks: %d", len(variable.Chunks))
	for i, chunk := range variable.Chunks {
		log.Printf("Client Chunk %d: ChunkID=%s, Index=%d", i, chunk.ChunkID, chunk.Index)
	}
	log.Printf("===========================")

	uploadPlan := make(map[string][]string)
	raftChunks := []ChunkStruct{}

	// Build the raft chunks
	for _, chunk := range variable.Chunks {
		var locationsSlice []string
		for i := 0; i < replication_factor; i++ {
			server := models.SelectServer()
			if server == nil {
				log.Println("HandleUpload: ERROR: Not enough datanodes available")
				c.JSON(404, gin.H{"error" : "No datanodes available"})
				return
			}
			locationsSlice = append(locationsSlice, server.URL)
		}
		uploadPlan[chunk.ChunkID] = locationsSlice
		
		raftChunks = append(raftChunks, ChunkStruct{
			ChunkID:    chunk.ChunkID,
			ChunkIndex: chunk.Index,
			Locations:  locationsSlice,
		})
	}

	// DEBUG: Log what we're sending to namenode (MOVED HERE)
	log.Printf("=== RAFT CHUNKS DEBUG ===")
	for i, chunk := range raftChunks {
		log.Printf("Raft Chunk %d: ChunkID=%s, Index=%d, Locations=%v", 
			i, chunk.ChunkID, chunk.ChunkIndex, chunk.Locations)
	}
	log.Printf("========================")

	log.Printf("the filename is: %s", variable.Filename)
	raftCommand := RaftCommand {
		Operation: "REGISTER_FILE",
		Filename: variable.Filename,
		Chunks: raftChunks,
	}
	
	JSONByteSlice, err := json.Marshal(raftCommand)
	log.Printf("JSON being sent to namenode: %s", string(JSONByteSlice))
	
	if err != nil {
		log.Printf("HandleUpload: ERROR: Failed to marshal RaftCommand: %v\n", err)
		c.JSON(500, gin.H{"error" : "error converting raftCommand to JSONByteSlice"})
		return
	}

	leaderAddr := GetCurrentLeader()
	if leaderAddr == "" {
		log.Println("HandleUpload: ERROR: No Namenode Leader found") 
		c.JSON(http.StatusServiceUnavailable, gin.H{"error" : "No NameNode Leader found"})
		return
	}

	ReqURL := leaderAddr + "/raft/propose"
	resp, err := http.Post(ReqURL, "application/json", bytes.NewBuffer(JSONByteSlice))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("HandleUpload: ERROR: Namenode returned error. Status: %s, Err: %v\n", resp.Status, err)
		c.JSON(500, gin.H{"error" : "couldnt send plan to namenode"})
		return
	}

	resp.Body.Close()
	c.JSON(200, gin.H{"success" : true, "upload_plan" : uploadPlan})
}


// func HandleUpload (c *gin.Context) {
//
// 	// this function basically recieves Filename, ChunkID and Index and decides on which chunk is stored where and informs the raft leader the same
// 	// each hcunk is stored in replicatation_factor number of datanodes
// 	const replication_factor = 3
//
// 	var variable ClientUploadRequest// the variable stores the client request content
// 	if err := c.BindJSON(&variable); err != nil {
// 		log.Printf("HandleUpload: ERROR: Failed to bind JSON: %v\n", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request:" + err.Error()})
// 		return
// 	}
//
// // ADD THIS DEBUG LOGGING:
// log.Printf("=== CLIENT REQUEST DEBUG ===")
// log.Printf("Filename: %s", variable.Filename)
// log.Printf("Number of chunks: %d", len(variable.Chunks))
// for i, chunk := range variable.Chunks {
//     log.Printf("Client Chunk %d: ChunkID=%s, Index=%d", i, chunk.ChunkID, chunk.Index)
// }
// log.Printf("===========================")
//
// 	uploadPlan := make(map[string][]string) // plan for the client -> chunkID - [url1, url2, url3] map
// 	// shared.chunkStruct is the struct imported from the foodo project, it basically looks like this,
// //	type ChunkStruct struct {
// //	ChunkID string
// //	ChunkIndex int
// //	Locations []string
// //} 
// 	raftChunks := []shared.ChunkStruct{} // plan for the Namenode -> the official data that we'll send and will be stored in the NameNode
// 	// the namenode will use this metadata to map chunks to datanodes
//
// log.Printf("=== RAFT CHUNKS DEBUG ===")
// for i, chunk := range raftChunks {
//     log.Printf("Raft Chunk %d: ChunkID=%s, Index=%d, Locations=%v", 
//         i, chunk.ChunkID, chunk.ChunkIndex, chunk.Locations)
// }
// log.Printf("========================")
//
//
// 	// now we iterate through every node
// 	for _, chunk := range variable.Chunks {
// 		// for every chunk we make a new 'locations' array which contains the addresses of the datanodes in which its running
// 		var locationsSlice []string
// 		// now we run the server selection for each chunk 'replicatation_factor' number of times
// 		for i := 0; i < replication_factor; i++ {
// 			server := models.SelectServer() // this selects the least active server connections from the heap
// 			if server == nil { // heap is empty, so no server added
// 				log.Println("HandleUpload: ERROR: Not enough datanodes available in heap (SelectServer returned nil).")
// 				    log.Printf("Raft Chunk : ChunkID=%s, Index=%d, Locations=%v", chunk.ChunkID, chunk.ChunkIndex, chunk.Locations)
// 				c.JSON(404, gin.H{"error" : "No datanodes available"})
// 				return
// 			}
// 			locationsSlice = append(locationsSlice, server.URL)
// 		}
// 		uploadPlan[chunk.ChunkID] = locationsSlice // we push the chunkID as the key and its locationSLice as the value, do this for every chunkID
// 		// this is the raftChunks slice that is sent to the namenode, using this only the namenode will map
// 		// this is how it looks,
// // 		raftChunks = [
// //     ChunkStruct{
// //         ChunkID:    "123",
// //         ChunkIndex: 0,
// //         Locations:  ["http://dn1:9000", "http://dn3:9000", "http://dn2:9000"]
// //     },
// //     
// //     ChunkStruct{
// //         ChunkID:    "789",
// //         ChunkIndex: 1,
// //         Locations:  ["http://dn3:9000", "http://dn1:9000", "http://dn2:9000"]
// //     }
// // ]
//
// 	raftChunks = append(raftChunks, shared.ChunkStruct{
//     ChunkID:    chunk.ChunkID,
//     ChunkIndex: chunk.Index,
//     Locations:  locationsSlice,
// 	})
//
// 	}
//
// 	// creates a raftCommand that will be sent to the namenode
// 	// it contains the operation (adding file to DN), filename and the raftChunks (slice of chunk metadata)
// 	log.Printf("the filename i: %s", variable.Filename)
// 	raftCommand := shared.RaftCommand {
// 		Operation: "REGISTER_FILE",
// 		Filename: variable.Filename,
// 		Chunks: raftChunks,
// 	}
// 	
// 	// now we have raft Command which is a struct of metadata
// 	// we should sent this metadata to the namenode now, but this is not the right format
// 	// we'll convert out go struct into json byte slice and send that over the network (use the "marshal()" function for this)
//
// 	JSONByteSlice, err := json.Marshal(raftCommand) // JSONByteSlice is the JSON byte slice of our raftCommand struct
//
// 	log.Printf("the filename i: %s", variable.Filename)
// 	if err != nil {
// 		log.Printf("HandleUpload: ERROR: Failed to marshal RaftCommand: %v\n", err)
// 		c.JSON(500, gin.H{"error" : "error converting raftCommand to JSONByteSlice in the loadb"})
// 		return
// 	}
//
// 	leaderAddr := GetCurrentLeader() // func in discovery.go
// 	if leaderAddr == "" {
// 		log.Println("HandleUpload: ERROR: No Namenode Leader found (GetCurrentLeader returned empty string).") 
// 		c.JSON(http.StatusServiceUnavailable, gin.H{"error" : "No NameNode Leader found"})
// 		return
// 	}
//
// 	// now that we have found the leader NameNode we send the JSONByteSlice to it, so it can store the metadata
// 	// we'll use a HTTP post req for this, we'll construct the ReqURL for the req
// 	ReqURL := leaderAddr + "/raft/propose" // so it'll look something like http://127.0.0.1:8080/raft/propose
// 	resp, err := http.Post(ReqURL, "application/json", bytes.NewBuffer(JSONByteSlice)) // sending the post req to ReqURl
// 	if err != nil || resp.StatusCode != http.StatusOK {
// 		log.Printf("HandleUpload: ERROR: Namenode leader at %s returned an error. Status: %s, Err: %v\n", ReqURL, resp.Status, err)
// 		c.JSON(500, gin.H{"error" : "couldnt send the loadb plan to the namenode thru http post"})
// 		return
// 	}
//
// 	resp.Body.Close()
//
// 	// if everything is successful, the plan is sent
// 	c.JSON(200, gin.H{"success" : true, "upload_plan" : uploadPlan})
//
//
// }
//
// HandleGetFileLocations proxies a download request to the Namenode leader.
// func HandleGetFileLocations(c *gin.Context) {
// 	// 1. Find the Namenode leader
// 	leaderAddr := GetCurrentLeader() // From your discovery.go
// 	if leaderAddr == "" {
// 		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no namenode leader available"})
// 		return
// 	}
//
// 	// 2. Get the filename from the client's query
// 	// (e.g., /api/get-file-locations?filename=foo.txt)
// 	fileName := c.Query("filename")
// 	if fileName == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'filename' query parameter"})
// 		return
// 	}
//
// 	// 3. Create a reverse proxy to the leader
// 	targetUrl, _ := url.Parse(leaderAddr)
// 	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
//
// 	// 4. Update the request URL to point to the *Namenode's* endpoint
// 	//    This will forward the query ?filename=foo.txt
// 	c.Request.URL.Path = "/get-metadata"
// 	c.Request.URL.Host = targetUrl.Host
// 	c.Request.URL.Scheme = targetUrl.Scheme
// 	c.Request.Host = targetUrl.Host
//
// 	// 5. Serve the request. The proxy streams the Namenode's
// 	//    JSON response directly back to the client.
// 	proxy.ServeHTTP(c.Writer, c.Request)
// }
//

func HandleGetFileLocations(c *gin.Context) {
	log.Println("[INFO] HandleGetFileLocations called")

	// 1. Find the Namenode leader
	leaderAddr := GetCurrentLeader() // From your discovery.go
	if leaderAddr == "" {
		log.Println("[ERROR] No Namenode leader available")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no namenode leader available"})
		return
	}
	log.Printf("[DEBUG] Current Namenode leader: %s\n", leaderAddr)

	// 2. Get the filename from the client's query
	fileName := c.Query("filename")
	if fileName == "" {
		log.Println("[WARN] Missing 'filename' query parameter in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'filename' query parameter"})
		return
	}
	log.Printf("[INFO] Requested file metadata: filename=%s\n", fileName)

	// 3. Create a reverse proxy to the leader
	targetUrl, err := url.Parse(leaderAddr)
	if err != nil {
		log.Printf("[ERROR] Failed to parse leader address '%s': %v\n", leaderAddr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid leader address"})
		return
	}
	log.Printf("[DEBUG] Parsed leader URL: scheme=%s, host=%s\n", targetUrl.Scheme, targetUrl.Host)

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	log.Println("[INFO] Reverse proxy created successfully")

	// 4. Modify the request to target the Namenode's /get-metadata endpoint
	originalPath := c.Request.URL.Path
	originalHost := c.Request.Host
	c.Request.URL.Path = "/get-metadata"
	c.Request.URL.Host = targetUrl.Host
	c.Request.URL.Scheme = targetUrl.Scheme
	c.Request.Host = targetUrl.Host

	log.Printf("[DEBUG] Updated request path from '%s' → '%s'\n", originalPath, c.Request.URL.Path)
	log.Printf("[DEBUG] Updated request host from '%s' → '%s'\n", originalHost, c.Request.Host)
	log.Printf("[INFO] Forwarding request to Namenode leader: %s%s?%s\n", targetUrl.String(), c.Request.URL.Path, c.Request.URL.RawQuery)

	// 5. Optional: capture proxy errors
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("[ERROR] Proxy error while forwarding to leader '%s': %v\n", leaderAddr, err)
		http.Error(rw, "failed to contact namenode leader", http.StatusBadGateway)
	}

	// 6. Serve the proxied request
	log.Println("[INFO] Proxying request to Namenode leader...")
	proxy.ServeHTTP(c.Writer, c.Request)
	log.Println("[INFO] Request successfully proxied to Namenode leader")
}

