package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/middleware"
	"github.com/routestorage/backend/internal/service"
)

type downloadChunkInfo struct {
	Index       int
	DriveFileID string
	AccountID   int64
	ChunkSize   int64
}

type downloadSession struct {
	ID          string
	FileID      int64
	FileName    string
	FileType    string
	TotalSize   int64
	Status      string
	Progress    int
	ChunksDone  int
	ChunksTotal int
	FilePath    string
	Error       string
	CreatedAt   time.Time
}

var downloadSessions = sync.Map{}

func generateSessionID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano()>>uint(i*4)) ^ byte(i*37)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:16])
}

func (h *FileHandler) DownloadByName(c *gin.Context) {
	userID := c.GetInt64("user_id")
	action := c.Query("action")

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	var f struct {
		ID        int64
		Name      string
		MimeType  string
		SizeTotal int64
	}
	err := h.DB.QueryRow(
		"SELECT id, name, mime_type, size_total FROM files WHERE name = ? AND user_id = ? AND is_folder = FALSE",
		req.Name, userID,
	).Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeTotal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT chunk_index, drive_file_id, account_id, chunk_size FROM file_chunks WHERE file_id = ? ORDER BY chunk_index ASC",
		f.ID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}
	defer rows.Close()

	var chunks []downloadChunkInfo
	for rows.Next() {
		var ci downloadChunkInfo
		if err := rows.Scan(&ci.Index, &ci.DriveFileID, &ci.AccountID, &ci.ChunkSize); err == nil {
			chunks = append(chunks, ci)
		}
	}

	if len(chunks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}

	// action=download → serve the assembled file directly (no async session)
	if action == "download" {
		sort.Slice(chunks, func(i, j int) bool {
			return chunks[i].Index < chunks[j].Index
		})

		type chunkData struct {
			data []byte
			err  error
			done chan struct{}
		}

		results := make([]chunkData, len(chunks))
		for i := range results {
			results[i].done = make(chan struct{})
		}

		for i, ch := range chunks {
			go func(idx int, ci downloadChunkInfo) {
				defer close(results[idx].done)

				accessToken, err := service.GetAccessTokenForAccount(h.DB, ci.AccountID)
				if err != nil {
					results[idx].err = fmt.Errorf("token error chunk %d: %w", ci.Index, err)
					return
				}

				driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", ci.DriveFileID)
				req, err := http.NewRequest("GET", driveURL, nil)
				if err != nil {
					results[idx].err = err
					return
				}
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					results[idx].err = fmt.Errorf("drive download failed chunk %d: %w", ci.Index, err)
					return
				}
				defer resp.Body.Close()

				data, err := io.ReadAll(resp.Body)
				results[idx].data = data
				results[idx].err = err
			}(i, ch)
		}

		for _, res := range results {
			<-res.done
			if res.err != nil {
				log.Printf("Chunk download error: %v", res.err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("chunk download failed: %v", res.err)})
				return
			}
		}

		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, f.Name))
		c.Header("Content-Type", f.MimeType)
		c.Status(http.StatusOK)

		for _, res := range results {
			c.Writer.Write(res.data)
		}
		return
	}

	// Default: start async download session
	sessionID := generateSessionID()
	tmpDir := filepath.Join(os.TempDir(), "routestorage_downloads")
	os.MkdirAll(tmpDir, 0755)
	tmpFile := filepath.Join(tmpDir, sessionID)

	sess := &downloadSession{
		ID:          sessionID,
		FileID:      f.ID,
		FileName:    f.Name,
		FileType:    f.MimeType,
		TotalSize:   f.SizeTotal,
		Status:      "downloading",
		ChunksTotal: len(chunks),
		FilePath:    tmpFile,
		CreatedAt:   time.Now(),
	}
	downloadSessions.Store(sessionID, sess)

	go h.processDownload(sess, chunks)

	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID,
		"file_name":  f.Name,
		"file_size":  f.SizeTotal,
		"chunks":     len(chunks),
	})
}

func (h *FileHandler) StartDownload(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")

	var f struct {
		ID        int64
		Name      string
		MimeType  string
		SizeTotal int64
	}
	err := h.DB.QueryRow(
		"SELECT id, name, mime_type, size_total FROM files WHERE id = ? AND user_id = ? AND is_folder = FALSE",
		id, userID,
	).Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeTotal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT chunk_index, drive_file_id, account_id, chunk_size FROM file_chunks WHERE file_id = ? ORDER BY chunk_index ASC",
		f.ID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}
	defer rows.Close()

	var chunks []downloadChunkInfo
	for rows.Next() {
		var c downloadChunkInfo
		if err := rows.Scan(&c.Index, &c.DriveFileID, &c.AccountID, &c.ChunkSize); err == nil {
			chunks = append(chunks, c)
		}
	}

	if len(chunks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}

	sessionID := generateSessionID()
	tmpDir := filepath.Join(os.TempDir(), "routestorage_downloads")
	os.MkdirAll(tmpDir, 0755)
	tmpFile := filepath.Join(tmpDir, sessionID)

	sess := &downloadSession{
		ID:          sessionID,
		FileID:      f.ID,
		FileName:    f.Name,
		FileType:    f.MimeType,
		TotalSize:   f.SizeTotal,
		Status:      "downloading",
		ChunksTotal: len(chunks),
		FilePath:    tmpFile,
		CreatedAt:   time.Now(),
	}
	downloadSessions.Store(sessionID, sess)

	go h.processDownload(sess, chunks)

	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID,
		"file_name":  f.Name,
		"file_size":  f.SizeTotal,
		"chunks":     len(chunks),
	})
}

func (h *FileHandler) processDownload(sess *downloadSession, chunks []downloadChunkInfo) {
	type chunkResult struct {
		Index int
		Data  []byte
		Error error
	}

	results := make([]chunkResult, len(chunks))
	var wg sync.WaitGroup

	for i, ch := range chunks {
		wg.Add(1)
		go func(idx int, ci downloadChunkInfo) {
			defer func() {
				sess.ChunksDone++
				sess.Progress = int(float64(sess.ChunksDone) / float64(sess.ChunksTotal) * 100)
				wg.Done()
			}()

			accessToken, err := service.GetAccessTokenForAccount(h.DB, ci.AccountID)
			if err != nil {
				results[idx] = chunkResult{Index: idx, Error: fmt.Errorf("token error: %w", err)}
				return
			}

			driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", ci.DriveFileID)
			req, err := http.NewRequest("GET", driveURL, nil)
			if err != nil {
				results[idx] = chunkResult{Index: idx, Error: err}
				return
			}
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results[idx] = chunkResult{Index: idx, Error: fmt.Errorf("drive download failed: %w", err)}
				return
			}
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			results[idx] = chunkResult{Index: idx, Data: data, Error: err}
		}(i, ch)
	}

	wg.Wait()

	failed := false
	for _, r := range results {
		if r.Error != nil {
			failed = true
			log.Printf("Download chunk %d failed: %v", r.Index, r.Error)
		}
	}

	if failed {
		sess.Status = "error"
		sess.Error = "one or more chunks failed to download"
		return
	}

	f, err := os.Create(sess.FilePath)
	if err != nil {
		sess.Status = "error"
		sess.Error = "failed to create temp file"
		return
	}
	defer f.Close()

	for _, r := range results {
		f.Write(r.Data)
	}

	sess.Progress = 100
	sess.ChunksDone = sess.ChunksTotal
	sess.Status = "ready"
}

func (h *FileHandler) CancelDownload(c *gin.Context) {
	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session required"})
		return
	}

	val, ok := downloadSessions.Load(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	sess := val.(*downloadSession)
	sess.Status = "cancelled"
	downloadSessions.Delete(sessionID)

	if sess.FilePath != "" {
		os.Remove(sess.FilePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}

func (h *FileHandler) Thumbnail(c *gin.Context) {
	var userID int64
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(int64)
	} else if tokenStr := c.Query("token"); tokenStr != "" {
		uid, err := middleware.ParseToken(tokenStr, []byte(getEnv("JWT_SECRET", "default-secret")))
		if err == nil {
			userID = uid
		}
	}
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var driveFileID string
	var accountID int64
	var mimeType string
	err := h.DB.QueryRow(
		"SELECT fc.drive_file_id, fc.account_id, f.mime_type FROM file_chunks fc JOIN files f ON fc.file_id = f.id WHERE fc.file_id = ? AND f.user_id = ? ORDER BY fc.chunk_index ASC LIMIT 1",
		id, userID,
	).Scan(&driveFileID, &accountID, &mimeType)
	if err != nil || driveFileID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	accessToken, err := service.GetAccessTokenForAccount(h.DB, accountID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "drive access unavailable"})
		return
	}

	driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", driveFileID)
	req, err := http.NewRequest("GET", driveURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch thumbnail"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

func (h *FileHandler) DownloadProgress(c *gin.Context) {
	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session required"})
		return
	}

	val, ok := downloadSessions.Load(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	sess := val.(*downloadSession)

	c.JSON(http.StatusOK, gin.H{
		"status":       sess.Status,
		"progress":     sess.Progress,
		"chunks_done":  sess.ChunksDone,
		"chunks_total": sess.ChunksTotal,
		"file_name":    sess.FileName,
		"file_size":    sess.TotalSize,
		"error":        sess.Error,
	})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")

	var f struct {
		ID        int64
		Name      string
		MimeType  string
		SizeTotal int64
	}
	err := h.DB.QueryRow(
		"SELECT id, name, mime_type, size_total FROM files WHERE id = ? AND user_id = ? AND is_folder = FALSE",
		id, userID,
	).Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeTotal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT chunk_index, drive_file_id, account_id, chunk_size FROM file_chunks WHERE file_id = ? ORDER BY chunk_index ASC",
		f.ID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}
	defer rows.Close()

	var chunks []downloadChunkInfo
	for rows.Next() {
		var ch downloadChunkInfo
		if err := rows.Scan(&ch.Index, &ch.DriveFileID, &ch.AccountID, &ch.ChunkSize); err == nil {
			chunks = append(chunks, ch)
		}
	}

	if len(chunks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no drive files associated"})
		return
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	type chunkData struct {
		data []byte
		err  error
		done chan struct{}
	}

	results := make([]chunkData, len(chunks))
	for i := range results {
		results[i].done = make(chan struct{})
	}

	for i, ch := range chunks {
		go func(idx int, ci downloadChunkInfo) {
			defer close(results[idx].done)

			accessToken, err := service.GetAccessTokenForAccount(h.DB, ci.AccountID)
			if err != nil {
				results[idx].err = fmt.Errorf("token error chunk %d: %w", ci.Index, err)
				return
			}

			driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", ci.DriveFileID)
			req, err := http.NewRequest("GET", driveURL, nil)
			if err != nil {
				results[idx].err = err
				return
			}
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results[idx].err = fmt.Errorf("drive download failed chunk %d: %w", ci.Index, err)
				return
			}
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			results[idx].data = data
			results[idx].err = err
		}(i, ch)
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, f.Name))
	c.Header("Content-Type", f.MimeType)
	c.Status(http.StatusOK)

	for _, res := range results {
		<-res.done
		if res.err != nil {
			log.Printf("Chunk download error: %v", res.err)
			continue
		}
		c.Writer.Write(res.data)
	}
}
