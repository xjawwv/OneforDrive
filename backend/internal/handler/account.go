package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/model"
	"github.com/routestorage/backend/internal/service"
	redispkg "github.com/routestorage/backend/pkg/redis"
)

type AccountHandler struct {
	DB *sql.DB
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	userID := c.GetInt64("user_id")
	rows, err := h.DB.Query(
		"SELECT id, user_id, email, capacity_total, capacity_used, is_active FROM drive_accounts WHERE user_id = ?",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query accounts"})
		return
	}
	defer rows.Close()

	var accounts []model.DriveAccount
	for rows.Next() {
		var a model.DriveAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Email, &a.CapacityTotal, &a.CapacityUsed, &a.IsActive); err != nil {
			continue
		}
		accounts = append(accounts, a)
	}
	if accounts == nil {
		accounts = []model.DriveAccount{}
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *AccountHandler) ConnectAccount(c *gin.Context) {
	clientID := GetEnv("GOOGLE_CLIENT_ID", "")
	redirectURL := GetEnv("GOOGLE_REDIRECT_URL", "http://localhost:8081/api/accounts/oauth/callback")

	if clientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google OAuth not configured"})
		return
	}

	userID := c.GetInt64("user_id")
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	ctx := context.Background()
	redispkg.Client.Set(ctx, "oauth_state:"+state, fmt.Sprintf("%d", userID), 10*time.Minute)

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("state", state)

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
	c.JSON(http.StatusOK, gin.H{"url": authURL})
}

func (h *AccountHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errParam := c.Query("error")
	frontendURL := GetEnv("FRONTEND_URL", "http://localhost:3000")

	if errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error="+errParam)
		return
	}

	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=no_code")
		return
	}

	ctx := context.Background()
	stateKey := "oauth_state:" + state
	val, err := redispkg.Client.Get(ctx, stateKey).Result()
	if err != nil || val == "" {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=invalid_state")
		return
	}
	redispkg.Client.Del(ctx, stateKey)

	stateUserID, _ := strconv.ParseInt(val, 10, 64)
	if stateUserID == 0 {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=invalid_user")
		return
	}

	tokenRes, err := service.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=token_exchange_failed")
		return
	}

	userEmail, _, err := service.FetchGoogleUserInfo(tokenRes.AccessToken)
	if err != nil {
		log.Printf("Failed to fetch user info: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=userinfo_failed")
		return
	}

	capacityTotal, capacityUsed, _ := service.GetDriveCapacity(tokenRes.AccessToken)

	routeStorageFolderID, err := service.CreateRouteStorageFolder(tokenRes.AccessToken)
	if err != nil {
		log.Printf("Failed to create RouteStorage folder: %v", err)
	}

	var expiryTime *time.Time
	if tokenRes.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tokenRes.ExpiresIn) * time.Second)
		expiryTime = &t
	}

	_, err = h.DB.Exec(
		`INSERT INTO drive_accounts (user_id, email, access_token, refresh_token, token_expiry, capacity_total, capacity_used, route_storage_folder_id, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, TRUE)`,
		stateUserID, userEmail, tokenRes.AccessToken, tokenRes.RefreshToken, expiryTime, capacityTotal, capacityUsed, routeStorageFolderID,
	)
	if err != nil {
		log.Printf("Failed to save drive account: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?error=save_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/settings?connected=1")
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id := c.Param("id")
	result, err := h.DB.Exec("DELETE FROM drive_accounts WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
