package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// ActivityEvent represents an activity log entry
type ActivityEvent struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	BookID      *uint     `json:"bookId,omitempty"`
	BookTitle   string    `json:"bookTitle,omitempty"`
	Status      string    `json:"status"` // success, warning, error
	Timestamp   time.Time `json:"timestamp"`
}

// getActivity returns recent activity
func (s *Server) getActivity(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get recent downloads
	var downloads []db.Download
	s.db.Order("added_at DESC").Limit(limit).Find(&downloads)

	activities := make([]ActivityEvent, 0)

	for _, d := range downloads {
		status := "success"
		if d.Status == "failed" {
			status = "error"
		} else if d.Status == "downloading" || d.Status == "queued" {
			status = "warning"
		}

		activities = append(activities, ActivityEvent{
			ID:        d.ID,
			Type:      "download",
			Title:     "Download " + d.Status,
			Message:   d.Title,
			BookID:    &d.BookID,
			Status:    status,
			Timestamp: time.Unix(d.AddedAt, 0),
		})
	}

	// Get recent imports (media files)
	var mediaFiles []db.MediaFile
	s.db.Preload("Book").Order("imported_at DESC").Limit(limit).Find(&mediaFiles)

	for _, mf := range mediaFiles {
		activities = append(activities, ActivityEvent{
			ID:        mf.ID,
			Type:      "import",
			Title:     "Imported",
			Message:   mf.FileName,
			BookID:    &mf.BookID,
			BookTitle: mf.Book.Title,
			Status:    "success",
			Timestamp: mf.ImportedAt,
		})
	}

	// Sort by timestamp
	// In production, this would be done in the database query

	return c.JSON(http.StatusOK, activities)
}

// getActivityHistory returns activity history with pagination
func (s *Server) getActivityHistory(c echo.Context) error {
	page := 1
	pageSize := 50

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	// Get downloads with pagination
	var downloads []db.Download
	var total int64
	
	s.db.Model(&db.Download{}).Count(&total)
	s.db.Order("added_at DESC").Offset(offset).Limit(pageSize).Find(&downloads)

	activities := make([]ActivityEvent, 0, len(downloads))
	for _, d := range downloads {
		status := "success"
		if d.Status == "failed" {
			status = "error"
		} else if d.Status == "downloading" || d.Status == "queued" {
			status = "warning"
		}

		activities = append(activities, ActivityEvent{
			ID:        d.ID,
			Type:      "download",
			Title:     "Download " + d.Status,
			Message:   d.Title,
			BookID:    &d.BookID,
			Status:    status,
			Timestamp: time.Unix(d.AddedAt, 0),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"activities": activities,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
	})
}

