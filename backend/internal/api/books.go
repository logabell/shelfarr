package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"github.com/shelfarr/shelfarr/internal/openlibrary"
)

type AddBookRequest struct {
	HardcoverID       string `json:"hardcoverId"`
	OpenLibraryWorkID string `json:"openLibraryWorkId"`
	Monitored         bool   `json:"monitored"`
}

// UpdateBookRequest represents the request body for updating a book
type UpdateBookRequest struct {
	Monitored bool   `json:"monitored"`
	Status    string `json:"status,omitempty"`
}

// getBooks returns all books with optional filtering
func (s *Server) getBooks(c echo.Context) error {
	var books []db.Book

	query := s.db.Preload("Author").Preload("Series").Preload("MediaFiles")

	// Optional filters
	if monitored := c.QueryParam("monitored"); monitored != "" {
		query = query.Where("monitored = ?", monitored == "true")
	}
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&books).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]BookResponse, len(books))
	for i, book := range books {
		responses[i] = bookToResponse(book)
	}

	return c.JSON(http.StatusOK, responses)
}

// getBook returns a single book by ID
func (s *Server) getBook(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.Preload("Author").Preload("Series").Preload("MediaFiles").First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	return c.JSON(http.StatusOK, bookToResponse(book))
}

func (s *Server) addBook(c echo.Context) error {
	var req AddBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.HardcoverID == "" && req.OpenLibraryWorkID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either hardcoverId or openLibraryWorkId is required"})
	}

	if req.OpenLibraryWorkID != "" {
		return s.addBookFromOpenLibrary(c, req)
	}

	return s.addBookFromHardcover(c, req)
}

func (s *Server) addBookFromHardcover(c echo.Context, req AddBookRequest) error {
	var existing db.Book
	if err := s.db.Where("hardcover_id = ?", req.HardcoverID).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Book already exists"})
	}

	client := hardcover.NewClient(s.config.HardcoverAPIURL)
	bookData, err := client.GetBook(req.HardcoverID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch book from Hardcover: " + err.Error()})
	}

	var author db.Author
	if bookData.AuthorID != "" {
		if err := s.db.Where("hardcover_id = ?", bookData.AuthorID).First(&author).Error; err != nil {
			authorData, err := client.GetAuthor(bookData.AuthorID)
			if err == nil {
				author = db.Author{
					HardcoverID: authorData.ID,
					Name:        authorData.Name,
					SortName:    authorData.SortName,
					Biography:   authorData.Biography,
					ImageURL:    authorData.ImageURL,
					Monitored:   false,
				}
				s.db.Create(&author)
			}
		}
	}

	var series *db.Series
	if bookData.SeriesID != "" {
		var seriesRecord db.Series
		if err := s.db.Where("hardcover_id = ?", bookData.SeriesID).First(&seriesRecord).Error; err != nil {
			seriesRecord = db.Series{
				HardcoverID: bookData.SeriesID,
				Name:        bookData.SeriesName,
			}
			s.db.Create(&seriesRecord)
		}
		series = &seriesRecord
	}

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

	if series != nil {
		book.SeriesID = &series.ID
	}

	if err := s.db.Create(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create book"})
	}

	s.db.Preload("Author").Preload("Series").First(&book, book.ID)

	return c.JSON(http.StatusCreated, bookToResponse(book))
}

func (s *Server) addBookFromOpenLibrary(c echo.Context, req AddBookRequest) error {
	workID := req.OpenLibraryWorkID
	if strings.HasPrefix(workID, "/works/") {
		workID = strings.TrimPrefix(workID, "/works/")
	}

	var existing db.Book
	if err := s.db.Where("open_library_work_id = ?", workID).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Book already exists"})
	}

	olClient := openlibrary.NewClient()
	work, err := olClient.GetWork(workID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch work from Open Library: " + err.Error()})
	}

	log.Printf("[INFO] addBookFromOpenLibrary: Adding work '%s' (OL: %s)", work.Title, workID)

	var author db.Author
	if len(work.Authors) > 0 && work.Authors[0].Author.Key != "" {
		authorOLID := openlibrary.ExtractOLID(work.Authors[0].Author.Key)

		if err := s.db.Where("open_library_id = ?", authorOLID).First(&author).Error; err != nil {
			olAuthor, err := olClient.GetAuthor(authorOLID)
			if err == nil {
				bio := openlibrary.ExtractDescription(olAuthor.Bio)
				var photoURL string
				if len(olAuthor.Photos) > 0 && olAuthor.Photos[0] > 0 {
					photoURL = openlibrary.GetAuthorPhotoURL(olAuthor.Photos[0], "M")
				}

				author = db.Author{
					OpenLibraryID: authorOLID,
					Name:          olAuthor.Name,
					SortName:      olAuthor.PersonalName,
					Biography:     bio,
					ImageURL:      photoURL,
					Monitored:     false,
				}
				s.db.Create(&author)
				log.Printf("[INFO] addBookFromOpenLibrary: Created author '%s' (OL: %s)", author.Name, authorOLID)
			}
		}
	}

	var coverURL string
	if len(work.Covers) > 0 && work.Covers[0] > 0 {
		coverURL = openlibrary.GetCoverURL(work.Covers[0], "L")
	}

	var releaseDate *time.Time
	if work.FirstPublishDate != "" {
		if t, err := time.Parse("2006", work.FirstPublishDate); err == nil {
			releaseDate = &t
		} else if t, err := time.Parse("January 2, 2006", work.FirstPublishDate); err == nil {
			releaseDate = &t
		} else if t, err := time.Parse("2006-01-02", work.FirstPublishDate); err == nil {
			releaseDate = &t
		}
	}

	description := openlibrary.ExtractDescription(work.Description)

	var isbn, isbn13 string
	editions, err := olClient.GetWorkEditions(workID, 10, 0)
	if err == nil && len(editions.Entries) > 0 {
		for _, ed := range editions.Entries {
			if len(ed.ISBN13) > 0 && isbn13 == "" {
				isbn13 = ed.ISBN13[0]
			}
			if len(ed.ISBN10) > 0 && isbn == "" {
				isbn = ed.ISBN10[0]
			}
			if isbn != "" && isbn13 != "" {
				break
			}
		}
	}

	book := db.Book{
		OpenLibraryWorkID: workID,
		Title:             work.Title,
		SortTitle:         work.Title,
		Description:       description,
		CoverURL:          coverURL,
		ReleaseDate:       releaseDate,
		ISBN:              isbn,
		ISBN13:            isbn13,
		AuthorID:          author.ID,
		Status:            db.StatusMissing,
		Monitored:         req.Monitored,
	}

	if isbn13 != "" || isbn != "" {
		gbClient := s.getGoogleBooksClient()
		if gbClient != nil {
			checkISBN := isbn13
			if checkISBN == "" {
				checkISBN = isbn
			}
			gbInfo, err := gbClient.GetVolumeByISBN(checkISBN)
			if err == nil {
				book.GoogleVolumeID = gbInfo.VolumeID
				book.IsEbook = &gbInfo.IsEbook
				book.HasEpub = &gbInfo.HasEpub
				book.HasPdf = &gbInfo.HasPdf
				book.BuyLink = gbInfo.BuyLink
				now := time.Now()
				book.EbookCheckedAt = &now
				log.Printf("[INFO] addBookFromOpenLibrary: Google Books enrichment - isEbook=%v, hasEpub=%v", gbInfo.IsEbook, gbInfo.HasEpub)
			}
		}
	}

	if err := s.db.Create(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create book"})
	}

	s.db.Preload("Author").Preload("Series").First(&book, book.ID)
	log.Printf("[INFO] addBookFromOpenLibrary: Successfully added book '%s' (ID: %d)", book.Title, book.ID)

	return c.JSON(http.StatusCreated, bookToResponse(book))
}

// updateBook updates an existing book
func (s *Server) updateBook(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	var req UpdateBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	book.Monitored = req.Monitored
	if req.Status != "" {
		book.Status = db.BookStatus(req.Status)
	}

	if err := s.db.Save(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update book"})
	}

	s.db.Preload("Author").Preload("Series").Preload("MediaFiles").First(&book, book.ID)

	return c.JSON(http.StatusOK, bookToResponse(book))
}

// deleteBook removes a book from the library
func (s *Server) deleteBook(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	// Soft delete media files (move to recycle bin)
	s.db.Where("book_id = ?", id).Delete(&db.MediaFile{})

	// Delete the book
	if err := s.db.Delete(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete book"})
	}

	return c.NoContent(http.StatusNoContent)
}

// BulkUpdateRequest represents a request to update multiple books
type BulkUpdateRequest struct {
	BookIDs   []uint `json:"bookIds"`
	Monitored *bool  `json:"monitored,omitempty"`
	Status    string `json:"status,omitempty"`
}

// BulkDeleteRequest represents a request to delete multiple books
type BulkDeleteRequest struct {
	BookIDs     []uint `json:"bookIds"`
	DeleteFiles bool   `json:"deleteFiles"`
}

// bulkUpdateBooks updates multiple books at once
func (s *Server) bulkUpdateBooks(c echo.Context) error {
	var req BulkUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if len(req.BookIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No book IDs provided"})
	}

	updates := make(map[string]interface{})
	if req.Monitored != nil {
		updates["monitored"] = *req.Monitored
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No updates provided"})
	}

	result := s.db.Model(&db.Book{}).Where("id IN ?", req.BookIDs).Updates(updates)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update books"})
	}

	return c.JSON(http.StatusOK, map[string]int64{"updated": result.RowsAffected})
}

// bulkDeleteBooks deletes multiple books at once
func (s *Server) bulkDeleteBooks(c echo.Context) error {
	var req BulkDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if len(req.BookIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No book IDs provided"})
	}

	// If deleteFiles is true, we'd need to actually delete files from disk
	// For now, just soft-delete the media files
	if req.DeleteFiles {
		// TODO: Implement actual file deletion
		s.db.Where("book_id IN ?", req.BookIDs).Delete(&db.MediaFile{})
	} else {
		// Just soft-delete media files (they stay in recycle bin)
		s.db.Where("book_id IN ?", req.BookIDs).Delete(&db.MediaFile{})
	}

	// Delete the books (soft delete)
	result := s.db.Where("id IN ?", req.BookIDs).Delete(&db.Book{})
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete books"})
	}

	return c.JSON(http.StatusOK, map[string]int64{"deleted": result.RowsAffected})
}
