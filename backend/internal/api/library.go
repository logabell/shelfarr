package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// LibraryResponse represents the library grid data
type LibraryResponse struct {
	Books    []BookResponse `json:"books"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// BookResponse represents a book in API responses
type BookResponse struct {
	ID           uint                `json:"id"`
	HardcoverID  string              `json:"hardcoverId"`
	Title        string              `json:"title"`
	SortTitle    string              `json:"sortTitle"`
	ISBN         string              `json:"isbn"`
	Description  string              `json:"description"`
	CoverURL     string              `json:"coverUrl"`
	Rating       float32             `json:"rating"`
	ReleaseDate  string              `json:"releaseDate,omitempty"`
	PageCount    int                 `json:"pageCount"`
	Status       string              `json:"status"`
	Monitored    bool                `json:"monitored"`
	Author       *AuthorResponse     `json:"author,omitempty"`
	Series       *SeriesResponse     `json:"series,omitempty"`
	SeriesIndex  *float32            `json:"seriesIndex,omitempty"`
	MediaFiles   []MediaFileResponse `json:"mediaFiles,omitempty"`
	HasEbook     bool                `json:"hasEbook"`
	HasAudiobook bool                `json:"hasAudiobook"`
	Format       string              `json:"format,omitempty"` // Primary format badge
}

// AuthorResponse represents an author in API responses
type AuthorResponse struct {
	ID              uint   `json:"id"`
	HardcoverID     string `json:"hardcoverId,omitempty"`
	OpenLibraryID   string `json:"openLibraryId,omitempty"`
	Name            string `json:"name"`
	SortName        string `json:"sortName"`
	ImageURL        string `json:"imageUrl"`
	Monitored       bool   `json:"monitored"`
	BookCount       int    `json:"bookCount,omitempty"`
	TotalBooksCount int    `json:"totalBooksCount,omitempty"`
	DownloadedCount int    `json:"downloadedCount,omitempty"`
}

// SeriesResponse represents a series in API responses
type SeriesResponse struct {
	ID              uint   `json:"id"`
	HardcoverID     string `json:"hardcoverId"`
	Name            string `json:"name"`
	BookCount       int    `json:"bookCount,omitempty"`       // Books in library
	TotalBooksCount int    `json:"totalBooksCount,omitempty"` // Total books from Hardcover (cached)
	DownloadedCount int    `json:"downloadedCount,omitempty"` // Books with files
}

// MediaFileResponse represents a media file in API responses
type MediaFileResponse struct {
	ID          uint   `json:"id"`
	FilePath    string `json:"filePath"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	Format      string `json:"format"`
	MediaType   string `json:"mediaType"`
	Bitrate     int    `json:"bitrate,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	EditionName string `json:"editionName,omitempty"`
}

// LibraryStatsResponse contains library statistics
type LibraryStatsResponse struct {
	TotalBooks      int64 `json:"totalBooks"`
	MonitoredBooks  int64 `json:"monitoredBooks"`
	DownloadedBooks int64 `json:"downloadedBooks"`
	MissingBooks    int64 `json:"missingBooks"`
	TotalAuthors    int64 `json:"totalAuthors"`
	TotalSeries     int64 `json:"totalSeries"`
	TotalEbooks     int64 `json:"totalEbooks"`
	TotalAudiobooks int64 `json:"totalAudiobooks"`
	TotalFileSize   int64 `json:"totalFileSize"`
}

// getLibrary returns the main library grid data
func (s *Server) getLibrary(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Filters
	status := c.QueryParam("status")
	authorID := c.QueryParam("authorId")
	seriesID := c.QueryParam("seriesId")
	mediaType := c.QueryParam("mediaType")
	sortBy := c.QueryParam("sortBy")
	if sortBy == "" {
		sortBy = "title"
	}
	sortOrder := c.QueryParam("sortOrder")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	var books []db.Book
	var total int64

	query := s.db.Model(&db.Book{}).Preload("Author").Preload("Series").Preload("MediaFiles")

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}
	if seriesID != "" {
		query = query.Where("series_id = ?", seriesID)
	}
	if mediaType != "" {
		query = query.Joins("JOIN media_files ON media_files.book_id = books.id").
			Where("media_files.media_type = ?", mediaType).
			Distinct()
	}

	// Get total count
	query.Count(&total)

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	query = query.Order(sortBy + " " + sortOrder).Offset(offset).Limit(pageSize)

	if err := query.Find(&books).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert to response format
	bookResponses := make([]BookResponse, len(books))
	for i, book := range books {
		bookResponses[i] = bookToResponse(book)
	}

	return c.JSON(http.StatusOK, LibraryResponse{
		Books:    bookResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// getLibraryStats returns library statistics
func (s *Server) getLibraryStats(c echo.Context) error {
	var stats LibraryStatsResponse

	s.db.Model(&db.Book{}).Count(&stats.TotalBooks)
	s.db.Model(&db.Book{}).Where("monitored = ?", true).Count(&stats.MonitoredBooks)
	s.db.Model(&db.Book{}).Where("status = ?", db.StatusDownloaded).Count(&stats.DownloadedBooks)
	s.db.Model(&db.Book{}).Where("status = ?", db.StatusMissing).Count(&stats.MissingBooks)
	s.db.Model(&db.Author{}).Count(&stats.TotalAuthors)
	s.db.Model(&db.Series{}).Count(&stats.TotalSeries)
	s.db.Model(&db.MediaFile{}).Where("media_type = ?", db.MediaTypeEbook).Count(&stats.TotalEbooks)
	s.db.Model(&db.MediaFile{}).Where("media_type = ?", db.MediaTypeAudiobook).Count(&stats.TotalAudiobooks)

	var totalSize struct{ Total int64 }
	s.db.Model(&db.MediaFile{}).Select("COALESCE(SUM(file_size), 0) as total").Scan(&totalSize)
	stats.TotalFileSize = totalSize.Total

	return c.JSON(http.StatusOK, stats)
}

// Helper function to convert Book model to BookResponse
func bookToResponse(book db.Book) BookResponse {
	resp := BookResponse{
		ID:          book.ID,
		HardcoverID: book.HardcoverID,
		Title:       book.Title,
		SortTitle:   book.SortTitle,
		ISBN:        book.ISBN,
		Description: book.Description,
		CoverURL:    book.CoverURL,
		Rating:      book.Rating,
		PageCount:   book.PageCount,
		Status:      string(book.Status),
		Monitored:   book.Monitored,
		SeriesIndex: book.SeriesIndex,
	}

	if book.ReleaseDate != nil {
		resp.ReleaseDate = book.ReleaseDate.Format("2006-01-02")
	}

	if book.Author.ID != 0 {
		resp.Author = &AuthorResponse{
			ID:          book.Author.ID,
			HardcoverID: book.Author.HardcoverID,
			Name:        book.Author.Name,
			SortName:    book.Author.SortName,
			ImageURL:    book.Author.ImageURL,
			Monitored:   book.Author.Monitored,
		}
	}

	if book.Series != nil && book.Series.ID != 0 {
		resp.Series = &SeriesResponse{
			ID:          book.Series.ID,
			HardcoverID: book.Series.HardcoverID,
			Name:        book.Series.Name,
		}
	}

	// Process media files
	for _, mf := range book.MediaFiles {
		resp.MediaFiles = append(resp.MediaFiles, MediaFileResponse{
			ID:          mf.ID,
			FilePath:    mf.FilePath,
			FileName:    mf.FileName,
			FileSize:    mf.FileSize,
			Format:      mf.Format,
			MediaType:   string(mf.MediaType),
			Bitrate:     mf.Bitrate,
			Duration:    mf.Duration,
			EditionName: mf.EditionName,
		})

		if mf.MediaType == db.MediaTypeEbook {
			resp.HasEbook = true
			if resp.Format == "" {
				resp.Format = mf.Format
			}
		} else if mf.MediaType == db.MediaTypeAudiobook {
			resp.HasAudiobook = true
		}
	}

	return resp
}
