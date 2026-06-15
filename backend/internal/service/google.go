package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"time"
)

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

type googleUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

type googleAboutResponse struct {
	StorageQuota struct {
		Limit string `json:"limit"`
		Usage string `json:"usage"`
	} `json:"storageQuota"`
}

func ExchangeCodeForToken(code string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", getEnv("GOOGLE_CLIENT_ID", ""))
	data.Set("client_secret", getEnv("GOOGLE_CLIENT_SECRET", ""))
	data.Set("redirect_uri", getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8081/api/accounts/oauth/callback"))
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange error (%d): %s", resp.StatusCode, string(body))
	}

	var tokenRes GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenRes, nil
}

func FetchGoogleUserInfo(accessToken string) (string, string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("userinfo error (%d): %s", resp.StatusCode, string(body))
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", err
	}

	return info.Email, info.Name, nil
}

func GetDriveCapacity(accessToken string) (int64, int64, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/drive/v3/about?fields=storageQuota", nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var about googleAboutResponse
	if err := json.Unmarshal(body, &about); err != nil {
		return 0, 0, err
	}

	limit, _ := strconv.ParseInt(about.StorageQuota.Limit, 10, 64)
	usage, _ := strconv.ParseInt(about.StorageQuota.Usage, 10, 64)

	return limit, usage, nil
}

func RefreshAccessToken(refreshToken string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", getEnv("GOOGLE_CLIENT_ID", ""))
	data.Set("client_secret", getEnv("GOOGLE_CLIENT_SECRET", ""))
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenRes GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, err
	}

	return &tokenRes, nil
}

func GetValidAccessToken(db *sql.DB, userID int64) (string, error) {
	var accessToken, refreshToken string
	var tokenExpiry *time.Time
	err := db.QueryRow(
		"SELECT access_token, refresh_token, token_expiry FROM drive_accounts WHERE user_id = ? AND is_active = TRUE LIMIT 1",
		userID,
	).Scan(&accessToken, &refreshToken, &tokenExpiry)
	if err != nil {
		return "", fmt.Errorf("no active drive account")
	}

	if tokenExpiry != nil && time.Now().Before(*tokenExpiry) {
		return accessToken, nil
	}

	newToken, err := RefreshAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("token refresh failed: %w", err)
	}

	var newExpiry *time.Time
	if newToken.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
		newExpiry = &t
	}

	db.Exec(
		"UPDATE drive_accounts SET access_token = ?, token_expiry = ? WHERE user_id = ? AND is_active = TRUE",
		newToken.AccessToken, newExpiry, userID,
	)

	return newToken.AccessToken, nil
}

type DriveAccountInfo struct {
	ID                    int64
	Email                 string
	Capacity              int64
	Used                  int64
	RouteStorageFolderID  string
}

func GetAllDriveAccounts(db *sql.DB, userID int64) ([]DriveAccountInfo, error) {
	rows, err := db.Query(
		"SELECT id, email, capacity_total, capacity_used, COALESCE(route_storage_folder_id, '') FROM drive_accounts WHERE user_id = ? AND is_active = TRUE",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []DriveAccountInfo
	for rows.Next() {
		var a DriveAccountInfo
		if err := rows.Scan(&a.ID, &a.Email, &a.Capacity, &a.Used, &a.RouteStorageFolderID); err == nil {
			accounts = append(accounts, a)
		}
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no active drive accounts")
	}
	return accounts, nil
}

func GetBestDriveAccount(db *sql.DB, userID int64, minBytes int64) (*DriveAccountInfo, error) {
	rows, err := db.Query(
		"SELECT id, email, capacity_total, capacity_used FROM drive_accounts WHERE user_id = ? AND is_active = TRUE ORDER BY (capacity_total - capacity_used) DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a DriveAccountInfo
		if err := rows.Scan(&a.ID, &a.Email, &a.Capacity, &a.Used); err != nil {
			continue
		}
		free := a.Capacity - a.Used
		if free >= minBytes || a.Capacity == 0 {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("no drive account with enough free space")
}

func GetAccessTokenForAccount(db *sql.DB, accountID int64) (string, error) {
	var accessToken, refreshToken string
	var tokenExpiry *time.Time
	err := db.QueryRow(
		"SELECT access_token, refresh_token, token_expiry FROM drive_accounts WHERE id = ? AND is_active = TRUE",
		accountID,
	).Scan(&accessToken, &refreshToken, &tokenExpiry)
	if err != nil {
		return "", err
	}

	if tokenExpiry != nil && time.Now().Before(*tokenExpiry) {
		return accessToken, nil
	}

	newToken, err := RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	var newExpiry *time.Time
	if newToken.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
		newExpiry = &t
	}

	db.Exec(
		"UPDATE drive_accounts SET access_token = ?, token_expiry = ? WHERE id = ?",
		newToken.AccessToken, newExpiry, accountID,
	)

	return newToken.AccessToken, nil
}

func UpdateAccountUsage(db *sql.DB, accountID int64, delta int64) {
	db.Exec(
		"UPDATE drive_accounts SET capacity_used = capacity_used + ? WHERE id = ? AND capacity_used + ? >= 0",
		delta, accountID, delta,
	)
}

func CreateRouteStorageFolder(accessToken string) (string, error) {
	existingID, err := findRouteStorageFolder(accessToken)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	metadata := map[string]interface{}{
		"name":     "RouteStorage",
		"mimeType": "application/vnd.google-apps.folder",
	}
	metaJSON, _ := json.Marshal(metadata)

	resp, err := http.PostForm("https://www.googleapis.com/drive/v3/files",
		url.Values{
			"uploadType": {"multipart"},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create folder request: %w", err)
	}
	defer resp.Body.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadataPart, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/json; charset=UTF-8"},
		"Content-Disposition": {"form-data; name=\"metadata\""},
	})
	metadataPart.Write(metaJSON)

	part, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/vnd.google-apps.folder"},
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"meta\""},
	})
	part.Write([]byte(" "))
	writer.Close()

	req, err := http.NewRequest("POST", "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp2.Body.Close()

	respBody, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create folder error (%d): %s", resp2.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	json.Unmarshal(respBody, &result)

	return result.ID, nil
}

func findRouteStorageFolder(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/drive/v3/files?q=name%3D%27RouteStorage%27+and+mimeType%3D%27application/vnd.google-apps.folder%27+and+trashed%3Dfalse&fields=files(id)", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	json.Unmarshal(body, &result)

	if len(result.Files) > 0 {
		return result.Files[0].ID, nil
	}
	return "", fmt.Errorf("not found")
}

func SyncOrphanedFiles(accessToken string, folderID string, knownDriveFileIDs map[string]bool) (int, error) {
	if folderID == "" {
		return 0, nil
	}

	var deleted int
	var pageToken string

	for {
		query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
		if pageToken != "" {
			query += fmt.Sprintf("&pageToken=%s", pageToken)
		}
		reqURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?q=%s&fields=nextPageToken,files(id,name)", url.QueryEscape(query))
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return deleted, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return deleted, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var result struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"files"`
		}
		json.Unmarshal(body, &result)

		for _, f := range result.Files {
			if !knownDriveFileIDs[f.ID] {
				delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", f.ID), nil)
				delReq.Header.Set("Authorization", "Bearer "+accessToken)
				delResp, err := http.DefaultClient.Do(delReq)
				if err == nil {
					delResp.Body.Close()
					if delResp.StatusCode == 204 || delResp.StatusCode == 200 {
						deleted++
					}
				}
			}
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return deleted, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
