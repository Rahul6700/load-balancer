package core

import (
	"log"
	"net/http"
	"sync"
	"time"
)

var mutex sync.Mutex // mutex lock to lock the currentLeader var

var currentLeaderAddress string // hold the str url of the current leader

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
				lock.Lock()
				if currentLeaderAddress != peer {
					log.Printf("New leader discovered: %s\n", peer)
					currentLeaderAddress = peer
				}
				lock.Unlock()
				
				foundLeader = true
				break // Stop polling once we find the leader
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		if !foundLeader {
			log.Println("No Namenode leader found in this poll cycle.")
			lock.Lock()
			currentLeaderAddress = "" 
			lock.Unlock()
	}
}
}

// GetCurrentLeader is a thread-safe way to read the leader's address
func GetCurrentLeader() string {
	leaderLock.RLock()
	defer leaderLock.RUnlock()
	return currentLeaderAddress
}

