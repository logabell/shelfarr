package indexer

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TorznabIndexer implements the Torznab protocol for Prowlarr/Jackett
type TorznabIndexer struct {
	name       string
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewTorznabIndexer creates a new Torznab indexer
func NewTorznabIndexer(name, baseURL, apiKey string) *TorznabIndexer {
	return &TorznabIndexer{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TorznabIndexer) Name() string {
	return t.name
}

func (t *TorznabIndexer) Type() string {
	return "torznab"
}

// torznabResponse represents the XML response from Torznab API
type torznabResponse struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []torznabItem `xml:"item"`
	} `xml:"channel"`
}

type torznabItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Size        int64  `xml:"size"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	Enclosure   struct {
		URL    string `xml:"url,attr"`
		Length int64  `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"enclosure"`
	Attrs []struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	} `xml:"attr"`
}

func (t *TorznabIndexer) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	// Build search URL
	u, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("apikey", t.apiKey)
	params.Set("t", "search")
	
	// Category: 7000 = Books, 3030 = Audiobooks
	if query.MediaType == "audiobook" {
		params.Set("cat", "3030")
	} else {
		params.Set("cat", "7000")
	}

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

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse XML response
	var response torznabResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to SearchResult
	results := make([]SearchResult, 0, len(response.Channel.Items))
	for _, item := range response.Channel.Items {
		result := SearchResult{
			Title:       item.Title,
			Size:        item.Size,
			DownloadURL: item.Enclosure.URL,
			PublishDate: item.PubDate,
			Indexer:     t.name,
		}

		// Parse attributes for additional info
		for _, attr := range item.Attrs {
			switch attr.Name {
			case "seeders":
				result.Seeders, _ = strconv.Atoi(attr.Value)
			case "leechers":
				result.Leechers, _ = strconv.Atoi(attr.Value)
			}
		}

		// Detect format from title
		result.Format = detectFormat(item.Title)

		results = append(results, result)
	}

	return results, nil
}

func (t *TorznabIndexer) Test(ctx context.Context) error {
	// Test with caps endpoint
	u, err := url.Parse(t.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("apikey", t.apiKey)
	params.Set("t", "caps")
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func (t *TorznabIndexer) Download(ctx context.Context, result SearchResult) (string, error) {
	return result.DownloadURL, nil
}

