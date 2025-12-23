package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"github.com/shelfarr/shelfarr/internal/indexer"
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
