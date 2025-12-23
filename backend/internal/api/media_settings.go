package api

import (
	"net/http"
	"os"
	"strconv"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// MediaSettingsResponse represents the media management settings
type MediaSettingsResponse struct {
	EbookRootFolder     string `json:"ebookRootFolder"`
	AudiobookRootFolder string `json:"audiobookRootFolder"`
	FileNamingEbook     string `json:"fileNamingEbook"`
	FileNamingAudiobook string `json:"fileNamingAudiobook"`
	FolderNaming        string `json:"folderNaming"`
	UseHardlinks        bool   `json:"useHardlinks"`
	RecycleBinEnabled   bool   `json:"recycleBinEnabled"`
	RecycleBinPath      string `json:"recycleBinPath"`
	RescanAfterImport   bool   `json:"rescanAfterImport"`
}

// MediaSettingsRequest represents the request body for updating media settings
type MediaSettingsRequest struct {
	EbookRootFolder     *string `json:"ebookRootFolder,omitempty"`
	AudiobookRootFolder *string `json:"audiobookRootFolder,omitempty"`
	FileNamingEbook     *string `json:"fileNamingEbook,omitempty"`
	FileNamingAudiobook *string `json:"fileNamingAudiobook,omitempty"`
	FolderNaming        *string `json:"folderNaming,omitempty"`
	UseHardlinks        *bool   `json:"useHardlinks,omitempty"`
	RecycleBinEnabled   *bool   `json:"recycleBinEnabled,omitempty"`
	RecycleBinPath      *string `json:"recycleBinPath,omitempty"`
	RescanAfterImport   *bool   `json:"rescanAfterImport,omitempty"`
}

// RootFolderResponse represents a root folder in API responses
type RootFolderResponse struct {
	ID         uint   `json:"id"`
	Path       string `json:"path"`
	MediaType  string `json:"mediaType"`
	Name       string `json:"name"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
	Accessible bool   `json:"accessible"`
}

// RootFolderRequest represents the request body for creating a root folder
type RootFolderRequest struct {
	Path      string `json:"path" validate:"required"`
	MediaType string `json:"mediaType" validate:"required"` // ebook or audiobook
	Name      string `json:"name,omitempty"`
}

// getMediaSettings returns the current media management settings
func (s *Server) getMediaSettings(c echo.Context) error {
	settings := MediaSettingsResponse{
		// Defaults
		FileNamingEbook:     "{Author}/{Title}",
		FileNamingAudiobook: "{Author}/{Title}",
		FolderNaming:        "{Author}/{Series}",
		UseHardlinks:        false,
		RecycleBinEnabled:   false,
		RecycleBinPath:      "",
		RescanAfterImport:   true,
	}

	// Load settings from database
	var dbSettings []db.Setting
	s.db.Where("key LIKE ?", "media_%").Find(&dbSettings)

	for _, setting := range dbSettings {
		switch setting.Key {
		case "media_ebook_root_folder":
			settings.EbookRootFolder = setting.Value
		case "media_audiobook_root_folder":
			settings.AudiobookRootFolder = setting.Value
		case "media_file_naming_ebook":
			settings.FileNamingEbook = setting.Value
		case "media_file_naming_audiobook":
			settings.FileNamingAudiobook = setting.Value
		case "media_folder_naming":
			settings.FolderNaming = setting.Value
		case "media_use_hardlinks":
			settings.UseHardlinks = setting.Value == "true"
		case "media_recycle_bin_enabled":
			settings.RecycleBinEnabled = setting.Value == "true"
		case "media_recycle_bin_path":
			settings.RecycleBinPath = setting.Value
		case "media_rescan_after_import":
			settings.RescanAfterImport = setting.Value != "false" // Default true
		}
	}

	return c.JSON(http.StatusOK, settings)
}

// updateMediaSettings updates media management settings
func (s *Server) updateMediaSettings(c echo.Context) error {
	var req MediaSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Update settings that are provided
	updates := map[string]*string{
		"media_ebook_root_folder":     req.EbookRootFolder,
		"media_audiobook_root_folder": req.AudiobookRootFolder,
		"media_file_naming_ebook":     req.FileNamingEbook,
		"media_file_naming_audiobook": req.FileNamingAudiobook,
		"media_folder_naming":         req.FolderNaming,
		"media_recycle_bin_path":      req.RecycleBinPath,
	}

	for key, valuePtr := range updates {
		if valuePtr != nil {
			setting := db.Setting{Key: key, Value: *valuePtr}
			s.db.Where("key = ?", key).Assign(setting).FirstOrCreate(&setting)
		}
	}

	// Handle boolean settings
	boolUpdates := map[string]*bool{
		"media_use_hardlinks":       req.UseHardlinks,
		"media_recycle_bin_enabled": req.RecycleBinEnabled,
		"media_rescan_after_import": req.RescanAfterImport,
	}

	for key, valuePtr := range boolUpdates {
		if valuePtr != nil {
			value := "false"
			if *valuePtr {
				value = "true"
			}
			setting := db.Setting{Key: key, Value: value}
			s.db.Where("key = ?", key).Assign(setting).FirstOrCreate(&setting)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated"})
}

// getRootFolders returns all configured root folders
func (s *Server) getRootFolders(c echo.Context) error {
	var rootFolders []db.RootFolder
	if err := s.db.Find(&rootFolders).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]RootFolderResponse, len(rootFolders))
	for i, rf := range rootFolders {
		freeSpace, totalSpace, accessible := getDiskSpace(rf.Path)
		responses[i] = RootFolderResponse{
			ID:         rf.ID,
			Path:       rf.Path,
			MediaType:  string(rf.MediaType),
			Name:       rf.Name,
			FreeSpace:  freeSpace,
			TotalSpace: totalSpace,
			Accessible: accessible,
		}
	}

	return c.JSON(http.StatusOK, responses)
}

// addRootFolder creates a new root folder
func (s *Server) addRootFolder(c echo.Context) error {
	var req RootFolderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Path is required"})
	}

	if req.MediaType != "ebook" && req.MediaType != "audiobook" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Media type must be 'ebook' or 'audiobook'"})
	}

	// Check if path exists and is a directory
	info, err := os.Stat(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Path does not exist"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot access path: " + err.Error()})
	}
	if !info.IsDir() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Path is not a directory"})
	}

	// Check if already exists
	var existing db.RootFolder
	if err := s.db.Where("path = ?", req.Path).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Root folder already exists"})
	}

	rootFolder := db.RootFolder{
		Path:      req.Path,
		MediaType: db.MediaType(req.MediaType),
		Name:      req.Name,
	}

	if err := s.db.Create(&rootFolder).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create root folder"})
	}

	freeSpace, totalSpace, accessible := getDiskSpace(rootFolder.Path)
	return c.JSON(http.StatusCreated, RootFolderResponse{
		ID:         rootFolder.ID,
		Path:       rootFolder.Path,
		MediaType:  string(rootFolder.MediaType),
		Name:       rootFolder.Name,
		FreeSpace:  freeSpace,
		TotalSpace: totalSpace,
		Accessible: accessible,
	})
}

// deleteRootFolder removes a root folder
func (s *Server) deleteRootFolder(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid root folder ID"})
	}

	var rootFolder db.RootFolder
	if err := s.db.First(&rootFolder, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Root folder not found"})
	}

	if err := s.db.Delete(&rootFolder).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete root folder"})
	}

	return c.NoContent(http.StatusNoContent)
}

// getDiskSpace returns free and total space for a path
func getDiskSpace(path string) (freeSpace, totalSpace int64, accessible bool) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, false
	}

	// Available space to non-root users
	freeSpace = int64(stat.Bavail) * int64(stat.Bsize)
	totalSpace = int64(stat.Blocks) * int64(stat.Bsize)
	accessible = true

	return
}

// getNamingPreview returns a preview of how a file would be named
func (s *Server) getNamingPreview(c echo.Context) error {
	template := c.QueryParam("template")
	if template == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Template is required"})
	}

	// Sample data for preview
	preview := map[string]string{
		"template": template,
		"preview": applyNamingTemplate(template, map[string]string{
			"Author":      "Brandon Sanderson",
			"Title":       "The Way of Kings",
			"Series":      "The Stormlight Archive",
			"SeriesIndex": "1",
			"Year":        "2010",
			"Quality":     "EPUB",
			"Format":      "epub",
		}),
	}

	return c.JSON(http.StatusOK, preview)
}

// applyNamingTemplate applies template variables to a naming template
func applyNamingTemplate(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := "{" + key + "}"
		// Simple string replacement
		for i := 0; i < len(result)-len(placeholder)+1; i++ {
			if result[i:i+len(placeholder)] == placeholder {
				result = result[:i] + value + result[i+len(placeholder):]
			}
		}
	}
	return result
}

// DirectoryInfo represents information about a directory
type DirectoryInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	HasChildren bool   `json:"hasChildren"`
}

// browseFilesystem returns directories at a given path for browsing
func (s *Server) browseFilesystem(c echo.Context) error {
	path := c.QueryParam("path")
	if path == "" {
		path = "/"
	}

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Path does not exist"})
		}
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Cannot access path"})
	}
	if !info.IsDir() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Path is not a directory"})
	}

	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Cannot read directory"})
	}

	directories := make([]DirectoryInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// Skip hidden directories on Unix systems
			if len(name) > 0 && name[0] == '.' {
				continue
			}

			fullPath := path
			if path == "/" {
				fullPath = "/" + name
			} else {
				fullPath = path + "/" + name
			}

			// Check if this directory has subdirectories
			hasChildren := false
			subEntries, err := os.ReadDir(fullPath)
			if err == nil {
				for _, subEntry := range subEntries {
					if subEntry.IsDir() {
						hasChildren = true
						break
					}
				}
			}

			directories = append(directories, DirectoryInfo{
				Name:        name,
				Path:        fullPath,
				HasChildren: hasChildren,
			})
		}
	}

	// Return response with current path and directories
	response := map[string]interface{}{
		"currentPath": path,
		"parent":      getParentPath(path),
		"directories": directories,
	}

	return c.JSON(http.StatusOK, response)
}

// getParentPath returns the parent directory path
func getParentPath(path string) string {
	if path == "/" || path == "" {
		return ""
	}

	// Find the last slash
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == len(path)-1 {
				// Skip trailing slash
				continue
			}
			lastSlash = i
			break
		}
	}

	if lastSlash <= 0 {
		return "/"
	}

	return path[:lastSlash]
}
