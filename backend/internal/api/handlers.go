package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/media"
)

// ========================
// Download Client Handlers
// ========================

// getDownloadClients returns all configured download clients
// DownloadClientResponse is the API response format for download clients
type DownloadClientResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Category string `json:"category"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
	Settings string `json:"settings"`
}

func toDownloadClientResponse(c db.DownloadClient) DownloadClientResponse {
	return DownloadClientResponse{
		ID:       c.ID,
		Name:     c.Name,
		Type:     c.Type,
		URL:      c.URL,
		Username: c.Username,
		Password: c.Password,
		Category: c.Category,
		Priority: c.Priority,
		Enabled:  c.Enabled,
		Settings: c.Settings,
	}
}

func (s *Server) getDownloadClients(c echo.Context) error {
	var clients []db.DownloadClient
	if err := s.db.Order("priority ASC").Find(&clients).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert to response format with lowercase JSON keys
	response := make([]DownloadClientResponse, len(clients))
	for i, client := range clients {
		response[i] = toDownloadClientResponse(client)
	}

	return c.JSON(http.StatusOK, response)
}

// addDownloadClient creates a new download client
func (s *Server) addDownloadClient(c echo.Context) error {
	var req DownloadClientResponse
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	client := db.DownloadClient{
		Name:     req.Name,
		Type:     req.Type,
		URL:      req.URL,
		Username: req.Username,
		Password: req.Password,
		Category: req.Category,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Settings: req.Settings,
	}

	if err := s.db.Create(&client).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create download client"})
	}

	return c.JSON(http.StatusCreated, toDownloadClientResponse(client))
}

// updateDownloadClient updates an existing download client
func (s *Server) updateDownloadClient(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var client db.DownloadClient
	if err := s.db.First(&client, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Download client not found"})
	}

	var req DownloadClientResponse
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Update fields
	client.Name = req.Name
	client.Type = req.Type
	client.URL = req.URL
	client.Username = req.Username
	client.Password = req.Password
	client.Category = req.Category
	client.Priority = req.Priority
	client.Enabled = req.Enabled
	client.Settings = req.Settings

	if err := s.db.Save(&client).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update download client"})
	}

	return c.JSON(http.StatusOK, toDownloadClientResponse(client))
}

// deleteDownloadClient removes a download client
func (s *Server) deleteDownloadClient(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := s.db.Delete(&db.DownloadClient{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete download client"})
	}

	return c.NoContent(http.StatusNoContent)
}

// testDownloadClient tests connectivity to a saved download client
func (s *Server) testDownloadClient(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var client db.DownloadClient
	if err := s.db.First(&client, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Download client not found"})
	}

	testErr := testDownloadClientConnection(client.Type, client.URL, client.Username, client.Password)
	if testErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": testErr.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Connection successful"})
}

// testDownloadClientConfig tests connectivity without saving the client first
func (s *Server) testDownloadClientConfig(c echo.Context) error {
	var req struct {
		Type     string `json:"type"`
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.URL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "URL is required"})
	}

	testErr := testDownloadClientConnection(req.Type, req.URL, req.Username, req.Password)
	if testErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": testErr.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Connection successful"})
}

// testDownloadClientConnection tests connection to a download client
func testDownloadClientConnection(clientType, url, username, password string) error {
	switch clientType {
	case "qbittorrent":
		return testQBittorrent(url, username, password)
	case "transmission":
		return testTransmission(url, username, password)
	case "deluge":
		return testDeluge(url, password)
	case "rtorrent":
		return testRTorrent(url, username, password)
	case "sabnzbd":
		return testSABnzbd(url, password)
	case "nzbget":
		return testNZBGet(url, username, password)
	default:
		return fmt.Errorf("unknown client type: %s", clientType)
	}
}

// getInsecureHTTPClient returns an HTTP client that skips TLS verification
// This is needed for seedboxes and self-signed certificates
func getInsecureHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}
}

// testQBittorrent tests qBittorrent connection
func testQBittorrent(url, username, password string) error {
	// Login to qBittorrent WebUI
	loginURL := strings.TrimRight(url, "/") + "/api/v2/auth/login"
	data := strings.NewReader("username=" + username + "&password=" + password)

	req, _ := http.NewRequest("POST", loginURL, data)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed - check username and password")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

// testTransmission tests Transmission RPC connection
func testTransmission(url, username, password string) error {
	rpcURL := strings.TrimRight(url, "/") + "/transmission/rpc"

	req, _ := http.NewRequest("POST", rpcURL, strings.NewReader(`{"method":"session-get"}`))
	req.Header.Set("Content-Type", "application/json")
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	// Transmission may return 409 first time (CSRF token), that's OK
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed - check username and password")
	}
	return fmt.Errorf("server returned status %d", resp.StatusCode)
}

// testDeluge tests Deluge Web UI connection
func testDeluge(url, password string) error {
	rpcURL := strings.TrimRight(url, "/") + "/json"

	// First, authenticate with Deluge
	authPayload := `{"method":"auth.login","params":["` + password + `"],"id":1}`
	req, err := http.NewRequest("POST", rpcURL, strings.NewReader(authPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Parse the response to check if auth succeeded
	// Deluge returns {"result": true/false, "error": null, "id": 1}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Result bool        `json:"result"`
		Error  interface{} `json:"error"`
		ID     int         `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("invalid response from Deluge: %v", err)
	}

	if result.Error != nil {
		return fmt.Errorf("Deluge error: %v", result.Error)
	}

	if !result.Result {
		return fmt.Errorf("authentication failed - check password")
	}

	return nil
}

// testSABnzbd tests SABnzbd API connection
func testSABnzbd(url, apiKey string) error {
	testURL := strings.TrimRight(url, "/") + "/api?mode=version&apikey=" + apiKey + "&output=json"

	req, _ := http.NewRequest("GET", testURL, nil)
	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed - check API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

// testNZBGet tests NZBGet XML-RPC connection
func testNZBGet(url, username, password string) error {
	rpcURL := strings.TrimRight(url, "/") + "/jsonrpc"

	req, _ := http.NewRequest("POST", rpcURL, strings.NewReader(`{"method":"version","params":[]}`))
	req.Header.Set("Content-Type", "application/json")
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed - check username and password")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

// testRTorrent tests ruTorrent/rTorrent XML-RPC connection
func testRTorrent(url, username, password string) error {
	// ruTorrent uses XML-RPC, typically at /plugins/rpc/rpc.php
	// Send a simple system.listMethods call to verify connection
	xmlPayload := `<?xml version="1.0"?><methodCall><methodName>system.listMethods</methodName></methodCall>`

	req, _ := http.NewRequest("POST", url, strings.NewReader(xmlPayload))
	req.Header.Set("Content-Type", "text/xml")
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := getInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed - check username and password")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

// deleteMediaFile deletes a media file
func (s *Server) deleteMediaFile(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	// Soft delete - moves to recycle bin
	if err := s.db.Delete(&db.MediaFile{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media file"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ========================
// Media File Handlers
// ========================

// getMediaFiles returns media files with optional filtering
func (s *Server) getMediaFiles(c echo.Context) error {
	var files []db.MediaFile

	query := s.db.Preload("Book")

	if bookID := c.QueryParam("bookId"); bookID != "" {
		query = query.Where("book_id = ?", bookID)
	}
	if mediaType := c.QueryParam("mediaType"); mediaType != "" {
		query = query.Where("media_type = ?", mediaType)
	}

	if err := query.Find(&files).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, files)
}

// streamMediaFile serves a media file for reading/listening
func (s *Server) streamMediaFile(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var file db.MediaFile
	if err := s.db.First(&file, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Media file not found"})
	}

	// Serve the file
	return c.File(file.FilePath)
}

// sendToKindle sends a book to a Kindle device via email
func (s *Server) sendToKindle(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var file db.MediaFile
	if err := s.db.First(&file, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Media file not found"})
	}

	// TODO: Implement SMTP sending logic
	// 1. Check format - convert if needed (EPUB -> MOBI)
	// 2. Get user's Kindle email from settings
	// 3. Send via SMTP

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Send to Kindle queued (not yet implemented)",
	})
}

// ========================
// Import Handlers
// ========================

// getPendingImports returns files in the downloads folder awaiting import
func (s *Server) getPendingImports(c echo.Context) error {
	importer := media.NewImporter(s.db, s.config.BooksPath, s.config.AudiobooksPath, media.OpHardlink)

	pending, err := importer.ScanDownloadsFolder(s.config.DownloadsPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, pending)
}

// manualImport manually maps a file to a book
func (s *Server) manualImport(c echo.Context) error {
	var req struct {
		FilePath    string `json:"filePath"`
		BookID      uint   `json:"bookId"`
		MediaType   string `json:"mediaType"`
		EditionName string `json:"editionName,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Get book details for path building
	var book db.Book
	if err := s.db.Preload("Author").Preload("Series").First(&book, req.BookID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	// Determine format from file extension
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(req.FilePath)), ".")
	if format == "" {
		format = "unknown"
	}

	// Build import request
	importReq := media.ImportRequest{
		SourcePath:  req.FilePath,
		BookID:      req.BookID,
		AuthorName:  book.Author.Name,
		BookTitle:   book.Title,
		MediaType:   req.MediaType,
		Format:      format,
		EditionName: req.EditionName,
	}

	// Add series info if available
	if book.Series != nil {
		importReq.SeriesName = book.Series.Name
		if book.SeriesIndex != nil {
			importReq.SeriesIndex = int(*book.SeriesIndex)
		}
	}

	// Perform import
	importer := media.NewImporter(s.db, s.config.BooksPath, s.config.AudiobooksPath, media.OpHardlink)
	result, err := importer.Import(importReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":     result.Success,
		"newPath":     result.NewPath,
		"mediaFileId": result.MediaFileID,
	})
}

// ========================
// User Handlers
// ========================

// getUsers returns all users (admin only)
func (s *Server) getUsers(c echo.Context) error {
	var users []db.User
	if err := s.db.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Remove password hashes from response
	for i := range users {
		users[i].PasswordHash = ""
	}

	return c.JSON(http.StatusOK, users)
}

// createUser creates a new user
func (s *Server) createUser(c echo.Context) error {
	var user db.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// TODO: Hash password before storing

	if err := s.db.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	user.PasswordHash = ""
	return c.JSON(http.StatusCreated, user)
}

// getCurrentUser returns the currently authenticated user
func (s *Server) getCurrentUser(c echo.Context) error {
	// TODO: Implement proper auth - for now return first user or create default
	var user db.User
	if err := s.db.First(&user).Error; err != nil {
		// Create default admin user
		user = db.User{
			Username:  "admin",
			IsAdmin:   true,
			CanRead:   true,
			CanDelete: true,
		}
		s.db.Create(&user)
	}

	user.PasswordHash = ""
	return c.JSON(http.StatusOK, user)
}

// updateUser updates user details
func (s *Server) updateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var user db.User
	if err := s.db.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := s.db.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
	}

	user.PasswordHash = ""
	return c.JSON(http.StatusOK, user)
}

// ========================
// Progress Tracking Handlers
// ========================

// getProgress returns reading/listening progress for a media file
func (s *Server) getProgress(c echo.Context) error {
	mediaFileID, err := strconv.ParseUint(c.Param("mediaFileId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media file ID"})
	}

	// TODO: Get current user from auth context
	userID := uint(1) // Placeholder

	var progress db.ReadProgress
	if err := s.db.Where("user_id = ? AND media_file_id = ?", userID, mediaFileID).First(&progress).Error; err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"progress": 0,
			"position": 0,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"progress":   progress.Progress,
		"position":   progress.Position,
		"lastReadAt": progress.LastReadAt,
	})
}

// updateProgress saves reading/listening progress
func (s *Server) updateProgress(c echo.Context) error {
	mediaFileID, err := strconv.ParseUint(c.Param("mediaFileId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media file ID"})
	}

	var req struct {
		Progress float32 `json:"progress"`
		Position int     `json:"position"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// TODO: Get current user from auth context
	userID := uint(1) // Placeholder

	var progress db.ReadProgress
	result := s.db.Where("user_id = ? AND media_file_id = ?", userID, mediaFileID).First(&progress)

	if result.Error != nil {
		// Create new progress record
		progress = db.ReadProgress{
			UserID:      userID,
			MediaFileID: uint(mediaFileID),
		}
	}

	progress.Progress = req.Progress
	progress.Position = req.Position

	if err := s.db.Save(&progress).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save progress"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Progress saved"})
}

// ========================
// Settings Handlers
// ========================

// getSettings returns application settings
func (s *Server) getSettings(c echo.Context) error {
	// Fetch hardcover API key from database
	var hardcoverSetting db.Setting
	hardcoverAPIKey := ""
	if err := s.db.Where("key = ?", "hardcover_api_key").First(&hardcoverSetting).Error; err == nil {
		hardcoverAPIKey = hardcoverSetting.Value
	}

	settings := map[string]interface{}{
		"general": map[string]interface{}{
			"instanceName": "Shelfarr",
		},
		"paths": map[string]interface{}{
			"books":      s.config.BooksPath,
			"audiobooks": s.config.AudiobooksPath,
			"downloads":  s.config.DownloadsPath,
		},
		"kindle": map[string]interface{}{
			"enabled": false,
			"email":   "",
		},
		"librarySearchProviders": map[string]interface{}{
			"hardcover": map[string]interface{}{
				"enabled":    hardcoverAPIKey != "",
				"apiKey":     hardcoverAPIKey,
				"apiUrl":     s.config.HardcoverAPIURL,
				"rateLimit":  60, // requests per minute
				"maxDepth":   3,  // max query depth
				"maxTimeout": 30, // seconds
			},
		},
	}

	return c.JSON(http.StatusOK, settings)
}

// updateSettings updates application settings
func (s *Server) updateSettings(c echo.Context) error {
	var req map[string]interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Handle librarySearchProviders.hardcover.apiKey
	if providers, ok := req["librarySearchProviders"].(map[string]interface{}); ok {
		if hardcover, ok := providers["hardcover"].(map[string]interface{}); ok {
			if apiKey, ok := hardcover["apiKey"].(string); ok {
				setting := db.Setting{Key: "hardcover_api_key", Value: apiKey}
				if err := s.db.Save(&setting).Error; err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save API key"})
				}
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated"})
}

// getProfiles returns quality profiles
func (s *Server) getProfiles(c echo.Context) error {
	var profiles []db.QualityProfile
	if err := s.db.Find(&profiles).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// If no profiles exist, create defaults
	if len(profiles) == 0 {
		defaults := []db.QualityProfile{
			{
				Name:          "Default Ebook",
				MediaType:     db.MediaTypeEbook,
				FormatRanking: "epub,azw3,mobi,pdf",
			},
			{
				Name:          "Default Audiobook",
				MediaType:     db.MediaTypeAudiobook,
				FormatRanking: "m4b,mp3",
				MinBitrate:    64,
			},
		}
		for _, p := range defaults {
			s.db.Create(&p)
		}
		profiles = defaults
	}

	return c.JSON(http.StatusOK, profiles)
}

// createProfile creates a new quality profile
func (s *Server) createProfile(c echo.Context) error {
	var profile db.QualityProfile
	if err := c.Bind(&profile); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := s.db.Create(&profile).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create profile"})
	}

	return c.JSON(http.StatusCreated, profile)
}

// getProfile returns a single quality profile by ID
func (s *Server) getProfile(c echo.Context) error {
	id := c.Param("id")

	var profile db.QualityProfile
	if err := s.db.First(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Profile not found"})
	}

	return c.JSON(http.StatusOK, profile)
}

// updateProfile updates an existing quality profile
func (s *Server) updateProfile(c echo.Context) error {
	id := c.Param("id")

	var profile db.QualityProfile
	if err := s.db.First(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Profile not found"})
	}

	var updates struct {
		Name          string `json:"name"`
		MediaType     string `json:"mediaType"`
		FormatRanking string `json:"formatRanking"`
		MinBitrate    int    `json:"minBitrate"`
	}

	if err := c.Bind(&updates); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if updates.Name != "" {
		profile.Name = updates.Name
	}
	if updates.MediaType != "" {
		profile.MediaType = db.MediaType(updates.MediaType)
	}
	if updates.FormatRanking != "" {
		profile.FormatRanking = updates.FormatRanking
	}
	if updates.MinBitrate >= 0 {
		profile.MinBitrate = updates.MinBitrate
	}

	if err := s.db.Save(&profile).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

// deleteProfile removes a quality profile
func (s *Server) deleteProfile(c echo.Context) error {
	id := c.Param("id")

	var profile db.QualityProfile
	if err := s.db.First(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Profile not found"})
	}

	// Don't allow deleting if it's the only profile of its type
	var count int64
	s.db.Model(&db.QualityProfile{}).Where("media_type = ?", profile.MediaType).Count(&count)
	if count <= 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete the only profile for this media type"})
	}

	if err := s.db.Delete(&profile).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete profile"})
	}

	return c.NoContent(http.StatusNoContent)
}
