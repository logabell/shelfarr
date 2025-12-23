package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/indexer"
)

// IndexerRequest represents the request body for creating/updating an indexer
type IndexerRequest struct {
	Name          string `json:"name" validate:"required"`
	Type          string `json:"type" validate:"required"` // torznab, mam, anna
	URL           string `json:"url" validate:"required"`
	APIKey        string `json:"apiKey,omitempty"`
	Cookie        string `json:"cookie,omitempty"`
	Priority      int    `json:"priority"`
	Enabled       bool   `json:"enabled"`
	VIPOnly       bool   `json:"vipOnly,omitempty"`
	FreeleechOnly bool   `json:"freeleechOnly,omitempty"`
}

// IndexerResponse represents an indexer in API responses
type IndexerResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	Priority      int    `json:"priority"`
	Enabled       bool   `json:"enabled"`
	VIPOnly       bool   `json:"vipOnly,omitempty"`
	FreeleechOnly bool   `json:"freeleechOnly,omitempty"`
}

// getIndexers returns all configured indexers
func (s *Server) getIndexers(c echo.Context) error {
	var indexers []db.Indexer
	
	if err := s.db.Order("priority ASC").Find(&indexers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]IndexerResponse, len(indexers))
	for i, idx := range indexers {
		responses[i] = IndexerResponse{
			ID:            idx.ID,
			Name:          idx.Name,
			Type:          idx.Type,
			URL:           idx.URL,
			Priority:      idx.Priority,
			Enabled:       idx.Enabled,
			VIPOnly:       idx.VIPOnly,
			FreeleechOnly: idx.FreeleechOnly,
		}
	}

	return c.JSON(http.StatusOK, responses)
}

// addIndexer creates a new indexer
func (s *Server) addIndexer(c echo.Context) error {
	var req IndexerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Type == "" || req.URL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name, type, and url are required"})
	}

	// Validate type
	validTypes := map[string]bool{"torznab": true, "mam": true, "anna": true}
	if !validTypes[req.Type] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid indexer type. Must be: torznab, mam, or anna"})
	}

	indexer := db.Indexer{
		Name:          req.Name,
		Type:          req.Type,
		URL:           req.URL,
		APIKey:        req.APIKey,
		Cookie:        req.Cookie,
		Priority:      req.Priority,
		Enabled:       req.Enabled,
		VIPOnly:       req.VIPOnly,
		FreeleechOnly: req.FreeleechOnly,
	}

	if err := s.db.Create(&indexer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create indexer"})
	}

	return c.JSON(http.StatusCreated, IndexerResponse{
		ID:            indexer.ID,
		Name:          indexer.Name,
		Type:          indexer.Type,
		URL:           indexer.URL,
		Priority:      indexer.Priority,
		Enabled:       indexer.Enabled,
		VIPOnly:       indexer.VIPOnly,
		FreeleechOnly: indexer.FreeleechOnly,
	})
}

// updateIndexer updates an existing indexer
func (s *Server) updateIndexer(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid indexer ID"})
	}

	var indexer db.Indexer
	if err := s.db.First(&indexer, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Indexer not found"})
	}

	var req IndexerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	indexer.Name = req.Name
	indexer.Type = req.Type
	indexer.URL = req.URL
	indexer.APIKey = req.APIKey
	indexer.Cookie = req.Cookie
	indexer.Priority = req.Priority
	indexer.Enabled = req.Enabled
	indexer.VIPOnly = req.VIPOnly
	indexer.FreeleechOnly = req.FreeleechOnly

	if err := s.db.Save(&indexer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update indexer"})
	}

	return c.JSON(http.StatusOK, IndexerResponse{
		ID:            indexer.ID,
		Name:          indexer.Name,
		Type:          indexer.Type,
		URL:           indexer.URL,
		Priority:      indexer.Priority,
		Enabled:       indexer.Enabled,
		VIPOnly:       indexer.VIPOnly,
		FreeleechOnly: indexer.FreeleechOnly,
	})
}

// deleteIndexer removes an indexer
func (s *Server) deleteIndexer(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid indexer ID"})
	}

	var indexer db.Indexer
	if err := s.db.First(&indexer, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Indexer not found"})
	}

	if err := s.db.Delete(&indexer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete indexer"})
	}

	return c.NoContent(http.StatusNoContent)
}

// testIndexer tests connectivity to an indexer
func (s *Server) testIndexer(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid indexer ID"})
	}

	var dbIndexer db.Indexer
	if err := s.db.First(&dbIndexer, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Indexer not found"})
	}

	// Create the appropriate indexer instance
	var idx indexer.Indexer
	switch dbIndexer.Type {
	case "mam":
		idx = indexer.NewMAMIndexer(dbIndexer.Name, dbIndexer.Cookie, dbIndexer.VIPOnly, dbIndexer.FreeleechOnly)
	case "torznab":
		idx = indexer.NewTorznabIndexer(dbIndexer.Name, dbIndexer.URL, dbIndexer.APIKey)
	case "anna":
		idx = indexer.NewAnnaIndexer(dbIndexer.Name)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown indexer type"})
	}

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()

	if err := idx.Test(ctx); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": "Connection test failed: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connection test successful",
	})
}

