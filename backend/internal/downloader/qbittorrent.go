package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// QBittorrentClient handles communication with qBittorrent Web API
type QBittorrentClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	sid        string // Session ID
}

// NewQBittorrentClient creates a new qBittorrent client
func NewQBittorrentClient(baseURL, username, password string) *QBittorrentClient {
	jar, _ := cookiejar.New(nil)
	
	return &QBittorrentClient{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// Login authenticates with qBittorrent
func (q *QBittorrentClient) Login(ctx context.Context) error {
	data := url.Values{
		"username": {q.username},
		"password": {q.password},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Ok." {
		return fmt.Errorf("login failed: %s", string(body))
	}

	// Extract SID from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			q.sid = cookie.Value
			break
		}
	}

	return nil
}

// Test checks if the connection is working
func (q *QBittorrentClient) Test(ctx context.Context) error {
	if err := q.Login(ctx); err != nil {
		return err
	}

	// Try to get version
	_, err := q.GetVersion(ctx)
	return err
}

// GetVersion returns the qBittorrent version
func (q *QBittorrentClient) GetVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", q.baseURL+"/api/v2/app/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// TorrentInfo represents information about a torrent in qBittorrent
type TorrentInfo struct {
	Hash           string  `json:"hash"`
	Name           string  `json:"name"`
	Size           int64   `json:"size"`
	Progress       float64 `json:"progress"`
	State          string  `json:"state"`
	DownloadSpeed  int64   `json:"dlspeed"`
	UploadSpeed    int64   `json:"upspeed"`
	ETA            int64   `json:"eta"`
	Category       string  `json:"category"`
	SavePath       string  `json:"save_path"`
	AddedOn        int64   `json:"added_on"`
	CompletedOn    int64   `json:"completion_on"`
	Ratio          float64 `json:"ratio"`
	Downloaded     int64   `json:"downloaded"`
	Uploaded       int64   `json:"uploaded"`
	AmountLeft     int64   `json:"amount_left"`
	NumSeeds       int     `json:"num_seeds"`
	NumLeeches     int     `json:"num_leechs"`
}

// GetTorrents returns all torrents, optionally filtered by category
func (q *QBittorrentClient) GetTorrents(ctx context.Context, category string) ([]TorrentInfo, error) {
	params := url.Values{}
	if category != "" {
		params.Set("category", category)
	}

	endpoint := q.baseURL + "/api/v2/torrents/info"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		// Session expired, re-login
		if err := q.Login(ctx); err != nil {
			return nil, err
		}
		return q.GetTorrents(ctx, category)
	}

	var torrents []TorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}

	return torrents, nil
}

// GetTorrent returns information about a specific torrent
func (q *QBittorrentClient) GetTorrent(ctx context.Context, hash string) (*TorrentInfo, error) {
	torrents, err := q.GetTorrents(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, t := range torrents {
		if strings.EqualFold(t.Hash, hash) {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("torrent not found: %s", hash)
}

// AddTorrentOptions holds options for adding a torrent
type AddTorrentOptions struct {
	Category     string
	SavePath     string
	Paused       bool
	Tags         []string
	AutoTMM      bool   // Automatic Torrent Management
	ContentLayout string // "Original", "Subfolder", "NoSubfolder"
}

// Type returns the client type
func (q *QBittorrentClient) Type() string {
	return "qbittorrent"
}

// AddDownload implements the Client interface
func (q *QBittorrentClient) AddDownload(ctx context.Context, url string, opts *DownloadOptions) (string, error) {
	addOpts := &AddTorrentOptions{}
	if opts != nil {
		addOpts.Category = opts.Category
		addOpts.SavePath = opts.SavePath
		addOpts.Paused = opts.Paused
	}
	
	if err := q.AddTorrentByURL(ctx, url, addOpts); err != nil {
		return "", err
	}
	
	// qBittorrent doesn't return the hash on add, so we return the URL as identifier
	// The actual hash would need to be fetched separately
	return url, nil
}

// GetDownload implements the Client interface
func (q *QBittorrentClient) GetDownload(ctx context.Context, id string) (*DownloadInfo, error) {
	torrent, err := q.GetTorrent(ctx, id)
	if err != nil {
		return nil, err
	}
	
	status := StatusDownloading
	switch torrent.State {
	case "pausedDL", "pausedUP":
		status = StatusPaused
	case "stalledDL", "stalledUP", "uploading", "seeding":
		status = StatusCompleted
	case "error":
		status = StatusFailed
	case "queuedDL", "queuedUP", "checkingDL", "checkingUP":
		status = StatusQueued
	}
	
	return &DownloadInfo{
		ID:            torrent.Hash,
		Name:          torrent.Name,
		Size:          torrent.Size,
		Downloaded:    torrent.Downloaded,
		Progress:      torrent.Progress,
		Status:        status,
		DownloadSpeed: torrent.DownloadSpeed,
		ETA:           torrent.ETA,
		SavePath:      torrent.SavePath,
		Category:      torrent.Category,
	}, nil
}

// GetAllDownloads implements the Client interface
func (q *QBittorrentClient) GetAllDownloads(ctx context.Context, category string) ([]DownloadInfo, error) {
	torrents, err := q.GetTorrents(ctx, category)
	if err != nil {
		return nil, err
	}
	
	downloads := make([]DownloadInfo, 0, len(torrents))
	for _, t := range torrents {
		info, _ := q.GetDownload(ctx, t.Hash)
		if info != nil {
			downloads = append(downloads, *info)
		}
	}
	
	return downloads, nil
}

// RemoveDownload implements the Client interface
func (q *QBittorrentClient) RemoveDownload(ctx context.Context, id string, deleteFiles bool) error {
	return q.DeleteTorrent(ctx, id, deleteFiles)
}

// PauseDownload implements the Client interface
func (q *QBittorrentClient) PauseDownload(ctx context.Context, id string) error {
	return q.PauseTorrent(ctx, id)
}

// ResumeDownload implements the Client interface
func (q *QBittorrentClient) ResumeDownload(ctx context.Context, id string) error {
	return q.ResumeTorrent(ctx, id)
}

// AddTorrentByURL adds a torrent from a URL or magnet link
func (q *QBittorrentClient) AddTorrentByURL(ctx context.Context, torrentURL string, opts *AddTorrentOptions) error {
	data := url.Values{
		"urls": {torrentURL},
	}

	if opts != nil {
		if opts.Category != "" {
			data.Set("category", opts.Category)
		}
		if opts.SavePath != "" {
			data.Set("savepath", opts.SavePath)
		}
		if opts.Paused {
			data.Set("paused", "true")
		}
		if len(opts.Tags) > 0 {
			data.Set("tags", strings.Join(opts.Tags, ","))
		}
		if opts.AutoTMM {
			data.Set("autoTMM", "true")
		}
		if opts.ContentLayout != "" {
			data.Set("contentLayout", opts.ContentLayout)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/add", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("add torrent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		if err := q.Login(ctx); err != nil {
			return err
		}
		return q.AddTorrentByURL(ctx, torrentURL, opts)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add torrent failed: %s", string(body))
	}

	return nil
}

// AddTorrentFile adds a torrent from a .torrent file
func (q *QBittorrentClient) AddTorrentFile(ctx context.Context, filename string, torrentData []byte, opts *AddTorrentOptions) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add torrent file
	part, err := writer.CreateFormFile("torrents", filename)
	if err != nil {
		return err
	}
	part.Write(torrentData)

	// Add options
	if opts != nil {
		if opts.Category != "" {
			writer.WriteField("category", opts.Category)
		}
		if opts.SavePath != "" {
			writer.WriteField("savepath", opts.SavePath)
		}
		if opts.Paused {
			writer.WriteField("paused", "true")
		}
		if len(opts.Tags) > 0 {
			writer.WriteField("tags", strings.Join(opts.Tags, ","))
		}
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/add", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("add torrent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add torrent failed: %s", string(respBody))
	}

	return nil
}

// PauseTorrent pauses a torrent
func (q *QBittorrentClient) PauseTorrent(ctx context.Context, hash string) error {
	return q.torrentAction(ctx, "pause", hash)
}

// ResumeTorrent resumes a torrent
func (q *QBittorrentClient) ResumeTorrent(ctx context.Context, hash string) error {
	return q.torrentAction(ctx, "resume", hash)
}

// DeleteTorrent removes a torrent
func (q *QBittorrentClient) DeleteTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	data := url.Values{
		"hashes":      {hash},
		"deleteFiles": {fmt.Sprintf("%t", deleteFiles)},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/delete", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (q *QBittorrentClient) torrentAction(ctx context.Context, action, hash string) error {
	data := url.Values{
		"hashes": {hash},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/"+action, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetCategories returns all categories
func (q *QBittorrentClient) GetCategories(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", q.baseURL+"/api/v2/torrents/categories", nil)
	if err != nil {
		return nil, err
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var categories map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateCategory creates a new category
func (q *QBittorrentClient) CreateCategory(ctx context.Context, name, savePath string) error {
	data := url.Values{
		"category": {name},
		"savePath": {savePath},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/createCategory", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SetTorrentCategory sets the category for a torrent
func (q *QBittorrentClient) SetTorrentCategory(ctx context.Context, hash, category string) error {
	data := url.Values{
		"hashes":   {hash},
		"category": {category},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/api/v2/torrents/setCategory", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// TorrentContent represents a file within a torrent
type TorrentContent struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Progress     float64 `json:"progress"`
	Priority     int     `json:"priority"`
	IsSeed       bool    `json:"is_seed"`
	PieceRange   []int   `json:"piece_range"`
	Availability float64 `json:"availability"`
}

// GetTorrentContents returns the files within a torrent
func (q *QBittorrentClient) GetTorrentContents(ctx context.Context, hash string) ([]TorrentContent, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", q.baseURL+"/api/v2/torrents/files?hash="+hash, nil)
	if err != nil {
		return nil, err
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var contents []TorrentContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	return contents, nil
}

