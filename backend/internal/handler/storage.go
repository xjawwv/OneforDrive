package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	DB *sql.DB
}

func (h *StorageHandler) GetStorageStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var totalFiles, totalSize sql.NullInt64
	h.DB.QueryRow("SELECT COUNT(*) FROM files WHERE user_id = ? AND is_folder = FALSE", userID).Scan(&totalFiles)
	h.DB.QueryRow("SELECT COALESCE(SUM(size_total), 0) FROM files WHERE user_id = ? AND is_folder = FALSE", userID).Scan(&totalSize)

	var totalDriveAccounts, activeDriveAccounts sql.NullInt64
	h.DB.QueryRow("SELECT COUNT(*) FROM drive_accounts WHERE user_id = ?", userID).Scan(&totalDriveAccounts)
	h.DB.QueryRow("SELECT COUNT(*) FROM drive_accounts WHERE user_id = ? AND is_active = TRUE", userID).Scan(&activeDriveAccounts)

	var totalCapacity, totalUsed sql.NullInt64
	h.DB.QueryRow("SELECT COALESCE(SUM(capacity_total), 0), COALESCE(SUM(capacity_used), 0) FROM drive_accounts WHERE user_id = ? AND is_active = TRUE", userID).Scan(&totalCapacity, &totalUsed)

	c.JSON(http.StatusOK, gin.H{
		"total_files":          totalFiles.Int64,
		"total_size_bytes":     totalSize.Int64,
		"total_drive_accounts": totalDriveAccounts.Int64,
		"active_drive_accounts": activeDriveAccounts.Int64,
		"total_capacity_bytes": totalCapacity.Int64,
		"total_used_bytes":     totalUsed.Int64,
	})
}
