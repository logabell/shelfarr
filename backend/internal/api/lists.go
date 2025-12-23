package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"gorm.io/gorm"
)

// ListRequest represents a request to create/update a Hardcover list
type ListRequest struct {
	Name           string `json:"name"`
	HardcoverURL   string `json:"hardcoverUrl"`
	HardcoverID    string `json:"hardcoverId"`
	Enabled        bool   `json:"enabled"`
	AutoAdd        bool   `json:"autoAdd"`
	Monitor        bool   `json:"monitor"`
	SyncInterval   int    `json:"syncInterval"` // hours
	QualityProfile uint   `json:"qualityProfile,omitempty"`
}

// getLists returns all configured Hardcover lists
func (s *Server) getLists(c echo.Context) error {
	var lists []db.HardcoverList
	s.db.Find(&lists)
	return c.JSON(http.StatusOK, lists)
}

// addList creates a new Hardcover list configuration
func (s *Server) addList(c echo.Context) error {
	var req ListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	list := db.HardcoverList{
		Name:            req.Name,
		HardcoverURL:    req.HardcoverURL,
		HardcoverID:     req.HardcoverID,
		Enabled:         req.Enabled,
		AutoAdd:         req.AutoAdd,
		Monitor:         req.Monitor,
		SyncIntervalHrs: req.SyncInterval,
		QualityProfile:  req.QualityProfile,
	}

	if err := s.db.Create(&list).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create list"})
	}

	return c.JSON(http.StatusCreated, list)
}

// updateList updates an existing Hardcover list configuration
func (s *Server) updateList(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var list db.HardcoverList
	if err := s.db.First(&list, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
	}

	var req ListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	list.Name = req.Name
	list.HardcoverURL = req.HardcoverURL
	list.HardcoverID = req.HardcoverID
	list.Enabled = req.Enabled
	list.AutoAdd = req.AutoAdd
	list.Monitor = req.Monitor
	list.SyncIntervalHrs = req.SyncInterval
	list.QualityProfile = req.QualityProfile

	if err := s.db.Save(&list).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update list"})
	}

	return c.JSON(http.StatusOK, list)
}

// deleteList removes a Hardcover list configuration
func (s *Server) deleteList(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := s.db.Delete(&db.HardcoverList{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete list"})
	}

	return c.NoContent(http.StatusNoContent)
}

// syncList manually syncs a Hardcover list
func (s *Server) syncList(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var list db.HardcoverList
	if err := s.db.First(&list, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
	}

	// Create Hardcover client
	client := hardcover.NewClientWithAPIKey(s.config.HardcoverAPIURL, s.config.HardcoverAPIKey)

	// Sync the list
	addedCount, err := syncHardcoverList(s.db, client, &list)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to sync list: " + err.Error()})
	}

	// Update last synced time
	now := time.Now()
	list.LastSyncedAt = &now
	s.db.Save(&list)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "List synced successfully",
		"booksAdded": addedCount,
	})
}

// syncHardcoverList syncs books from a Hardcover list
func syncHardcoverList(gdb *gorm.DB, client *hardcover.Client, list *db.HardcoverList) (int, error) {
	// Get books from the Hardcover list
	// This would call the Hardcover API to fetch list contents
	// For now, we'll use a placeholder implementation

	result, err := client.GetListBooks(list.HardcoverID, false)
	if err != nil {
		return 0, err
	}

	addedCount := 0

	for _, hcBook := range result.Books {
		// Check if book already exists
		var existingBook db.Book
		result := gdb.Where("hardcover_id = ?", hcBook.ID).First(&existingBook)
		
		if result.Error == nil {
			// Book exists, skip
			continue
		}

		// Check if auto-add is enabled
		if !list.AutoAdd {
			continue
		}

		// Get or create author
		var author db.Author
		if len(hcBook.Authors) > 0 {
			authorName := hcBook.Authors[0]
			result := gdb.Where("name = ?", authorName).First(&author)
			if result.Error != nil {
				author = db.Author{
					Name: authorName,
				}
				gdb.Create(&author)
			}
		}

		// Create the book
		status := db.StatusMissing
		if !list.Monitor {
			status = db.StatusUnmonitored
		}

		newBook := db.Book{
			Title:       hcBook.Title,
			HardcoverID: hcBook.ID,
			AuthorID:    author.ID,
			Description: hcBook.Description,
			CoverURL:    hcBook.CoverURL,
			Status:      status,
			Monitored:   list.Monitor,
		}

		if err := gdb.Create(&newBook).Error; err != nil {
			continue
		}

		addedCount++
	}

	return addedCount, nil
}

// ListSyncService handles automatic list syncing
type ListSyncService struct {
	db              *gorm.DB
	hardcoverClient *hardcover.Client
}

// NewListSyncService creates a new list sync service
func NewListSyncService(database *gorm.DB, hcClient *hardcover.Client) *ListSyncService {
	return &ListSyncService{
		db:              database,
		hardcoverClient: hcClient,
	}
}

// SyncAllLists syncs all enabled lists
func (ls *ListSyncService) SyncAllLists() error {
	var lists []db.HardcoverList
	ls.db.Where("enabled = ?", true).Find(&lists)

	for _, list := range lists {
		// Check if sync is needed based on interval
		if list.LastSyncedAt != nil {
			nextSync := list.LastSyncedAt.Add(time.Duration(list.SyncIntervalHrs) * time.Hour)
			if time.Now().Before(nextSync) {
				continue
			}
		}

		_, err := syncHardcoverList(ls.db, ls.hardcoverClient, &list)
		if err != nil {
			// Log error but continue with other lists
			continue
		}

		// Update last synced time
		now := time.Now()
		list.LastSyncedAt = &now
		ls.db.Save(&list)
	}

	return nil
}

