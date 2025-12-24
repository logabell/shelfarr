package googlebooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL    = "https://www.googleapis.com/books/v1"
	DailyQuota = 1000
	UserAgent  = "Shelfarr/1.0 (https://github.com/shelfarr/shelfarr)"
)

var (
	ErrQuotaExhausted = errors.New("google books daily quota exhausted")
	ErrNotFound       = errors.New("book not found")
	ErrNoAPIKey       = errors.New("google books API key not configured")
)

type QuotaTracker struct {
	mu        sync.RWMutex
	count     int
	resetDate string
}

func (q *QuotaTracker) Check() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	today := time.Now().UTC().Format("2006-01-02")
	if q.resetDate != today {
		q.count = 0
		q.resetDate = today
	}

	if q.count >= DailyQuota {
		return ErrQuotaExhausted
	}
	return nil
}

func (q *QuotaTracker) Increment() {
	q.mu.Lock()
	defer q.mu.Unlock()

	today := time.Now().UTC().Format("2006-01-02")
	if q.resetDate != today {
		q.count = 0
		q.resetDate = today
	}
	q.count++
}

func (q *QuotaTracker) Remaining() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	today := time.Now().UTC().Format("2006-01-02")
	if q.resetDate != today {
		return DailyQuota
	}
	return DailyQuota - q.count
}

func (q *QuotaTracker) Count() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
	quota      *QuotaTracker
}

func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   BaseURL,
		apiKey:    apiKey,
		userAgent: UserAgent,
		quota:     &QuotaTracker{},
	}
}

func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

func (c *Client) HasAPIKey() bool {
	return c.apiKey != ""
}

func (c *Client) QuotaRemaining() int {
	return c.quota.Remaining()
}

func (c *Client) QuotaUsed() int {
	return c.quota.Count()
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, ErrNoAPIKey
	}

	if err := c.quota.Check(); err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	c.quota.Increment()
	return resp, nil
}

func (c *Client) GetVolumeByISBN(isbn string) (*EbookInfo, error) {
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.TrimSpace(isbn)

	if isbn == "" {
		return nil, errors.New("ISBN is required")
	}

	params := url.Values{}
	params.Set("q", "isbn:"+isbn)
	params.Set("maxResults", "1")
	params.Set("key", c.apiKey)

	reqURL := fmt.Sprintf("%s/volumes?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result VolumesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, ErrNotFound
	}

	return volumeToEbookInfo(&result.Items[0]), nil
}

func (c *Client) SearchVolumes(query string, maxResults int) ([]EbookInfo, error) {
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 40 {
		maxResults = 40
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))
	params.Set("key", c.apiKey)

	reqURL := fmt.Sprintf("%s/volumes?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result VolumesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	books := make([]EbookInfo, len(result.Items))
	for i, item := range result.Items {
		books[i] = *volumeToEbookInfo(&item)
	}

	return books, nil
}

func (c *Client) GetVolume(volumeID string) (*EbookInfo, error) {
	params := url.Values{}
	params.Set("key", c.apiKey)

	reqURL := fmt.Sprintf("%s/volumes/%s?%s", c.baseURL, volumeID, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var volume Volume
	if err := json.NewDecoder(resp.Body).Decode(&volume); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return volumeToEbookInfo(&volume), nil
}

func (c *Client) CheckEbookStatus(isbn string) (isEbook bool, hasEpub bool, hasPdf bool, buyLink string, err error) {
	info, err := c.GetVolumeByISBN(isbn)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, false, false, "", nil
		}
		return false, false, false, "", err
	}

	return info.IsEbook, info.HasEpub, info.HasPdf, info.BuyLink, nil
}

func (c *Client) Test() error {
	params := url.Values{}
	params.Set("q", "test")
	params.Set("maxResults", "1")
	params.Set("key", c.apiKey)

	reqURL := fmt.Sprintf("%s/volumes?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API test failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func volumeToEbookInfo(v *Volume) *EbookInfo {
	info := &EbookInfo{
		VolumeID:    v.ID,
		IsEbook:     v.SaleInfo.IsEbook,
		HasEpub:     v.AccessInfo.Epub.IsAvailable,
		HasPdf:      v.AccessInfo.Pdf.IsAvailable,
		BuyLink:     v.SaleInfo.BuyLink,
		Title:       v.VolumeInfo.Title,
		Authors:     v.VolumeInfo.Authors,
		Description: v.VolumeInfo.Description,
		PageCount:   v.VolumeInfo.PageCount,
		Language:    v.VolumeInfo.Language,
		Categories:  v.VolumeInfo.Categories,
		Rating:      v.VolumeInfo.AverageRating,
	}

	for _, id := range v.VolumeInfo.IndustryIdentifiers {
		switch id.Type {
		case "ISBN_10":
			info.ISBN10 = id.Identifier
		case "ISBN_13":
			info.ISBN13 = id.Identifier
		}
	}

	if v.VolumeInfo.ImageLinks != nil {
		if v.VolumeInfo.ImageLinks.Large != "" {
			info.CoverURL = v.VolumeInfo.ImageLinks.Large
		} else if v.VolumeInfo.ImageLinks.Medium != "" {
			info.CoverURL = v.VolumeInfo.ImageLinks.Medium
		} else if v.VolumeInfo.ImageLinks.Thumbnail != "" {
			info.CoverURL = v.VolumeInfo.ImageLinks.Thumbnail
		}
	}

	return info
}
