package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/handler"
	"github.com/routestorage/backend/internal/middleware"
	"github.com/routestorage/backend/internal/repository"
	redispkg "github.com/routestorage/backend/pkg/redis"
)

func main() {
	jwtSecret := []byte(getEnv("JWT_SECRET", "default-secret"))

	repository.InitDB()
	redispkg.InitRedis()

	// Global concurrency limiter — max 10 concurrent chunk operations (uploads/downloads)
	handler.InitChunkSemaphore(10)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	authH := &handler.AuthHandler{DB: repository.DB, JWTSecret: jwtSecret}
	accountH := &handler.AccountHandler{DB: repository.DB}
	fileH := &handler.FileHandler{DB: repository.DB}
	storageH := &handler.StorageHandler{DB: repository.DB}
	shareH := &handler.ShareHandler{DB: repository.DB}
	rbacH := &handler.RBACHandler{DB: repository.DB}
	featureRouteH := &handler.FeatureRouteHandler{DB: repository.DB}

	rbac := func(perm string) gin.HandlerFunc {
		return middleware.RequirePermission(perm)
	}

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	storage := r.Group("/api/storage")
	storage.Use(middleware.AuthMiddleware(jwtSecret))
	{
		storage.GET("/stats", storageH.GetStorageStats)
	}

	r.GET("/api/routes", featureRouteH.ListRoutes)
	r.GET("/api/routes/:path", featureRouteH.GetRoute)
	r.PUT("/api/routes/:id", middleware.AuthMiddleware(jwtSecret), rbac("users.manage"), featureRouteH.UpdateRoute)
	r.PUT("/api/routes/:id/exempt-roles", middleware.AuthMiddleware(jwtSecret), rbac("users.manage"), featureRouteH.SetExemptRoles)

	r.GET("/api/accounts/oauth/callback", accountH.OAuthCallback)

	accounts := r.Group("/api/accounts")
	accounts.Use(middleware.AuthMiddleware(jwtSecret))
	{
		accounts.GET("/connect", accountH.ConnectAccount)
		accounts.GET("", accountH.GetAccounts)
		accounts.POST("/:id/sync", accountH.SyncDrive)
		accounts.DELETE("/:id", accountH.DeleteAccount)
	}

	files := r.Group("/api/files")
	files.Use(middleware.AuthMiddleware(jwtSecret))
	{
		files.GET("/breadcrumb", fileH.GetBreadcrumb)
		files.GET("", fileH.GetFiles)
		files.POST("/upload", fileH.UploadFile)
		files.POST("/folder", fileH.CreateFolder)
		files.GET("/:id/download", fileH.DownloadFile)
		files.POST("/:id/download", fileH.StartDownload)
		files.POST("/download-by-name", fileH.DownloadByName)
		files.GET("/:id/download-progress", fileH.DownloadProgress)
		files.DELETE("/download-cancel", fileH.CancelDownload)
		files.GET("/:id/info", fileH.FileInfo)
		files.GET("/:id/progress", fileH.UploadProgress)
		files.POST("/:id/share", shareH.CreateShareLink)
		files.GET("/:id/shares", shareH.GetShareLinks)
		files.DELETE("/:id/share/:linkId", shareH.RevokeShareLink)
		files.DELETE("/:id", fileH.DeleteFile)
	}

	r.GET("/api/files/:id/thumbnail", fileH.Thumbnail)

	r.GET("/shared/:token", shareH.AccessShared)
	r.GET("/shared/:token/download", shareH.SharedDownload)
	r.GET("/shared/:token/download-all", shareH.SharedDownloadAll)
	r.GET("/shared/:token/thumbnail", shareH.SharedThumbnail)

	rbacRoutes := r.Group("/api/rbac")
	rbacRoutes.Use(middleware.AuthMiddleware(jwtSecret))
	{
		rbacRoutes.GET("/me/permissions", rbacH.GetMyPermissions)
		rbacRoutes.GET("/me/roles", rbacH.GetMyRoles)
		rbacRoutes.GET("/users", rbac("users.manage"), rbacH.ListUsers)
		rbacRoutes.GET("/roles", rbac("users.manage"), rbacH.ListRoles)
		rbacRoutes.POST("/roles", rbac("users.manage"), rbacH.CreateRole)
		rbacRoutes.GET("/permissions", rbac("users.manage"), rbacH.ListPermissions)
		rbacRoutes.GET("/roles/:id/permissions", rbacH.GetRolePermissions)
		rbacRoutes.PUT("/roles/:id/permissions", rbac("users.manage"), rbacH.SetRolePermissions)
		rbacRoutes.GET("/users/:id/roles", rbacH.GetUserRoles)
		rbacRoutes.POST("/users/:id/roles", rbac("users.manage"), rbacH.AssignRole)
		rbacRoutes.DELETE("/users/:id/roles/:role_id", rbac("users.manage"), rbacH.RemoveRole)
	}

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // no timeout — uploads/downloads stream indefinitely
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Graceful shutdown — wait for in-flight requests to finish
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
