package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SABnzbdClient handles communication with SABnzbd API
type SABnzbdClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSABnzbdClient creates a new SABnzbd client
func NewSABnzbdClient(baseURL, apiKey string) *SABnzbdClient {
	return &SABnzbdClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SABnzbdClient) Type() string {
	return "sabnzbd"
}

// Test checks if the connection is working
func (s *SABnzbdClient) Test(ctx context.Context) error {
	_, err := s.GetVersion(ctx)
	return err
}

// GetVersion returns the SABnzbd version
func (s *SABnzbdClient) GetVersion(ctx context.Context) (string, error) {
	resp, err := s.request(ctx, "version", nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	return result.Version, nil
}

// sabQueue represents the SABnzbd queue response
type sabQueue struct {
	Queue struct {
		Slots []sabSlot `json:"slots"`
	} `json:"queue"`
}

type sabSlot struct {
	NZOID       string `json:"nzo_id"`
	Filename    string `json:"filename"`
	Size        string `json:"size"`       // Human readable size
	SizeLeft    string `json:"sizeleft"`   // Human readable
	Percentage  string `json:"percentage"` // "0" to "100"
	Status      string `json:"status"`     // Downloading, Paused, Queued, etc.
	TimeLeft    string `json:"timeleft"`   // "0:00:00"
	Category    string `json:"cat"`
	MBLeft      string `json:"mbleft"`
	MB          string `json:"mb"`
}

// sabHistory represents the SABnzbd history response
type sabHistory struct {
	History struct {
		Slots []sabHistorySlot `json:"slots"`
	} `json:"history"`
}

type sabHistorySlot struct {
	NZOID      string `json:"nzo_id"`
	Name       string `json:"name"`
	Size       int64  `json:"bytes"`
	Status     string `json:"status"` // Completed, Failed, etc.
	Category   string `json:"category"`
	Storage    string `json:"storage"` // Path to downloaded files
	FailMessage string `json:"fail_message"`
}

// AddDownload adds an NZB download from URL
func (s *SABnzbdClient) AddDownload(ctx context.Context, downloadURL string, opts *DownloadOptions) (string, error) {
	params := map[string]string{
		"name": downloadURL,
	}

	if opts != nil {
		if opts.Category != "" {
			params["cat"] = opts.Category
		}
		if opts.Priority != 0 {
			params["priority"] = fmt.Sprintf("%d", opts.Priority)
		}
		if opts.Paused {
			params["pp"] = "-1" // Process priority -1 = paused
		}
	}

	resp, err := s.request(ctx, "addurl", params)
	if err != nil {
		return "", err
	}

	var result struct {
		Status bool     `json:"status"`
		NZOIDs []string `json:"nzo_ids"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if !result.Status || len(result.NZOIDs) == 0 {
		return "", fmt.Errorf("failed to add NZB")
	}

	return result.NZOIDs[0], nil
}

// GetDownload returns information about a specific download
func (s *SABnzbdClient) GetDownload(ctx context.Context, id string) (*DownloadInfo, error) {
	// Check queue first
	queueResp, err := s.request(ctx, "queue", nil)
	if err == nil {
		var queue sabQueue
		if json.Unmarshal(queueResp, &queue) == nil {
			for _, slot := range queue.Queue.Slots {
				if slot.NZOID == id {
					return s.queueSlotToInfo(slot), nil
				}
			}
		}
	}

	// Check history
	histResp, err := s.request(ctx, "history", nil)
	if err != nil {
		return nil, err
	}

	var history sabHistory
	if err := json.Unmarshal(histResp, &history); err != nil {
		return nil, err
	}

	for _, slot := range history.History.Slots {
		if slot.NZOID == id {
			return s.historySlotToInfo(slot), nil
		}
	}

	return nil, fmt.Errorf("download not found: %s", id)
}

// GetAllDownloads returns all downloads in queue and recent history
func (s *SABnzbdClient) GetAllDownloads(ctx context.Context, category string) ([]DownloadInfo, error) {
	var downloads []DownloadInfo

	// Get queue
	queueResp, err := s.request(ctx, "queue", nil)
	if err == nil {
		var queue sabQueue
		if json.Unmarshal(queueResp, &queue) == nil {
			for _, slot := range queue.Queue.Slots {
				if category == "" || slot.Category == category {
					downloads = append(downloads, *s.queueSlotToInfo(slot))
				}
			}
		}
	}

	// Get history
	histResp, err := s.request(ctx, "history", map[string]string{"limit": "50"})
	if err == nil {
		var history sabHistory
		if json.Unmarshal(histResp, &history) == nil {
			for _, slot := range history.History.Slots {
				if category == "" || slot.Category == category {
					downloads = append(downloads, *s.historySlotToInfo(slot))
				}
			}
		}
	}

	return downloads, nil
}

// RemoveDownload removes a download
func (s *SABnzbdClient) RemoveDownload(ctx context.Context, id string, deleteFiles bool) error {
	// Try queue first
	_, err := s.request(ctx, "queue", map[string]string{
		"name":  "delete",
		"value": id,
	})
	if err == nil {
		return nil
	}

	// Try history
	delFiles := "0"
	if deleteFiles {
		delFiles = "1"
	}
	_, err = s.request(ctx, "history", map[string]string{
		"name":       "delete",
		"value":      id,
		"del_files":  delFiles,
	})
	return err
}

// PauseDownload pauses a download
func (s *SABnzbdClient) PauseDownload(ctx context.Context, id string) error {
	_, err := s.request(ctx, "queue", map[string]string{
		"name":  "pause",
		"value": id,
	})
	return err
}

// ResumeDownload resumes a download
func (s *SABnzbdClient) ResumeDownload(ctx context.Context, id string) error {
	_, err := s.request(ctx, "queue", map[string]string{
		"name":  "resume",
		"value": id,
	})
	return err
}

// GetCategories returns all categories
func (s *SABnzbdClient) GetCategories(ctx context.Context) ([]string, error) {
	resp, err := s.request(ctx, "get_cats", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Categories, nil
}

func (s *SABnzbdClient) request(ctx context.Context, mode string, params map[string]string) ([]byte, error) {
	reqURL, err := url.Parse(s.baseURL + "/api")
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Set("output", "json")
	q.Set("apikey", s.apiKey)
	q.Set("mode", mode)
	
	for k, v := range params {
		q.Set(k, v)
	}
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for API error
	var apiError struct {
		Status bool   `json:"status"`
		Error  string `json:"error"`
	}
	if json.Unmarshal(body, &apiError) == nil && !apiError.Status && apiError.Error != "" {
		return nil, fmt.Errorf("SABnzbd error: %s", apiError.Error)
	}

	return body, nil
}

func (s *SABnzbdClient) queueSlotToInfo(slot sabSlot) *DownloadInfo {
	var progress float64
	fmt.Sscanf(slot.Percentage, "%f", &progress)

	var mb, mbleft float64
	fmt.Sscanf(slot.MB, "%f", &mb)
	fmt.Sscanf(slot.MBLeft, "%f", &mbleft)

	status := StatusQueued
	switch strings.ToLower(slot.Status) {
	case "downloading":
		status = StatusDownloading
	case "paused":
		status = StatusPaused
	}

	return &DownloadInfo{
		ID:         slot.NZOID,
		Name:       slot.Filename,
		Size:       int64(mb * 1024 * 1024),
		Downloaded: int64((mb - mbleft) * 1024 * 1024),
		Progress:   progress / 100,
		Status:     status,
		Category:   slot.Category,
	}
}

func (s *SABnzbdClient) historySlotToInfo(slot sabHistorySlot) *DownloadInfo {
	status := StatusCompleted
	if strings.ToLower(slot.Status) == "failed" {
		status = StatusFailed
	}

	return &DownloadInfo{
		ID:       slot.NZOID,
		Name:     slot.Name,
		Size:     slot.Size,
		Downloaded: slot.Size,
		Progress: 1.0,
		Status:   status,
		SavePath: slot.Storage,
		Category: slot.Category,
	}
}

