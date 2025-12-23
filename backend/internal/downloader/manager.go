package downloader

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DownloadStatus represents the status of a download
type DownloadStatus string

const (
	StatusQueued      DownloadStatus = "queued"
	StatusDownloading DownloadStatus = "downloading"
	StatusPaused      DownloadStatus = "paused"
	StatusCompleted   DownloadStatus = "completed"
	StatusFailed      DownloadStatus = "failed"
	StatusImporting   DownloadStatus = "importing"
)

// Download represents an active or completed download
type Download struct {
	ID           uint   `gorm:"primaryKey"`
	ClientType   string // qbittorrent, transmission, sabnzbd, etc.
	ClientID     uint   // ID of the download client in the database
	ExternalID   string // Hash for torrents, NZB ID for usenet
	MediaType    string // ebook or audiobook - allows both types per book
	BookID       uint   `gorm:"index"`
	Title        string
	Size         int64
	Downloaded   int64
	Progress     float64
	Status       DownloadStatus
	Category     string
	DownloadURL  string
	OutputPath   string
	ErrorMessage string
	AddedAt      int64
	CompletedAt  int64
}

func (Download) TableName() string {
	return "downloads"
}

// Client is the interface that all download clients must implement
type Client interface {
	Type() string
	Test(ctx context.Context) error
	AddDownload(ctx context.Context, url string, opts *DownloadOptions) (string, error) // Returns external ID
	GetDownload(ctx context.Context, id string) (*DownloadInfo, error)
	GetAllDownloads(ctx context.Context, category string) ([]DownloadInfo, error)
	RemoveDownload(ctx context.Context, id string, deleteFiles bool) error
	PauseDownload(ctx context.Context, id string) error
	ResumeDownload(ctx context.Context, id string) error
}

// DownloadOptions holds options for adding a download
type DownloadOptions struct {
	Category string
	SavePath string
	Paused   bool
	Priority int // 0 = normal, 1 = high, -1 = low
}

// DownloadInfo holds information about a download
type DownloadInfo struct {
	ID            string
	Name          string
	Size          int64
	Downloaded    int64
	Progress      float64
	Status        DownloadStatus
	DownloadSpeed int64
	ETA           int64 // seconds
	SavePath      string
	Category      string
}

// Manager manages multiple download clients and downloads
type Manager struct {
	db      *gorm.DB
	clients map[uint]Client
}

// NewManager creates a new download manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:      db,
		clients: make(map[uint]Client),
	}
}

// RegisterClient registers a download client
func (m *Manager) RegisterClient(clientID uint, client Client) {
	m.clients[clientID] = client
}

// GetClient returns a registered client by ID
func (m *Manager) GetClient(clientID uint) (Client, error) {
	client, ok := m.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("client not found: %d", clientID)
	}
	return client, nil
}

// AddDownload adds a new download
func (m *Manager) AddDownload(ctx context.Context, clientID uint, bookID uint, title, downloadURL string, opts *DownloadOptions) (*Download, error) {
	client, err := m.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	// Add to external client
	externalID, err := client.AddDownload(ctx, downloadURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to add download: %w", err)
	}

	// Create database record
	download := &Download{
		ClientType:  client.Type(),
		ClientID:    clientID,
		ExternalID:  externalID,
		BookID:      bookID,
		Title:       title,
		Status:      StatusQueued,
		DownloadURL: downloadURL,
	}

	if opts != nil {
		download.Category = opts.Category
	}

	if err := m.db.Create(download).Error; err != nil {
		return nil, err
	}

	return download, nil
}

// SyncDownloads synchronizes download status with external clients
func (m *Manager) SyncDownloads(ctx context.Context) error {
	var downloads []Download
	if err := m.db.Where("status IN ?", []DownloadStatus{StatusQueued, StatusDownloading, StatusPaused}).Find(&downloads).Error; err != nil {
		return err
	}

	for _, download := range downloads {
		client, err := m.GetClient(download.ClientID)
		if err != nil {
			continue
		}

		info, err := client.GetDownload(ctx, download.ExternalID)
		if err != nil {
			continue
		}

		// Update download record
		download.Status = info.Status
		download.Progress = info.Progress
		download.Downloaded = info.Downloaded
		download.Size = info.Size

		if info.Status == StatusCompleted {
			download.OutputPath = info.SavePath
		}

		m.db.Save(&download)
	}

	return nil
}

// GetActiveDownloads returns all active downloads
func (m *Manager) GetActiveDownloads(ctx context.Context) ([]Download, error) {
	var downloads []Download
	err := m.db.Where("status IN ?", []DownloadStatus{StatusQueued, StatusDownloading, StatusPaused}).Find(&downloads).Error
	return downloads, err
}

// GetDownloadByBookID returns downloads for a specific book
func (m *Manager) GetDownloadByBookID(ctx context.Context, bookID uint) ([]Download, error) {
	var downloads []Download
	err := m.db.Where("book_id = ?", bookID).Find(&downloads).Error
	return downloads, err
}

// RemoveDownload removes a download
func (m *Manager) RemoveDownload(ctx context.Context, downloadID uint, deleteFiles bool) error {
	var download Download
	if err := m.db.First(&download, downloadID).Error; err != nil {
		return err
	}

	client, err := m.GetClient(download.ClientID)
	if err == nil {
		// Try to remove from external client
		client.RemoveDownload(ctx, download.ExternalID, deleteFiles)
	}

	// Remove from database
	return m.db.Delete(&download).Error
}

// CreateClientFromDB creates a download client from database model
func CreateClientFromDB(clientType, url, username, password string) (Client, error) {
	switch clientType {
	case "qbittorrent":
		return NewQBittorrentClient(url, username, password), nil
	case "sabnzbd":
		return NewSABnzbdClient(url, password), nil // password is API key for SABnzbd
	case "deluge":
		return NewDelugeClient(url, password), nil
	default:
		return nil, fmt.Errorf("unsupported client type: %s", clientType)
	}
}
