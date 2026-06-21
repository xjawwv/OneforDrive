package handler

import (
	"net/http"
	"sync"
	"time"
)

// fileSem limits how many files are processed simultaneously (uploads + downloads).
// Prevents overwhelming MySQL and Google Drive token refresh.
var fileSem chan struct{}

// chunkSem limits concurrent chunk transfers within a single file.
var chunkSem chan struct{}

// HTTPClient is a shared client with timeouts to prevent goroutine leaks.
var HTTPClient *http.Client

// uploadQueue tracks pending uploads so the frontend can see queue position.
var uploadQueue = struct {
	sync.Mutex
	items []uploadQueueItem
}{}

type uploadQueueItem struct {
	FileID int64
	UserID int64
	Name   string
}

// InitConcurrency initializes all semaphores and the HTTP client.
// maxFiles = max simultaneous file operations (uploads + downloads).
// maxChunks = max concurrent chunk transfers per file.
func InitConcurrency(maxFiles, maxChunks int) {
	fileSem = make(chan struct{}, maxFiles)
	chunkSem = make(chan struct{}, maxChunks)
	HTTPClient = &http.Client{
		Timeout: 10 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func acquireFileSlot()   { fileSem <- struct{}{} }
func releaseFileSlot()   { <-fileSem }
func acquireChunkSlot()  { chunkSem <- struct{}{} }
func releaseChunkSlot()  { <-chunkSem }
