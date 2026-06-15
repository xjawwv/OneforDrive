package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	ID        int64
	Email     string
	Capacity  int64
	Used      int64
}

func GetAllDriveAccounts(db *sql.DB, userID int64) ([]DriveAccountInfo, error) {
	rows, err := db.Query(
		"SELECT id, email, capacity_total, capacity_used FROM drive_accounts WHERE user_id = ? AND is_active = TRUE",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []DriveAccountInfo
	for rows.Next() {
		var a DriveAccountInfo
		if err := rows.Scan(&a.ID, &a.Email, &a.Capacity, &a.Used); err == nil {
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
