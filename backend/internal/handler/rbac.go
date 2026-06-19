package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/repository"
)

type RBACHandler struct {
	DB *sql.DB
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := repository.ListRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := repository.CreateRole(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "description": req.Description})
}

func (h *RBACHandler) ListPermissions(c *gin.Context) {
	perms, err := repository.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list permissions"})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	permIDs, err := repository.GetRolePermissions(roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission_ids": permIDs})
}

func (h *RBACHandler) SetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var req struct {
		PermissionIDs []int64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.SetRolePermissions(roleID, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permissions updated"})
}

func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	roles, err := repository.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) AssignRole(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		RoleID int64 `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.AssignRole(userID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
}

func (h *RBACHandler) RemoveRole(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	if err := repository.RemoveRole(userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role removed"})
}

func (h *RBACHandler) ListUsers(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, email, name FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	defer rows.Close()

	var users []gin.H
	for rows.Next() {
		var id int64
		var email, name string
		if err := rows.Scan(&id, &email, &name); err == nil {
			users = append(users, gin.H{"id": id, "email": email, "name": name})
		}
	}
	c.JSON(http.StatusOK, users)
}

func (h *RBACHandler) GetMyPermissions(c *gin.Context) {
	userID := c.GetInt64("user_id")
	perms, err := repository.GetUserPermissions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}

func (h *RBACHandler) GetMyRoles(c *gin.Context) {
	userID := c.GetInt64("user_id")
	roleIDs, err := repository.GetUserRoleIDs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role_ids": roleIDs})
}
