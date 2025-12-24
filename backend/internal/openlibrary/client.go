package openlibrary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL       = "https://openlibrary.org"
	CoversBaseURL = "https://covers.openlibrary.org"
	UserAgent     = "Shelfarr/1.0 (https://github.com/shelfarr/shelfarr)"
)

// Client is an Open Library API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewClient creates a new Open Library client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   BaseURL,
		userAgent: UserAgent,
	}
}

// doRequest performs an HTTP request with proper headers
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}

// SearchBooks searches for books using the search API
func (c *Client) SearchBooks(query string, limit, offset int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("fields", "key,title,author_name,author_key,first_publish_year,edition_count,cover_i,isbn,language,subject,publisher,number_of_pages_median,ratings_average,ratings_count,has_fulltext")

	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// SearchByTitle searches for books by title
func (c *Client) SearchByTitle(title string, limit int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("title", title)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("fields", "key,title,author_name,author_key,first_publish_year,edition_count,cover_i,isbn,language,subject,publisher,number_of_pages_median,ratings_average,ratings_count,has_fulltext")

	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// SearchByAuthor searches for books by author
func (c *Client) SearchByAuthor(author string, limit int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("author", author)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("fields", "key,title,author_name,author_key,first_publish_year,edition_count,cover_i,isbn,language,subject,publisher,number_of_pages_median,ratings_average,ratings_count,has_fulltext")

	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// SearchByISBN searches for a book by ISBN
func (c *Client) SearchByISBN(isbn string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("isbn", isbn)
	params.Set("limit", "1")
	params.Set("fields", "key,title,author_name,author_key,first_publish_year,edition_count,cover_i,isbn,language,subject,publisher,number_of_pages_median,ratings_average,ratings_count,has_fulltext")

	reqURL := fmt.Sprintf("%s/search.json?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetBookByISBN gets book details using the Books API
func (c *Client) GetBookByISBN(isbn string) (*BookData, error) {
	// Clean ISBN
	isbn = strings.ReplaceAll(isbn, "-", "")

	params := url.Values{}
	params.Set("bibkeys", "ISBN:"+isbn)
	params.Set("jscmd", "data")
	params.Set("format", "json")

	reqURL := fmt.Sprintf("%s/api/books?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Response is a map with the bibkey as key
	var result map[string]*BookData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Get the book data from the map
	bookData, ok := result["ISBN:"+isbn]
	if !ok {
		return nil, fmt.Errorf("book not found for ISBN: %s", isbn)
	}

	return bookData, nil
}

// GetWork gets a work by its OLID
func (c *Client) GetWork(olid string) (*Work, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/works/") {
		olid = "/works/" + olid
	}

	reqURL := fmt.Sprintf("%s%s.json", c.baseURL, olid)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("work not found: %s", olid)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var work Work
	if err := json.NewDecoder(resp.Body).Decode(&work); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &work, nil
}

// GetWorkEditions gets editions for a work
func (c *Client) GetWorkEditions(olid string, limit, offset int) (*EditionsResponse, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/works/") {
		olid = "/works/" + olid
	}

	reqURL := fmt.Sprintf("%s%s/editions.json?limit=%d&offset=%d", c.baseURL, olid, limit, offset)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result EditionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetEdition gets a specific edition by OLID
func (c *Client) GetEdition(olid string) (*Edition, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/books/") {
		olid = "/books/" + olid
	}

	reqURL := fmt.Sprintf("%s%s.json", c.baseURL, olid)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("edition not found: %s", olid)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var edition Edition
	if err := json.NewDecoder(resp.Body).Decode(&edition); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &edition, nil
}

// GetAuthor gets an author by OLID
func (c *Client) GetAuthor(olid string) (*Author, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/authors/") {
		olid = "/authors/" + olid
	}

	reqURL := fmt.Sprintf("%s%s.json", c.baseURL, olid)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("author not found: %s", olid)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var author Author
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &author, nil
}

// GetAuthorWorks gets works by an author
func (c *Client) GetAuthorWorks(olid string, limit, offset int) (*AuthorWorksResponse, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/authors/") {
		olid = "/authors/" + olid
	}

	reqURL := fmt.Sprintf("%s%s/works.json?limit=%d&offset=%d", c.baseURL, olid, limit, offset)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result AuthorWorksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// SearchAuthors searches for authors
func (c *Client) SearchAuthors(query string, limit int) (*AuthorSearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	reqURL := fmt.Sprintf("%s/search/authors.json?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var result AuthorSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetSubject gets books for a subject
func (c *Client) GetSubject(subject string, limit, offset int, ebooksOnly bool) (*SubjectResponse, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if ebooksOnly {
		params.Set("ebooks", "true")
	}

	// Subject should be URL-safe
	subject = strings.ReplaceAll(strings.ToLower(subject), " ", "_")

	reqURL := fmt.Sprintf("%s/subjects/%s.json?%s", c.baseURL, subject, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("subject not found: %s", subject)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result SubjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetWorkRatings gets ratings for a work
func (c *Client) GetWorkRatings(olid string) (*RatingsResponse, error) {
	// Ensure OLID format
	if !strings.HasPrefix(olid, "/works/") {
		olid = "/works/" + olid
	}

	reqURL := fmt.Sprintf("%s%s/ratings.json", c.baseURL, olid)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result RatingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetCoverURL returns the URL for a cover image
func GetCoverURL(coverID int, size string) string {
	// size can be "S", "M", or "L"
	if size == "" {
		size = "M"
	}
	return fmt.Sprintf("%s/b/id/%d-%s.jpg", CoversBaseURL, coverID, size)
}

// GetCoverURLByISBN returns the URL for a cover by ISBN (rate limited - prefer GetCoverURL)
func GetCoverURLByISBN(isbn, size string) string {
	if size == "" {
		size = "M"
	}
	return fmt.Sprintf("%s/b/isbn/%s-%s.jpg", CoversBaseURL, isbn, size)
}

// GetAuthorPhotoURL returns the URL for an author photo
func GetAuthorPhotoURL(photoID int, size string) string {
	if size == "" {
		size = "M"
	}
	return fmt.Sprintf("%s/a/id/%d-%s.jpg", CoversBaseURL, photoID, size)
}

// GetAuthorPhotoURLByOLID returns the URL for an author photo by OLID
func GetAuthorPhotoURLByOLID(olid, size string) string {
	if size == "" {
		size = "M"
	}
	// Remove /authors/ prefix if present
	olid = strings.TrimPrefix(olid, "/authors/")
	return fmt.Sprintf("%s/a/olid/%s-%s.jpg", CoversBaseURL, olid, size)
}

// ExtractDescription extracts the description string from a work or author
// Description can be a string or an object with "value" key
func ExtractDescription(desc interface{}) string {
	if desc == nil {
		return ""
	}

	switch v := desc.(type) {
	case string:
		return v
	case map[string]interface{}:
		if value, ok := v["value"].(string); ok {
			return value
		}
	}

	return ""
}

// ExtractOLID extracts the OLID from a key path like "/works/OL123W"
func ExtractOLID(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return key
}

// Test tests the connection to Open Library
func (c *Client) Test() error {
	req, err := http.NewRequest("GET", c.baseURL+"/search.json?q=test&limit=1", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
