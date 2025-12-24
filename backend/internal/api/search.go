package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/googlebooks"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"github.com/shelfarr/shelfarr/internal/indexer"
	"github.com/shelfarr/shelfarr/internal/openlibrary"
)

// SearchResult represents a search result from Hardcover.app
type SearchResult struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	AuthorID    string  `json:"authorId"`
	CoverURL    string  `json:"coverUrl"`
	Rating      float32 `json:"rating"`
	ReleaseYear int     `json:"releaseYear,omitempty"`
	ISBN        string  `json:"isbn,omitempty"`
	Description string  `json:"description,omitempty"`
	InLibrary   bool    `json:"inLibrary"`
}

// AuthorSearchResult represents an author search result
type AuthorSearchResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ImageURL   string `json:"imageUrl"`
	BooksCount int    `json:"booksCount"`
	Biography  string `json:"biography,omitempty"`
	InLibrary  bool   `json:"inLibrary"`
}

// SeriesSearchResult represents a series search result
type SeriesSearchResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	BooksCount int    `json:"booksCount"`
	AuthorID   string `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	InLibrary  bool   `json:"inLibrary"`
}

// ListSearchResult represents a list search result
type ListSearchResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	BooksCount  int    `json:"booksCount"`
	Username    string `json:"username,omitempty"`
}

// UnifiedSearchResponse contains results from all search types
type UnifiedSearchResponse struct {
	Books   []SearchResult       `json:"books,omitempty"`
	Authors []AuthorSearchResult `json:"authors,omitempty"`
	Series  []SeriesSearchResult `json:"series,omitempty"`
	Lists   []ListSearchResult   `json:"lists,omitempty"`
}

// IndexerSearchResult represents a search result from an indexer
type IndexerSearchResult struct {
	Indexer     string `json:"indexer"`
	Title       string `json:"title"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
	Seeders     int    `json:"seeders,omitempty"`
	Leechers    int    `json:"leechers,omitempty"`
	DownloadURL string `json:"downloadUrl"`
	InfoURL     string `json:"infoUrl,omitempty"`
	PublishDate string `json:"publishDate,omitempty"`
	Quality     string `json:"quality"` // Calculated quality score
	Freeleech   bool   `json:"freeleech,omitempty"`
	VIP         bool   `json:"vip,omitempty"`
	Author      string `json:"author,omitempty"`
	Narrator    string `json:"narrator,omitempty"`
	Category    string `json:"category,omitempty"`
	LangCode    string `json:"langCode,omitempty"` // 3-letter language code
}

// searchHardcover searches Hardcover.app for books, authors, series, or lists
func (s *Server) searchHardcover(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Query parameter 'q' is required"})
	}

	searchType := c.QueryParam("type") // "book", "author", "series", "list", "all"
	if searchType == "" {
		searchType = "book"
	}

	// Get API key from database first, fallback to config
	apiKey := s.config.HardcoverAPIKey
	var setting db.Setting
	if err := s.db.Where("key = ?", "hardcover_api_key").First(&setting).Error; err == nil && setting.Value != "" {
		apiKey = setting.Value
	}

	// Check if API key is configured
	if apiKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Hardcover.app API key not configured. Please add your API key in Settings > Library Search Providers."})
	}

	// Create client with API key
	client := hardcover.NewClientWithAPIKey(s.config.HardcoverAPIURL, apiKey)

	// Handle unified search (all types)
	if searchType == "all" {
		return s.searchHardcoverAll(c, client, query)
	}

	// Handle specific type searches
	switch searchType {
	case "book":
		return s.searchHardcoverBooks(c, client, query)
	case "author":
		return s.searchHardcoverAuthors(c, client, query)
	case "series":
		return s.searchHardcoverSeries(c, client, query)
	case "list":
		return s.searchHardcoverLists(c, client, query)
	default:
		return s.searchHardcoverBooks(c, client, query)
	}
}

// searchHardcoverBooks searches for books
func (s *Server) searchHardcoverBooks(c echo.Context, client *hardcover.Client, query string) error {
	languages := s.GetPreferredLanguages()
	books, err := client.SearchBooks(query, languages)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []SearchResult
	for _, book := range books {
		// Check if already in library
		var count int64
		s.db.Model(&db.Book{}).Where("hardcover_id = ?", book.ID).Count(&count)

		results = append(results, SearchResult{
			ID:          book.ID,
			Title:       book.Title,
			Author:      book.AuthorName,
			AuthorID:    book.AuthorID,
			CoverURL:    book.CoverURL,
			Rating:      book.Rating,
			ReleaseYear: book.ReleaseYear,
			ISBN:        book.ISBN,
			Description: book.Description,
			InLibrary:   count > 0,
		})
	}

	return c.JSON(http.StatusOK, results)
}

// searchHardcoverAuthors searches for authors
func (s *Server) searchHardcoverAuthors(c echo.Context, client *hardcover.Client, query string) error {
	authors, err := client.SearchAuthors(query)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []AuthorSearchResult
	for _, author := range authors {
		// Check if already in library
		var count int64
		s.db.Model(&db.Author{}).Where("hardcover_id = ?", author.ID).Count(&count)

		results = append(results, AuthorSearchResult{
			ID:         author.ID,
			Name:       author.Name,
			ImageURL:   author.ImageURL,
			BooksCount: author.BooksCount,
			Biography:  author.Biography,
			InLibrary:  count > 0,
		})
	}

	return c.JSON(http.StatusOK, results)
}

// searchHardcoverSeries searches for series
func (s *Server) searchHardcoverSeries(c echo.Context, client *hardcover.Client, query string) error {
	seriesList, err := client.SearchSeries(query)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []SeriesSearchResult
	for _, series := range seriesList {
		// Check if already in library
		var count int64
		s.db.Model(&db.Series{}).Where("hardcover_id = ?", series.ID).Count(&count)

		results = append(results, SeriesSearchResult{
			ID:         series.ID,
			Name:       series.Name,
			BooksCount: series.BooksCount,
			AuthorID:   series.AuthorID,
			AuthorName: series.AuthorName,
			InLibrary:  count > 0,
		})
	}

	return c.JSON(http.StatusOK, results)
}

// searchHardcoverLists searches for lists
func (s *Server) searchHardcoverLists(c echo.Context, client *hardcover.Client, query string) error {
	lists, err := client.SearchLists(query)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []ListSearchResult
	for _, list := range lists {
		results = append(results, ListSearchResult{
			ID:          list.ID,
			Name:        list.Name,
			Description: list.Description,
			BooksCount:  list.BooksCount,
			Username:    list.Username,
		})
	}

	return c.JSON(http.StatusOK, results)
}

// searchHardcoverAll performs a unified search across all types
func (s *Server) searchHardcoverAll(c echo.Context, client *hardcover.Client, query string) error {
	// Use the client's SearchAll which handles errors properly
	languages := s.GetPreferredLanguages()
	results, err := client.SearchAll(query, languages)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	response := UnifiedSearchResponse{}

	// Process books and check library status
	for _, book := range results.Books {
		var count int64
		s.db.Model(&db.Book{}).Where("hardcover_id = ?", book.ID).Count(&count)

		response.Books = append(response.Books, SearchResult{
			ID:          book.ID,
			Title:       book.Title,
			Author:      book.AuthorName,
			AuthorID:    book.AuthorID,
			CoverURL:    book.CoverURL,
			Rating:      book.Rating,
			ReleaseYear: book.ReleaseYear,
			ISBN:        book.ISBN,
			Description: book.Description,
			InLibrary:   count > 0,
		})
	}

	// Process authors and check library status
	for _, author := range results.Authors {
		var count int64
		s.db.Model(&db.Author{}).Where("hardcover_id = ?", author.ID).Count(&count)

		response.Authors = append(response.Authors, AuthorSearchResult{
			ID:         author.ID,
			Name:       author.Name,
			ImageURL:   author.ImageURL,
			BooksCount: author.BooksCount,
			Biography:  author.Biography,
			InLibrary:  count > 0,
		})
	}

	// Process series and check library status
	for _, series := range results.Series {
		var count int64
		s.db.Model(&db.Series{}).Where("hardcover_id = ?", series.ID).Count(&count)

		response.Series = append(response.Series, SeriesSearchResult{
			ID:         series.ID,
			Name:       series.Name,
			BooksCount: series.BooksCount,
			AuthorID:   series.AuthorID,
			AuthorName: series.AuthorName,
			InLibrary:  count > 0,
		})
	}

	// Process lists (no library status check needed)
	for _, list := range results.Lists {
		response.Lists = append(response.Lists, ListSearchResult{
			ID:          list.ID,
			Name:        list.Name,
			Description: list.Description,
			BooksCount:  list.BooksCount,
			Username:    list.Username,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// testHardcover tests the Hardcover.app API connection
func (s *Server) testHardcover(c echo.Context) error {
	// Get API key from database first, fallback to config
	apiKey := s.config.HardcoverAPIKey
	var setting db.Setting
	if err := s.db.Where("key = ?", "hardcover_api_key").First(&setting).Error; err == nil && setting.Value != "" {
		apiKey = setting.Value
	}

	// Check if API key is configured
	if apiKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Hardcover.app API key not configured"})
	}

	// Create client and test connection
	client := hardcover.NewClientWithAPIKey(s.config.HardcoverAPIURL, apiKey)
	if err := client.Test(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Connection failed: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Connection successful"})
}

// searchIndexers searches configured indexers for downloads
func (s *Server) searchIndexers(c echo.Context) error {
	bookID := c.QueryParam("bookId")
	query := c.QueryParam("q")
	mediaType := c.QueryParam("mediaType") // "ebook" or "audiobook"

	log.Printf("[DEBUG] searchIndexers called: bookId=%s, q=%s, mediaType=%s", bookID, query, mediaType)

	if bookID == "" && query == "" {
		log.Printf("[DEBUG] searchIndexers: missing required parameters")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either 'bookId' or 'q' parameter is required"})
	}

	// Build search query
	var searchQuery indexer.SearchQuery
	searchQuery.MediaType = mediaType

	// If bookId provided, get book details for search
	if bookID != "" {
		var book db.Book
		if err := s.db.Preload("Author").First(&book, bookID).Error; err != nil {
			log.Printf("[DEBUG] searchIndexers: book not found with id=%s, error=%v", bookID, err)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
		}
		searchQuery.Title = book.Title
		searchQuery.Author = book.Author.Name
		searchQuery.ISBN = book.ISBN
		searchQuery.BookID = book.HardcoverID
		log.Printf("[DEBUG] searchIndexers: searching for book '%s' by '%s' (ISBN: %s)", searchQuery.Title, searchQuery.Author, searchQuery.ISBN)
	} else {
		searchQuery.Title = query
		log.Printf("[DEBUG] searchIndexers: free-text search for '%s'", query)
	}

	// Load enabled indexers from database
	var dbIndexers []db.Indexer
	if err := s.db.Where("enabled = ?", true).Order("priority ASC").Find(&dbIndexers).Error; err != nil {
		log.Printf("[DEBUG] searchIndexers: failed to load indexers, error=%v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load indexers"})
	}

	log.Printf("[DEBUG] searchIndexers: found %d enabled indexers", len(dbIndexers))
	for _, idx := range dbIndexers {
		log.Printf("[DEBUG]   - Indexer: %s (type=%s, enabled=%v)", idx.Name, idx.Type, idx.Enabled)
	}

	if len(dbIndexers) == 0 {
		log.Printf("[DEBUG] searchIndexers: no enabled indexers configured, returning empty results")
		return c.JSON(http.StatusOK, []IndexerSearchResult{})
	}

	// Create indexer manager and add indexers
	manager := indexer.NewManager()
	for _, dbIdx := range dbIndexers {
		idx := createIndexerFromDB(dbIdx)
		if idx != nil {
			manager.AddIndexer(idx)
		}
	}

	// Perform search with timeout
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	log.Printf("[DEBUG] searchIndexers: starting search across %d indexers", len(dbIndexers))
	results, err := manager.SearchAll(ctx, searchQuery)
	if err != nil {
		log.Printf("[DEBUG] searchIndexers: search failed, error=%v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Search failed: " + err.Error()})
	}

	log.Printf("[DEBUG] searchIndexers: received %d total results from indexers", len(results))

	// Convert to API response format
	apiResults := make([]IndexerSearchResult, 0, len(results))
	for _, r := range results {
		apiResults = append(apiResults, IndexerSearchResult{
			Indexer:     r.Indexer,
			Title:       r.Title,
			Size:        r.Size,
			Format:      r.Format,
			Seeders:     r.Seeders,
			Leechers:    r.Leechers,
			DownloadURL: r.DownloadURL,
			InfoURL:     r.InfoURL,
			PublishDate: r.PublishDate,
			Quality:     calculateQualityLabel(r),
			Freeleech:   r.Freeleech,
			VIP:         r.VIP,
			Author:      r.Author,
			Narrator:    r.Narrator,
			Category:    r.Category,
			LangCode:    r.LangCode,
		})
	}

	return c.JSON(http.StatusOK, apiResults)
}

// createIndexerFromDB creates an Indexer instance from database model
func createIndexerFromDB(dbIdx db.Indexer) indexer.Indexer {
	switch dbIdx.Type {
	case "mam":
		return indexer.NewMAMIndexer(dbIdx.Name, dbIdx.Cookie, dbIdx.VIPOnly, dbIdx.FreeleechOnly)
	case "torznab":
		return indexer.NewTorznabIndexer(dbIdx.Name, dbIdx.URL, dbIdx.APIKey)
	case "anna":
		return indexer.NewAnnaIndexer(dbIdx.Name)
	default:
		return nil
	}
}

// calculateQualityLabel generates a quality label based on result attributes
func calculateQualityLabel(r indexer.SearchResult) string {
	// Score based on seeders, format, and freeleech status
	if r.Seeders >= 10 && (r.Format == "EPUB" || r.Format == "M4B") {
		if r.Freeleech {
			return "Excellent (FL)"
		}
		return "Excellent"
	}
	if r.Seeders >= 5 {
		if r.Freeleech {
			return "Good (FL)"
		}
		return "Good"
	}
	if r.Seeders >= 1 {
		return "Available"
	}
	return "Low Seeds"
}

type OpenLibrarySearchResult struct {
	Key                string   `json:"key"`
	Title              string   `json:"title"`
	Authors            []string `json:"authors,omitempty"`
	AuthorKeys         []string `json:"authorKeys,omitempty"`
	FirstPublishYear   int      `json:"firstPublishYear,omitempty"`
	CoverURL           string   `json:"coverUrl,omitempty"`
	ISBN               string   `json:"isbn,omitempty"`
	ISBN13             string   `json:"isbn13,omitempty"`
	Language           string   `json:"language,omitempty"`
	PageCount          int      `json:"pageCount,omitempty"`
	Rating             float32  `json:"rating,omitempty"`
	Subjects           []string `json:"subjects,omitempty"`
	HasFulltext        bool     `json:"hasFulltext"`
	InLibrary          bool     `json:"inLibrary"`
	IsEbook            *bool    `json:"isEbook,omitempty"`
	HasEpub            *bool    `json:"hasEpub,omitempty"`
	GoogleBooksChecked bool     `json:"googleBooksChecked"`
}

type OpenLibraryAuthorSearchResult struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	BirthDate   string   `json:"birthDate,omitempty"`
	DeathDate   string   `json:"deathDate,omitempty"`
	TopWork     string   `json:"topWork,omitempty"`
	WorkCount   int      `json:"workCount,omitempty"`
	TopSubjects []string `json:"topSubjects,omitempty"`
	ImageURL    string   `json:"imageUrl,omitempty"`
	InLibrary   bool     `json:"inLibrary"`
}

type OpenLibrarySearchResponse struct {
	Results           []OpenLibrarySearchResult       `json:"results,omitempty"`
	AuthorResults     []OpenLibraryAuthorSearchResult `json:"authorResults,omitempty"`
	Total             int                             `json:"total"`
	GoogleQuotaUsed   int                             `json:"googleQuotaUsed"`
	GoogleQuotaRemain int                             `json:"googleQuotaRemaining"`
}

func (s *Server) searchOpenLibrary(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Query parameter 'q' is required"})
	}

	searchType := c.QueryParam("type")
	if searchType == "" {
		searchType = "book"
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := parseIntParam(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	client := s.getOpenLibraryClient()

	if searchType == "author" {
		return s.searchOpenLibraryAuthors(c, client, query, limit)
	}

	return s.searchOpenLibraryBooks(c, client, query, limit)
}

func (s *Server) searchOpenLibraryAuthors(c echo.Context, client *openlibrary.Client, query string, limit int) error {
	result, err := client.SearchAuthors(query, limit)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []OpenLibraryAuthorSearchResult
	for _, doc := range result.Docs {
		olResult := OpenLibraryAuthorSearchResult{
			Key:         extractOLID(doc.Key),
			Name:        doc.Name,
			BirthDate:   doc.BirthDate,
			DeathDate:   doc.DeathDate,
			TopWork:     doc.TopWork,
			WorkCount:   doc.WorkCount,
			TopSubjects: doc.TopSubjects,
		}

		olResult.ImageURL = openlibrary.GetAuthorPhotoURLByOLID(olResult.Key, "M")

		var count int64
		s.db.Model(&db.Author{}).Where("open_library_id = ?", olResult.Key).Count(&count)
		olResult.InLibrary = count > 0

		results = append(results, olResult)
	}

	return c.JSON(http.StatusOK, OpenLibrarySearchResponse{
		AuthorResults:     results,
		Total:             result.NumFound,
		GoogleQuotaUsed:   0,
		GoogleQuotaRemain: 0,
	})
}

func (s *Server) searchOpenLibraryBooks(c echo.Context, client *openlibrary.Client, query string, limit int) error {
	result, err := client.SearchBooks(query, limit, 0)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Search failed: " + err.Error()})
	}

	var results []OpenLibrarySearchResult
	for _, doc := range result.Docs {
		olResult := OpenLibrarySearchResult{
			Key:              extractOLID(doc.Key),
			Title:            doc.Title,
			Authors:          doc.AuthorName,
			AuthorKeys:       doc.AuthorKey,
			FirstPublishYear: doc.FirstPublishYear,
			PageCount:        doc.NumberOfPagesMedian,
			Rating:           float32(doc.RatingsAverage),
			Subjects:         doc.Subject,
			HasFulltext:      doc.HasFulltext,
		}

		if len(doc.ISBN) > 0 {
			for _, isbn := range doc.ISBN {
				if len(isbn) == 10 && olResult.ISBN == "" {
					olResult.ISBN = isbn
				} else if len(isbn) == 13 && olResult.ISBN13 == "" {
					olResult.ISBN13 = isbn
				}
				if olResult.ISBN != "" && olResult.ISBN13 != "" {
					break
				}
			}
		}

		if doc.CoverI > 0 {
			olResult.CoverURL = getCoverURL(doc.CoverI, "M")
		}

		if len(doc.Language) > 0 {
			olResult.Language = doc.Language[0]
		}

		var count int64
		if olResult.ISBN13 != "" {
			s.db.Model(&db.Book{}).Where("isbn13 = ?", olResult.ISBN13).Count(&count)
		}
		if count == 0 && olResult.ISBN != "" {
			s.db.Model(&db.Book{}).Where("isbn = ?", olResult.ISBN).Count(&count)
		}
		olResult.InLibrary = count > 0

		results = append(results, olResult)
	}

	gbClient := s.getGoogleBooksClient()
	quotaUsed := 0
	quotaRemain := 0
	if gbClient != nil {
		quotaUsed = gbClient.QuotaUsed()
		quotaRemain = gbClient.QuotaRemaining()
	}

	return c.JSON(http.StatusOK, OpenLibrarySearchResponse{
		Results:           results,
		Total:             result.NumFound,
		GoogleQuotaUsed:   quotaUsed,
		GoogleQuotaRemain: quotaRemain,
	})
}

func (s *Server) testOpenLibrary(c echo.Context) error {
	client := s.getOpenLibraryClient()
	if err := client.Test(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Connection failed: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Connection successful"})
}

func parseIntParam(s string) (int, error) {
	var n int
	_, err := time.ParseDuration("0s")
	if err != nil {
		return 0, err
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid integer")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func extractOLID(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return key
}

func getCoverURL(coverID int, size string) string {
	if size == "" {
		size = "M"
	}
	return "https://covers.openlibrary.org/b/id/" + strconv.Itoa(coverID) + "-" + size + ".jpg"
}

func (s *Server) getOpenLibraryClient() *openlibrary.Client {
	return openlibrary.NewClient()
}

func (s *Server) getGoogleBooksClient() *googlebooks.Client {
	var setting db.Setting
	if err := s.db.Where("key = ?", "google_books_api_key").First(&setting).Error; err != nil || setting.Value == "" {
		return nil
	}
	return googlebooks.NewClient(setting.Value)
}

func (s *Server) testGoogleBooks(c echo.Context) error {
	client := s.getGoogleBooksClient()
	if client == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Google Books API key not configured"})
	}

	if err := client.Test(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Connection failed: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":        "Connection successful",
		"quotaUsed":      client.QuotaUsed(),
		"quotaRemaining": client.QuotaRemaining(),
	})
}

func (s *Server) getGoogleBooksQuota(c echo.Context) error {
	client := s.getGoogleBooksClient()
	if client == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"configured": false,
			"quotaUsed":  0,
			"quotaLimit": 1000,
			"remaining":  0,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"configured": true,
		"quotaUsed":  client.QuotaUsed(),
		"quotaLimit": 1000,
		"remaining":  client.QuotaRemaining(),
	})
}

type EbookStatusResponse struct {
	ISBN           string `json:"isbn"`
	Checked        bool   `json:"checked"`
	IsEbook        *bool  `json:"isEbook,omitempty"`
	HasEpub        *bool  `json:"hasEpub,omitempty"`
	HasPdf         *bool  `json:"hasPdf,omitempty"`
	BuyLink        string `json:"buyLink,omitempty"`
	GoogleVolumeID string `json:"googleVolumeId,omitempty"`
	QuotaExhausted bool   `json:"quotaExhausted"`
}

func (s *Server) checkEbookStatus(c echo.Context) error {
	isbn := c.QueryParam("isbn")
	if isbn == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ISBN parameter is required"})
	}

	client := s.getGoogleBooksClient()
	if client == nil {
		return c.JSON(http.StatusOK, EbookStatusResponse{
			ISBN:           isbn,
			Checked:        false,
			QuotaExhausted: true,
		})
	}

	info, err := client.GetVolumeByISBN(isbn)
	if err != nil {
		if err == googlebooks.ErrQuotaExhausted {
			return c.JSON(http.StatusOK, EbookStatusResponse{
				ISBN:           isbn,
				Checked:        false,
				QuotaExhausted: true,
			})
		}
		if err == googlebooks.ErrNotFound {
			isEbook := false
			return c.JSON(http.StatusOK, EbookStatusResponse{
				ISBN:    isbn,
				Checked: true,
				IsEbook: &isEbook,
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, EbookStatusResponse{
		ISBN:           isbn,
		Checked:        true,
		IsEbook:        &info.IsEbook,
		HasEpub:        &info.HasEpub,
		HasPdf:         &info.HasPdf,
		BuyLink:        info.BuyLink,
		GoogleVolumeID: info.VolumeID,
	})
}
