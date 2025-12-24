package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
)

type HardcoverBookResponse struct {
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	Description          string        `json:"description,omitempty"`
	CoverURL             string        `json:"coverUrl,omitempty"`
	Rating               float32       `json:"rating"`
	ReleaseDate          string        `json:"releaseDate,omitempty"`
	ReleaseYear          int           `json:"releaseYear,omitempty"`
	PageCount            int           `json:"pageCount,omitempty"`
	ISBN                 string        `json:"isbn,omitempty"`
	ISBN13               string        `json:"isbn13,omitempty"`
	AuthorID             string        `json:"authorId,omitempty"`
	AuthorName           string        `json:"authorName,omitempty"`
	AuthorImage          string        `json:"authorImage,omitempty"`
	SeriesID             string        `json:"seriesId,omitempty"`
	SeriesName           string        `json:"seriesName,omitempty"`
	SeriesIndex          *float32      `json:"seriesIndex,omitempty"`
	Genres               []string      `json:"genres,omitempty"`
	LanguageCode         string        `json:"languageCode,omitempty"`
	HasAudiobook         bool          `json:"hasAudiobook"`
	HasEbook             bool          `json:"hasEbook"`
	HasDigitalEdition    bool          `json:"hasDigitalEdition"`
	DigitalEditionCount  int           `json:"digitalEditionCount"`
	PhysicalEditionCount int           `json:"physicalEditionCount"`
	EditionCount         int           `json:"editionCount,omitempty"`
	AudioDuration        int           `json:"audioDuration,omitempty"`
	InLibrary            bool          `json:"inLibrary"`
	LibraryBook          *BookResponse `json:"libraryBook,omitempty"`
}

type HardcoverAuthorResponse struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Biography         string                  `json:"biography,omitempty"`
	ImageURL          string                  `json:"imageUrl,omitempty"`
	BooksCount        int                     `json:"booksCount"`
	TotalBooksCount   int                     `json:"totalBooksCount"`
	DigitalBooksCount int                     `json:"digitalBooksCount"`
	PhysicalOnlyCount int                     `json:"physicalOnlyCount"`
	InLibrary         bool                    `json:"inLibrary"`
	Books             []HardcoverBookResponse `json:"books,omitempty"`
}

type HardcoverSeriesResponse struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	BooksCount        int                     `json:"booksCount"`
	TotalBooksCount   int                     `json:"totalBooksCount"`
	DigitalBooksCount int                     `json:"digitalBooksCount"`
	PhysicalOnlyCount int                     `json:"physicalOnlyCount"`
	InLibrary         bool                    `json:"inLibrary"`
	Books             []HardcoverBookResponse `json:"books,omitempty"`
}

// getHardcoverClient creates a Hardcover client with the configured API key
func (s *Server) getHardcoverClient() (*hardcover.Client, error) {
	// Get API key from database first, fallback to config
	apiKey := s.config.HardcoverAPIKey
	var setting db.Setting
	if err := s.db.Where("key = ?", "hardcover_api_key").First(&setting).Error; err == nil && setting.Value != "" {
		apiKey = setting.Value
	}

	if apiKey == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Hardcover.app API key not configured")
	}

	return hardcover.NewClientWithAPIKey(s.config.HardcoverAPIURL, apiKey), nil
}

// getHardcoverBook returns detailed book info from Hardcover
func (s *Server) getHardcoverBook(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Book ID is required"})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to initialize Hardcover client: " + err.Error(),
		})
	}

	book, err := client.GetBook(id)
	if err != nil {
		// Provide more detailed error information
		errMsg := err.Error()
		if strings.Contains(errMsg, "authentication failed") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Hardcover API authentication failed. Please check your API key in Settings.",
			})
		}
		if strings.Contains(errMsg, "book not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Book not found in Hardcover database",
			})
		}
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error": "Failed to fetch book from Hardcover: " + errMsg,
		})
	}

	// Check if in library
	var count int64
	s.db.Model(&db.Book{}).Where("hardcover_id = ?", book.ID).Count(&count)

	releaseDate := ""
	releaseYear := 0
	if book.ReleaseDate != nil {
		releaseDate = book.ReleaseDate.Format("2006-01-02")
		releaseYear = book.ReleaseDate.Year()
	}

	return c.JSON(http.StatusOK, HardcoverBookResponse{
		ID:                   book.ID,
		Title:                book.Title,
		Description:          book.Description,
		CoverURL:             book.CoverURL,
		Rating:               book.Rating,
		ReleaseDate:          releaseDate,
		ReleaseYear:          releaseYear,
		PageCount:            book.PageCount,
		ISBN:                 book.ISBN,
		ISBN13:               book.ISBN13,
		AuthorID:             book.AuthorID,
		AuthorName:           book.AuthorName,
		AuthorImage:          book.AuthorImage,
		SeriesID:             book.SeriesID,
		SeriesName:           book.SeriesName,
		SeriesIndex:          book.SeriesIndex,
		Genres:               book.Genres,
		LanguageCode:         book.LanguageCode,
		HasAudiobook:         book.HasAudiobook,
		HasEbook:             book.HasEbook,
		HasDigitalEdition:    book.HasDigitalEdition,
		DigitalEditionCount:  book.DigitalEditionCount,
		PhysicalEditionCount: book.PhysicalEditionCount,
		EditionCount:         book.EditionCount,
		AudioDuration:        book.AudioDuration,
		InLibrary:            count > 0,
	})
}

func (s *Server) getHardcoverAuthor(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Author ID is required"})
	}

	showPhysical := c.QueryParam("showPhysical") == "true"

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	author, err := client.GetAuthor(id)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch author: " + err.Error()})
	}

	languages := s.GetPreferredLanguages()
	result, err := client.GetBooksByAuthor(id, languages, showPhysical)
	if err != nil {
		result = &hardcover.FilteredBooksResult{}
	}

	result = s.enrichBooksWithGoogleBooks(result, author.Name)

	var authorCount int64
	s.db.Model(&db.Author{}).Where("hardcover_id = ?", author.ID).Count(&authorCount)

	bookResponses := make([]HardcoverBookResponse, len(result.Books))
	for i, book := range result.Books {
		var libBook db.Book
		err := s.db.Where("hardcover_id = ?", book.ID).First(&libBook).Error
		inLibrary := err == nil

		releaseDate := ""
		releaseYear := 0
		if book.ReleaseDate != nil {
			releaseDate = book.ReleaseDate.Format("2006-01-02")
			releaseYear = book.ReleaseDate.Year()
		}

		resp := HardcoverBookResponse{
			ID:                   book.ID,
			Title:                book.Title,
			Description:          book.Description,
			CoverURL:             book.CoverURL,
			Rating:               book.Rating,
			ReleaseDate:          releaseDate,
			ReleaseYear:          releaseYear,
			PageCount:            book.PageCount,
			ISBN:                 book.ISBN,
			ISBN13:               book.ISBN13,
			SeriesID:             book.SeriesID,
			SeriesName:           book.SeriesName,
			SeriesIndex:          book.SeriesIndex,
			Genres:               book.Genres,
			LanguageCode:         book.LanguageCode,
			HasDigitalEdition:    book.HasDigitalEdition,
			DigitalEditionCount:  book.DigitalEditionCount,
			PhysicalEditionCount: book.PhysicalEditionCount,
			InLibrary:            inLibrary,
		}

		if inLibrary {
			libResp := bookToResponse(libBook)
			resp.LibraryBook = &libResp
		}
		bookResponses[i] = resp
	}

	return c.JSON(http.StatusOK, HardcoverAuthorResponse{
		ID:                author.ID,
		Name:              author.Name,
		Biography:         author.Biography,
		ImageURL:          author.ImageURL,
		BooksCount:        len(result.Books),
		TotalBooksCount:   result.TotalCount,
		DigitalBooksCount: result.DigitalCount,
		PhysicalOnlyCount: result.PhysicalOnlyCount,
		InLibrary:         authorCount > 0,
		Books:             bookResponses,
	})
}

func (s *Server) getHardcoverSeries(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Series ID is required"})
	}

	showPhysical := c.QueryParam("showPhysical") == "true"

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	languages := s.GetPreferredLanguages()
	result, err := client.GetSeries(id, languages, showPhysical)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch series: " + err.Error()})
	}

	var seriesCount int64
	s.db.Model(&db.Series{}).Where("hardcover_id = ?", result.Series.ID).Count(&seriesCount)

	bookResponses := make([]HardcoverBookResponse, len(result.Books))
	for i, book := range result.Books {
		var libBook db.Book
		err := s.db.Where("hardcover_id = ?", book.ID).First(&libBook).Error
		inLibrary := err == nil

		releaseDate := ""
		releaseYear := 0
		if book.ReleaseDate != nil {
			releaseDate = book.ReleaseDate.Format("2006-01-02")
			releaseYear = book.ReleaseDate.Year()
		}

		resp := HardcoverBookResponse{
			ID:                   book.ID,
			Title:                book.Title,
			Description:          book.Description,
			CoverURL:             book.CoverURL,
			Rating:               book.Rating,
			ReleaseDate:          releaseDate,
			ReleaseYear:          releaseYear,
			PageCount:            book.PageCount,
			ISBN:                 book.ISBN,
			ISBN13:               book.ISBN13,
			AuthorID:             book.AuthorID,
			AuthorName:           book.AuthorName,
			SeriesIndex:          book.SeriesIndex,
			Genres:               book.Genres,
			LanguageCode:         book.LanguageCode,
			HasDigitalEdition:    book.HasDigitalEdition,
			DigitalEditionCount:  book.DigitalEditionCount,
			PhysicalEditionCount: book.PhysicalEditionCount,
			InLibrary:            inLibrary,
		}

		if inLibrary {
			libResp := bookToResponse(libBook)
			resp.LibraryBook = &libResp
		}
		bookResponses[i] = resp
	}

	return c.JSON(http.StatusOK, HardcoverSeriesResponse{
		ID:                result.Series.ID,
		Name:              result.Series.Name,
		BooksCount:        len(result.Books),
		TotalBooksCount:   result.TotalCount,
		DigitalBooksCount: result.DigitalCount,
		PhysicalOnlyCount: result.PhysicalOnlyCount,
		InLibrary:         seriesCount > 0,
		Books:             bookResponses,
	})
}

// addHardcoverBook adds a book from Hardcover to the library
func (s *Server) addHardcoverBook(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Book ID is required"})
	}

	var req struct {
		Monitored bool   `json:"monitored"`
		MediaType string `json:"mediaType"` // ebook, audiobook, both
	}
	if err := c.Bind(&req); err != nil {
		req.Monitored = true
		req.MediaType = "both"
	}

	// Check if already exists
	var existing db.Book
	if err := s.db.Where("hardcover_id = ?", id).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Book already in library", "bookId": strconv.FormatUint(uint64(existing.ID), 10)})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	book, err := client.GetBook(id)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch book: " + err.Error()})
	}

	// Create or get author
	var author db.Author
	if book.AuthorID != "" {
		if err := s.db.Where("hardcover_id = ?", book.AuthorID).First(&author).Error; err != nil {
			// Create author
			author = db.Author{
				HardcoverID: book.AuthorID,
				Name:        book.AuthorName,
				SortName:    book.AuthorName,
			}
			s.db.Create(&author)
		}
	}

	// Create or get series
	var seriesID *uint
	if book.SeriesID != "" {
		var series db.Series
		if err := s.db.Where("hardcover_id = ?", book.SeriesID).First(&series).Error; err != nil {
			// Create series
			series = db.Series{
				HardcoverID: book.SeriesID,
				Name:        book.SeriesName,
			}
			s.db.Create(&series)
		}
		seriesID = &series.ID
	}

	// Create book
	newBook := db.Book{
		HardcoverID: book.ID,
		Title:       book.Title,
		SortTitle:   book.SortTitle,
		ISBN:        book.ISBN,
		ISBN13:      book.ISBN13,
		Description: book.Description,
		CoverURL:    book.CoverURL,
		Rating:      book.Rating,
		ReleaseDate: book.ReleaseDate,
		PageCount:   book.PageCount,
		AuthorID:    author.ID,
		SeriesID:    seriesID,
		SeriesIndex: book.SeriesIndex,
		Status:      db.StatusMissing,
		Monitored:   req.Monitored,
	}

	if err := s.db.Create(&newBook).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add book"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Book added to library",
		"bookId":  newBook.ID,
	})
}

func (s *Server) enrichBooksWithGoogleBooks(result *hardcover.FilteredBooksResult, authorName string) *hardcover.FilteredBooksResult {
	gbClient := s.getGoogleBooksClient()
	if gbClient == nil || result == nil {
		return result
	}

	for i := range result.Books {
		book := &result.Books[i]
		if book.HasDigitalEdition {
			continue
		}

		query := book.Title
		if authorName != "" {
			query = "intitle:" + book.Title + " inauthor:" + authorName
		}

		books, err := gbClient.SearchVolumes(query, 3)
		if err != nil {
			continue
		}

		for _, gbBook := range books {
			if gbBook.IsEbook || gbBook.HasEpub {
				book.HasDigitalEdition = true
				book.HasEbook = true
				result.DigitalCount++
				result.PhysicalOnlyCount--
				break
			}
		}
	}

	return result
}
