package media

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success     bool
	FilePath    string
	NewPath     string
	MediaFileID uint
	Error       string
}

// Importer handles importing media files into the library
type Importer struct {
	db          *gorm.DB
	scanner     *Scanner
	fileOps     *FileOperator
	pathBuilder *PathBuilder
	operation   FileOperation
}

// NewImporter creates a new importer
func NewImporter(db *gorm.DB, booksPath, audiobooksPath string, operation FileOperation) *Importer {
	return &Importer{
		db:          db,
		scanner:     NewScanner(),
		fileOps:     NewFileOperator(operation),
		pathBuilder: NewPathBuilder(booksPath, audiobooksPath),
		operation:   operation,
	}
}

// ImportRequest represents a request to import a file
type ImportRequest struct {
	SourcePath  string
	BookID      uint
	AuthorName  string
	BookTitle   string
	SeriesName  string
	SeriesIndex int
	MediaType   string // "ebook" or "audiobook"
	Format      string
	EditionName string
}

// Import imports a single file or folder into the library
func (i *Importer) Import(req ImportRequest) (*ImportResult, error) {
	result := &ImportResult{
		FilePath: req.SourcePath,
	}

	// Validate source exists
	info, err := os.Stat(req.SourcePath)
	if err != nil {
		result.Error = fmt.Sprintf("source not found: %s", req.SourcePath)
		return result, fmt.Errorf(result.Error)
	}

	var destPath string
	var importErr error

	if info.IsDir() {
		// Import folder (typically audiobook)
		destPath = i.pathBuilder.BuildAudiobookPath(req.AuthorName, req.BookTitle)
		importErr = i.fileOps.ImportFolder(req.SourcePath, destPath, i.operation)
	} else {
		// Import single file
		if req.MediaType == "audiobook" {
			destPath = i.pathBuilder.BuildAudiobookFilePath(req.AuthorName, req.BookTitle, req.Format)
		} else {
			if req.SeriesName != "" && req.SeriesIndex > 0 {
				destPath = i.pathBuilder.BuildSeriesBookPath(req.AuthorName, req.SeriesName, req.SeriesIndex, req.BookTitle, req.Format)
			} else {
				destPath = i.pathBuilder.BuildBookPath(req.AuthorName, req.BookTitle, req.Format)
			}
		}
		importErr = i.fileOps.ImportFile(req.SourcePath, destPath, i.operation)
	}

	if importErr != nil {
		result.Error = importErr.Error()
		return result, importErr
	}

	// Create MediaFile record
	mediaFile := &MediaFileRecord{
		BookID:      req.BookID,
		FilePath:    destPath,
		FileName:    filepath.Base(destPath),
		Format:      req.Format,
		MediaType:   req.MediaType,
		EditionName: req.EditionName,
		ImportedAt:  time.Now(),
	}

	// Get file size
	if fileInfo, err := os.Stat(destPath); err == nil {
		mediaFile.FileSize = fileInfo.Size()
	}

	if err := i.db.Create(mediaFile).Error; err != nil {
		result.Error = fmt.Sprintf("failed to create database record: %v", err)
		return result, err
	}

	// Update book status
	if err := i.db.Model(&BookRecord{}).Where("id = ?", req.BookID).Updates(map[string]interface{}{
		"status": "downloaded",
	}).Error; err != nil {
		// Log but don't fail the import
		fmt.Printf("Warning: failed to update book status: %v\n", err)
	}

	result.Success = true
	result.NewPath = destPath
	result.MediaFileID = mediaFile.ID

	return result, nil
}

// ScanDownloadsFolder scans the downloads folder for pending imports
func (i *Importer) ScanDownloadsFolder(downloadsPath string) ([]PendingImport, error) {
	var pending []PendingImport

	// Scan for individual files
	files, err := i.scanner.ScanDirectory(downloadsPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		author, title, series, seriesNum := i.scanner.ExtractMetadataFromFilename(file.Name)
		
		pending = append(pending, PendingImport{
			Path:           file.Path,
			Name:           file.Name,
			Size:           file.Size,
			Format:         file.Format,
			MediaType:      file.MediaType,
			IsFolder:       file.IsDir,
			ExtractedAuthor: author,
			ExtractedTitle:  title,
			ExtractedSeries: series,
			ExtractedSeriesNum: seriesNum,
		})
	}

	// Scan for audiobook folders
	folders, err := i.scanner.ScanForAudiobookFolders(downloadsPath)
	if err == nil {
		for _, folder := range folders {
			author, title, series, seriesNum := i.scanner.ExtractMetadataFromFilename(folder.Name)
			size, _ := i.scanner.CalculateFolderSize(folder.Path)
			
			pending = append(pending, PendingImport{
				Path:           folder.Path,
				Name:           folder.Name,
				Size:           size,
				Format:         "folder",
				MediaType:      "audiobook",
				IsFolder:       true,
				ExtractedAuthor: author,
				ExtractedTitle:  title,
				ExtractedSeries: series,
				ExtractedSeriesNum: seriesNum,
			})
		}
	}

	return pending, nil
}

// PendingImport represents a file awaiting import
type PendingImport struct {
	Path               string `json:"path"`
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	Format             string `json:"format"`
	MediaType          string `json:"mediaType"`
	IsFolder           bool   `json:"isFolder"`
	ExtractedAuthor    string `json:"extractedAuthor,omitempty"`
	ExtractedTitle     string `json:"extractedTitle,omitempty"`
	ExtractedSeries    string `json:"extractedSeries,omitempty"`
	ExtractedSeriesNum int    `json:"extractedSeriesNum,omitempty"`
}

// MediaFileRecord for database operations
type MediaFileRecord struct {
	ID          uint      `gorm:"primaryKey"`
	BookID      uint      `gorm:"index"`
	FilePath    string    `gorm:"uniqueIndex"`
	FileName    string
	FileSize    int64
	Format      string
	MediaType   string
	Bitrate     int
	Duration    int
	EditionName string
	ImportedAt  time.Time
}

func (MediaFileRecord) TableName() string {
	return "media_files"
}

// BookRecord for database operations
type BookRecord struct {
	ID     uint   `gorm:"primaryKey"`
	Status string
}

func (BookRecord) TableName() string {
	return "books"
}

// RenameFile renames a media file following the naming conventions
func (i *Importer) RenameFile(mediaFileID uint, authorName, bookTitle, seriesName string, seriesIndex int) error {
	var mediaFile MediaFileRecord
	if err := i.db.First(&mediaFile, mediaFileID).Error; err != nil {
		return fmt.Errorf("media file not found: %w", err)
	}

	var newPath string
	format := strings.ToLower(mediaFile.Format)

	if mediaFile.MediaType == "audiobook" {
		newPath = i.pathBuilder.BuildAudiobookFilePath(authorName, bookTitle, format)
	} else {
		if seriesName != "" && seriesIndex > 0 {
			newPath = i.pathBuilder.BuildSeriesBookPath(authorName, seriesName, seriesIndex, bookTitle, format)
		} else {
			newPath = i.pathBuilder.BuildBookPath(authorName, bookTitle, format)
		}
	}

	// Move file to new location
	if err := i.fileOps.ImportFile(mediaFile.FilePath, newPath, OpMove); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Update database record
	mediaFile.FilePath = newPath
	mediaFile.FileName = filepath.Base(newPath)
	
	return i.db.Save(&mediaFile).Error
}

// CleanupEmptyFolders removes empty folders from a path
func (i *Importer) CleanupEmptyFolders(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == rootPath {
			return nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		if len(entries) == 0 {
			os.Remove(path)
		}

		return nil
	})
}

