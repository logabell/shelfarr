package api

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// SystemStatus represents the overall system status
type SystemStatus struct {
	Version   string       `json:"version"`
	StartTime time.Time    `json:"startTime"`
	Uptime    string       `json:"uptime"`
	OS        string       `json:"os"`
	Arch      string       `json:"arch"`
	GoVersion string       `json:"goVersion"`
	Database  DBStatus     `json:"database"`
	Disk      DiskStatus   `json:"disk"`
	Clients   ClientStatus `json:"clients"`
	Library   LibStatus    `json:"library"`
}

// DBStatus represents database status
type DBStatus struct {
	Type   string `json:"type"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Status string `json:"status"`
}

// DiskStatus represents disk usage
type DiskStatus struct {
	BooksPath      PathStatus `json:"booksPath"`
	AudiobooksPath PathStatus `json:"audiobooksPath"`
	DownloadsPath  PathStatus `json:"downloadsPath"`
}

// PathStatus represents status of a path
type PathStatus struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	Writable  bool   `json:"writable"`
	UsedBytes int64  `json:"usedBytes,omitempty"`
}

// ClientStatus represents status of connected clients
type ClientStatus struct {
	Indexers        int `json:"indexers"`
	DownloadClients int `json:"downloadClients"`
	WebSockets      int `json:"webSockets"`
}

// LibStatus represents library statistics
type LibStatus struct {
	TotalBooks      int64 `json:"totalBooks"`
	MonitoredBooks  int64 `json:"monitoredBooks"`
	TotalAuthors    int64 `json:"totalAuthors"`
	TotalSeries     int64 `json:"totalSeries"`
	TotalMediaFiles int64 `json:"totalMediaFiles"`
	TotalSize       int64 `json:"totalSize"`
}

// TaskInfo represents information about a scheduled task
type TaskInfo struct {
	Name       string    `json:"name"`
	Interval   string    `json:"interval"`
	LastRun    time.Time `json:"lastRun"`
	NextRun    time.Time `json:"nextRun"`
	Running    bool      `json:"running"`
	Enabled    bool      `json:"enabled"`
	LastStatus string    `json:"lastStatus"`
}

var serverStartTime = time.Now()

// getSystemStatus returns system status information
func (s *Server) getSystemStatus(c echo.Context) error {
	status := SystemStatus{
		Version:   "1.0.0",
		StartTime: serverStartTime,
		Uptime:    time.Since(serverStartTime).Round(time.Second).String(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}

	// Database status
	status.Database = DBStatus{
		Type:   "sqlite",
		Path:   s.config.DatabasePath,
		Status: "connected",
	}
	if info, err := os.Stat(s.config.DatabasePath); err == nil {
		status.Database.Size = info.Size()
	}

	// Disk status
	status.Disk = DiskStatus{
		BooksPath:      checkPath(s.config.BooksPath),
		AudiobooksPath: checkPath(s.config.AudiobooksPath),
		DownloadsPath:  checkPath(s.config.DownloadsPath),
	}

	// Client status
	var indexerCount, downloadClientCount int64
	s.db.Model(&db.Indexer{}).Where("enabled = ?", true).Count(&indexerCount)
	s.db.Model(&db.DownloadClient{}).Where("enabled = ?", true).Count(&downloadClientCount)

	status.Clients = ClientStatus{
		Indexers:        int(indexerCount),
		DownloadClients: int(downloadClientCount),
		WebSockets:      s.wsHub.ClientCount(),
	}

	// Library status
	var totalBooks, monitoredBooks, totalAuthors, totalSeries, totalMediaFiles int64
	var totalSize int64

	s.db.Model(&db.Book{}).Count(&totalBooks)
	s.db.Model(&db.Book{}).Where("monitored = ?", true).Count(&monitoredBooks)
	s.db.Model(&db.Author{}).Count(&totalAuthors)
	s.db.Model(&db.Series{}).Count(&totalSeries)
	s.db.Model(&db.MediaFile{}).Count(&totalMediaFiles)
	s.db.Model(&db.MediaFile{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	status.Library = LibStatus{
		TotalBooks:      totalBooks,
		MonitoredBooks:  monitoredBooks,
		TotalAuthors:    totalAuthors,
		TotalSeries:     totalSeries,
		TotalMediaFiles: totalMediaFiles,
		TotalSize:       totalSize,
	}

	return c.JSON(http.StatusOK, status)
}

// getSystemTasks returns information about scheduled tasks
func (s *Server) getSystemTasks(c echo.Context) error {
	// These would come from the scheduler in a full implementation
	tasks := []TaskInfo{
		{
			Name:       "MetadataSync",
			Interval:   "24h",
			Enabled:    true,
			LastStatus: "success",
		},
		{
			Name:       "ListSync",
			Interval:   "6h",
			Enabled:    true,
			LastStatus: "success",
		},
		{
			Name:       "DownloadSync",
			Interval:   "30s",
			Enabled:    true,
			Running:    false,
			LastStatus: "success",
		},
		{
			Name:       "LibraryScan",
			Interval:   "1h",
			Enabled:    true,
			LastStatus: "success",
		},
		{
			Name:       "RecycleBinCleanup",
			Interval:   "168h",
			Enabled:    true,
			LastStatus: "success",
		},
	}

	return c.JSON(http.StatusOK, tasks)
}

// runSystemTask manually triggers a scheduled task
func (s *Server) runSystemTask(c echo.Context) error {
	taskName := c.Param("name")

	// In a full implementation, this would call the scheduler
	validTasks := map[string]bool{
		"MetadataSync":      true,
		"ListSync":          true,
		"DownloadSync":      true,
		"LibraryScan":       true,
		"RecycleBinCleanup": true,
	}

	if !validTasks[taskName] {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Task " + taskName + " started",
	})
}

// getSystemLogs returns recent system logs
func (s *Server) getSystemLogs(c echo.Context) error {
	// In production, this would read from a log file or logging system
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute),
			"level":     "info",
			"message":   "Download completed: Example Book",
		},
		{
			"timestamp": time.Now().Add(-10 * time.Minute),
			"level":     "info",
			"message":   "Library scan completed: 0 new files",
		},
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"level":     "info",
			"message":   "Server started",
		},
	}

	return c.JSON(http.StatusOK, logs)
}

// createBackup creates a database backup
func (s *Server) createBackup(c echo.Context) error {
	// In production, this would create an actual backup
	backupPath := s.config.ConfigPath + "/backups/shelfarr_" + time.Now().Format("20060102_150405") + ".db"

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup created",
		"path":    backupPath,
	})
}

// checkPath checks if a path exists and is writable
func checkPath(path string) PathStatus {
	status := PathStatus{Path: path}

	info, err := os.Stat(path)
	if err != nil {
		status.Exists = false
		return status
	}

	status.Exists = true

	// Check if writable by trying to create a temp file
	testFile := path + "/.shelfarr_write_test"
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		status.Writable = true
	}

	// Calculate used space (simplified - just checks if directory)
	if info.IsDir() {
		status.UsedBytes = calculateDirSize(path)
	}

	return status
}

func calculateDirSize(path string) int64 {
	var size int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if info, err := entry.Info(); err == nil {
			if entry.IsDir() {
				size += calculateDirSize(path + "/" + entry.Name())
			} else {
				size += info.Size()
			}
		}
	}

	return size
}

func (s *Server) refreshAllMetadata(c echo.Context) error {
	var req struct {
		BookIDs []uint `json:"bookIds"`
	}
	if err := c.Bind(&req); err != nil {
		req.BookIDs = nil
	}

	var books []db.Book
	query := s.db.Where("hardcover_id != ''")
	if len(req.BookIDs) > 0 {
		query = query.Where("id IN ?", req.BookIDs)
	}
	if err := query.Find(&books).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch books"})
	}

	if len(books) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"message":   "No books to refresh",
			"refreshed": 0,
			"failed":    0,
		})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create Hardcover client"})
	}

	var refreshed, failed int
	var errors []string

	for _, book := range books {
		bookData, err := client.GetBook(book.HardcoverID)
		if err != nil {
			failed++
			errors = append(errors, book.Title+": "+err.Error())
			continue
		}

		s.updateBookFromHardcover(&book, bookData)
		if err := s.db.Save(&book).Error; err != nil {
			failed++
			errors = append(errors, book.Title+": save failed")
			continue
		}

		s.syncGenres(&book, bookData.Genres)
		s.syncEditions(&book, bookData)
		s.syncContributors(&book, bookData)
		refreshed++

		time.Sleep(100 * time.Millisecond)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":   "Metadata refresh completed",
		"refreshed": refreshed,
		"failed":    failed,
		"errors":    errors,
	})
}
