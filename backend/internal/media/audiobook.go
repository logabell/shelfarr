package media

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AudiobookProcessor handles audiobook processing using FFmpeg
type AudiobookProcessor struct {
	ffmpegPath  string
	ffprobePath string
}

// NewAudiobookProcessor creates a new audiobook processor
func NewAudiobookProcessor() *AudiobookProcessor {
	// Find FFmpeg and FFprobe
	ffmpegPath, _ := exec.LookPath("ffmpeg")
	ffprobePath, _ := exec.LookPath("ffprobe")

	return &AudiobookProcessor{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

// IsAvailable checks if FFmpeg is available
func (a *AudiobookProcessor) IsAvailable() bool {
	return a.ffmpegPath != "" && a.ffprobePath != ""
}

// GetVersion returns the FFmpeg version
func (a *AudiobookProcessor) GetVersion(ctx context.Context) (string, error) {
	if !a.IsAvailable() {
		return "", fmt.Errorf("ffmpeg not found")
	}

	cmd := exec.CommandContext(ctx, a.ffmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}
	return "", fmt.Errorf("unable to parse version")
}

// AudioFileInfo contains information about an audio file
type AudioFileInfo struct {
	Path     string
	Duration float64 // seconds
	Bitrate  int     // kbps
	Format   string
	Codec    string
	Channels int
	SampleRate int
}

// Chapter represents an audiobook chapter
type Chapter struct {
	Title     string
	StartTime float64 // seconds
	EndTime   float64 // seconds
}

// GetAudioInfo retrieves information about an audio file
func (a *AudiobookProcessor) GetAudioInfo(ctx context.Context, path string) (*AudioFileInfo, error) {
	if !a.IsAvailable() {
		return nil, fmt.Errorf("ffprobe not found")
	}

	cmd := exec.CommandContext(ctx, a.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to probe file: %w", err)
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecName  string `json:"codec_name"`
			CodecType  string `json:"codec_type"`
			Channels   int    `json:"channels"`
			SampleRate string `json:"sample_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse probe output: %w", err)
	}

	info := &AudioFileInfo{
		Path:   path,
		Format: strings.ToLower(filepath.Ext(path)[1:]),
	}

	if duration, err := strconv.ParseFloat(result.Format.Duration, 64); err == nil {
		info.Duration = duration
	}
	if bitrate, err := strconv.Atoi(result.Format.BitRate); err == nil {
		info.Bitrate = bitrate / 1000
	}

	for _, stream := range result.Streams {
		if stream.CodecType == "audio" {
			info.Codec = stream.CodecName
			info.Channels = stream.Channels
			if sr, err := strconv.Atoi(stream.SampleRate); err == nil {
				info.SampleRate = sr
			}
			break
		}
	}

	return info, nil
}

// M4BConversionOptions holds options for M4B conversion
type M4BConversionOptions struct {
	Title       string
	Author      string
	Album       string
	Genre       string
	Year        string
	CoverPath   string
	Chapters    []Chapter
	Bitrate     int  // kbps, 0 for auto
	SampleRate  int  // Hz, 0 for auto
	Normalize   bool
}

// ConvertToM4B converts audio files to a single M4B audiobook
func (a *AudiobookProcessor) ConvertToM4B(ctx context.Context, inputPaths []string, outputPath string, opts *M4BConversionOptions) (*ConversionResult, error) {
	if !a.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg not found")
	}

	start := time.Now()
	result := &ConversionResult{
		InputPath:  strings.Join(inputPaths, ", "),
		OutputPath: outputPath,
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create output directory: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	// Sort input files
	sort.Strings(inputPaths)

	// Create concat file for FFmpeg
	concatFile, err := os.CreateTemp("", "audiobook-concat-*.txt")
	if err != nil {
		result.Error = fmt.Sprintf("failed to create concat file: %v", err)
		return result, fmt.Errorf(result.Error)
	}
	defer os.Remove(concatFile.Name())

	for _, path := range inputPaths {
		// Escape single quotes in paths
		escapedPath := strings.ReplaceAll(path, "'", "'\\''")
		fmt.Fprintf(concatFile, "file '%s'\n", escapedPath)
	}
	concatFile.Close()

	// Build FFmpeg arguments
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile.Name(),
	}

	// Add cover image if provided
	if opts != nil && opts.CoverPath != "" {
		if _, err := os.Stat(opts.CoverPath); err == nil {
			args = append(args, "-i", opts.CoverPath)
			args = append(args, "-map", "0:a", "-map", "1:v")
			args = append(args, "-c:v", "copy")
			args = append(args, "-disposition:v:0", "attached_pic")
		}
	}

	// Audio encoding settings
	args = append(args, "-c:a", "aac")
	
	if opts != nil && opts.Bitrate > 0 {
		args = append(args, "-b:a", fmt.Sprintf("%dk", opts.Bitrate))
	} else {
		args = append(args, "-b:a", "128k")
	}

	if opts != nil && opts.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(opts.SampleRate))
	}

	// Normalization
	if opts != nil && opts.Normalize {
		args = append(args, "-af", "loudnorm=I=-16:LRA=11:TP=-1.5")
	}

	// Metadata
	if opts != nil {
		if opts.Title != "" {
			args = append(args, "-metadata", fmt.Sprintf("title=%s", opts.Title))
		}
		if opts.Author != "" {
			args = append(args, "-metadata", fmt.Sprintf("artist=%s", opts.Author))
			args = append(args, "-metadata", fmt.Sprintf("album_artist=%s", opts.Author))
		}
		if opts.Album != "" {
			args = append(args, "-metadata", fmt.Sprintf("album=%s", opts.Album))
		}
		if opts.Genre != "" {
			args = append(args, "-metadata", fmt.Sprintf("genre=%s", opts.Genre))
		} else {
			args = append(args, "-metadata", "genre=Audiobook")
		}
		if opts.Year != "" {
			args = append(args, "-metadata", fmt.Sprintf("date=%s", opts.Year))
		}
	}

	// Output
	args = append(args, "-y", outputPath)

	// Run FFmpeg
	cmd := exec.CommandContext(ctx, a.ffmpegPath, args...)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("conversion failed: %v - %s", err, stderr.String())
		return result, fmt.Errorf(result.Error)
	}

	// Add chapters if provided
	if opts != nil && len(opts.Chapters) > 0 {
		if err := a.addChapters(ctx, outputPath, opts.Chapters); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to add chapters: %v\n", err)
		}
	} else if len(inputPaths) > 1 {
		// Auto-generate chapters from input files
		chapters, err := a.generateChaptersFromFiles(ctx, inputPaths)
		if err == nil && len(chapters) > 0 {
			a.addChapters(ctx, outputPath, chapters)
		}
	}

	result.Success = true
	result.Duration = time.Since(start)

	return result, nil
}

// generateChaptersFromFiles creates chapters based on input file names and durations
func (a *AudiobookProcessor) generateChaptersFromFiles(ctx context.Context, paths []string) ([]Chapter, error) {
	var chapters []Chapter
	var currentTime float64

	for i, path := range paths {
		info, err := a.GetAudioInfo(ctx, path)
		if err != nil {
			continue
		}

		// Get chapter name from filename
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		// Clean up common patterns
		name = strings.TrimPrefix(name, fmt.Sprintf("%02d", i+1))
		name = strings.TrimPrefix(name, fmt.Sprintf("%d", i+1))
		name = strings.TrimPrefix(name, " - ")
		name = strings.TrimPrefix(name, "- ")
		name = strings.TrimPrefix(name, "_")
		name = strings.TrimSpace(name)
		
		if name == "" {
			name = fmt.Sprintf("Chapter %d", i+1)
		}

		chapters = append(chapters, Chapter{
			Title:     name,
			StartTime: currentTime,
			EndTime:   currentTime + info.Duration,
		})

		currentTime += info.Duration
	}

	return chapters, nil
}

// addChapters adds chapter metadata to an M4B file
func (a *AudiobookProcessor) addChapters(ctx context.Context, m4bPath string, chapters []Chapter) error {
	// Create chapters file in FFmetadata format
	chaptersFile, err := os.CreateTemp("", "chapters-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(chaptersFile.Name())

	fmt.Fprintln(chaptersFile, ";FFMETADATA1")
	
	for _, ch := range chapters {
		fmt.Fprintln(chaptersFile, "[CHAPTER]")
		fmt.Fprintf(chaptersFile, "TIMEBASE=1/1000\n")
		fmt.Fprintf(chaptersFile, "START=%d\n", int64(ch.StartTime*1000))
		fmt.Fprintf(chaptersFile, "END=%d\n", int64(ch.EndTime*1000))
		fmt.Fprintf(chaptersFile, "title=%s\n", ch.Title)
	}
	chaptersFile.Close()

	// Create temp output file
	tempOutput := m4bPath + ".temp.m4b"

	// Run FFmpeg to add chapters
	cmd := exec.CommandContext(ctx, a.ffmpegPath,
		"-i", m4bPath,
		"-i", chaptersFile.Name(),
		"-map_metadata", "1",
		"-codec", "copy",
		"-y", tempOutput,
	)

	if err := cmd.Run(); err != nil {
		os.Remove(tempOutput)
		return fmt.Errorf("failed to add chapters: %w", err)
	}

	// Replace original with chaptered version
	os.Remove(m4bPath)
	return os.Rename(tempOutput, m4bPath)
}

// SplitByChapters splits an audiobook into individual chapter files
func (a *AudiobookProcessor) SplitByChapters(ctx context.Context, inputPath string, outputDir string) ([]string, error) {
	if !a.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg not found")
	}

	// Get chapters from file
	chapters, err := a.ExtractChapters(ctx, inputPath)
	if err != nil || len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found in file")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	var outputPaths []string
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	for i, ch := range chapters {
		outputName := fmt.Sprintf("%02d - %s.m4a", i+1, sanitizeFilename(ch.Title))
		outputPath := filepath.Join(outputDir, outputName)

		cmd := exec.CommandContext(ctx, a.ffmpegPath,
			"-i", inputPath,
			"-ss", fmt.Sprintf("%.3f", ch.StartTime),
			"-to", fmt.Sprintf("%.3f", ch.EndTime),
			"-c", "copy",
			"-y", outputPath,
		)

		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to extract chapter %d: %v\n", i+1, err)
			continue
		}

		outputPaths = append(outputPaths, outputPath)
	}

	_ = baseName // Suppress unused warning

	return outputPaths, nil
}

// ExtractChapters extracts chapter information from an audio file
func (a *AudiobookProcessor) ExtractChapters(ctx context.Context, path string) ([]Chapter, error) {
	if !a.IsAvailable() {
		return nil, fmt.Errorf("ffprobe not found")
	}

	cmd := exec.CommandContext(ctx, a.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_chapters",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to probe chapters: %w", err)
	}

	var result struct {
		Chapters []struct {
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			Tags      struct {
				Title string `json:"title"`
			} `json:"tags"`
		} `json:"chapters"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse chapters: %w", err)
	}

	var chapters []Chapter
	for i, ch := range result.Chapters {
		title := ch.Tags.Title
		if title == "" {
			title = fmt.Sprintf("Chapter %d", i+1)
		}

		start, _ := strconv.ParseFloat(ch.StartTime, 64)
		end, _ := strconv.ParseFloat(ch.EndTime, 64)

		chapters = append(chapters, Chapter{
			Title:     title,
			StartTime: start,
			EndTime:   end,
		})
	}

	return chapters, nil
}

// ParseCueSheet parses a CUE sheet file for chapter information
func ParseCueSheet(path string) ([]Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chapters []Chapter
	scanner := bufio.NewScanner(file)
	
	var currentTitle string
	var currentIndex float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "TITLE") {
			// Extract title (remove quotes)
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				currentTitle = strings.Trim(parts[1], "\"")
			}
		} else if strings.HasPrefix(line, "INDEX 01") {
			// Parse time (format: MM:SS:FF where FF is frames, 75 frames per second)
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				timeParts := strings.Split(parts[2], ":")
				if len(timeParts) == 3 {
					min, _ := strconv.Atoi(timeParts[0])
					sec, _ := strconv.Atoi(timeParts[1])
					frames, _ := strconv.Atoi(timeParts[2])
					
					startTime := float64(min*60) + float64(sec) + float64(frames)/75.0
					
					if len(chapters) > 0 {
						chapters[len(chapters)-1].EndTime = startTime
					}
					
					chapters = append(chapters, Chapter{
						Title:     currentTitle,
						StartTime: startTime,
					})
				}
			}
			currentIndex++
		}
	}

	return chapters, scanner.Err()
}

