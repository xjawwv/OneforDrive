package handler

import (
	"net/http"
	"time"
)

// ChunkSemaphore limits concurrent Google Drive API calls (uploads + downloads).
// Prevents OOM when processing many files simultaneously.
var chunkSemaphore chan struct{}

// HTTPClient is a shared client with timeouts to prevent goroutine leaks.
var HTTPClient *http.Client

func InitChunkSemaphore(maxConcurrent int) {
	chunkSemaphore = make(chan struct{}, maxConcurrent)
	HTTPClient = &http.Client{
		Timeout: 10 * time.Minute, // max time for a single chunk transfer
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func acquireChunkSlot() {
	chunkSemaphore <- struct{}{}
}

func releaseChunkSlot() {
	<-chunkSemaphore
}
