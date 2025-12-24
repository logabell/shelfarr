package api

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"github.com/shelfarr/shelfarr/internal/openlibrary"
)

// AuthorDetailResponse represents an author with all their books (including those not in library)
type AuthorDetailResponse struct {
	ID              uint              `json:"id"`
	HardcoverID     string            `json:"hardcoverId,omitempty"`
	OpenLibraryID   string            `json:"openLibraryId,omitempty"`
	Name            string            `json:"name"`
	SortName        string            `json:"sortName"`
	Biography       string            `json:"biography"`
	ImageURL        string            `json:"imageUrl"`
	Monitored       bool              `json:"monitored"`
	Books           []AuthorBookEntry `json:"books"`
	TotalBooks      int               `json:"totalBooks"`
	InLibrary       int               `json:"inLibrary"`
	DownloadedCount int               `json:"downloadedCount"`
}

// AuthorBookEntry represents a book by an author (may or may not be in library)
type AuthorBookEntry struct {
	OpenLibraryWorkID string        `json:"openLibraryWorkId,omitempty"`
	HardcoverID       string        `json:"hardcoverId,omitempty"`
	Title             string        `json:"title"`
	CoverURL          string        `json:"coverUrl,omitempty"`
	AuthorName        string        `json:"authorName,omitempty"`
	Rating            float32       `json:"rating"`
	ReleaseYear       int           `json:"releaseYear,omitempty"`
	SeriesID          string        `json:"seriesId,omitempty"`
	SeriesName        string        `json:"seriesName,omitempty"`
	SeriesIndex       *float32      `json:"seriesIndex,omitempty"`
	Compilation       bool          `json:"compilation"`
	InLibrary         bool          `json:"inLibrary"`
	Book              *BookResponse `json:"book,omitempty"`
	IsEbook           *bool         `json:"isEbook,omitempty"`
	HasEpub           *bool         `json:"hasEpub,omitempty"`
	HasPdf            *bool         `json:"hasPdf,omitempty"`
	IsAudiobook       *bool         `json:"isAudiobook,omitempty"`
}

// AddAuthorRequest represents the request body for adding an author
// Supports either hardcoverId OR openLibraryId (at least one required)
type AddAuthorRequest struct {
	HardcoverID   string `json:"hardcoverId"`
	OpenLibraryID string `json:"openLibraryId"`
	Monitored     bool   `json:"monitored"`
	AddAllBooks   bool   `json:"addAllBooks"`
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

// getAuthor returns a single author with ALL their books using Open Library as PRIMARY source
// Per API Integration Strategy: Open Library = primary, Google Books = ebook detection, Hardcover = series/audiobook only
func (s *Server) getAuthor(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid author ID"})
	}

	var author db.Author
	if err := s.db.First(&author, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Author not found"})
	}

	// Auto-migrate: If no OpenLibraryID, search Open Library by name
	if author.OpenLibraryID == "" {
		olClient := openlibrary.NewClient()
		searchResult, err := olClient.SearchAuthors(author.Name, 10)
		if err != nil {
			log.Printf("[WARN] getAuthor: failed to search Open Library for '%s': %v", author.Name, err)
		} else if len(searchResult.Docs) > 0 {
			// Pick the author entry with the most works (better heuristic for canonical entry)
			bestMatch := searchResult.Docs[0]
			for _, doc := range searchResult.Docs[1:] {
				if doc.WorkCount > bestMatch.WorkCount {
					bestMatch = doc
				}
			}
			olid := openlibrary.ExtractOLID(bestMatch.Key)
			author.OpenLibraryID = olid
			s.db.Save(&author)
			log.Printf("[INFO] Auto-migrated author '%s' to OpenLibraryID: %s (work_count=%d, found %d candidates)",
				author.Name, olid, bestMatch.WorkCount, len(searchResult.Docs))
		} else {
			log.Printf("[WARN] getAuthor: no Open Library results for author '%s'", author.Name)
		}
	}

	// Get all library books by this author for quick lookup
	var libraryBooks []db.Book
	s.db.Preload("Series").Preload("MediaFiles").Where("author_id = ?", author.ID).Find(&libraryBooks)

	libraryBooksByOLWorkID := make(map[string]db.Book)
	libraryBooksByTitle := make(map[string]db.Book)
	for _, book := range libraryBooks {
		if book.OpenLibraryWorkID != "" {
			libraryBooksByOLWorkID[book.OpenLibraryWorkID] = book
		}
		libraryBooksByTitle[book.Title] = book
	}

	entries := make([]AuthorBookEntry, 0)
	inLibraryCount := 0
	downloadedCount := 0
	totalBooks := 0

	// PRIMARY: Fetch author's works from Open Library
	if author.OpenLibraryID != "" {
		log.Printf("[DEBUG] getAuthor: fetching works for OpenLibraryID=%s (author: %s)", author.OpenLibraryID, author.Name)
		olClient := openlibrary.NewClient()
		worksResult, err := olClient.GetAuthorWorks(author.OpenLibraryID, 100, 0)
		if err != nil {
			log.Printf("[ERROR] getAuthor: failed to fetch works from Open Library for '%s' (OLID: %s): %v", author.Name, author.OpenLibraryID, err)
		} else {
			log.Printf("[INFO] getAuthor: fetched %d works from Open Library for '%s' (OLID: %s)", worksResult.Size, author.Name, author.OpenLibraryID)

			// Update cached count
			now := time.Now()
			author.TotalBooksCount = worksResult.Size
			author.CachedAt = &now
			s.db.Save(&author)

			// Get Google Books client for ebook detection
			gbClient := s.getGoogleBooksClient()

			for _, work := range worksResult.Entries {
				workID := openlibrary.ExtractOLID(work.Key)

				// Get cover URL
				coverURL := ""
				if len(work.Covers) > 0 && work.Covers[0] > 0 {
					coverURL = openlibrary.GetCoverURL(work.Covers[0], "M")
				} else if work.CoverID > 0 {
					coverURL = openlibrary.GetCoverURL(work.CoverID, "M")
				}

				entry := AuthorBookEntry{
					OpenLibraryWorkID: workID,
					Title:             work.Title,
					CoverURL:          coverURL,
					AuthorName:        author.Name,
					InLibrary:         false,
				}

				// Check if in library (by OL Work ID or title)
				if libBook, exists := libraryBooksByOLWorkID[workID]; exists {
					entry.InLibrary = true
					entry.HardcoverID = libBook.HardcoverID
					inLibraryCount++
					resp := bookToResponse(libBook)
					entry.Book = &resp
					entry.IsEbook = libBook.IsEbook
					entry.HasEpub = libBook.HasEpub
					entry.IsAudiobook = libBook.IsAudiobook
					if libBook.Series != nil {
						entry.SeriesID = libBook.Series.HardcoverID
						entry.SeriesName = libBook.Series.Name
					}
					entry.SeriesIndex = libBook.SeriesIndex
					if libBook.Status == db.StatusDownloaded {
						downloadedCount++
					}
				} else if libBook, exists := libraryBooksByTitle[work.Title]; exists {
					entry.InLibrary = true
					entry.HardcoverID = libBook.HardcoverID
					inLibraryCount++
					resp := bookToResponse(libBook)
					entry.Book = &resp
					entry.IsEbook = libBook.IsEbook
					entry.HasEpub = libBook.HasEpub
					entry.IsAudiobook = libBook.IsAudiobook
					if libBook.Series != nil {
						entry.SeriesID = libBook.Series.HardcoverID
						entry.SeriesName = libBook.Series.Name
					}
					entry.SeriesIndex = libBook.SeriesIndex
					if libBook.Status == db.StatusDownloaded {
						downloadedCount++
					}
				}

				// ENRICH: Check Google Books for ebook status (if not already known)
				if entry.IsEbook == nil && gbClient != nil {
					query := "intitle:" + work.Title + " inauthor:" + author.Name
					gbBooks, err := gbClient.SearchVolumes(query, 3)
					if err == nil {
						for _, gbBook := range gbBooks {
							if gbBook.IsEbook || gbBook.HasEpub {
								isEbook := true
								entry.IsEbook = &isEbook
								entry.HasEpub = &gbBook.HasEpub
								hasPdf := gbBook.HasPdf
								entry.HasPdf = &hasPdf
								break
							}
						}
					}
				}

				entries = append(entries, entry)
				totalBooks++
			}
		}
	}

	// FALLBACK: If Open Library returned nothing, use library books
	if len(entries) == 0 && len(libraryBooks) > 0 {
		log.Printf("[INFO] getAuthor: using library-only view for '%s'", author.Name)
		for _, book := range libraryBooks {
			resp := bookToResponse(book)
			entry := AuthorBookEntry{
				OpenLibraryWorkID: book.OpenLibraryWorkID,
				HardcoverID:       book.HardcoverID,
				Title:             book.Title,
				CoverURL:          book.CoverURL,
				AuthorName:        author.Name,
				Rating:            book.Rating,
				InLibrary:         true,
				Book:              &resp,
				IsEbook:           book.IsEbook,
				HasEpub:           book.HasEpub,
				IsAudiobook:       book.IsAudiobook,
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

	response := AuthorDetailResponse{
		ID:              author.ID,
		HardcoverID:     author.HardcoverID,
		OpenLibraryID:   author.OpenLibraryID,
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

// addAuthor adds a new author from Open Library (preferred) or Hardcover.app
func (s *Server) addAuthor(c echo.Context) error {
	var req AddAuthorRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.HardcoverID == "" && req.OpenLibraryID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either hardcoverId or openLibraryId is required"})
	}

	// Prefer Open Library if provided
	if req.OpenLibraryID != "" {
		return s.addAuthorFromOpenLibrary(c, req)
	}

	return s.addAuthorFromHardcover(c, req)
}

// addAuthorFromOpenLibrary adds an author using Open Library as the source
func (s *Server) addAuthorFromOpenLibrary(c echo.Context, req AddAuthorRequest) error {
	olClient := openlibrary.NewClient()

	olAuthor, err := olClient.GetAuthor(req.OpenLibraryID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch author from Open Library: " + err.Error()})
	}

	olid := openlibrary.ExtractOLID(olAuthor.Key)

	var existing db.Author
	if err := s.db.Where("open_library_id = ?", olid).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Author already exists"})
	}

	biography := openlibrary.ExtractDescription(olAuthor.Bio)

	imageURL := ""
	if len(olAuthor.Photos) > 0 && olAuthor.Photos[0] > 0 {
		imageURL = openlibrary.GetAuthorPhotoURL(olAuthor.Photos[0], "M")
	} else {
		imageURL = openlibrary.GetAuthorPhotoURLByOLID(olid, "M")
	}

	author := db.Author{
		OpenLibraryID: olid,
		Name:          olAuthor.Name,
		SortName:      olAuthor.Name,
		Biography:     biography,
		ImageURL:      imageURL,
		Monitored:     req.Monitored,
	}

	if err := s.db.Create(&author).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create author"})
	}

	if req.AddAllBooks {
		worksResult, err := olClient.GetAuthorWorks(olid, 100, 0)
		if err == nil {
			for _, work := range worksResult.Entries {
				workID := openlibrary.ExtractOLID(work.Key)

				var existingBook db.Book
				if s.db.Where("open_library_work_id = ?", workID).First(&existingBook).Error != nil {
					coverURL := ""
					if len(work.Covers) > 0 && work.Covers[0] > 0 {
						coverURL = openlibrary.GetCoverURL(work.Covers[0], "M")
					}

					book := db.Book{
						OpenLibraryWorkID: workID,
						Title:             work.Title,
						SortTitle:         work.Title,
						CoverURL:          coverURL,
						AuthorID:          author.ID,
						Status:            db.StatusMissing,
						Monitored:         req.Monitored,
					}
					s.db.Create(&book)
				}
			}
		}
	}

	now := time.Now()
	worksResult, err := olClient.GetAuthorWorks(olid, 1, 0)
	if err == nil {
		author.TotalBooksCount = worksResult.Size
		author.CachedAt = &now
		s.db.Save(&author)
	}

	log.Printf("[INFO] Added author '%s' from Open Library (OLID: %s)", author.Name, olid)

	return c.JSON(http.StatusCreated, AuthorResponse{
		ID:            author.ID,
		OpenLibraryID: author.OpenLibraryID,
		Name:          author.Name,
		SortName:      author.SortName,
		ImageURL:      author.ImageURL,
		Monitored:     author.Monitored,
	})
}

// addAuthorFromHardcover adds an author using Hardcover.app as the source
func (s *Server) addAuthorFromHardcover(c echo.Context, req AddAuthorRequest) error {
	var existing db.Author
	if err := s.db.Where("hardcover_id = ?", req.HardcoverID).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Author already exists"})
	}

	client := hardcover.NewClient(s.config.HardcoverAPIURL)
	authorData, err := client.GetAuthor(req.HardcoverID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch author from Hardcover: " + err.Error()})
	}

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

	if req.AddAllBooks {
		languages := s.GetPreferredLanguages()
		result, err := client.GetBooksByAuthor(req.HardcoverID, languages, false)
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

	log.Printf("[INFO] Added author '%s' from Hardcover (ID: %s)", author.Name, author.HardcoverID)

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

func (s *Server) deleteAuthor(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid author ID"})
	}

	deleteFiles := c.QueryParam("deleteFiles") == "true"

	var author db.Author
	if err := s.db.First(&author, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Author not found"})
	}

	if deleteFiles {
		var books []db.Book
		s.db.Preload("MediaFiles").Where("author_id = ?", author.ID).Find(&books)

		for _, book := range books {
			for _, mf := range book.MediaFiles {
				if err := os.Remove(mf.FilePath); err != nil && !os.IsNotExist(err) {
					log.Printf("[WARN] Failed to delete file %s: %v", mf.FilePath, err)
				}
				s.db.Delete(&mf)
			}
			s.db.Delete(&book)
		}
	}

	if err := s.db.Delete(&author).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete author"})
	}

	log.Printf("[INFO] Deleted author '%s' (ID: %d), deleteFiles=%v", author.Name, author.ID, deleteFiles)

	return c.NoContent(http.StatusNoContent)
}
