package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/service"
)

func getChunkSizeMB() int64 {
	v := getEnv("CHUNK_SIZE_MB", "256")
	n, _ := strconv.ParseInt(v, 10, 64)
	if n <= 0 {
		n = 256
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	parentIDStr := c.PostForm("parent_id")
	var parentID *int64
	if parentIDStr != "" && parentIDStr != "null" {
		pid, _ := strconv.ParseInt(parentIDStr, 10, 64)
		parentID = &pid
	}

	tmpDir := filepath.Join(os.TempDir(), "routestorage_uploads")
	os.MkdirAll(tmpDir, 0755)
	tmpFile, err := os.CreateTemp(tmpDir, "upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	written, err := io.Copy(tmpFile, file)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save temp file"})
		return
	}

	result, err := h.DB.Exec(
		"INSERT INTO files (user_id, name, mime_type, size_total, status, upload_progress, parent_id, is_folder) VALUES (?, ?, ?, ?, 'uploading', 0, ?, FALSE)",
		userID, header.Filename, header.Header.Get("Content-Type"), written, parentID,
	)
	if err != nil {
		os.Remove(tmpFile.Name())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file record"})
		return
	}
	fileID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{
		"id":     fileID,
		"name":   header.Filename,
		"size":   written,
		"status": "uploading",
	})

	go h.processUpload(userID, fileID, header.Filename, tmpFile.Name(), written)
}

func (h *FileHandler) processUpload(userID, fileID int64, filename, tmpPath string, totalSize int64) {
	defer os.Remove(tmpPath)

	chunkSizeBytes := getChunkSizeMB() * 1024 * 1024
	numChunks := int((totalSize + chunkSizeBytes - 1) / chunkSizeBytes)
	if numChunks < 1 {
		numChunks = 1
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		h.DB.Exec("UPDATE files SET status = 'error', upload_progress = 0 WHERE id = ?", fileID)
		return
	}
	defer f.Close()

	results := make([]chunkResult, numChunks)
	var wg sync.WaitGroup

	for i := 0; i < numChunks; i++ {
		buf := make([]byte, chunkSizeBytes)
		n, readErr := f.Read(buf)
		if n == 0 {
			if readErr != nil && readErr != io.EOF {
				h.DB.Exec("UPDATE files SET status = 'error', upload_progress = 0 WHERE id = ?", fileID)
			}
			break
		}
		chunkData := buf[:n]
		wg.Add(1)
		go func(idx int, data []byte) {
			defer wg.Done()
			results[idx] = h.uploadChunk(userID, fileID, filename, data, idx)
		}(i, chunkData)

		progress := int(float64(i+1) / float64(numChunks) * 100)
		if progress > 99 {
			progress = 99
		}
		h.DB.Exec("UPDATE files SET upload_progress = ? WHERE id = ?", progress, fileID)
	}

	wg.Wait()

	failed := false
	for _, r := range results {
		if r.err != nil {
			failed = true
			log.Printf("Chunk %d upload failed for file %d: %v", r.index, fileID, r.err)
		}
	}

	if failed {
		for _, r := range results {
			if r.err == nil && r.driveFileID != "" {
				accessToken, err := service.GetAccessTokenForAccount(h.DB, r.accountID)
				if err == nil {
					deleteURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", r.driveFileID)
					req, _ := http.NewRequest("DELETE", deleteURL, nil)
					req.Header.Set("Authorization", "Bearer "+accessToken)
					http.DefaultClient.Do(req)
				}
			}
		}
		h.DB.Exec("DELETE FROM file_chunks WHERE file_id = ?", fileID)
		h.DB.Exec("UPDATE files SET status = 'error', upload_progress = 0 WHERE id = ?", fileID)
		return
	}

	h.DB.Exec("UPDATE files SET status = 'active', upload_progress = 100 WHERE id = ?", fileID)
}

func (h *FileHandler) uploadChunk(userID, fileID int64, filename string, data []byte, index int) chunkResult {
	account, err := service.GetBestDriveAccount(h.DB, userID, int64(len(data)))
	if err != nil {
		return chunkResult{index: index, err: err}
	}

	accessToken, err := service.GetAccessTokenForAccount(h.DB, account.ID)
	if err != nil {
		return chunkResult{index: index, err: fmt.Errorf("no access token for account %d", account.ID)}
	}

	chunkName := fmt.Sprintf("%s.part%d", filename, index)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadataJSON, _ := json.Marshal(map[string]string{
		"name":     chunkName,
		"mimeType": "application/octet-stream",
	})

	metadataPart, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/json; charset=UTF-8"},
		"Content-Disposition": {"form-data; name=\"metadata\""},
	})
	metadataPart.Write(metadataJSON)

	part, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/octet-stream"},
		"Content-Disposition": {fmt.Sprintf("form-data; name=\"file\"; filename=\"%s\"", chunkName)},
	})
	part.Write(data)
	writer.Close()

	uploadURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return chunkResult{index: index, err: err}
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return chunkResult{index: index, err: fmt.Errorf("drive upload failed: %w", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return chunkResult{index: index, err: fmt.Errorf("drive error %d: %s", resp.StatusCode, string(respBody))}
	}

	var driveResp struct {
		ID string `json:"id"`
	}
	json.Unmarshal(respBody, &driveResp)

	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	h.DB.Exec(
		"INSERT INTO file_chunks (file_id, chunk_index, chunk_size, drive_file_id, account_id, checksum) VALUES (?, ?, ?, ?, ?, ?)",
		fileID, index, int64(len(data)), driveResp.ID, account.ID, checksum,
	)

	service.UpdateAccountUsage(h.DB, account.ID, int64(len(data)))

	return chunkResult{
		index:       index,
		driveFileID: driveResp.ID,
		accountID:   account.ID,
		size:        int64(len(data)),
		checksum:    checksum,
	}
}

type chunkResult struct {
	index       int
	driveFileID string
	accountID   int64
	size        int64
	checksum    string
	err         error
}
