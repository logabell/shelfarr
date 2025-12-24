package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// AuthorDetailResponse represents an author with all their books (including those not in library)
type AuthorDetailResponse struct {
	ID              uint              `json:"id"`
	HardcoverID     string            `json:"hardcoverId"`
	Name            string            `json:"name"`
	SortName        string            `json:"sortName"`
	Biography       string            `json:"biography"`
	ImageURL        string            `json:"imageUrl"`
	Monitored       bool              `json:"monitored"`
	Books           []AuthorBookEntry `json:"books"`
	TotalBooks      int               `json:"totalBooks"`      // Total books from Hardcover
	InLibrary       int               `json:"inLibrary"`       // Books added to library
	DownloadedCount int               `json:"downloadedCount"` // Books with files downloaded
}

// AuthorBookEntry represents a book by an author (may or may not be in library)
type AuthorBookEntry struct {
	HardcoverID string        `json:"hardcoverId"`
	Title       string        `json:"title"`
	CoverURL    string        `json:"coverUrl,omitempty"`
	AuthorName  string        `json:"authorName,omitempty"`
	Rating      float32       `json:"rating"`
	ReleaseYear int           `json:"releaseYear,omitempty"`
	SeriesID    string        `json:"seriesId,omitempty"`
	SeriesName  string        `json:"seriesName,omitempty"`
	SeriesIndex *float32      `json:"seriesIndex,omitempty"`
	Compilation bool          `json:"compilation"`
	InLibrary   bool          `json:"inLibrary"`
	Book        *BookResponse `json:"book,omitempty"`
}

// AddAuthorRequest represents the request body for adding an author
type AddAuthorRequest struct {
	HardcoverID string `json:"hardcoverId" validate:"required"`
	Monitored   bool   `json:"monitored"`
	AddAllBooks bool   `json:"addAllBooks"`
}

// UpdateAuthorRequest represents the request body for updating an author
type UpdateAuthorRequest struct {
	Monitored bool `json:"monitored"`
}

// getAuthors returns all authors
func (s *Server) getAuthors(c echo.Context) error {
	var authors []db.Author

	query := s.db.Model(&db.Author{})

	// Optional filters
	if monitored := c.QueryParam("monitored"); monitored != "" {
		query = query.Where("monitored = ?", monitored == "true")
	}

	if err := query.Find(&authors).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]AuthorResponse, len(authors))
	for i, author := range authors {
		var bookCount int64
		s.db.Model(&db.Book{}).Where("author_id = ?", author.ID).Count(&bookCount)

		// Count downloaded books by this author
		var downloadedCount int64
		s.db.Model(&db.Book{}).Where("author_id = ? AND status = ?", author.ID, db.StatusDownloaded).Count(&downloadedCount)

		responses[i] = AuthorResponse{
			ID:              author.ID,
			HardcoverID:     author.HardcoverID,
			Name:            author.Name,
			SortName:        author.SortName,
			ImageURL:        author.ImageURL,
			Monitored:       author.Monitored,
			BookCount:       int(bookCount),
			TotalBooksCount: author.TotalBooksCount, // Cached from Hardcover
			DownloadedCount: int(downloadedCount),
		}
	}

	return c.JSON(http.StatusOK, responses)
}

// getAuthor returns a single author with ALL their books from Hardcover, marking which are in library
func (s *Server) getAuthor(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid author ID"})
	}

	var author db.Author
	if err := s.db.First(&author, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Author not found"})
	}

	entries := make([]AuthorBookEntry, 0)
	inLibraryCount := 0
	downloadedCount := 0
	totalBooks := 0

	// Try to fetch author's books from Hardcover (digital editions only by default)
	if author.HardcoverID != "" {
		client, err := s.getHardcoverClient()
		if err == nil {
			languages := s.GetPreferredLanguages()
			result, err := client.GetBooksByAuthorWithCounts(author.HardcoverID, languages)
			if err == nil {
				log.Printf("[DEBUG] getAuthor: fetched %d books from Hardcover for author '%s' (languages: %v)", len(result.Books), author.Name, languages)

				now := time.Now()
				author.TotalBooksCount = result.TotalCount
				author.CachedAt = &now
				s.db.Save(&author)

				// Collect all HardcoverIDs from the Hardcover response
				hardcoverIDs := make([]string, 0, len(result.Books))
				for _, hcBook := range result.Books {
					if hcBook.ID != "" {
						hardcoverIDs = append(hardcoverIDs, hcBook.ID)
					}
				}

				// Query library books by HardcoverID (not author_id) to catch all books
				// regardless of which author they're associated with in our DB
				libraryBooksByHardcoverID := make(map[string]db.Book)
				if len(hardcoverIDs) > 0 {
					var libraryBooks []db.Book
					s.db.Preload("Series").Preload("MediaFiles").
						Where("hardcover_id IN ?", hardcoverIDs).
						Find(&libraryBooks)

					for _, book := range libraryBooks {
						if book.HardcoverID != "" {
							libraryBooksByHardcoverID[book.HardcoverID] = book
						}
					}
					log.Printf("[DEBUG] getAuthor: found %d library books matching Hardcover IDs for author '%s'", len(libraryBooks), author.Name)
				}

				for _, hcBook := range result.Books {
					entry := AuthorBookEntry{
						HardcoverID: hcBook.ID,
						Title:       hcBook.Title,
						CoverURL:    hcBook.CoverURL,
						AuthorName:  author.Name,
						Rating:      hcBook.Rating,
						SeriesID:    hcBook.SeriesID,
						SeriesName:  hcBook.SeriesName,
						SeriesIndex: hcBook.SeriesIndex,
						Compilation: hcBook.Compilation,
						InLibrary:   false,
					}

					if hcBook.ReleaseDate != nil {
						entry.ReleaseYear = hcBook.ReleaseDate.Year()
					}

					if libBook, exists := libraryBooksByHardcoverID[hcBook.ID]; exists {
						entry.InLibrary = true
						inLibraryCount++
						resp := bookToResponse(libBook)
						entry.Book = &resp
						if libBook.Status == db.StatusDownloaded {
							downloadedCount++
						}
					}

					entries = append(entries, entry)
					totalBooks++
				}
			} else {
				log.Printf("[DEBUG] getAuthor: failed to fetch from Hardcover: %v", err)
			}
		}
	}

	// If we didn't get Hardcover data, fall back to library-only view (query by author_id)
	if len(entries) == 0 {
		var libraryBooks []db.Book
		s.db.Preload("Series").Preload("MediaFiles").
			Where("author_id = ?", author.ID).
			Find(&libraryBooks)

		if len(libraryBooks) > 0 {
			log.Printf("[DEBUG] getAuthor: using library-only view for author '%s' (%d books)", author.Name, len(libraryBooks))
			for _, book := range libraryBooks {
				resp := bookToResponse(book)
				entry := AuthorBookEntry{
					HardcoverID: book.HardcoverID,
					Title:       book.Title,
					CoverURL:    book.CoverURL,
					AuthorName:  author.Name,
					Rating:      book.Rating,
					InLibrary:   true,
					Book:        &resp,
				}
				if book.Series != nil {
					entry.SeriesID = book.Series.HardcoverID
					entry.SeriesName = book.Series.Name
				}
				entry.SeriesIndex = book.SeriesIndex
				if book.ReleaseDate != nil {
					entry.ReleaseYear = book.ReleaseDate.Year()
				}
				entries = append(entries, entry)
				totalBooks++
				inLibraryCount++
				if book.Status == db.StatusDownloaded {
					downloadedCount++
				}
			}
		}
	}

	response := AuthorDetailResponse{
		ID:              author.ID,
		HardcoverID:     author.HardcoverID,
		Name:            author.Name,
		SortName:        author.SortName,
		Biography:       author.Biography,
		ImageURL:        author.ImageURL,
		Monitored:       author.Monitored,
		Books:           entries,
		TotalBooks:      totalBooks,
		InLibrary:       inLibraryCount,
		DownloadedCount: downloadedCount,
	}

	return c.JSON(http.StatusOK, response)
}

// addAuthor adds a new author from Hardcover.app
func (s *Server) addAuthor(c echo.Context) error {
	var req AddAuthorRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.HardcoverID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "hardcoverId is required"})
	}

	// Check if author already exists
	var existing db.Author
	if err := s.db.Where("hardcover_id = ?", req.HardcoverID).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Author already exists"})
	}

	// Fetch author data from Hardcover.app
	client, err := s.getHardcoverClient()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Hardcover API key not configured"})
	}
	authorData, err := client.GetAuthor(req.HardcoverID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch author from Hardcover: " + err.Error()})
	}

	// Create the author
	author := db.Author{
		HardcoverID: authorData.ID,
		Name:        authorData.Name,
		SortName:    authorData.SortName,
		Biography:   authorData.Biography,
		ImageURL:    authorData.ImageURL,
		Monitored:   req.Monitored,
	}

	if err := s.db.Create(&author).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create author"})
	}

	// Optionally add all books by this author
	if req.AddAllBooks {
		languages := s.GetPreferredLanguages()
		result, err := client.GetBooksByAuthor(req.HardcoverID, languages)
		if err == nil {
			for _, bookData := range result.Books {
				var existingBook db.Book
				if s.db.Where("hardcover_id = ?", bookData.ID).First(&existingBook).Error != nil {
					book := db.Book{
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
						SeriesIndex: bookData.SeriesIndex,
						Status:      db.StatusMissing,
						Monitored:   req.Monitored,
					}
					s.db.Create(&book)
				}
			}
		}
	}

	return c.JSON(http.StatusCreated, AuthorResponse{
		ID:          author.ID,
		HardcoverID: author.HardcoverID,
		Name:        author.Name,
		SortName:    author.SortName,
		ImageURL:    author.ImageURL,
		Monitored:   author.Monitored,
	})
}

// updateAuthor updates an existing author
func (s *Server) updateAuthor(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid author ID"})
	}

	var author db.Author
	if err := s.db.First(&author, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Author not found"})
	}

	var req UpdateAuthorRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	author.Monitored = req.Monitored

	if err := s.db.Save(&author).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update author"})
	}

	var bookCount int64
	s.db.Model(&db.Book{}).Where("author_id = ?", author.ID).Count(&bookCount)

	return c.JSON(http.StatusOK, AuthorResponse{
		ID:          author.ID,
		HardcoverID: author.HardcoverID,
		Name:        author.Name,
		SortName:    author.SortName,
		ImageURL:    author.ImageURL,
		Monitored:   author.Monitored,
		BookCount:   int(bookCount),
	})
}

// deleteAuthor removes an author from the library
func (s *Server) deleteAuthor(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid author ID"})
	}

	var author db.Author
	if err := s.db.First(&author, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Author not found"})
	}

	// Delete the author (books remain but become orphaned)
	if err := s.db.Delete(&author).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete author"})
	}

	return c.NoContent(http.StatusNoContent)
}
