package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
)

type EditionResponse struct {
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

type ContributorResponse struct {
	AuthorID    uint   `json:"authorId"`
	AuthorName  string `json:"authorName"`
	AuthorImage string `json:"authorImage,omitempty"`
	Role        string `json:"role"`
	Position    int    `json:"position"`
}

type HardcoverBookResponse struct {
	ID                    string                `json:"id"`
	Title                 string                `json:"title"`
	Subtitle              string                `json:"subtitle,omitempty"`
	Description           string                `json:"description,omitempty"`
	CoverURL              string                `json:"coverUrl,omitempty"`
	Rating                float32               `json:"rating"`
	ReleaseDate           string                `json:"releaseDate,omitempty"`
	ReleaseYear           int                   `json:"releaseYear,omitempty"`
	PageCount             int                   `json:"pageCount,omitempty"`
	ISBN                  string                `json:"isbn,omitempty"`
	ISBN13                string                `json:"isbn13,omitempty"`
	AuthorID              string                `json:"authorId,omitempty"`
	AuthorName            string                `json:"authorName,omitempty"`
	AuthorImage           string                `json:"authorImage,omitempty"`
	SeriesID              string                `json:"seriesId,omitempty"`
	SeriesName            string                `json:"seriesName,omitempty"`
	SeriesIndex           *float32              `json:"seriesIndex,omitempty"`
	Genres                []string              `json:"genres,omitempty"`
	LanguageCode          string                `json:"languageCode,omitempty"`
	Language              string                `json:"language,omitempty"`
	HasAudiobook          bool                  `json:"hasAudiobook"`
	HasEbook              bool                  `json:"hasEbook"`
	HasPhysical           bool                  `json:"hasPhysical"`
	HasDigitalEdition     bool                  `json:"hasDigitalEdition"`
	DigitalEditionCount   int                   `json:"digitalEditionCount"`
	PhysicalEditionCount  int                   `json:"physicalEditionCount"`
	EbookEditionCount     int                   `json:"ebookEditionCount"`
	AudiobookEditionCount int                   `json:"audiobookEditionCount"`
	EditionCount          int                   `json:"editionCount,omitempty"`
	AudioDuration         int                   `json:"audioDuration,omitempty"`
	Compilation           bool                  `json:"compilation"`
	Editions              []EditionResponse     `json:"editions,omitempty"`
	Contributors          []ContributorResponse `json:"contributors,omitempty"`
	InLibrary             bool                  `json:"inLibrary"`
	LibraryBook           *BookResponse         `json:"libraryBook,omitempty"`
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

	var libBook db.Book
	err = s.db.Where("hardcover_id = ?", book.ID).First(&libBook).Error
	inLibrary := err == nil

	releaseDate := ""
	releaseYear := book.ReleaseYear
	if book.ReleaseDate != nil {
		releaseDate = book.ReleaseDate.Format("2006-01-02")
		if releaseYear == 0 {
			releaseYear = book.ReleaseDate.Year()
		}
	}

	editions := make([]EditionResponse, 0, len(book.Editions))
	for _, ed := range book.Editions {
		edResp := EditionResponse{
			HardcoverID:   ed.ID,
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
			edResp.ReleaseDate = ed.ReleaseDate.Format("2006-01-02")
		}
		editions = append(editions, edResp)
	}

	contributors := make([]ContributorResponse, 0, len(book.Contributors))
	for _, c := range book.Contributors {
		contributors = append(contributors, ContributorResponse{
			AuthorName:  c.AuthorName,
			AuthorImage: c.AuthorImage,
			Role:        c.Role,
			Position:    c.Position,
		})
	}

	resp := HardcoverBookResponse{
		ID:                    book.ID,
		Title:                 book.Title,
		Subtitle:              book.Subtitle,
		Description:           book.Description,
		CoverURL:              book.CoverURL,
		Rating:                book.Rating,
		ReleaseDate:           releaseDate,
		ReleaseYear:           releaseYear,
		PageCount:             book.PageCount,
		ISBN:                  book.ISBN,
		ISBN13:                book.ISBN13,
		AuthorID:              book.AuthorID,
		AuthorName:            book.AuthorName,
		AuthorImage:           book.AuthorImage,
		SeriesID:              book.SeriesID,
		SeriesName:            book.SeriesName,
		SeriesIndex:           book.SeriesIndex,
		Genres:                book.Genres,
		LanguageCode:          book.LanguageCode,
		Language:              book.Language,
		HasAudiobook:          book.HasAudiobook,
		HasEbook:              book.HasEbook,
		HasPhysical:           book.HasPhysical,
		HasDigitalEdition:     book.HasDigitalEdition,
		DigitalEditionCount:   book.DigitalEditionCount,
		PhysicalEditionCount:  book.PhysicalEditionCount,
		EbookEditionCount:     book.EbookEditionCount,
		AudiobookEditionCount: book.AudiobookEditionCount,
		EditionCount:          book.EditionCount,
		AudioDuration:         book.AudioDuration,
		Compilation:           book.Compilation,
		Editions:              editions,
		Contributors:          contributors,
		InLibrary:             inLibrary,
	}

	if inLibrary {
		libResp := bookToResponse(libBook)
		resp.LibraryBook = &libResp
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) getHardcoverAuthor(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Author ID is required"})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	author, err := client.GetAuthor(id)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch author: " + err.Error()})
	}

	languages := s.GetPreferredLanguages()
	result, err := client.GetBooksByAuthor(id, languages)
	if err != nil {
		result = &hardcover.FilteredBooksResult{}
	}

	var authorCount int64
	s.db.Model(&db.Author{}).Where("hardcover_id = ?", author.ID).Count(&authorCount)

	// Debug: log how many books have series info
	booksWithSeries := 0
	for _, b := range result.Books {
		if b.SeriesID != "" {
			booksWithSeries++
		}
	}
	log.Printf("[DEBUG] getHardcoverAuthor: Author '%s' has %d books, %d with series info", author.Name, len(result.Books), booksWithSeries)

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

		// Debug: log series info for books with series
		if book.SeriesID != "" {
			seriesIdx := "nil"
			if book.SeriesIndex != nil {
				seriesIdx = fmt.Sprintf("%.1f", *book.SeriesIndex)
			}
			log.Printf("[DEBUG] getHardcoverAuthor: Book '%s' - Series: %s (#%s)", book.Title, book.SeriesName, seriesIdx)
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

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	languages := s.GetPreferredLanguages()
	result, err := client.GetSeries(id, languages)
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

func (s *Server) addHardcoverBook(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Book ID is required"})
	}

	var req struct {
		Monitored     bool   `json:"monitored"`
		MediaType     string `json:"mediaType"`
		ForceAuthorID uint   `json:"forceAuthorId"`
		ForceSeriesID uint   `json:"forceSeriesId"`
	}
	if err := c.Bind(&req); err != nil {
		req.Monitored = true
		req.MediaType = "both"
	}

	var existing db.Book
	if err := s.db.Unscoped().Where("hardcover_id = ?", id).First(&existing).Error; err == nil {
		if existing.DeletedAt.Valid {
			if err := s.db.Unscoped().Model(&existing).Updates(map[string]any{
				"deleted_at": nil,
				"status":     db.StatusMissing,
				"monitored":  req.Monitored,
			}).Error; err != nil {
				log.Printf("[ERROR] addHardcoverBook: failed to restore book: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to restore book"})
			}
			log.Printf("[DEBUG] addHardcoverBook: restored soft-deleted book '%s' (ID: %d)", existing.Title, existing.ID)
			return c.JSON(http.StatusCreated, map[string]any{
				"message": "Book restored to library",
				"bookId":  existing.ID,
			})
		}
		return c.JSON(http.StatusConflict, map[string]any{
			"error":       "Book already in library",
			"bookId":      existing.ID,
			"hardcoverId": existing.HardcoverID,
			"status":      existing.Status,
			"authorId":    existing.AuthorID,
			"seriesId":    existing.SeriesID,
		})
	}

	client, err := s.getHardcoverClient()
	if err != nil {
		return err
	}

	book, err := client.GetBook(id)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch book: " + err.Error()})
	}

	var authorID uint
	if req.ForceAuthorID > 0 {
		var forcedAuthor db.Author
		if err := s.db.First(&forcedAuthor, req.ForceAuthorID).Error; err == nil {
			authorID = forcedAuthor.ID
			log.Printf("[DEBUG] addHardcoverBook: using forced author ID %d for book '%s'", authorID, book.Title)
		}
	}
	if authorID == 0 && book.AuthorID != "" {
		authorID = s.getOrCreateAuthor(book)
	}

	var seriesID *uint
	if req.ForceSeriesID > 0 {
		var forcedSeries db.Series
		if err := s.db.First(&forcedSeries, req.ForceSeriesID).Error; err == nil {
			seriesID = &forcedSeries.ID
			log.Printf("[DEBUG] addHardcoverBook: using forced series ID %d for book '%s'", *seriesID, book.Title)
		}
	}
	if seriesID == nil && book.SeriesID != "" {
		seriesID = s.getOrCreateSeries(book)
	}

	now := timeNow()
	newBook := db.Book{
		HardcoverID:           book.ID,
		Title:                 book.Title,
		SortTitle:             book.SortTitle,
		Subtitle:              book.Subtitle,
		Headline:              book.Headline,
		Slug:                  book.Slug,
		ISBN:                  book.ISBN,
		ISBN13:                book.ISBN13,
		Description:           book.Description,
		CoverURL:              book.CoverURL,
		Rating:                book.Rating,
		RatingsCount:          book.RatingsCount,
		ReviewsCount:          book.ReviewsCount,
		ReleaseDate:           book.ReleaseDate,
		ReleaseYear:           book.ReleaseYear,
		PageCount:             book.PageCount,
		LanguageCode:          book.LanguageCode,
		Language:              book.Language,
		AudioDuration:         book.AudioDuration,
		HasEbook:              book.HasEbook,
		HasAudiobook:          book.HasAudiobook,
		HasPhysical:           book.HasPhysical,
		EditionCount:          book.EditionCount,
		EbookEditionCount:     book.EbookEditionCount,
		AudiobookEditionCount: book.AudiobookEditionCount,
		PhysicalEditionCount:  book.PhysicalEditionCount,
		LiteraryType:          book.LiteraryType,
		Category:              book.Category,
		Compilation:           book.Compilation,
		AuthorID:              authorID,
		SeriesID:              seriesID,
		SeriesIndex:           book.SeriesIndex,
		Status:                db.StatusMissing,
		Monitored:             req.Monitored,
		LastSyncedAt:          &now,
	}

	if err := s.db.Create(&newBook).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add book"})
	}

	s.syncGenres(&newBook, book.Genres)
	s.syncEditions(&newBook, book)
	s.syncContributors(&newBook, book)

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Book added to library",
		"bookId":  newBook.ID,
	})
}

func timeNow() time.Time {
	return time.Now()
}

func (s *Server) getOrCreateAuthor(book *hardcover.BookData) uint {
	var author db.Author
	if err := s.db.Where("hardcover_id = ?", book.AuthorID).First(&author).Error; err != nil {
		author = db.Author{
			HardcoverID: book.AuthorID,
			Name:        book.AuthorName,
			SortName:    book.AuthorName,
		}
		if len(book.Contributors) > 0 {
			for _, c := range book.Contributors {
				if c.AuthorID == book.AuthorID {
					author.Slug = c.AuthorSlug
					author.Biography = c.AuthorBio
					author.ImageURL = c.AuthorImage
					author.BornDate = c.BornDate
					author.BornYear = c.BornYear
					author.DeathDate = c.DeathDate
					author.DeathYear = c.DeathYear
					author.Location = c.Location
					author.IsBIPOC = c.IsBIPOC
					author.IsLGBTQ = c.IsLGBTQ
					break
				}
			}
		}
		s.db.Create(&author)
	}
	return author.ID
}

func (s *Server) getOrCreateSeries(book *hardcover.BookData) *uint {
	var series db.Series
	if err := s.db.Where("hardcover_id = ?", book.SeriesID).First(&series).Error; err != nil {
		series = db.Series{
			HardcoverID:       book.SeriesID,
			Name:              book.SeriesName,
			Slug:              book.SeriesSlug,
			IsCompleted:       book.SeriesIsCompleted,
			PrimaryBooksCount: book.SeriesPrimaryBooksCount,
		}
		if book.SeriesAuthorID != "" {
			var seriesAuthor db.Author
			if err := s.db.Where("hardcover_id = ?", book.SeriesAuthorID).First(&seriesAuthor).Error; err == nil {
				series.AuthorID = &seriesAuthor.ID
			}
		}
		s.db.Create(&series)
	}
	return &series.ID
}

func (s *Server) syncGenres(dbBook *db.Book, genres []string) {
	if len(genres) == 0 {
		return
	}

	var genreModels []*db.Genre
	for _, name := range genres {
		var genre db.Genre
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		if err := s.db.Where("name = ?", name).First(&genre).Error; err != nil {
			genre = db.Genre{Name: name, Slug: slug}
			s.db.Create(&genre)
		}
		genreModels = append(genreModels, &genre)
	}

	s.db.Model(dbBook).Association("Genres").Replace(genreModels)
}

func (s *Server) syncEditions(dbBook *db.Book, book *hardcover.BookData) {
	if len(book.Editions) == 0 {
		return
	}

	s.db.Where("book_id = ?", dbBook.ID).Delete(&db.Edition{})

	for _, ed := range book.Editions {
		var publisherID *uint
		if ed.PublisherID != "" && ed.PublisherName != "" {
			var publisher db.Publisher
			if err := s.db.Where("hardcover_id = ?", ed.PublisherID).First(&publisher).Error; err != nil {
				publisher = db.Publisher{HardcoverID: ed.PublisherID, Name: ed.PublisherName}
				s.db.Create(&publisher)
			}
			publisherID = &publisher.ID
		}

		edition := db.Edition{
			HardcoverID:   ed.ID,
			BookID:        dbBook.ID,
			ISBN10:        ed.ISBN10,
			ISBN13:        ed.ISBN13,
			ASIN:          ed.ASIN,
			Title:         ed.Title,
			Subtitle:      ed.Subtitle,
			EditionFormat: ed.EditionFormat,
			Format:        ed.Format,
			LanguageCode:  ed.LanguageCode,
			Language:      ed.Language,
			PublisherID:   publisherID,
			PublisherName: ed.PublisherName,
			PageCount:     ed.PageCount,
			AudioSeconds:  ed.AudioSeconds,
			ReleaseDate:   ed.ReleaseDate,
			CoverURL:      ed.CoverURL,
		}
		s.db.Create(&edition)
	}
}

func (s *Server) syncContributors(dbBook *db.Book, book *hardcover.BookData) {
	if len(book.Contributors) == 0 {
		return
	}

	s.db.Where("book_id = ?", dbBook.ID).Delete(&db.Contributor{})

	for _, c := range book.Contributors {
		var author db.Author
		if err := s.db.Where("hardcover_id = ?", c.AuthorID).First(&author).Error; err != nil {
			author = db.Author{
				HardcoverID: c.AuthorID,
				Name:        c.AuthorName,
				SortName:    c.AuthorName,
				Slug:        c.AuthorSlug,
				Biography:   c.AuthorBio,
				ImageURL:    c.AuthorImage,
				BornDate:    c.BornDate,
				BornYear:    c.BornYear,
				DeathDate:   c.DeathDate,
				DeathYear:   c.DeathYear,
				Location:    c.Location,
				IsBIPOC:     c.IsBIPOC,
				IsLGBTQ:     c.IsLGBTQ,
			}
			s.db.Create(&author)
		}

		role := db.ContributorRole(c.Role)
		if role == "" {
			role = db.RoleAuthor
		}

		contributor := db.Contributor{
			BookID:   dbBook.ID,
			AuthorID: author.ID,
			Role:     role,
			Position: c.Position,
		}
		s.db.Create(&contributor)
	}
}
