package media

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ConversionResult represents the result of a format conversion
type ConversionResult struct {
	Success    bool
	InputPath  string
	OutputPath string
	Duration   time.Duration
	Error      string
}

// EbookConverter handles ebook format conversions using Calibre's ebook-convert
type EbookConverter struct {
	calibrePath string // Path to ebook-convert binary
}

// NewEbookConverter creates a new ebook converter
func NewEbookConverter() *EbookConverter {
	// Try to find ebook-convert in common locations
	paths := []string{
		"ebook-convert",                           // In PATH
		"/usr/bin/ebook-convert",                  // Linux
		"/Applications/calibre.app/Contents/MacOS/ebook-convert", // macOS
	}

	var calibrePath string
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			calibrePath = p
			break
		}
	}

	return &EbookConverter{
		calibrePath: calibrePath,
	}
}

// IsAvailable checks if ebook-convert is available
func (e *EbookConverter) IsAvailable() bool {
	return e.calibrePath != ""
}

// GetVersion returns the Calibre version
func (e *EbookConverter) GetVersion(ctx context.Context) (string, error) {
	if !e.IsAvailable() {
		return "", fmt.Errorf("ebook-convert not found")
	}

	cmd := exec.CommandContext(ctx, e.calibrePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ConversionOptions holds options for ebook conversion
type ConversionOptions struct {
	// Output format (epub, mobi, azw3, pdf)
	OutputFormat string
	
	// Metadata options
	Title  string
	Author string
	Series string
	SeriesIndex float64
	
	// EPUB-specific
	EPUBFlatten bool
	
	// MOBI/AZW3-specific
	MobiFileType string // "old" or "both"
	
	// PDF-specific
	PaperSize string
	PageMargin float64
	
	// General
	PreserveCovers bool
}

// Convert converts an ebook from one format to another
func (e *EbookConverter) Convert(ctx context.Context, inputPath string, outputPath string, opts *ConversionOptions) (*ConversionResult, error) {
	if !e.IsAvailable() {
		return nil, fmt.Errorf("ebook-convert not found - please install Calibre")
	}

	start := time.Now()
	result := &ConversionResult{
		InputPath:  inputPath,
		OutputPath: outputPath,
	}

	// Validate input exists
	if _, err := os.Stat(inputPath); err != nil {
		result.Error = fmt.Sprintf("input file not found: %s", inputPath)
		return result, fmt.Errorf(result.Error)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create output directory: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	// Build command arguments
	args := []string{inputPath, outputPath}

	if opts != nil {
		// Metadata options
		if opts.Title != "" {
			args = append(args, "--title", opts.Title)
		}
		if opts.Author != "" {
			args = append(args, "--authors", opts.Author)
		}
		if opts.Series != "" {
			args = append(args, "--series", opts.Series)
		}
		if opts.SeriesIndex > 0 {
			args = append(args, "--series-index", fmt.Sprintf("%.1f", opts.SeriesIndex))
		}

		// Format-specific options
		outputExt := strings.ToLower(filepath.Ext(outputPath))
		
		if outputExt == ".epub" {
			if opts.EPUBFlatten {
				args = append(args, "--epub-flatten")
			}
			args = append(args, "--epub-version", "3")
		}

		if outputExt == ".mobi" || outputExt == ".azw3" {
			if opts.MobiFileType != "" {
				args = append(args, "--mobi-file-type", opts.MobiFileType)
			}
		}

		if outputExt == ".pdf" {
			if opts.PaperSize != "" {
				args = append(args, "--paper-size", opts.PaperSize)
			}
			if opts.PageMargin > 0 {
				margin := fmt.Sprintf("%.1f", opts.PageMargin)
				args = append(args, "--pdf-page-margin-left", margin)
				args = append(args, "--pdf-page-margin-right", margin)
				args = append(args, "--pdf-page-margin-top", margin)
				args = append(args, "--pdf-page-margin-bottom", margin)
			}
		}

		// General options
		if opts.PreserveCovers {
			args = append(args, "--preserve-cover-aspect-ratio")
		}
	}

	// Run conversion
	cmd := exec.CommandContext(ctx, e.calibrePath, args...)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("conversion failed: %v - %s", err, stderr.String())
		return result, fmt.Errorf(result.Error)
	}

	// Verify output was created
	if _, err := os.Stat(outputPath); err != nil {
		result.Error = "conversion completed but output file not found"
		return result, fmt.Errorf(result.Error)
	}

	result.Success = true
	result.Duration = time.Since(start)

	return result, nil
}

// ConvertToFormat is a convenience method to convert to a specific format
func (e *EbookConverter) ConvertToFormat(ctx context.Context, inputPath string, outputFormat string, opts *ConversionOptions) (*ConversionResult, error) {
	// Generate output path
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputDir := filepath.Dir(inputPath)
	outputPath := filepath.Join(outputDir, baseName+"."+strings.ToLower(outputFormat))

	if opts == nil {
		opts = &ConversionOptions{}
	}
	opts.OutputFormat = outputFormat

	return e.Convert(ctx, inputPath, outputPath, opts)
}

// GetSupportedInputFormats returns formats that can be converted from
func (e *EbookConverter) GetSupportedInputFormats() []string {
	return []string{
		"epub", "mobi", "azw", "azw3", "pdf", "txt", "rtf", "html", "htmlz",
		"docx", "odt", "fb2", "lit", "lrf", "pdb", "rb", "cbz", "cbr",
	}
}

// GetSupportedOutputFormats returns formats that can be converted to
func (e *EbookConverter) GetSupportedOutputFormats() []string {
	return []string{
		"epub", "mobi", "azw3", "pdf", "txt", "rtf", "html", "htmlz",
		"docx", "odt", "fb2", "lit", "lrf", "pdb", "rb", "cbz", "cbr",
	}
}

// CanConvert checks if a conversion between two formats is possible
func (e *EbookConverter) CanConvert(inputFormat, outputFormat string) bool {
	inputFormat = strings.ToLower(strings.TrimPrefix(inputFormat, "."))
	outputFormat = strings.ToLower(strings.TrimPrefix(outputFormat, "."))

	inputSupported := false
	for _, f := range e.GetSupportedInputFormats() {
		if f == inputFormat {
			inputSupported = true
			break
		}
	}

	outputSupported := false
	for _, f := range e.GetSupportedOutputFormats() {
		if f == outputFormat {
			outputSupported = true
			break
		}
	}

	return inputSupported && outputSupported
}

// ExtractCover extracts the cover image from an ebook
func (e *EbookConverter) ExtractCover(ctx context.Context, inputPath, outputPath string) error {
	if !e.IsAvailable() {
		return fmt.Errorf("ebook-convert not found")
	}

	// Use ebook-meta to extract cover
	metaPath := strings.Replace(e.calibrePath, "ebook-convert", "ebook-meta", 1)
	
	cmd := exec.CommandContext(ctx, metaPath, inputPath, "--get-cover", outputPath)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract cover: %w", err)
	}

	return nil
}

// GetMetadata extracts metadata from an ebook
func (e *EbookConverter) GetMetadata(ctx context.Context, inputPath string) (map[string]string, error) {
	if !e.IsAvailable() {
		return nil, fmt.Errorf("ebook-convert not found")
	}

	metaPath := strings.Replace(e.calibrePath, "ebook-convert", "ebook-meta", 1)
	
	cmd := exec.CommandContext(ctx, metaPath, inputPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	// Parse output
	metadata := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				metadata[key] = value
			}
		}
	}

	return metadata, nil
}

