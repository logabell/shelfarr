package metadata

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/shelfarr/shelfarr/internal/db"
	"github.com/shelfarr/shelfarr/internal/googlebooks"
	"github.com/shelfarr/shelfarr/internal/hardcover"
	"github.com/shelfarr/shelfarr/internal/openlibrary"

	"gorm.io/gorm"
)

const (
	EbookCacheDuration     = 24 * time.Hour
	AudiobookCacheDuration = 24 * time.Hour
	MetadataCacheDuration  = 7 * 24 * time.Hour
)

type Aggregator struct {
	db          *gorm.DB
	openLibrary *openlibrary.Client
	googleBooks *googlebooks.Client
	hardcover   *hardcover.Client
}

type AggregatedBook struct {
	Title                string
	SortTitle            string
	Description          string
	ISBN                 string
	ISBN13               string
	CoverURL             string
	Rating               float32
	ReleaseDate          *time.Time
	PageCount            int
	Language             string
	Genres               []string
	OpenLibraryWorkID    string
	OpenLibraryEditionID string
	GoogleVolumeID       string
	HardcoverID          string
	AuthorName           string
	AuthorID             string
	SeriesName           string
	SeriesID             string
	SeriesIndex          *float32
	IsEbook              *bool
	IsAudiobook          *bool
	HasEpub              *bool
	HasPdf               *bool
	BuyLink              string
	AudioDuration        int
}

type AggregatedAuthor struct {
	Name          string
	SortName      string
	Biography     string
	ImageURL      string
	BirthDate     *time.Time
	DeathDate     *time.Time
	OpenLibraryID string
	HardcoverID   string
}

func NewAggregator(database *gorm.DB, olClient *openlibrary.Client, gbClient *googlebooks.Client, hcClient *hardcover.Client) *Aggregator {
	return &Aggregator{
		db:          database,
		openLibrary: olClient,
		googleBooks: gbClient,
		hardcover:   hcClient,
	}
}

func (a *Aggregator) SearchBooks(query string, limit int) ([]AggregatedBook, error) {
	if a.openLibrary == nil {
		return nil, errors.New("open library client not configured")
	}

	result, err := a.openLibrary.SearchBooks(query, limit, 0)
	if err != nil {
		return nil, err
	}

	books := make([]AggregatedBook, 0, len(result.Docs))
	for _, doc := range result.Docs {
		book := AggregatedBook{
			Title:             doc.Title,
			Description:       "",
			Rating:            float32(doc.RatingsAverage),
			PageCount:         doc.NumberOfPagesMedian,
			OpenLibraryWorkID: openlibrary.ExtractOLID(doc.Key),
		}

		if doc.FirstPublishYear > 0 {
			t := time.Date(doc.FirstPublishYear, 1, 1, 0, 0, 0, 0, time.UTC)
			book.ReleaseDate = &t
		}

		if len(doc.ISBN) > 0 {
			for _, isbn := range doc.ISBN {
				if len(isbn) == 10 && book.ISBN == "" {
					book.ISBN = isbn
				} else if len(isbn) == 13 && book.ISBN13 == "" {
					book.ISBN13 = isbn
				}
				if book.ISBN != "" && book.ISBN13 != "" {
					break
				}
			}
		}

		if doc.CoverI > 0 {
			book.CoverURL = openlibrary.GetCoverURL(doc.CoverI, "L")
		}

		if len(doc.AuthorName) > 0 {
			book.AuthorName = doc.AuthorName[0]
		}
		if len(doc.AuthorKey) > 0 {
			book.AuthorID = doc.AuthorKey[0]
		}

		if len(doc.Language) > 0 {
			book.Language = doc.Language[0]
		}

		book.Genres = doc.Subject

		books = append(books, book)
	}

	return books, nil
}

func (a *Aggregator) GetBookByISBN(isbn string) (*AggregatedBook, error) {
	if a.openLibrary == nil {
		return nil, errors.New("open library client not configured")
	}

	olBook, err := a.openLibrary.GetBookByISBN(isbn)
	if err != nil {
		return nil, err
	}

	book := &AggregatedBook{
		Title:       olBook.Title,
		Description: "",
		PageCount:   olBook.NumberOfPages,
	}

	if olBook.Cover != nil {
		if olBook.Cover.Large != "" {
			book.CoverURL = olBook.Cover.Large
		} else if olBook.Cover.Medium != "" {
			book.CoverURL = olBook.Cover.Medium
		}
	}

	if len(olBook.Authors) > 0 {
		book.AuthorName = olBook.Authors[0].Name
	}

	if len(olBook.Publishers) > 0 {
		// Publishers available but not stored in AggregatedBook
	}

	if olBook.PublishDate != "" {
		if t, err := time.Parse("2006", olBook.PublishDate); err == nil {
			book.ReleaseDate = &t
		} else if t, err := time.Parse("January 2, 2006", olBook.PublishDate); err == nil {
			book.ReleaseDate = &t
		}
	}

	if olBook.Identifiers != nil {
		if len(olBook.Identifiers.ISBN10) > 0 && book.ISBN == "" {
			book.ISBN = olBook.Identifiers.ISBN10[0]
		}
		if len(olBook.Identifiers.ISBN13) > 0 && book.ISBN13 == "" {
			book.ISBN13 = olBook.Identifiers.ISBN13[0]
		}
	}

	if book.ISBN == "" {
		book.ISBN = isbn
	}

	if a.googleBooks != nil && a.googleBooks.HasAPIKey() {
		a.enrichWithGoogleBooks(book, isbn)
	}

	return book, nil
}

func (a *Aggregator) enrichWithGoogleBooks(book *AggregatedBook, isbn string) {
	if isbn == "" && book.ISBN13 != "" {
		isbn = book.ISBN13
	}
	if isbn == "" && book.ISBN != "" {
		isbn = book.ISBN
	}
	if isbn == "" {
		return
	}

	info, err := a.googleBooks.GetVolumeByISBN(isbn)
	if err != nil {
		if errors.Is(err, googlebooks.ErrQuotaExhausted) {
			// Quota exhausted - book will have unknown ebook status
		}
		return
	}

	book.GoogleVolumeID = info.VolumeID
	book.IsEbook = &info.IsEbook
	book.HasEpub = &info.HasEpub
	book.HasPdf = &info.HasPdf
	book.BuyLink = info.BuyLink

	if book.CoverURL == "" && info.CoverURL != "" {
		book.CoverURL = info.CoverURL
	}
	if book.Description == "" && info.Description != "" {
		book.Description = info.Description
	}
}

func (a *Aggregator) EnrichWithHardcover(book *AggregatedBook, hardcoverID string) error {
	if a.hardcover == nil {
		return errors.New("hardcover client not configured")
	}

	hcBook, err := a.hardcover.GetBook(hardcoverID)
	if err != nil {
		return err
	}

	book.HardcoverID = hcBook.ID
	book.IsAudiobook = &hcBook.HasAudiobook
	book.AudioDuration = hcBook.AudioDuration

	if hcBook.SeriesID != "" {
		book.SeriesID = hcBook.SeriesID
		book.SeriesName = hcBook.SeriesName
		book.SeriesIndex = hcBook.SeriesIndex
	}

	if book.CoverURL == "" && hcBook.CoverURL != "" {
		book.CoverURL = hcBook.CoverURL
	}
	if book.Description == "" && hcBook.Description != "" {
		book.Description = hcBook.Description
	}
	if book.Rating == 0 && hcBook.Rating > 0 {
		book.Rating = hcBook.Rating
	}

	return nil
}

func (a *Aggregator) CheckEbookStatus(book *db.Book) error {
	if book.EbookCheckedAt != nil && time.Since(*book.EbookCheckedAt) < EbookCacheDuration {
		return nil
	}

	isbn := book.ISBN13
	if isbn == "" {
		isbn = book.ISBN
	}
	if isbn == "" {
		return errors.New("no ISBN available")
	}

	if a.googleBooks == nil || !a.googleBooks.HasAPIKey() {
		return errors.New("google books client not configured")
	}

	isEbook, hasEpub, hasPdf, buyLink, err := a.googleBooks.CheckEbookStatus(isbn)
	if err != nil {
		if errors.Is(err, googlebooks.ErrQuotaExhausted) {
			return err
		}
		return err
	}

	now := time.Now()
	book.IsEbook = &isEbook
	book.HasEpub = &hasEpub
	book.HasPdf = &hasPdf
	book.BuyLink = buyLink
	book.EbookCheckedAt = &now

	return a.db.Save(book).Error
}

func (a *Aggregator) CheckAudiobookStatus(book *db.Book) error {
	if book.AudiobookCheckedAt != nil && time.Since(*book.AudiobookCheckedAt) < AudiobookCacheDuration {
		return nil
	}

	if book.HardcoverID == "" {
		return errors.New("no hardcover ID available")
	}

	if a.hardcover == nil {
		return errors.New("hardcover client not configured")
	}

	hcBook, err := a.hardcover.GetBook(book.HardcoverID)
	if err != nil {
		return err
	}

	now := time.Now()
	book.IsAudiobook = &hcBook.HasAudiobook
	book.AudioDuration = hcBook.AudioDuration
	book.AudiobookCheckedAt = &now

	return a.db.Save(book).Error
}

func (a *Aggregator) RefreshBookMetadata(book *db.Book) error {
	if book.MetadataUpdatedAt != nil && time.Since(*book.MetadataUpdatedAt) < MetadataCacheDuration {
		return nil
	}

	isbn := book.ISBN13
	if isbn == "" {
		isbn = book.ISBN
	}

	if isbn != "" && a.openLibrary != nil {
		olBook, err := a.openLibrary.GetBookByISBN(isbn)
		if err == nil {
			if book.Description == "" {
				book.Description = ""
			}
			if book.PageCount == 0 {
				book.PageCount = olBook.NumberOfPages
			}
			if olBook.Cover != nil && book.CoverURL == "" {
				if olBook.Cover.Large != "" {
					book.CoverURL = olBook.Cover.Large
				} else if olBook.Cover.Medium != "" {
					book.CoverURL = olBook.Cover.Medium
				}
			}
		}
	}

	now := time.Now()
	book.MetadataUpdatedAt = &now

	return a.db.Save(book).Error
}

func (a *Aggregator) GetAuthor(openLibraryID string) (*AggregatedAuthor, error) {
	if a.openLibrary == nil {
		return nil, errors.New("open library client not configured")
	}

	olAuthor, err := a.openLibrary.GetAuthor(openLibraryID)
	if err != nil {
		return nil, err
	}

	author := &AggregatedAuthor{
		Name:          olAuthor.Name,
		OpenLibraryID: openlibrary.ExtractOLID(olAuthor.Key),
	}

	if olAuthor.Bio != nil {
		author.Biography = openlibrary.ExtractDescription(olAuthor.Bio)
	}

	if len(olAuthor.Photos) > 0 {
		author.ImageURL = openlibrary.GetAuthorPhotoURL(olAuthor.Photos[0], "L")
	}

	if olAuthor.BirthDate != "" {
		if t, err := time.Parse("2 January 2006", olAuthor.BirthDate); err == nil {
			author.BirthDate = &t
		} else if t, err := time.Parse("January 2, 2006", olAuthor.BirthDate); err == nil {
			author.BirthDate = &t
		} else if t, err := time.Parse("2006", olAuthor.BirthDate); err == nil {
			author.BirthDate = &t
		}
	}

	if olAuthor.DeathDate != "" {
		if t, err := time.Parse("2 January 2006", olAuthor.DeathDate); err == nil {
			author.DeathDate = &t
		} else if t, err := time.Parse("January 2, 2006", olAuthor.DeathDate); err == nil {
			author.DeathDate = &t
		} else if t, err := time.Parse("2006", olAuthor.DeathDate); err == nil {
			author.DeathDate = &t
		}
	}

	return author, nil
}

func (a *Aggregator) QuotaStatus() (remaining int, exhausted bool) {
	if a.googleBooks == nil {
		return 0, true
	}
	remaining = a.googleBooks.QuotaRemaining()
	return remaining, remaining <= 0
}

func ToDBBook(agg *AggregatedBook) *db.Book {
	book := &db.Book{
		HardcoverID:          agg.HardcoverID,
		OpenLibraryWorkID:    agg.OpenLibraryWorkID,
		OpenLibraryEditionID: agg.OpenLibraryEditionID,
		GoogleVolumeID:       agg.GoogleVolumeID,
		Title:                agg.Title,
		SortTitle:            agg.SortTitle,
		ISBN:                 agg.ISBN,
		ISBN13:               agg.ISBN13,
		Description:          agg.Description,
		CoverURL:             agg.CoverURL,
		Rating:               agg.Rating,
		ReleaseDate:          agg.ReleaseDate,
		PageCount:            agg.PageCount,
		Language:             agg.Language,
		SeriesIndex:          agg.SeriesIndex,
		IsEbook:              agg.IsEbook,
		IsAudiobook:          agg.IsAudiobook,
		HasEpub:              agg.HasEpub,
		HasPdf:               agg.HasPdf,
		BuyLink:              agg.BuyLink,
		AudioDuration:        agg.AudioDuration,
	}

	if len(agg.Genres) > 0 {
		genresJSON, _ := json.Marshal(agg.Genres)
		book.Genres = string(genresJSON)
	}

	return book
}

func ToDBAuthor(agg *AggregatedAuthor) *db.Author {
	return &db.Author{
		HardcoverID:   agg.HardcoverID,
		OpenLibraryID: agg.OpenLibraryID,
		Name:          agg.Name,
		SortName:      agg.SortName,
		Biography:     agg.Biography,
		ImageURL:      agg.ImageURL,
		BirthDate:     agg.BirthDate,
		DeathDate:     agg.DeathDate,
	}
}
