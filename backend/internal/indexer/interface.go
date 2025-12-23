package indexer

import (
	"context"
	"log"
)

// SearchResult represents a search result from an indexer
type SearchResult struct {
	Title       string
	Size        int64
	Format      string // epub, pdf, m4b, mp3, etc.
	Seeders     int
	Leechers    int
	DownloadURL string
	InfoURL     string
	PublishDate string
	Freeleech   bool
	VIP         bool
	Indexer     string

	// Additional metadata
	Author      string
	Narrator    string
	Category    string
	SeriesName  string
	SeriesIndex string
	LangCode    string // 3-letter language code (e.g., ENG, SPA)

	// Audiobook quality metadata
	Bitrate  int // kbps for audiobooks
	Duration int // seconds

	// Quality scoring
	Quality int // 0-100, calculated by the decision engine
}

// SearchQuery represents a search request
type SearchQuery struct {
	Title  string
	Author string
	ISBN   string
	BookID string // Hardcover ID for verification

	// Filters
	MediaType string // ebook, audiobook
}

// Indexer is the interface that all indexers must implement
type Indexer interface {
	// Name returns the display name of the indexer
	Name() string

	// Type returns the indexer type (torznab, mam, anna)
	Type() string

	// Search performs a search and returns results
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)

	// Test checks if the indexer is reachable and configured correctly
	Test(ctx context.Context) error

	// Download returns the actual download URL/magnet/NZB for a result
	Download(ctx context.Context, result SearchResult) (string, error)
}

// Manager handles multiple indexers and orchestrates searches
type Manager struct {
	indexers []Indexer
}

// NewManager creates a new indexer manager
func NewManager() *Manager {
	return &Manager{
		indexers: make([]Indexer, 0),
	}
}

// AddIndexer adds an indexer to the manager
func (m *Manager) AddIndexer(indexer Indexer) {
	m.indexers = append(m.indexers, indexer)
}

// SearchAll searches all enabled indexers using the waterfall method
func (m *Manager) SearchAll(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	var allResults []SearchResult

	log.Printf("[DEBUG] SearchAll: starting waterfall search with Title='%s', Author='%s', ISBN='%s', MediaType='%s'",
		query.Title, query.Author, query.ISBN, query.MediaType)

	// Waterfall search: Author+Title -> Title only -> ISBN (if available)
	// Changed order: Author+Title first since it's most specific with usable text
	searches := []SearchQuery{
		// First: Author + Title (most specific with search text)
		{Title: query.Title, Author: query.Author, ISBN: query.ISBN, MediaType: query.MediaType},
		// Second: Title only (broader search)
		{Title: query.Title, MediaType: query.MediaType},
		// Third: try ISBN if available (some indexers may support ISBN search)
		{ISBN: query.ISBN, MediaType: query.MediaType},
	}

	for _, indexer := range m.indexers {
		log.Printf("[DEBUG] SearchAll: searching indexer '%s'", indexer.Name())

		for i, search := range searches {
			if search.ISBN == "" && search.Title == "" {
				log.Printf("[DEBUG] SearchAll: skipping search #%d (empty)", i+1)
				continue // Skip empty searches
			}

			log.Printf("[DEBUG] SearchAll: trying search #%d: Title='%s', Author='%s', ISBN='%s'",
				i+1, search.Title, search.Author, search.ISBN)

			results, err := indexer.Search(ctx, search)
			if err != nil {
				log.Printf("[DEBUG] SearchAll: search #%d failed: %v", i+1, err)
				continue // Try next search strategy on error
			}

			log.Printf("[DEBUG] SearchAll: search #%d returned %d results", i+1, len(results))

			if len(results) > 0 {
				allResults = append(allResults, results...)
				break // Found results with this search type, move to next indexer
			}
		}
	}

	log.Printf("[DEBUG] SearchAll: completed with %d total results", len(allResults))

	// Sort by quality score
	// TODO: Implement quality scoring based on profiles

	return allResults, nil
}
