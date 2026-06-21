package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/model"
	"github.com/routestorage/backend/internal/service"
)

type FileHandler struct {
	DB *sql.DB
}

func (h *FileHandler) GetFiles(c *gin.Context) {
	userID := c.GetInt64("user_id")
	parentID := c.Query("parent_id")

	var rows *sql.Rows
	var err error
	if parentID == "" || parentID == "null" {
		rows, err = h.DB.Query(
			"SELECT id, user_id, name, mime_type, size_total, status, parent_id, is_folder, updated_at FROM files WHERE user_id = ? AND (parent_id IS NULL) ORDER BY is_folder DESC, name ASC",
			userID,
		)
	} else {
		rows, err = h.DB.Query(
			"SELECT id, user_id, name, mime_type, size_total, status, parent_id, is_folder, updated_at FROM files WHERE user_id = ? AND parent_id = ? ORDER BY is_folder DESC, name ASC",
			userID, parentID,
		)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query files"})
		return
	}
	defer rows.Close()

	var files []model.FileEntry
	for rows.Next() {
		var f model.FileEntry
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.MimeType, &f.SizeTotal, &f.Status, &f.ParentID, &f.IsFolder, &f.UpdatedAt); err != nil {
			continue
		}
		files = append(files, f)
	}
	if files == nil {
		files = []model.FileEntry{}
	}
	c.JSON(http.StatusOK, files)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")

	var isFolder bool
	err := h.DB.QueryRow("SELECT is_folder FROM files WHERE id = ? AND user_id = ?", id, userID).Scan(&isFolder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if !isFolder {
		chunkRows, err := h.DB.Query(
			"SELECT fc.drive_file_id, fc.account_id FROM file_chunks fc WHERE fc.file_id = ?",
			id,
		)
		if err == nil {
			defer chunkRows.Close()
			for chunkRows.Next() {
				var driveFileID string
				var accountID int64
				if err := chunkRows.Scan(&driveFileID, &accountID); err != nil {
					continue
				}
				if driveFileID == "" {
					continue
				}
				accessToken, err := service.GetAccessTokenForAccount(h.DB, accountID)
				if err != nil {
					continue
				}
				deleteURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", driveFileID)
				req, _ := http.NewRequest("DELETE", deleteURL, nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				http.DefaultClient.Do(req)
			}
		}
	}

	h.DB.Exec("DELETE FROM file_chunks WHERE file_id IN (SELECT id FROM files WHERE (id = ? OR parent_id = ?) AND user_id = ?)", id, id, userID)
	result, err := h.DB.Exec("DELETE FROM files WHERE (id = ? OR parent_id = ?) AND user_id = ?", id, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *FileHandler) FileInfo(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")

	var f model.FileEntry
	err := h.DB.QueryRow(
		"SELECT id, user_id, name, mime_type, size_total, status, parent_id, is_folder, updated_at FROM files WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&f.ID, &f.UserID, &f.Name, &f.MimeType, &f.SizeTotal, &f.Status, &f.ParentID, &f.IsFolder, &f.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	chunks, _ := h.DB.Query("SELECT id, chunk_index, chunk_size, account_id, checksum FROM file_chunks WHERE file_id = ?", f.ID)
	var chunkList []model.FileChunk
	if chunks != nil {
		defer chunks.Close()
		for chunks.Next() {
			var ch model.FileChunk
			if err := chunks.Scan(&ch.ID, &ch.ChunkIndex, &ch.ChunkSize, &ch.AccountID, &ch.Checksum); err == nil {
				chunkList = append(chunkList, ch)
			}
		}
	}
	if chunkList == nil {
		chunkList = []model.FileChunk{}
	}

	c.JSON(http.StatusOK, gin.H{
		"file":   f,
		"chunks": chunkList,
	})
}

func (h *FileHandler) UploadProgress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")

	var status string
	err := h.DB.QueryRow(
		"SELECT status FROM files WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

func (h *FileHandler) GetBreadcrumb(c *gin.Context) {
	userID := c.GetInt64("user_id")
	folderID := c.Query("folder_id")
	if folderID == "" || folderID == "null" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	type crumb struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	var crumbs []crumb
	currentID := folderID

	for currentID != "" && currentID != "null" {
		var id int64
		var name string
		var parentID sql.NullInt64
		err := h.DB.QueryRow(
			"SELECT id, name, parent_id FROM files WHERE id = ? AND user_id = ? AND is_folder = TRUE",
			currentID, userID,
		).Scan(&id, &name, &parentID)
		if err != nil {
			break
		}
		crumbs = append([]crumb{{ID: id, Name: name}}, crumbs...)
		if parentID.Valid {
			currentID = strconv.FormatInt(parentID.Int64, 10)
		} else {
			break
		}
	}

	c.JSON(http.StatusOK, crumbs)
}

func (h *FileHandler) CreateFolder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.DB.Exec(
		"INSERT INTO files (user_id, name, mime_type, size_total, status, parent_id, is_folder) VALUES (?, ?, 'folder', 0, 'active', ?, TRUE)",
		userID, req.Name, req.ParentID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "is_folder": true})
}
