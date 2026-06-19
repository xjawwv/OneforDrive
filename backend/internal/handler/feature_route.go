package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FeatureRouteHandler struct {
	DB *sql.DB
}

type featureRoute struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Icon         string  `json:"icon"`
	Enabled      bool    `json:"enabled"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	DisplayOrder int     `json:"display_order"`
	ExemptRoleIDs []int64 `json:"exempt_role_ids"`
}

func (h *FeatureRouteHandler) ListRoutes(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, name, path, icon, enabled, description, category, display_order, exempt_role_ids FROM feature_routes ORDER BY display_order, id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list routes"})
		return
	}
	defer rows.Close()

	var routes []featureRoute
	for rows.Next() {
		var r featureRoute
		var exemptJSON sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &r.Path, &r.Icon, &r.Enabled, &r.Description, &r.Category, &r.DisplayOrder, &exemptJSON); err == nil {
			if exemptJSON.Valid && exemptJSON.String != "" {
				json.Unmarshal([]byte(exemptJSON.String), &r.ExemptRoleIDs)
			}
			if r.ExemptRoleIDs == nil {
				r.ExemptRoleIDs = []int64{}
			}
			routes = append(routes, r)
		}
	}
	if routes == nil {
		routes = []featureRoute{}
	}
	c.JSON(http.StatusOK, routes)
}

func (h *FeatureRouteHandler) GetRoute(c *gin.Context) {
	path := c.Param("path")
	if path[0] != '/' {
		path = "/" + path
	}
	var r featureRoute
	var exemptJSON sql.NullString
	err := h.DB.QueryRow(
		"SELECT id, name, path, icon, enabled, description, category, display_order, exempt_role_ids FROM feature_routes WHERE path = ?",
		path,
	).Scan(&r.ID, &r.Name, &r.Path, &r.Icon, &r.Enabled, &r.Description, &r.Category, &r.DisplayOrder, &exemptJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}
	if exemptJSON.Valid && exemptJSON.String != "" {
		json.Unmarshal([]byte(exemptJSON.String), &r.ExemptRoleIDs)
	}
	if r.ExemptRoleIDs == nil {
		r.ExemptRoleIDs = []int64{}
	}
	c.JSON(http.StatusOK, r)
}

func (h *FeatureRouteHandler) UpdateRoute(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Enabled     *bool   `json:"enabled"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
		DisplayOrder *int   `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		h.DB.Exec("UPDATE feature_routes SET name = ? WHERE id = ?", *req.Name, id)
	}
	if req.Enabled != nil {
		h.DB.Exec("UPDATE feature_routes SET enabled = ? WHERE id = ?", *req.Enabled, id)
	}
	if req.Description != nil {
		h.DB.Exec("UPDATE feature_routes SET description = ? WHERE id = ?", *req.Description, id)
	}
	if req.Icon != nil {
		h.DB.Exec("UPDATE feature_routes SET icon = ? WHERE id = ?", *req.Icon, id)
	}
	if req.DisplayOrder != nil {
		h.DB.Exec("UPDATE feature_routes SET display_order = ? WHERE id = ?", *req.DisplayOrder, id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "route updated"})
}

func (h *FeatureRouteHandler) SetExemptRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	var req struct {
		RoleIDs []int64 `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exemptJSON interface{}
	if len(req.RoleIDs) > 0 {
		data, _ := json.Marshal(req.RoleIDs)
		exemptJSON = string(data)
	}

	_, err = h.DB.Exec("UPDATE feature_routes SET exempt_role_ids = ? WHERE id = ?", exemptJSON, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exempt roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exempt roles updated"})
}
