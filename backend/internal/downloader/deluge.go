package downloader

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// DelugeClient handles communication with Deluge Web API
type DelugeClient struct {
	baseURL    string
	password   string
	httpClient *http.Client
	requestID  int
}

// NewDelugeClient creates a new Deluge client
func NewDelugeClient(baseURL, password string) *DelugeClient {
	jar, _ := cookiejar.New(nil)

	return &DelugeClient{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		requestID: 0,
	}
}

// delugeRequest represents a JSON-RPC request to Deluge
type delugeRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// delugeResponse represents a JSON-RPC response from Deluge
type delugeResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *delugeError    `json:"error"`
}

type delugeError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// call makes a JSON-RPC call to Deluge
func (d *DelugeClient) call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	d.requestID++

	if params == nil {
		params = []interface{}{}
	}

	reqBody := delugeRequest{
		ID:     d.requestID,
		Method: method,
		Params: params,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var delugeResp delugeResponse
	if err := json.Unmarshal(body, &delugeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}

	if delugeResp.Error != nil {
		return nil, fmt.Errorf("deluge error: %s (code: %d)", delugeResp.Error.Message, delugeResp.Error.Code)
	}

	return delugeResp.Result, nil
}

// Login authenticates with Deluge
func (d *DelugeClient) Login(ctx context.Context) error {
	result, err := d.call(ctx, "auth.login", d.password)
	if err != nil {
		return err
	}

	var success bool
	if err := json.Unmarshal(result, &success); err != nil {
		return fmt.Errorf("failed to parse login result: %w", err)
	}

	if !success {
		return fmt.Errorf("login failed: invalid password")
	}

	// Connect to the daemon
	_, err = d.call(ctx, "web.connected")
	if err != nil {
		// Try to get hosts and connect to the first one
		hostsResult, err := d.call(ctx, "web.get_hosts")
		if err != nil {
			return fmt.Errorf("failed to get hosts: %w", err)
		}

		var hosts [][]interface{}
		if err := json.Unmarshal(hostsResult, &hosts); err != nil {
			return nil // May already be connected
		}

		if len(hosts) > 0 {
			hostID, ok := hosts[0][0].(string)
			if ok {
				_, _ = d.call(ctx, "web.connect", hostID)
			}
		}
	}

	return nil
}

// Test checks if the connection is working
func (d *DelugeClient) Test(ctx context.Context) error {
	if err := d.Login(ctx); err != nil {
		return err
	}

	// Try to get daemon info
	_, err := d.call(ctx, "daemon.info")
	if err != nil {
		// Fall back to checking web connection status
		_, err = d.call(ctx, "web.connected")
	}
	return err
}

// Type returns the client type
func (d *DelugeClient) Type() string {
	return "deluge"
}

// AddDownload implements the Client interface
func (d *DelugeClient) AddDownload(ctx context.Context, url string, opts *DownloadOptions) (string, error) {
	if err := d.Login(ctx); err != nil {
		return "", err
	}

	// Build options map
	options := make(map[string]interface{})
	if opts != nil && opts.SavePath != "" {
		options["download_location"] = opts.SavePath
	}

	// Add torrent by URL
	result, err := d.call(ctx, "web.add_torrents", []map[string]interface{}{
		{
			"path":    url,
			"options": options,
		},
	})
	if err != nil {
		// Try alternative method for magnet/URL
		result, err = d.call(ctx, "core.add_torrent_url", url, options)
		if err != nil {
			return "", fmt.Errorf("failed to add torrent: %w", err)
		}
	}

	// Parse the result to get torrent ID
	var torrentID string
	if err := json.Unmarshal(result, &torrentID); err != nil {
		// Result might be an array
		var results []interface{}
		if err := json.Unmarshal(result, &results); err == nil && len(results) > 0 {
			if arr, ok := results[0].([]interface{}); ok && len(arr) > 1 {
				if id, ok := arr[1].(string); ok {
					torrentID = id
				}
			}
		}
	}

	if torrentID == "" {
		torrentID = url // Use URL as fallback ID
	}

	return torrentID, nil
}

// GetDownload implements the Client interface
func (d *DelugeClient) GetDownload(ctx context.Context, id string) (*DownloadInfo, error) {
	if err := d.Login(ctx); err != nil {
		return nil, err
	}

	keys := []string{
		"name", "total_size", "progress", "state", "download_payload_rate",
		"eta", "save_path", "total_done", "label",
	}

	result, err := d.call(ctx, "core.get_torrent_status", id, keys)
	if err != nil {
		return nil, err
	}

	var status map[string]interface{}
	if err := json.Unmarshal(result, &status); err != nil {
		return nil, err
	}

	if len(status) == 0 {
		return nil, fmt.Errorf("torrent not found: %s", id)
	}

	info := &DownloadInfo{
		ID:       id,
		Name:     getString(status, "name"),
		Size:     getInt64(status, "total_size"),
		Progress: getFloat64(status, "progress") / 100.0,
		Status:   d.mapState(getString(status, "state")),
		SavePath: getString(status, "save_path"),
		Category: getString(status, "label"),
	}

	if eta, ok := status["eta"].(float64); ok && eta > 0 {
		info.ETA = int64(eta)
	}

	if speed, ok := status["download_payload_rate"].(float64); ok {
		info.DownloadSpeed = int64(speed)
	}

	if done, ok := status["total_done"].(float64); ok {
		info.Downloaded = int64(done)
	}

	return info, nil
}

func (d *DelugeClient) mapState(state string) DownloadStatus {
	switch strings.ToLower(state) {
	case "downloading":
		return StatusDownloading
	case "seeding":
		return StatusCompleted
	case "paused":
		return StatusPaused
	case "queued":
		return StatusQueued
	case "error":
		return StatusFailed
	case "checking":
		return StatusDownloading
	default:
		return StatusDownloading
	}
}

// GetAllDownloads implements the Client interface
func (d *DelugeClient) GetAllDownloads(ctx context.Context, category string) ([]DownloadInfo, error) {
	if err := d.Login(ctx); err != nil {
		return nil, err
	}

	keys := []string{
		"name", "total_size", "progress", "state", "download_payload_rate",
		"eta", "save_path", "total_done", "label", "hash",
	}

	filterDict := make(map[string]interface{})
	if category != "" {
		filterDict["label"] = category
	}

	result, err := d.call(ctx, "core.get_torrents_status", filterDict, keys)
	if err != nil {
		return nil, err
	}

	var torrents map[string]map[string]interface{}
	if err := json.Unmarshal(result, &torrents); err != nil {
		return nil, err
	}

	downloads := make([]DownloadInfo, 0, len(torrents))
	for hash, status := range torrents {
		info := DownloadInfo{
			ID:            hash,
			Name:          getString(status, "name"),
			Size:          getInt64(status, "total_size"),
			Progress:      getFloat64(status, "progress") / 100.0,
			Status:        d.mapState(getString(status, "state")),
			DownloadSpeed: getInt64(status, "download_payload_rate"),
			ETA:           getInt64(status, "eta"),
			SavePath:      getString(status, "save_path"),
			Downloaded:    getInt64(status, "total_done"),
			Category:      getString(status, "label"),
		}
		downloads = append(downloads, info)
	}

	return downloads, nil
}

// RemoveDownload implements the Client interface
func (d *DelugeClient) RemoveDownload(ctx context.Context, id string, deleteFiles bool) error {
	if err := d.Login(ctx); err != nil {
		return err
	}

	_, err := d.call(ctx, "core.remove_torrent", id, deleteFiles)
	return err
}

// PauseDownload implements the Client interface
func (d *DelugeClient) PauseDownload(ctx context.Context, id string) error {
	if err := d.Login(ctx); err != nil {
		return err
	}

	_, err := d.call(ctx, "core.pause_torrent", []string{id})
	return err
}

// ResumeDownload implements the Client interface
func (d *DelugeClient) ResumeDownload(ctx context.Context, id string) error {
	if err := d.Login(ctx); err != nil {
		return err
	}

	_, err := d.call(ctx, "core.resume_torrent", []string{id})
	return err
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
