package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/downloader"
	"github.com/shelfarr/shelfarr/internal/indexer"
)

// DownloadRequest represents a request to download a book
type DownloadRequest struct {
	BookID      uint   `json:"bookId"`
	IndexerName string `json:"indexer"`
	DownloadURL string `json:"downloadUrl"`
	Title       string `json:"title"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
	MediaType   string `json:"mediaType"` // ebook or audiobook
}

// DownloadResponse represents a download status response
type DownloadResponse struct {
	ID          uint    `json:"id"`
	BookID      uint    `json:"bookId"`
	Title       string  `json:"title"`
	MediaType   string  `json:"mediaType"` // ebook or audiobook
	Status      string  `json:"status"`
	Progress    float64 `json:"progress"`
	Size        int64   `json:"size"`
	Downloaded  int64   `json:"downloaded"`
	AddedAt     int64   `json:"addedAt"`
	CompletedAt int64   `json:"completedAt,omitempty"`
}

// triggerDownload initiates a download for a book
func (s *Server) triggerDownload(c echo.Context) error {
	var req DownloadRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[DEBUG] triggerDownload: invalid request body, error=%v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	log.Printf("[DEBUG] triggerDownload called: bookId=%d, indexer=%s, title=%s, url=%s", req.BookID, req.IndexerName, req.Title, req.DownloadURL)

	if req.BookID == 0 || req.DownloadURL == "" {
		log.Printf("[DEBUG] triggerDownload: missing required fields")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bookId and downloadUrl are required"})
	}

	// Verify book exists
	var book db.Book
	if err := s.db.First(&book, req.BookID).Error; err != nil {
		log.Printf("[DEBUG] triggerDownload: book not found with id=%d, error=%v", req.BookID, err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	log.Printf("[DEBUG] triggerDownload: found book '%s'", book.Title)

	// Get the first enabled download client
	var downloadClient db.DownloadClient
	if err := s.db.Where("enabled = ?", true).Order("priority ASC").First(&downloadClient).Error; err != nil {
		log.Printf("[DEBUG] triggerDownload: no enabled download client found, error=%v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No download client configured"})
	}

	log.Printf("[DEBUG] triggerDownload: using download client '%s' (type=%s, url=%s)", downloadClient.Name, downloadClient.Type, downloadClient.URL)

	// Default media type to ebook if not specified
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "ebook"
	}

	// Create download record
	download := db.Download{
		BookID:      req.BookID,
		ClientID:    downloadClient.ID,
		ClientType:  downloadClient.Type,
		MediaType:   mediaType,
		Title:       req.Title,
		DownloadURL: req.DownloadURL,
		Size:        req.Size,
		Status:      "queued",
		Category:    downloadClient.Category,
		AddedAt:     time.Now().Unix(),
	}

	// Create the download client
	client, err := downloader.CreateClientFromDB(downloadClient.Type, downloadClient.URL, downloadClient.Username, downloadClient.Password)
	if err != nil {
		log.Printf("[DEBUG] triggerDownload: failed to create download client, error=%v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create download client"})
	}

	// Add to download client
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	log.Printf("[DEBUG] triggerDownload: adding download to client with category=%s", downloadClient.Category)
	externalID, err := client.AddDownload(ctx, req.DownloadURL, &downloader.DownloadOptions{
		Category: downloadClient.Category,
	})
	if err != nil {
		log.Printf("[DEBUG] triggerDownload: failed to add download to client, error=%v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add download: " + err.Error()})
	}

	log.Printf("[DEBUG] triggerDownload: download added successfully, externalID=%s", externalID)

	download.ExternalID = externalID
	download.Status = "downloading"

	if err := s.db.Create(&download).Error; err != nil {
		log.Printf("[DEBUG] triggerDownload: failed to save download record, error=%v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save download"})
	}

	// Update book status
	s.db.Model(&book).Update("status", "downloading")

	log.Printf("[DEBUG] triggerDownload: download started successfully, downloadId=%d", download.ID)

	return c.JSON(http.StatusCreated, DownloadResponse{
		ID:        download.ID,
		BookID:    download.BookID,
		Title:     download.Title,
		MediaType: download.MediaType,
		Status:    download.Status,
		Progress:  0,
		Size:      download.Size,
		AddedAt:   download.AddedAt,
	})
}

// getDownloads returns all active downloads
func (s *Server) getDownloads(c echo.Context) error {
	var downloads []db.Download

	query := s.db.Order("added_at DESC")

	// Filter by status if provided
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Limit results
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	query = query.Limit(limit)

	if err := query.Find(&downloads).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]DownloadResponse, len(downloads))
	for i, d := range downloads {
		responses[i] = DownloadResponse{
			ID:          d.ID,
			BookID:      d.BookID,
			Title:       d.Title,
			MediaType:   d.MediaType,
			Status:      d.Status,
			Progress:    d.Progress,
			Size:        d.Size,
			Downloaded:  d.Downloaded,
			AddedAt:     d.AddedAt,
			CompletedAt: d.CompletedAt,
		}
	}

	return c.JSON(http.StatusOK, responses)
}

// getDownload returns a specific download
func (s *Server) getDownload(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid download ID"})
	}

	var download db.Download
	if err := s.db.First(&download, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Download not found"})
	}

	return c.JSON(http.StatusOK, DownloadResponse{
		ID:          download.ID,
		BookID:      download.BookID,
		Title:       download.Title,
		MediaType:   download.MediaType,
		Status:      download.Status,
		Progress:    download.Progress,
		Size:        download.Size,
		Downloaded:  download.Downloaded,
		AddedAt:     download.AddedAt,
		CompletedAt: download.CompletedAt,
	})
}

// deleteDownload removes a download
func (s *Server) deleteDownload(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid download ID"})
	}

	var download db.Download
	if err := s.db.First(&download, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Download not found"})
	}

	// Try to remove from download client
	if download.ExternalID != "" && download.ClientID > 0 {
		var downloadClient db.DownloadClient
		if s.db.First(&downloadClient, download.ClientID).Error == nil {
			client, _ := downloader.CreateClientFromDB(downloadClient.Type, downloadClient.URL, downloadClient.Username, downloadClient.Password)
			if client != nil {
				ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
				client.RemoveDownload(ctx, download.ExternalID, false)
				cancel()
			}
		}
	}

	// Delete from database
	if err := s.db.Delete(&download).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete download"})
	}

	return c.NoContent(http.StatusNoContent)
}

// automaticSearch performs an automatic search and download for a book
func (s *Server) automaticSearch(c echo.Context) error {
	bookID, err := strconv.ParseUint(c.Param("bookId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	// Get book details
	var book db.Book
	if err := s.db.Preload("Author").First(&book, bookID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	// Get enabled indexers
	var dbIndexers []db.Indexer
	if err := s.db.Where("enabled = ?", true).Order("priority ASC").Find(&dbIndexers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load indexers"})
	}

	if len(dbIndexers) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No indexers configured"})
	}

	// Create indexer manager and search
	manager := indexer.NewManager()
	for _, dbIdx := range dbIndexers {
		idx := createIndexerFromDB(dbIdx)
		if idx != nil {
			manager.AddIndexer(idx)
		}
	}

	searchQuery := indexer.SearchQuery{
		Title:     book.Title,
		Author:    book.Author.Name,
		ISBN:      book.ISBN,
		BookID:    book.HardcoverID,
		MediaType: c.QueryParam("mediaType"),
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	results, err := manager.SearchAll(ctx, searchQuery)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Search failed: " + err.Error()})
	}

	if len(results) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No results found"})
	}

	// Determine media type for scoring
	mediaType := c.QueryParam("mediaType")
	if mediaType == "" {
		mediaType = "ebook"
	}
	isAudiobook := mediaType == "audiobook"

	// Get quality profile for this media type
	var profile db.QualityProfile
	if err := s.db.Where("media_type = ? AND is_default = ?", mediaType, true).First(&profile).Error; err != nil {
		// If no default profile, try to get any profile for this media type
		if err := s.db.Where("media_type = ?", mediaType).First(&profile).Error; err != nil {
			// Fall back to simple scoring if no profiles configured
			profile = db.QualityProfile{
				FormatRanking: "epub,azw3,mobi,pdf",
				MinBitrate:    0,
			}
			if isAudiobook {
				profile.FormatRanking = "m4b,mp3"
			}
		}
	}

	// Select best result using quality profile scoring
	bestResult := indexer.GetBestResult(results, profile.FormatRanking, profile.MinBitrate, isAudiobook)
	if bestResult == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No suitable results found matching quality profile"})
	}

	// Get the first enabled download client
	var downloadClient db.DownloadClient
	if err := s.db.Where("enabled = ?", true).Order("priority ASC").First(&downloadClient).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No download client configured"})
	}

	// Create and initiate download
	client, err := downloader.CreateClientFromDB(downloadClient.Type, downloadClient.URL, downloadClient.Username, downloadClient.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create download client"})
	}

	externalID, err := client.AddDownload(ctx, bestResult.DownloadURL, &downloader.DownloadOptions{
		Category: downloadClient.Category,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add download: " + err.Error()})
	}

	// Save download record
	download := db.Download{
		BookID:      book.ID,
		ClientID:    downloadClient.ID,
		ClientType:  downloadClient.Type,
		ExternalID:  externalID,
		Title:       bestResult.Title,
		DownloadURL: bestResult.DownloadURL,
		Size:        bestResult.Size,
		Status:      "downloading",
		Category:    downloadClient.Category,
		AddedAt:     time.Now().Unix(),
	}

	if err := s.db.Create(&download).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save download"})
	}

	// Update book status
	s.db.Model(&book).Update("status", "downloading")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Download started",
		"downloadId": download.ID,
		"title":      bestResult.Title,
		"indexer":    bestResult.Indexer,
		"size":       bestResult.Size,
		"format":     bestResult.Format,
	})
}

// selectBestResult selects the best result based on quality criteria
func selectBestResult(results []indexer.SearchResult) *indexer.SearchResult {
	if len(results) == 0 {
		return nil
	}

	var best *indexer.SearchResult
	bestScore := -1

	for i := range results {
		r := &results[i]
		score := calculateScore(r)
		if score > bestScore {
			bestScore = score
			best = r
		}
	}

	return best
}

// calculateScore calculates a quality score for a search result
func calculateScore(r *indexer.SearchResult) int {
	score := 0

	// Prefer freeleech
	if r.Freeleech {
		score += 50
	}

	// Prefer more seeders
	score += r.Seeders * 2 // 2 points per seeder

	// Prefer certain formats
	formatScores := map[string]int{
		"EPUB": 20,
		"AZW3": 18,
		"M4B":  20,
		"MOBI": 15,
		"MP3":  10,
		"PDF":  5,
	}
	if fs, ok := formatScores[r.Format]; ok {
		score += fs
	}

	// Size consideration (prefer reasonable sizes)
	sizeMB := r.Size / (1024 * 1024)
	if sizeMB > 0 && sizeMB < 500 { // Reasonable ebook size
		score += 10
	} else if sizeMB >= 500 && sizeMB < 5000 { // Reasonable audiobook size
		score += 15
	}

	return score
}
