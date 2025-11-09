package core

import (
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	// The list of all your Namenodes
	NameNodeSlice = []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}

	// The lock that protects the leader address
	leaderLock sync.RWMutex
	
	// The variable to store the leader's address
	currentLeaderAddress string
)

// StartLeaderDiscovery is a background goroutine you'll start from main.go
func StartLeaderDiscovery() {
	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds

	for range ticker.C {
		foundLeader := false
		for _, peer := range NameNodeSlice {
			addr := peer + "/status"
			resp, err := http.Get(addr)
			
			// --- Use the 'http.StatusOK' constant ---
			if err == nil && resp.StatusCode == http.StatusOK {
				
				// --- We found the leader ---
				// Use the 'Write' lock
				leaderLock.Lock() 
				if currentLeaderAddress != peer {
					log.Printf("New leader discovered: %s\n", peer)
					currentLeaderAddress = peer
				}
				leaderLock.Unlock()
				
				foundLeader = true
				resp.Body.Close() // Close the body *inside* the success block
				break // Stop polling once we find the leader
			}
			
			if resp != nil {
				resp.Body.Close()
			}
		}
		
		if !foundLeader {
			log.Println("No Namenode leader found in this poll cycle.")
			// Use the 'Write' lock
			leaderLock.Lock()
			currentLeaderAddress = "" // No known leader
			leaderLock.Unlock()
		}
	}
}

// GetCurrentLeader is a thread-safe way to read the leader's address
func GetCurrentLeader() string {
	// Use the 'Read' lock
	leaderLock.RLock()
	defer leaderLock.RUnlock()
	return currentLeaderAddress
}
