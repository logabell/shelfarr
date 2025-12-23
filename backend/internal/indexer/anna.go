package indexer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// AnnaIndexer implements the Anna's Archive indexer (scraper)
type AnnaIndexer struct {
	name       string
	httpClient *http.Client
}

// NewAnnaIndexer creates a new Anna's Archive indexer
func NewAnnaIndexer(name string) *AnnaIndexer {
	return &AnnaIndexer{
		name: name,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *AnnaIndexer) Name() string {
	return a.name
}

func (a *AnnaIndexer) Type() string {
	return "anna"
}

func (a *AnnaIndexer) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	// Anna's Archive search URL
	baseURL := "https://annas-archive.org/search"
	
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	
	// Build search query
	searchTerms := query.Title
	if query.Author != "" {
		searchTerms = query.Author + " " + searchTerms
	}
	if query.ISBN != "" {
		searchTerms = query.ISBN
	}
	params.Set("q", searchTerms)

	u.RawQuery = params.Encode()

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anna's Archive request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse HTML response (basic scraping)
	results := a.parseSearchResults(string(body))
	
	return results, nil
}

// parseSearchResults extracts search results from Anna's Archive HTML
func (a *AnnaIndexer) parseSearchResults(html string) []SearchResult {
	var results []SearchResult

	// Basic regex patterns for scraping
	// Note: This is a simplified implementation - real scraping would need more robust parsing
	
	// Find result blocks
	resultPattern := regexp.MustCompile(`<a[^>]+href="(/md5/[^"]+)"[^>]*>.*?<div[^>]*>([^<]+)</div>`)
	matches := resultPattern.FindAllStringSubmatch(html, -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			result := SearchResult{
				Title:       strings.TrimSpace(match[2]),
				InfoURL:     "https://annas-archive.org" + match[1],
				DownloadURL: "https://annas-archive.org" + match[1],
				Indexer:     a.name,
				Format:      detectFormat(match[2]),
			}
			results = append(results, result)
		}
	}

	// Limit results
	if len(results) > 20 {
		results = results[:20]
	}

	return results
}

func (a *AnnaIndexer) Test(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://annas-archive.org/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func (a *AnnaIndexer) Download(ctx context.Context, result SearchResult) (string, error) {
	// For Anna's Archive, we need to fetch the detail page and extract download links
	// This would involve scraping the MD5 page for actual download URLs
	// For now, return the info URL
	return result.InfoURL, nil
}

