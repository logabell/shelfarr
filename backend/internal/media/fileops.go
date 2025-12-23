package media

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileOperation represents the type of file operation to perform
type FileOperation string

const (
	OpMove     FileOperation = "move"
	OpCopy     FileOperation = "copy"
	OpHardlink FileOperation = "hardlink"
)

// FileOperator handles file system operations
type FileOperator struct {
	defaultOperation FileOperation
}

// NewFileOperator creates a new file operator
func NewFileOperator(defaultOp FileOperation) *FileOperator {
	if defaultOp == "" {
		defaultOp = OpHardlink // Default to hardlinks for space efficiency
	}
	return &FileOperator{
		defaultOperation: defaultOp,
	}
}

// ImportFile imports a file to the library using the specified operation
func (f *FileOperator) ImportFile(sourcePath, destPath string, operation FileOperation) error {
	if operation == "" {
		operation = f.defaultOperation
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination file already exists: %s", destPath)
	}

	switch operation {
	case OpMove:
		return f.moveFile(sourcePath, destPath)
	case OpCopy:
		return f.copyFile(sourcePath, destPath)
	case OpHardlink:
		return f.hardlinkFile(sourcePath, destPath)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

// ImportFolder imports a folder (e.g., audiobook with multiple files)
func (f *FileOperator) ImportFolder(sourcePath, destPath string, operation FileOperation) error {
	if operation == "" {
		operation = f.defaultOperation
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destPath, err)
	}

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return f.ImportFile(path, targetPath, operation)
	})
}

func (f *FileOperator) moveFile(src, dst string) error {
	// Try rename first (fastest if on same filesystem)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fall back to copy + delete
	if err := f.copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func (f *FileOperator) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		os.Remove(dst) // Clean up on error
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Preserve permissions
	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

func (f *FileOperator) hardlinkFile(src, dst string) error {
	err := os.Link(src, dst)
	if err != nil {
		// Hardlinks may fail across filesystems, fall back to copy
		return f.copyFile(src, dst)
	}
	return nil
}

// DeleteFile removes a file (moves to recycle bin if configured)
func (f *FileOperator) DeleteFile(path string, recycleBinPath string) error {
	if recycleBinPath != "" {
		// Move to recycle bin instead of deleting
		binPath := filepath.Join(recycleBinPath, filepath.Base(path))
		return f.moveFile(path, binPath)
	}
	return os.Remove(path)
}

// DeleteFolder removes a folder and all contents
func (f *FileOperator) DeleteFolder(path string, recycleBinPath string) error {
	if recycleBinPath != "" {
		// Move to recycle bin instead of deleting
		binPath := filepath.Join(recycleBinPath, filepath.Base(path))
		return os.Rename(path, binPath)
	}
	return os.RemoveAll(path)
}

// PathBuilder generates organized paths for books
type PathBuilder struct {
	booksRoot      string
	audiobooksRoot string
}

// NewPathBuilder creates a new path builder
func NewPathBuilder(booksRoot, audiobooksRoot string) *PathBuilder {
	return &PathBuilder{
		booksRoot:      booksRoot,
		audiobooksRoot: audiobooksRoot,
	}
}

// BuildBookPath generates the path for a book file
// Pattern: /books/Author Name/Book Title/Book Title.epub
func (p *PathBuilder) BuildBookPath(authorName, bookTitle, format string) string {
	safeAuthor := sanitizeFilename(authorName)
	safeTitle := sanitizeFilename(bookTitle)
	filename := fmt.Sprintf("%s.%s", safeTitle, strings.ToLower(format))
	
	return filepath.Join(p.booksRoot, safeAuthor, safeTitle, filename)
}

// BuildSeriesBookPath generates the path for a book in a series
// Pattern: /books/Author Name/Series Name/01 - Book Title/01 - Book Title.epub
func (p *PathBuilder) BuildSeriesBookPath(authorName, seriesName string, seriesIndex int, bookTitle, format string) string {
	safeAuthor := sanitizeFilename(authorName)
	safeSeries := sanitizeFilename(seriesName)
	safeTitle := sanitizeFilename(bookTitle)
	
	folderName := fmt.Sprintf("%02d - %s", seriesIndex, safeTitle)
	filename := fmt.Sprintf("%02d - %s.%s", seriesIndex, safeTitle, strings.ToLower(format))
	
	return filepath.Join(p.booksRoot, safeAuthor, safeSeries, folderName, filename)
}

// BuildAudiobookPath generates the path for an audiobook
// Pattern: /audiobooks/Author Name/Book Title/
func (p *PathBuilder) BuildAudiobookPath(authorName, bookTitle string) string {
	safeAuthor := sanitizeFilename(authorName)
	safeTitle := sanitizeFilename(bookTitle)
	
	return filepath.Join(p.audiobooksRoot, safeAuthor, safeTitle)
}

// BuildAudiobookFilePath generates the path for an audiobook file
// Pattern: /audiobooks/Author Name/Book Title/Book Title.m4b
func (p *PathBuilder) BuildAudiobookFilePath(authorName, bookTitle, format string) string {
	basePath := p.BuildAudiobookPath(authorName, bookTitle)
	safeTitle := sanitizeFilename(bookTitle)
	filename := fmt.Sprintf("%s.%s", safeTitle, strings.ToLower(format))
	
	return filepath.Join(basePath, filename)
}

// sanitizeFilename removes or replaces characters that are problematic in filenames
func sanitizeFilename(name string) string {
	// Replace problematic characters
	replacements := map[string]string{
		"/":  "-",
		"\\": "-",
		":":  " -",
		"*":  "",
		"?":  "",
		"\"": "'",
		"<":  "",
		">":  "",
		"|":  "-",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Remove multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	result = spaceRegex.ReplaceAllString(result, " ")

	// Trim leading/trailing spaces and dots
	result = strings.Trim(result, " .")

	// Limit length
	if len(result) > 200 {
		result = result[:200]
	}

	return result
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

