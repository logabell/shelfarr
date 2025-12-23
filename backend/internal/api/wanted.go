package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// WantedBook represents a book that is wanted (missing or needs upgrade)
type WantedBook struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	AuthorID    uint    `json:"authorId"`
	AuthorName  string  `json:"authorName"`
	CoverURL    string  `json:"coverUrl"`
	Status      string  `json:"status"`
	SeriesName  string  `json:"seriesName,omitempty"`
	SeriesIndex float32 `json:"seriesIndex,omitempty"`
	Monitored   bool    `json:"monitored"`
	HasEbook    bool    `json:"hasEbook"`
	HasAudiobook bool   `json:"hasAudiobook"`
}

// getWanted returns all wanted books (missing + cutoff unmet)
func (s *Server) getWanted(c echo.Context) error {
	page := 1
	pageSize := 50

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	// Get books that are monitored and missing
	var books []db.Book
	var total int64

	query := s.db.Model(&db.Book{}).
		Where("monitored = ? AND status = ?", true, db.StatusMissing)
	
	query.Count(&total)
	query.Preload("Author").Preload("Series").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&books)

	wanted := make([]WantedBook, 0, len(books))
	for _, book := range books {
		w := WantedBook{
			ID:         book.ID,
			Title:      book.Title,
			AuthorID:   book.AuthorID,
			CoverURL:   book.CoverURL,
			Status:     string(book.Status),
			Monitored:  book.Monitored,
		}

		if book.Author.ID != 0 {
			w.AuthorName = book.Author.Name
		}
		if book.Series != nil {
			w.SeriesName = book.Series.Name
			if book.SeriesIndex != nil {
				w.SeriesIndex = *book.SeriesIndex
			}
		}

		// Check for existing media
		var ebookCount, audioCount int64
		s.db.Model(&db.MediaFile{}).Where("book_id = ? AND media_type = ?", book.ID, "ebook").Count(&ebookCount)
		s.db.Model(&db.MediaFile{}).Where("book_id = ? AND media_type = ?", book.ID, "audiobook").Count(&audioCount)
		w.HasEbook = ebookCount > 0
		w.HasAudiobook = audioCount > 0

		wanted = append(wanted, w)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"books":    wanted,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// getWantedMissing returns books completely missing media files
func (s *Server) getWantedMissing(c echo.Context) error {
	page := 1
	pageSize := 50
	mediaType := c.QueryParam("mediaType") // ebook, audiobook, or empty for both

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	// Subquery to find books with no media files of the specified type
	var books []db.Book
	var total int64

	query := s.db.Model(&db.Book{}).Where("monitored = ?", true)

	if mediaType != "" {
		// Books missing this specific media type
		query = query.Where("id NOT IN (SELECT DISTINCT book_id FROM media_files WHERE media_type = ? AND deleted_at IS NULL)", mediaType)
	} else {
		// Books missing all media types
		query = query.Where("id NOT IN (SELECT DISTINCT book_id FROM media_files WHERE deleted_at IS NULL)")
	}

	query.Count(&total)
	query.Preload("Author").Preload("Series").
		Offset(offset).Limit(pageSize).
		Order("title ASC").
		Find(&books)

	wanted := make([]WantedBook, 0, len(books))
	for _, book := range books {
		w := WantedBook{
			ID:        book.ID,
			Title:     book.Title,
			AuthorID:  book.AuthorID,
			CoverURL:  book.CoverURL,
			Status:    string(book.Status),
			Monitored: book.Monitored,
		}

		if book.Author.ID != 0 {
			w.AuthorName = book.Author.Name
		}
		if book.Series != nil {
			w.SeriesName = book.Series.Name
			if book.SeriesIndex != nil {
				w.SeriesIndex = *book.SeriesIndex
			}
		}

		wanted = append(wanted, w)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"books":    wanted,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// getWantedCutoff returns books that don't meet quality cutoff
func (s *Server) getWantedCutoff(c echo.Context) error {
	page := 1
	pageSize := 50

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// For cutoff unmet, we need to check quality profiles
	// This is a simplified implementation - in production, you'd compare
	// against actual quality profile rankings

	// Get books with media files that might need upgrade
	var books []db.Book
	var total int64

	// Books that have media but might need better quality
	query := s.db.Model(&db.Book{}).
		Where("monitored = ?", true).
		Where("id IN (SELECT DISTINCT book_id FROM media_files WHERE deleted_at IS NULL)")

	query.Count(&total)
	query.Preload("Author").Preload("Series").Preload("MediaFiles").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("title ASC").
		Find(&books)

	// Filter to books that don't meet cutoff (simplified logic)
	wanted := make([]WantedBook, 0)
	for _, book := range books {
		// Check if any media file is below cutoff
		needsUpgrade := false
		for _, mf := range book.MediaFiles {
			// Simplified: if format is PDF for ebook, needs upgrade
			if mf.MediaType == db.MediaTypeEbook && mf.Format == "pdf" {
				needsUpgrade = true
				break
			}
			// Simplified: if bitrate < 64kbps for audiobook, needs upgrade
			if mf.MediaType == db.MediaTypeAudiobook && mf.Bitrate < 64 {
				needsUpgrade = true
				break
			}
		}

		if needsUpgrade {
			w := WantedBook{
				ID:        book.ID,
				Title:     book.Title,
				AuthorID:  book.AuthorID,
				CoverURL:  book.CoverURL,
				Status:    "cutoff",
				Monitored: book.Monitored,
			}

			if book.Author.ID != 0 {
				w.AuthorName = book.Author.Name
			}
			if book.Series != nil {
				w.SeriesName = book.Series.Name
				if book.SeriesIndex != nil {
					w.SeriesIndex = *book.SeriesIndex
				}
			}

			wanted = append(wanted, w)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"books":    wanted,
		"total":    len(wanted),
		"page":     page,
		"pageSize": pageSize,
	})
}

