package db

import (
	"time"

	"gorm.io/gorm"
)

// BookStatus represents the current state of a book in the library
type BookStatus string

const (
	StatusMissing     BookStatus = "missing"
	StatusDownloading BookStatus = "downloading"
	StatusDownloaded  BookStatus = "downloaded"
	StatusUnmonitored BookStatus = "unmonitored"
	StatusUnreleased  BookStatus = "unreleased"
)

// MediaType distinguishes between ebooks and audiobooks
type MediaType string

const (
	MediaTypeEbook     MediaType = "ebook"
	MediaTypeAudiobook MediaType = "audiobook"
)

// Author represents a book author
type Author struct {
	gorm.Model
	HardcoverID string `gorm:"uniqueIndex"`
	Name        string `gorm:"index"`
	SortName    string
	Biography   string
	ImageURL    string
	Monitored   bool `gorm:"default:false"`
	Books       []Book

	// Cached metadata from Hardcover
	TotalBooksCount int        `gorm:"default:0"` // Total books by author from Hardcover
	CachedAt        *time.Time // When Hardcover data was last cached
}

// Series represents a book series
type Series struct {
	gorm.Model
	HardcoverID string `gorm:"uniqueIndex"`
	Name        string `gorm:"index"`
	Description string
	Books       []Book

	// Cached metadata from Hardcover
	TotalBooksCount int        `gorm:"default:0"` // Total books in series from Hardcover
	CachedAt        *time.Time // When Hardcover data was last cached
}

// Book represents a book entry in the library
type Book struct {
	gorm.Model
	HardcoverID string `gorm:"uniqueIndex"`
	Title       string `gorm:"index"`
	SortTitle   string
	ISBN        string `gorm:"index"`
	ISBN13      string `gorm:"index"`
	Description string
	CoverURL    string
	Rating      float32
	ReleaseDate *time.Time
	PageCount   int

	// Relationships
	AuthorID    uint
	Author      Author
	SeriesID    *uint
	Series      *Series
	SeriesIndex *float32

	// Status tracking
	Status    BookStatus `gorm:"default:'missing'"`
	Monitored bool       `gorm:"default:true"`

	// Media files
	MediaFiles []MediaFile

	// Metadata sync
	LastSyncedAt *time.Time
}

// MediaFile represents a physical file (ebook or audiobook)
type MediaFile struct {
	gorm.Model
	BookID uint
	Book   Book

	// File info
	FilePath  string `gorm:"uniqueIndex"`
	FileName  string
	FileSize  int64
	Format    string // epub, pdf, m4b, mp3, etc.
	MediaType MediaType

	// Quality info (for audiobooks)
	Bitrate  int
	Duration int // seconds

	// Edition info
	EditionName string // "US Edition", "Narrator A", etc.

	// Tracking
	ImportedAt time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"` // Soft delete for recycle bin
}

// User represents an application user
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Email        string
	IsAdmin      bool `gorm:"default:false"`
	CanRead      bool `gorm:"default:true"`
	CanDelete    bool `gorm:"default:false"`

	// SSO support
	RemoteUser string `gorm:"index"` // For header-based auth

	// Reading progress
	ReadProgress []ReadProgress
}

// ReadProgress tracks user progress through media
type ReadProgress struct {
	gorm.Model
	UserID      uint
	User        User
	MediaFileID uint
	MediaFile   MediaFile

	// Progress tracking
	Progress   float32 // 0.0 - 1.0 (percentage for ebooks, timestamp ratio for audio)
	Position   int     // Page number or seconds
	LastReadAt time.Time
}

// Indexer represents a configured search indexer
type Indexer struct {
	gorm.Model
	Name     string
	Type     string // "torznab", "mam", "anna"
	URL      string
	APIKey   string
	Cookie   string // For MAM
	Priority int    `gorm:"default:0"`
	Enabled  bool   `gorm:"default:true"`

	// MAM-specific
	VIPOnly       bool `gorm:"default:false"`
	FreeleechOnly bool `gorm:"default:false"`
}

// DownloadClient represents a configured download client
type DownloadClient struct {
	gorm.Model
	Name     string `json:"name"`
	Type     string `json:"type"` // "qbittorrent", "transmission", "deluge", "sabnzbd", "nzbget", "rtorrent"
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Category string `json:"category"`
	Priority int    `json:"priority" gorm:"default:0"`
	Enabled  bool   `json:"enabled" gorm:"default:true"`
	Settings string `json:"settings" gorm:"type:text"` // JSON for extra settings (SSL, port, seedbox type, path mappings)
}

// QualityProfile defines format/quality preferences
type QualityProfile struct {
	gorm.Model
	Name      string
	MediaType MediaType

	// For ebooks: comma-separated format ranking (e.g., "epub,azw3,mobi,pdf")
	// For audiobooks: comma-separated format ranking (e.g., "m4b,mp3")
	FormatRanking string

	// Audiobook specific
	MinBitrate int `gorm:"default:0"` // Minimum acceptable bitrate
}

// Notification represents a notification configuration
type Notification struct {
	gorm.Model
	Name    string
	Type    string // "webhook", "discord", "telegram", "email"
	Enabled bool   `gorm:"default:true"`

	// Connection settings
	WebhookURL       string
	DiscordWebhook   string
	TelegramBotToken string
	TelegramChatID   string
	EmailTo          string

	// Triggers
	OnGrab        bool `gorm:"default:false"`
	OnDownload    bool `gorm:"default:true"`
	OnUpgrade     bool `gorm:"default:false"`
	OnImport      bool `gorm:"default:true"`
	OnDelete      bool `gorm:"default:false"`
	OnHealthIssue bool `gorm:"default:true"`
}

// HardcoverList represents a monitored Hardcover.app list
type HardcoverList struct {
	gorm.Model
	Name            string
	HardcoverURL    string `gorm:"uniqueIndex"`
	HardcoverID     string `gorm:"index"`
	Enabled         bool   `gorm:"default:true"`
	Monitor         bool   `gorm:"default:true"`
	AutoAdd         bool   `gorm:"default:true"`
	SyncIntervalHrs int    `gorm:"default:6"`
	QualityProfile  uint
	LastSyncedAt    *time.Time
}

// Download represents an active or completed download
type Download struct {
	gorm.Model
	BookID       uint   `gorm:"index"`
	ClientID     uint   `gorm:"index"`
	ClientType   string // qbittorrent, transmission, sabnzbd, etc.
	ExternalID   string `gorm:"index"` // Hash for torrents, NZB ID for usenet
	MediaType    string // ebook or audiobook - allows both types per book
	Title        string
	DownloadURL  string
	OutputPath   string
	Size         int64
	Downloaded   int64
	Progress     float64
	Status       string `gorm:"default:'queued'"` // queued, downloading, paused, completed, failed, importing
	Category     string
	ErrorMessage string
	AddedAt      int64
	CompletedAt  int64
}

// Setting represents a key-value configuration setting stored in the database
type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// RootFolder represents a configured media library root folder
type RootFolder struct {
	gorm.Model
	Path       string    `gorm:"uniqueIndex"`
	MediaType  MediaType // ebook or audiobook
	Name       string    // Optional display name
	FreeSpace  int64     `gorm:"-"` // Calculated at runtime, not stored
	TotalSpace int64     `gorm:"-"` // Calculated at runtime, not stored
}
