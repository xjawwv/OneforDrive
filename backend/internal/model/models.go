package model

import "time"

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Name         string `json:"name"`
	QuotaLimit   int64  `json:"quota_limit"`
	QuotaUsed    int64  `json:"quota_used"`
}

type DriveAccount struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Email         string     `json:"email"`
	AccessToken   string     `json:"-"`
	RefreshToken  string     `json:"-"`
	TokenExpiry   *time.Time `json:"token_expiry"`
	CapacityTotal int64      `json:"capacity_total"`
	CapacityUsed  int64      `json:"capacity_used"`
	IsActive      bool       `json:"is_active"`
}

type FileEntry struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	SizeTotal int64     `json:"size_total"`
	Status    string    `json:"status"`
	ParentID  *int64    `json:"parent_id"`
	IsFolder  bool      `json:"is_folder"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileChunk struct {
	ID          int64  `json:"id"`
	FileID      int64  `json:"file_id"`
	ChunkIndex  int    `json:"chunk_index"`
	ChunkSize   int64  `json:"chunk_size"`
	DriveFileID string `json:"drive_file_id"`
	AccountID   int64  `json:"account_id"`
	Checksum    string `json:"checksum"`
}
