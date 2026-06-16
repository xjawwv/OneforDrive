package handler

import (
	"archive/zip"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/service"
)

type ShareHandler struct {
	DB *sql.DB
}

func generateShareToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *ShareHandler) CreateShareLink(c *gin.Context) {
	userID := c.GetInt64("user_id")
	fileIDStr := c.Param("id")
	fileID, _ := strconv.ParseInt(fileIDStr, 10, 64)

	var isFolder bool
	err := h.DB.QueryRow("SELECT is_folder FROM files WHERE id = ? AND user_id = ?", fileID, userID).Scan(&isFolder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var req struct {
		ExpiresIn string `json:"expires_in"`
	}
	c.ShouldBindJSON(&req)

	var expiresAt *time.Time
	switch req.ExpiresIn {
	case "1h":
		t := time.Now().Add(1 * time.Hour)
		expiresAt = &t
	case "24h":
		t := time.Now().Add(24 * time.Hour)
		expiresAt = &t
	case "7d":
		t := time.Now().Add(7 * 24 * time.Hour)
		expiresAt = &t
	case "30d":
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	default:
		expiresAt = nil
	}

	token := generateShareToken()
	result, err := h.DB.Exec(
		"INSERT INTO shared_links (file_id, token, expires_at, created_by) VALUES (?, ?, ?, ?)",
		fileID, token, expiresAt, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share link"})
		return
	}

	linkID, _ := result.LastInsertId()
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	c.JSON(http.StatusCreated, gin.H{
		"id":         linkID,
		"token":      token,
		"url":        fmt.Sprintf("%s/shared/%s", frontendURL, token),
		"expires_at": expiresAt,
		"is_folder":  isFolder,
	})
}

func (h *ShareHandler) GetShareLinks(c *gin.Context) {
	userID := c.GetInt64("user_id")
	fileIDStr := c.Param("id")
	fileID, _ := strconv.ParseInt(fileIDStr, 10, 64)

	rows, err := h.DB.Query(
		"SELECT id, token, expires_at, created_at FROM shared_links WHERE file_id = ? AND created_by = ? ORDER BY created_at DESC",
		fileID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query links"})
		return
	}
	defer rows.Close()

	type linkInfo struct {
		ID        int64      `json:"id"`
		Token     string     `json:"token"`
		URL       string     `json:"url"`
		ExpiresAt *time.Time `json:"expires_at"`
		CreatedAt time.Time  `json:"created_at"`
		IsValid   bool       `json:"is_valid"`
	}

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	var links []linkInfo
	for rows.Next() {
		var l linkInfo
		if err := rows.Scan(&l.ID, &l.Token, &l.ExpiresAt, &l.CreatedAt); err == nil {
			l.URL = fmt.Sprintf("%s/shared/%s", frontendURL, l.Token)
			l.IsValid = l.ExpiresAt == nil || l.ExpiresAt.After(time.Now())
			links = append(links, l)
		}
	}
	if links == nil {
		links = []linkInfo{}
	}

	c.JSON(http.StatusOK, links)
}

func (h *ShareHandler) RevokeShareLink(c *gin.Context) {
	userID := c.GetInt64("user_id")
	linkIDStr := c.Param("linkId")
	linkID, _ := strconv.ParseInt(linkIDStr, 10, 64)

	result, err := h.DB.Exec(
		"DELETE FROM shared_links WHERE id = ? AND created_by = ?",
		linkID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

func (h *ShareHandler) AccessShared(c *gin.Context) {
	token := c.Param("token")

	var fileID int64
	var expiresAt sql.NullTime
	err := h.DB.QueryRow(
		"SELECT file_id, expires_at FROM shared_links WHERE token = ?", token,
	).Scan(&fileID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share link not found or expired"})
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	var f struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		MimeType string `json:"mime_type"`
		Size     int64  `json:"size"`
		IsFolder bool   `json:"is_folder"`
	}
	err = h.DB.QueryRow(
		"SELECT id, name, mime_type, size_total, is_folder FROM files WHERE id = ?", fileID,
	).Scan(&f.ID, &f.Name, &f.MimeType, &f.Size, &f.IsFolder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var children []gin.H
	if f.IsFolder {
		rows, err := h.DB.Query(
			"SELECT id, name, mime_type, size_total, is_folder FROM files WHERE parent_id = ? ORDER BY is_folder DESC, name ASC",
			fileID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var child struct {
					ID       int64
					Name     string
					MimeType string
					Size     int64
					IsFolder bool
				}
				if err := rows.Scan(&child.ID, &child.Name, &child.MimeType, &child.Size, &child.IsFolder); err == nil {
					children = append(children, gin.H{
						"id":        child.ID,
						"name":      child.Name,
						"mime_type": child.MimeType,
						"size":      child.Size,
						"is_folder": child.IsFolder,
					})
				}
			}
		}
		if children == nil {
			children = []gin.H{}
		}
	}

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	c.JSON(http.StatusOK, gin.H{
		"file":       f,
		"children":   children,
		"token":      token,
		"expires_at": expiresAt.Time,
		"shared_url": fmt.Sprintf("%s/shared/%s", frontendURL, token),
	})
}

func (h *ShareHandler) SharedDownload(c *gin.Context) {
	token := c.Param("token")

	var fileID int64
	var expiresAt sql.NullTime
	err := h.DB.QueryRow(
		"SELECT file_id, expires_at FROM shared_links WHERE token = ?", token,
	).Scan(&fileID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	targetID := fileID
	if childIDStr := c.Query("child_id"); childIDStr != "" {
		if childID, parseErr := strconv.ParseInt(childIDStr, 10, 64); parseErr == nil {
			targetID = childID
		}
	}

	var f struct {
		Name     string
		MimeType string
		Size     int64
	}
	err = h.DB.QueryRow(
		"SELECT name, mime_type, size_total FROM files WHERE id = ? AND is_folder = FALSE", targetID,
	).Scan(&f.Name, &f.MimeType, &f.Size)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT chunk_index, drive_file_id, account_id, chunk_size FROM file_chunks WHERE file_id = ? ORDER BY chunk_index ASC",
		targetID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no chunks found"})
		return
	}
	defer rows.Close()

	type chunkInfo struct {
		Index       int
		DriveFileID string
		AccountID   int64
	}

	var chunks []chunkInfo
	for rows.Next() {
		var ch chunkInfo
		var chunkSize int64
		if err := rows.Scan(&ch.Index, &ch.DriveFileID, &ch.AccountID, &chunkSize); err == nil {
			chunks = append(chunks, ch)
		}
	}

	if len(chunks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no data"})
		return
	}

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
		go func(idx int, ci chunkInfo) {
			defer close(results[idx].done)
			accessToken, err := service.GetAccessTokenForAccount(h.DB, ci.AccountID)
			if err != nil {
				results[idx].err = err
				return
			}
			driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", ci.DriveFileID)
			req, _ := http.NewRequest("GET", driveURL, nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results[idx].err = err
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
	if f.Size > 0 {
		c.Header("Content-Length", fmt.Sprintf("%d", f.Size))
	}
	c.Status(http.StatusOK)

	for _, res := range results {
		<-res.done
		if res.err == nil {
			c.Writer.Write(res.data)
		}
	}
}

func (h *ShareHandler) SharedThumbnail(c *gin.Context) {
	token := c.Param("token")

	var fileID int64
	var expiresAt sql.NullTime
	err := h.DB.QueryRow(
		"SELECT file_id, expires_at FROM shared_links WHERE token = ?", token,
	).Scan(&fileID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "expired"})
		return
	}

	targetID := fileID
	if childIDStr := c.Query("child_id"); childIDStr != "" {
		if childID, parseErr := strconv.ParseInt(childIDStr, 10, 64); parseErr == nil {
			targetID = childID
		}
	}

	var driveFileID string
	var accountID int64
	var mimeType string
	err = h.DB.QueryRow(
		"SELECT fc.drive_file_id, fc.account_id, f.mime_type FROM file_chunks fc JOIN files f ON fc.file_id = f.id WHERE fc.file_id = ? ORDER BY fc.chunk_index ASC LIMIT 1",
		targetID,
	).Scan(&driveFileID, &accountID, &mimeType)
	if err != nil || driveFileID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	accessToken, err := service.GetAccessTokenForAccount(h.DB, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "drive access unavailable"})
		return
	}

	driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", driveFileID)
	req, _ := http.NewRequest("GET", driveURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

func (h *ShareHandler) SharedDownloadAll(c *gin.Context) {
	token := c.Param("token")

	var fileID int64
	var expiresAt sql.NullTime
	err := h.DB.QueryRow(
		"SELECT file_id, expires_at FROM shared_links WHERE token = ?", token,
	).Scan(&fileID, &expiresAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	var isFolder bool
	var folderName string
	h.DB.QueryRow("SELECT is_folder, name FROM files WHERE id = ?", fileID).Scan(&isFolder, &folderName)

	var fileIDs []int64
	if isFolder {
		rows, _ := h.DB.Query("SELECT id FROM files WHERE parent_id = ? AND is_folder = FALSE", fileID)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var id int64
				if rows.Scan(&id) == nil {
					fileIDs = append(fileIDs, id)
				}
			}
		}
	} else {
		fileIDs = append(fileIDs, fileID)
	}

	if len(fileIDs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no files to download"})
		return
	}

	zipName := folderName
	if zipName == "" {
		zipName = "download"
	}
	zipName += ".zip"

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, zipName))
	c.Status(http.StatusOK)

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	for _, fid := range fileIDs {
		var fname string
		h.DB.QueryRow("SELECT name FROM files WHERE id = ?", fid).Scan(&fname)
		if fname == "" {
			continue
		}

		chunkRows, _ := h.DB.Query(
			"SELECT chunk_index, drive_file_id, account_id FROM file_chunks WHERE file_id = ? ORDER BY chunk_index ASC", fid,
		)
		if chunkRows == nil {
			continue
		}

		partWriter, err := zipWriter.Create(fname)
		if err != nil {
			chunkRows.Close()
			continue
		}

		for chunkRows.Next() {
			var driveFileID string
			var accountID int64
			if chunkRows.Scan(new(int), &driveFileID, &accountID) != nil {
				continue
			}

			accessToken, err := service.GetAccessTokenForAccount(h.DB, accountID)
			if err != nil {
				continue
			}

			driveURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", driveFileID)
			req, _ := http.NewRequest("GET", driveURL, nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			io.Copy(partWriter, resp.Body)
			resp.Body.Close()
		}
		chunkRows.Close()
	}
}
