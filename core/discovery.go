package core

import (
	"log"
	"net/http"
	"sync"
	"time"
)

var mutex sync.RWMutex // mutex lock to lock the currentLeader var

var currentLeaderAddress string // hold the str url of the current leader

// defining statically for now, maybe implement an endpoint to add dynamically later ?
var NameNodeSlice = []string{
	"http://localhost:8001",
	"http://localhost:8002",
	"http://localhost:8003",
}

// we need to start a background go-routines that keeps running
// and once every few seconds we poll all the namenodes to check which is the leader namenode -> store this in a variable
func LeaderDiscovery(){
	// Time.ticker is used when you want something to keep happening once every few seconds/mins
	// you initialize the ticker with the time interval
	ticker := time.NewTicker(1*time.Second) // here we're running for every 1 sec

	for range ticker.C {
		foundLeader := false
		for _, peer := range NameNodeSlice { // for every URL in the namenode Url slice
			addr := peer + "/status" // so it becomes "https://localhost:8080/status"
			resp, err := http.Get(addr) // sends a get req to the addr, gets back response and err
			if err == nil && resp.StatusCode == http.StatusOK { // we successfully found a leader
				mutex.Lock()
				if currentLeaderAddress != peer {
					log.Printf("New leader discovered: %s\n", peer)
					currentLeaderAddress = peer
				}
				mutex.Unlock()
				
				foundLeader = true
				break // Stop polling once we find the leader
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		if !foundLeader {
			log.Println("No Namenode leader found in this poll cycle.")
			mutex.Lock()
			currentLeaderAddress = "" 
			mutex.Unlock()
	}
}
}

// GetCurrentLeader is a thread-safe way to read the leader's address
func GetCurrentLeader() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return currentLeaderAddress
}

