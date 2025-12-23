package media

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileInfo represents information about a discovered file
type FileInfo struct {
	Path      string
	Name      string
	Size      int64
	ModTime   time.Time
	Format    string
	MediaType string // "ebook" or "audiobook"
	IsDir     bool
}

// Scanner handles file system scanning operations
type Scanner struct {
	ebookFormats     []string
	audiobookFormats []string
}

// NewScanner creates a new file scanner
func NewScanner() *Scanner {
	return &Scanner{
		ebookFormats:     []string{".epub", ".pdf", ".mobi", ".azw", ".azw3", ".cbz", ".cbr", ".fb2", ".djvu"},
		audiobookFormats: []string{".mp3", ".m4a", ".m4b", ".flac", ".ogg", ".wma", ".aac"},
	}
}

// ScanDirectory scans a directory for media files
func (s *Scanner) ScanDirectory(path string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if d.IsDir() {
			// Check if this directory might be an audiobook folder (contains multiple audio files)
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		format := strings.TrimPrefix(ext, ".")
		mediaType := s.detectMediaType(ext)

		if mediaType != "" {
			files = append(files, FileInfo{
				Path:      filePath,
				Name:      d.Name(),
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				Format:    format,
				MediaType: mediaType,
				IsDir:     false,
			})
		}

		return nil
	})

	return files, err
}

// ScanForAudiobookFolders finds directories containing multiple audio files
func (s *Scanner) ScanForAudiobookFolders(path string) ([]FileInfo, error) {
	var folders []FileInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		folderPath := filepath.Join(path, entry.Name())
		audioFiles, _ := s.countAudioFiles(folderPath)

		if audioFiles >= 2 { // Likely an audiobook folder
			info, _ := entry.Info()
			folders = append(folders, FileInfo{
				Path:      folderPath,
				Name:      entry.Name(),
				Size:      0, // Will be calculated separately
				ModTime:   info.ModTime(),
				Format:    "folder",
				MediaType: "audiobook",
				IsDir:     true,
			})
		}
	}

	return folders, nil
}

func (s *Scanner) countAudioFiles(path string) (int, int64) {
	var count int
	var totalSize int64

	filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if s.isAudiobook(ext) {
			count++
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})

	return count, totalSize
}

func (s *Scanner) detectMediaType(ext string) string {
	ext = strings.ToLower(ext)
	
	for _, format := range s.ebookFormats {
		if ext == format {
			return "ebook"
		}
	}
	
	for _, format := range s.audiobookFormats {
		if ext == format {
			return "audiobook"
		}
	}

	return ""
}

func (s *Scanner) isAudiobook(ext string) bool {
	ext = strings.ToLower(ext)
	for _, format := range s.audiobookFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// ExtractMetadataFromFilename attempts to extract book info from filename
func (s *Scanner) ExtractMetadataFromFilename(filename string) (author, title, series string, seriesNum int) {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Common patterns:
	// "Author - Title"
	// "Author - Series #01 - Title"
	// "Title (Author)"
	// "Author_Title"

	// Try "Author - Title" pattern
	if parts := strings.SplitN(name, " - ", 2); len(parts) == 2 {
		author = strings.TrimSpace(parts[0])
		title = strings.TrimSpace(parts[1])
		
		// Check for series pattern in title: "Series #01 - Title" or "Series 01 - Title"
		seriesPattern := regexp.MustCompile(`^(.+?)\s*[#]?(\d+)\s*[-–]\s*(.+)$`)
		if matches := seriesPattern.FindStringSubmatch(title); len(matches) == 4 {
			series = matches[1]
			fmt.Sscanf(matches[2], "%d", &seriesNum)
			title = matches[3]
		}
		return
	}

	// Try "Title (Author)" pattern
	parenPattern := regexp.MustCompile(`^(.+?)\s*\(([^)]+)\)$`)
	if matches := parenPattern.FindStringSubmatch(name); len(matches) == 3 {
		title = strings.TrimSpace(matches[1])
		author = strings.TrimSpace(matches[2])
		return
	}

	// Try underscore pattern
	if parts := strings.SplitN(name, "_", 2); len(parts) == 2 {
		author = strings.TrimSpace(parts[0])
		title = strings.TrimSpace(parts[1])
		return
	}

	// Default: whole name is title
	title = name
	return
}

// CalculateFolderSize calculates total size of files in a folder
func (s *Scanner) CalculateFolderSize(path string) (int64, error) {
	var size int64
	
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

