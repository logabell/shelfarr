package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
)

// AddBookRequest represents the request body for adding a book
type AddBookRequest struct {
	HardcoverID string `json:"hardcoverId" validate:"required"`
	Monitored   bool   `json:"monitored"`
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

// addBook adds a new book from Hardcover.app
func (s *Server) addBook(c echo.Context) error {
	var req AddBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.HardcoverID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "hardcoverId is required"})
	}

	var existing db.Book
	if err := s.db.Unscoped().Where("hardcover_id = ?", req.HardcoverID).First(&existing).Error; err == nil {
		if existing.DeletedAt.Valid {
			if err := s.db.Unscoped().Model(&existing).Updates(map[string]interface{}{
				"deleted_at": nil,
				"status":     db.StatusMissing,
				"monitored":  req.Monitored,
			}).Error; err != nil {
				log.Printf("[ERROR] addBook: failed to restore book: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to restore book"})
			}
			s.db.Preload("Author").Preload("Series").First(&existing, existing.ID)
			log.Printf("[DEBUG] addBook: restored soft-deleted book '%s' (ID: %d)", existing.Title, existing.ID)
			return c.JSON(http.StatusCreated, bookToResponse(existing))
		}
		return c.JSON(http.StatusConflict, map[string]string{"error": "Book already exists"})
	}

	client := hardcover.NewClient(s.config.HardcoverAPIURL)
	bookData, err := client.GetBook(req.HardcoverID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch book from Hardcover: " + err.Error()})
	}

	// Create or find author
	var author db.Author
	if bookData.AuthorID != "" {
		if err := s.db.Where("hardcover_id = ?", bookData.AuthorID).First(&author).Error; err != nil {
			// Create new author
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

	// Create or find series
	var series *db.Series
	if bookData.SeriesID != "" {
		var seriesRecord db.Series
		if err := s.db.Where("hardcover_id = ?", bookData.SeriesID).First(&seriesRecord).Error; err != nil {
			// Note: We'd need to fetch series data from Hardcover
			seriesRecord = db.Series{
				HardcoverID: bookData.SeriesID,
				Name:        bookData.SeriesName,
			}
			s.db.Create(&seriesRecord)
		}
		series = &seriesRecord
	}

	// Create the book
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

	// Reload with associations
	s.db.Preload("Author").Preload("Series").First(&book, book.ID)

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

// DeleteBookResponse contains metadata for cache invalidation
type DeleteBookResponse struct {
	Message     string `json:"message"`
	BookID      uint   `json:"bookId"`
	HardcoverID string `json:"hardcoverId"`
	AuthorID    uint   `json:"authorId,omitempty"`
	SeriesID    *uint  `json:"seriesId,omitempty"`
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

	// Capture metadata before deletion for cache invalidation
	response := DeleteBookResponse{
		Message:     "Book deleted successfully",
		BookID:      book.ID,
		HardcoverID: book.HardcoverID,
		AuthorID:    book.AuthorID,
		SeriesID:    book.SeriesID,
	}

	// Soft delete media files (move to recycle bin)
	s.db.Where("book_id = ?", id).Delete(&db.MediaFile{})

	// Delete the book
	if err := s.db.Delete(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete book"})
	}

	return c.JSON(http.StatusOK, response)
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

func (s *Server) bulkDeleteBooks(c echo.Context) error {
	var req BulkDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if len(req.BookIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No book IDs provided"})
	}

	if req.DeleteFiles {
		s.db.Where("book_id IN ?", req.BookIDs).Delete(&db.MediaFile{})
	} else {
		s.db.Where("book_id IN ?", req.BookIDs).Delete(&db.MediaFile{})
	}

	result := s.db.Where("id IN ?", req.BookIDs).Delete(&db.Book{})
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete books"})
	}

	return c.JSON(http.StatusOK, map[string]int64{"deleted": result.RowsAffected})
}

func (s *Server) getBookEditions(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	var editions []db.Edition
	if err := s.db.Where("book_id = ?", id).Preload("Publisher").Find(&editions).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch editions"})
	}

	type EditionResp struct {
		ID            uint   `json:"id"`
		HardcoverID   string `json:"hardcoverId"`
		Format        string `json:"format"`
		EditionFormat string `json:"editionFormat,omitempty"`
		ISBN10        string `json:"isbn10,omitempty"`
		ISBN13        string `json:"isbn13,omitempty"`
		ASIN          string `json:"asin,omitempty"`
		Title         string `json:"title,omitempty"`
		Subtitle      string `json:"subtitle,omitempty"`
		LanguageCode  string `json:"languageCode,omitempty"`
		Language      string `json:"language,omitempty"`
		PublisherName string `json:"publisherName,omitempty"`
		PageCount     int    `json:"pageCount,omitempty"`
		AudioSeconds  int    `json:"audioSeconds,omitempty"`
		ReleaseDate   string `json:"releaseDate,omitempty"`
		CoverURL      string `json:"coverUrl,omitempty"`
	}

	editionResps := make([]EditionResp, 0, len(editions))
	for _, ed := range editions {
		resp := EditionResp{
			ID:            ed.ID,
			HardcoverID:   ed.HardcoverID,
			Format:        ed.Format,
			EditionFormat: ed.EditionFormat,
			ISBN10:        ed.ISBN10,
			ISBN13:        ed.ISBN13,
			ASIN:          ed.ASIN,
			Title:         ed.Title,
			Subtitle:      ed.Subtitle,
			LanguageCode:  ed.LanguageCode,
			Language:      ed.Language,
			PublisherName: ed.PublisherName,
			PageCount:     ed.PageCount,
			AudioSeconds:  ed.AudioSeconds,
			CoverURL:      ed.CoverURL,
		}
		if ed.ReleaseDate != nil {
			resp.ReleaseDate = ed.ReleaseDate.Format("2006-01-02")
		}
		editionResps = append(editionResps, resp)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"bookId":    book.ID,
		"bookTitle": book.Title,
		"editions":  editionResps,
	})
}

func (s *Server) getBookContributors(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	var contributors []db.Contributor
	if err := s.db.Where("book_id = ?", id).Preload("Author").Order("position ASC").Find(&contributors).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch contributors"})
	}

	type ContributorResp struct {
		ID          uint   `json:"id"`
		AuthorID    uint   `json:"authorId"`
		AuthorName  string `json:"authorName"`
		AuthorImage string `json:"authorImage,omitempty"`
		Role        string `json:"role"`
		Position    int    `json:"position"`
	}

	contribResps := make([]ContributorResp, 0, len(contributors))
	for _, c := range contributors {
		contribResps = append(contribResps, ContributorResp{
			ID:          c.ID,
			AuthorID:    c.AuthorID,
			AuthorName:  c.Author.Name,
			AuthorImage: c.Author.ImageURL,
			Role:        string(c.Role),
			Position:    c.Position,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"bookId":       book.ID,
		"bookTitle":    book.Title,
		"contributors": contribResps,
	})
}

func (s *Server) refreshBookMetadata(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	var book db.Book
	if err := s.db.First(&book, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	if book.HardcoverID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Book has no Hardcover ID"})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create Hardcover client"})
	}

	bookData, err := client.GetBook(book.HardcoverID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch book from Hardcover: " + err.Error()})
	}

	s.updateBookFromHardcover(&book, bookData)

	if err := s.db.Save(&book).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save book"})
	}

	s.syncGenres(&book, bookData.Genres)
	s.syncEditions(&book, bookData)
	s.syncContributors(&book, bookData)

	return c.JSON(http.StatusOK, map[string]any{
		"message":  "Metadata refreshed",
		"bookId":   book.ID,
		"title":    book.Title,
		"editions": len(bookData.Editions),
	})
}

func (s *Server) getGenres(c echo.Context) error {
	type GenreResponse struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Slug      string `json:"slug"`
		BookCount int64  `json:"bookCount"`
	}

	var genres []db.Genre
	if err := s.db.Find(&genres).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch genres"})
	}

	responses := make([]GenreResponse, 0, len(genres))
	for _, g := range genres {
		var count int64
		s.db.Model(&db.Book{}).Joins("JOIN book_genres ON book_genres.book_id = books.id").
			Where("book_genres.genre_id = ?", g.ID).Count(&count)

		responses = append(responses, GenreResponse{
			ID:        g.ID,
			Name:      g.Name,
			Slug:      g.Slug,
			BookCount: count,
		})
	}

	return c.JSON(http.StatusOK, responses)
}

func (s *Server) updateBookFromHardcover(book *db.Book, data *hardcover.BookData) {
	book.Title = data.Title
	book.Subtitle = data.Subtitle
	book.Headline = data.Headline
	book.Slug = data.Slug
	book.ISBN = data.ISBN
	book.ISBN13 = data.ISBN13
	book.Description = data.Description
	book.CoverURL = data.CoverURL
	book.Rating = data.Rating
	book.RatingsCount = data.RatingsCount
	book.ReviewsCount = data.ReviewsCount
	book.ReleaseDate = data.ReleaseDate
	book.ReleaseYear = data.ReleaseYear
	book.PageCount = data.PageCount
	book.LanguageCode = data.LanguageCode
	book.Language = data.Language
	book.AudioDuration = data.AudioDuration
	book.HasEbook = data.HasEbook
	book.HasAudiobook = data.HasAudiobook
	book.HasPhysical = data.HasPhysical
	book.EditionCount = data.EditionCount
	book.EbookEditionCount = data.EbookEditionCount
	book.AudiobookEditionCount = data.AudiobookEditionCount
	book.PhysicalEditionCount = data.PhysicalEditionCount
	book.LiteraryType = data.LiteraryType
	book.Category = data.Category
	book.Compilation = data.Compilation

	now := timeNow()
	book.LastSyncedAt = &now
}
