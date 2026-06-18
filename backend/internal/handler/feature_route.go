package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FeatureRouteHandler struct {
	DB *sql.DB
}

type featureRoute struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Icon        string `json:"icon"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
	Category    string `json:"category"`
	DisplayOrder int   `json:"display_order"`
}

func (h *FeatureRouteHandler) ListRoutes(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, name, path, icon, enabled, description, category, display_order FROM feature_routes ORDER BY display_order, id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list routes"})
		return
	}
	defer rows.Close()

	var routes []featureRoute
	for rows.Next() {
		var r featureRoute
		if err := rows.Scan(&r.ID, &r.Name, &r.Path, &r.Icon, &r.Enabled, &r.Description, &r.Category, &r.DisplayOrder); err == nil {
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
	err := h.DB.QueryRow(
		"SELECT id, name, path, icon, enabled, description, category, display_order FROM feature_routes WHERE path = ?",
		path,
	).Scan(&r.ID, &r.Name, &r.Path, &r.Icon, &r.Enabled, &r.Description, &r.Category, &r.DisplayOrder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
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
