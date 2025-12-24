package api

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
)

// SeriesDetailResponse represents a series with all its books (including those not in library)
type SeriesDetailResponse struct {
	ID              uint              `json:"id"`
	HardcoverID     string            `json:"hardcoverId"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Books           []SeriesBookEntry `json:"books"`
	TotalBooks      int               `json:"totalBooks"`      // Total books from Hardcover
	InLibrary       int               `json:"inLibrary"`       // Books added to library
	DownloadedCount int               `json:"downloadedCount"` // Books with files downloaded
	MissingBooks    int               `json:"missingBooks"`    // Not in library yet
}

// SeriesBookEntry represents a book in a series (may or may not be in library)
type SeriesBookEntry struct {
	Index       float32       `json:"index"`
	Book        *BookResponse `json:"book,omitempty"`        // Library book (if in library)
	HardcoverID string        `json:"hardcoverId,omitempty"` // Hardcover ID for adding
	Title       string        `json:"title"`                 // Title from Hardcover or library
	CoverURL    string        `json:"coverUrl,omitempty"`    // Cover from Hardcover
	AuthorName  string        `json:"authorName,omitempty"`  // Author name from Hardcover
	Rating      float32       `json:"rating"`                // Rating from Hardcover
	ReleaseYear int           `json:"releaseYear,omitempty"` // Release year from Hardcover
	Compilation bool          `json:"compilation"`           // Whether book is a compilation/box set
	InLibrary   bool          `json:"inLibrary"`             // Whether book is in library
}

// getSeries returns all series in the library
func (s *Server) getSeries(c echo.Context) error {
	var seriesList []db.Series

	if err := s.db.Find(&seriesList).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]SeriesResponse, len(seriesList))
	for i, series := range seriesList {
		var bookCount int64
		s.db.Model(&db.Book{}).Where("series_id = ?", series.ID).Count(&bookCount)

		// Count downloaded books in this series
		var downloadedCount int64
		s.db.Model(&db.Book{}).Where("series_id = ? AND status = ?", series.ID, db.StatusDownloaded).Count(&downloadedCount)

		responses[i] = SeriesResponse{
			ID:              series.ID,
			HardcoverID:     series.HardcoverID,
			Name:            series.Name,
			BookCount:       int(bookCount),
			TotalBooksCount: series.TotalBooksCount, // Cached from Hardcover
			DownloadedCount: int(downloadedCount),
		}
	}

	return c.JSON(http.StatusOK, responses)
}

// getSeriesDetail returns a series with all books from Hardcover, marking which are in library
func (s *Server) getSeriesDetail(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid series ID"})
	}

	var series db.Series
	if err := s.db.First(&series, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Series not found"})
	}

	// Get all library books in this series (indexed by HardcoverID for quick lookup)
	var libraryBooks []db.Book
	s.db.Preload("Author").Preload("MediaFiles").
		Where("series_id = ?", series.ID).
		Find(&libraryBooks)

	// Create a map for quick lookup
	libraryBooksByHardcoverID := make(map[string]db.Book)
	for _, book := range libraryBooks {
		if book.HardcoverID != "" {
			libraryBooksByHardcoverID[book.HardcoverID] = book
		}
	}

	entries := make([]SeriesBookEntry, 0)
	ownedCount := 0
	totalBooks := 0

	// Try to fetch series data from Hardcover
	if series.HardcoverID != "" {
		client, err := s.getHardcoverClient()
		if err == nil {
			languages := s.GetPreferredLanguages()
			result, err := client.GetSeries(series.HardcoverID, languages, false)
			if err == nil && result.Series != nil {
				log.Printf("[DEBUG] getSeriesDetail: fetched %d books from Hardcover for series '%s' (languages: %v)", len(result.Books), series.Name, languages)

				now := time.Now()
				if result.Series.PrimaryBooksCount > 0 {
					series.TotalBooksCount = result.Series.PrimaryBooksCount
				} else if result.TotalCount > 0 {
					series.TotalBooksCount = result.TotalCount
				} else {
					series.TotalBooksCount = len(result.Books)
				}
				series.CachedAt = &now
				s.db.Save(&series)

				for _, hcBook := range result.Books {
					var seriesIndex float32 = 0
					if hcBook.SeriesIndex != nil {
						seriesIndex = *hcBook.SeriesIndex
					}
					entry := SeriesBookEntry{
						Index:       seriesIndex,
						HardcoverID: hcBook.ID,
						Title:       hcBook.Title,
						CoverURL:    hcBook.CoverURL,
						AuthorName:  hcBook.AuthorName,
						Rating:      hcBook.Rating,
						Compilation: hcBook.Compilation,
						InLibrary:   false,
					}

					if hcBook.ReleaseDate != nil {
						entry.ReleaseYear = hcBook.ReleaseDate.Year()
					}

					// Check if this book is in our library
					if libBook, exists := libraryBooksByHardcoverID[hcBook.ID]; exists {
						entry.InLibrary = true
						resp := bookToResponse(libBook)
						entry.Book = &resp
						if libBook.Status == db.StatusDownloaded {
							ownedCount++
						}
					}

					entries = append(entries, entry)
					totalBooks++
				}
			} else {
				log.Printf("[DEBUG] getSeriesDetail: failed to fetch from Hardcover: %v", err)
			}
		}
	}

	// If we didn't get Hardcover data, fall back to library-only view
	if len(entries) == 0 && len(libraryBooks) > 0 {
		log.Printf("[DEBUG] getSeriesDetail: using library-only view for series '%s'", series.Name)
		for _, book := range libraryBooks {
			resp := bookToResponse(book)
			var index float32 = 0
			if book.SeriesIndex != nil {
				index = *book.SeriesIndex
			}
			entry := SeriesBookEntry{
				Index:       index,
				Book:        &resp,
				HardcoverID: book.HardcoverID,
				Title:       book.Title,
				CoverURL:    book.CoverURL,
				Rating:      book.Rating,
				InLibrary:   true,
			}
			if book.Author.Name != "" {
				entry.AuthorName = book.Author.Name
			}
			if book.ReleaseDate != nil {
				entry.ReleaseYear = book.ReleaseDate.Year()
			}
			entries = append(entries, entry)
			totalBooks++
			if book.Status == db.StatusDownloaded {
				ownedCount++
			}
		}
	}

	// Sort by index
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Index < entries[j].Index
	})

	// Count books in library
	inLibraryCount := 0
	for _, entry := range entries {
		if entry.InLibrary {
			inLibraryCount++
		}
	}

	response := SeriesDetailResponse{
		ID:              series.ID,
		HardcoverID:     series.HardcoverID,
		Name:            series.Name,
		Description:     series.Description,
		Books:           entries,
		TotalBooks:      totalBooks,
		InLibrary:       inLibraryCount,
		DownloadedCount: ownedCount,
		MissingBooks:    totalBooks - inLibraryCount,
	}

	return c.JSON(http.StatusOK, response)
}

type AddSeriesBooksRequest struct {
	BookIDs   []string `json:"bookIds"`
	Monitored bool     `json:"monitored"`
}

type AddSeriesBooksResponse struct {
	Message      string   `json:"message"`
	AddedCount   int      `json:"addedCount"`
	SkippedCount int      `json:"skippedCount"`
	Errors       []string `json:"errors,omitempty"`
}

func (s *Server) deleteSeries(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid series ID"})
	}

	deleteFiles := c.QueryParam("deleteFiles") == "true"

	var series db.Series
	if err := s.db.First(&series, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Series not found"})
	}

	if deleteFiles {
		var books []db.Book
		s.db.Preload("MediaFiles").Where("series_id = ?", series.ID).Find(&books)

		for _, book := range books {
			for _, mf := range book.MediaFiles {
				if err := os.Remove(mf.FilePath); err != nil && !os.IsNotExist(err) {
					log.Printf("[WARN] Failed to delete file %s: %v", mf.FilePath, err)
				}
				s.db.Delete(&mf)
			}
		}
	}

	s.db.Model(&db.Book{}).Where("series_id = ?", series.ID).Update("series_id", nil)

	if err := s.db.Delete(&series).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete series"})
	}

	log.Printf("[INFO] Deleted series '%s' (ID: %d), deleteFiles=%v", series.Name, series.ID, deleteFiles)

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) addSeriesBooks(c echo.Context) error {
	seriesIDParam := c.Param("id")
	if seriesIDParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Series ID is required"})
	}

	var req AddSeriesBooksRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if len(req.BookIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "At least one book ID is required"})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to initialize Hardcover client"})
	}

	languages := s.GetPreferredLanguages()

	seriesID, err := strconv.ParseUint(seriesIDParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid series ID"})
	}

	var series db.Series
	if err := s.db.First(&series, seriesID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Series not found"})
	}

	result, err := client.GetSeries(series.HardcoverID, languages, false)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch series from Hardcover"})
	}

	bookDataMap := make(map[string]*hardcover.BookData)
	for i := range result.Books {
		bookDataMap[result.Books[i].ID] = &result.Books[i]
	}

	addedCount := 0
	skippedCount := 0
	var errors []string

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, bookID := range req.BookIDs {
		var existingBook db.Book
		if err := tx.Where("hardcover_id = ?", bookID).First(&existingBook).Error; err == nil {
			skippedCount++
			continue
		}

		bookData, exists := bookDataMap[bookID]
		if !exists {
			fetchedBook, fetchErr := client.GetBook(bookID)
			if fetchErr != nil {
				errors = append(errors, "Failed to fetch book "+bookID+": "+fetchErr.Error())
				continue
			}
			bookData = fetchedBook
		}

		var author db.Author
		if bookData.AuthorID != "" {
			if err := tx.Where("hardcover_id = ?", bookData.AuthorID).First(&author).Error; err != nil {
				author = db.Author{
					HardcoverID: bookData.AuthorID,
					Name:        bookData.AuthorName,
					SortName:    bookData.AuthorName,
				}
				if err := tx.Create(&author).Error; err != nil {
					errors = append(errors, "Failed to create author for book "+bookID)
					continue
				}
			}
		}

		newBook := db.Book{
			HardcoverID: bookData.ID,
			Title:       bookData.Title,
			SortTitle:   bookData.SortTitle,
			ISBN:        bookData.ISBN,
			ISBN13:      bookData.ISBN13,
			Description: bookData.Description,
			CoverURL:    bookData.CoverURL,
			Rating:      bookData.Rating,
			ReleaseDate: bookData.ReleaseDate,
			PageCount:   bookData.PageCount,
			AuthorID:    author.ID,
			SeriesID:    &series.ID,
			SeriesIndex: bookData.SeriesIndex,
			Status:      db.StatusMissing,
			Monitored:   req.Monitored,
		}

		if err := tx.Create(&newBook).Error; err != nil {
			errors = append(errors, "Failed to add book "+bookData.Title)
			continue
		}

		addedCount++
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	response := AddSeriesBooksResponse{
		Message:      "Books added to library",
		AddedCount:   addedCount,
		SkippedCount: skippedCount,
	}

	if len(errors) > 0 {
		response.Errors = errors
	}

	return c.JSON(http.StatusCreated, response)
}
