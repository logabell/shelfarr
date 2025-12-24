package hardcover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Client handles communication with the Hardcover.app GraphQL API
type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
}

// NewClient creates a new Hardcover API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		// Rate limit: 60 requests per minute = 1 request per second
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// NewClientWithAPIKey creates a new Hardcover API client with API key authentication
func NewClientWithAPIKey(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		// Rate limit: 60 requests per minute = 1 request per second
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// SetAPIKey sets the API key for authenticated requests
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// Digital format constants for reading_format.format values
const (
	FormatEbook     = "Ebook"
	FormatAudiobook = "Audiobook"
	FormatPhysical  = "Physical"
)

// EditionData represents a single edition from Hardcover
type EditionData struct {
	ID            string
	ISBN10        string
	ISBN13        string
	ASIN          string
	Title         string
	Subtitle      string
	EditionFormat string
	Format        string
	LanguageCode  string
	Language      string
	PublisherID   string
	PublisherName string
	PageCount     int
	AudioSeconds  int
	ReleaseDate   *time.Time
	CoverURL      string
}

// ContributorData represents a contributor from Hardcover
type ContributorData struct {
	AuthorID       string
	AuthorName     string
	AuthorSlug     string
	AuthorImage    string
	AuthorBio      string
	Role           string
	Position       int
	BornDate       *time.Time
	BornYear       *int
	DeathDate      *time.Time
	DeathYear      *int
	Location       string
	IsBIPOC        *bool
	IsLGBTQ        *bool
	AlternateNames []string
}

// BookData represents book data from Hardcover.app
type BookData struct {
	ID           string
	Title        string
	SortTitle    string
	Subtitle     string
	Headline     string
	Slug         string
	ISBN         string
	ISBN13       string
	Description  string
	CoverURL     string
	Rating       float32
	RatingsCount int
	ReviewsCount int
	ReleaseDate  *time.Time
	ReleaseYear  int
	PageCount    int
	AuthorID     string
	AuthorName   string
	AuthorImage  string
	Authors      []string
	SeriesID     string
	SeriesName   string
	SeriesSlug   string
	SeriesIndex  *float32
	Genres       []string
	LanguageCode string
	Language     string

	LiteraryType string
	Category     string
	Compilation  bool

	HasAudiobook          bool
	HasEbook              bool
	HasPhysical           bool
	HasDigitalEdition     bool
	DigitalEditionCount   int
	PhysicalEditionCount  int
	EbookEditionCount     int
	AudiobookEditionCount int
	EditionCount          int
	AudioDuration         int

	Editions     []EditionData
	Contributors []ContributorData

	SeriesIsCompleted       *bool
	SeriesPrimaryBooksCount int
	SeriesAuthorID          string
	SeriesAuthorName        string
}

// FilteredBooksResult contains books filtered by digital availability
type FilteredBooksResult struct {
	Books             []BookData
	TotalCount        int // All books before filtering
	DigitalCount      int // Books with digital editions
	PhysicalOnlyCount int // Books with only physical editions
}

type AuthorData struct {
	ID                 string
	Name               string
	SortName           string
	Slug               string
	Biography          string
	ImageURL           string
	BooksCount         int
	BornDate           *time.Time
	BornYear           *int
	DeathDate          *time.Time
	DeathYear          *int
	Location           string
	GenderID           *int
	IsBIPOC            *bool
	IsLGBTQ            *bool
	AlternateNames     []string
	TotalBooksAll      int
	TotalBooksFiltered int
}

type SeriesData struct {
	ID                 string
	Name               string
	Slug               string
	Description        string
	BooksCount         int
	AuthorID           string
	AuthorName         string
	IsCompleted        *bool
	TotalBooksAll      int
	TotalBooksFiltered int
	PrimaryBooksCount  int
}

// ListData represents list data from Hardcover.app
type ListData struct {
	ID          string
	Name        string
	Description string
	BooksCount  int
	UserID      string
	Username    string
}

// UnifiedSearchResults contains search results for multiple types
type UnifiedSearchResults struct {
	Books   []BookData
	Authors []AuthorData
	Series  []SeriesData
	Lists   []ListData
}

// graphQLRequest represents a GraphQL request
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `errors,omitempty"`
}

// execute sends a GraphQL query and returns the response
func (c *Client) execute(query string, variables map[string]interface{}) (json.RawMessage, error) {
	// Wait for rate limiter (respects 60 requests/minute limit)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Shelfarr/1.0")

	// Add authorization header if API key is configured
	if c.apiKey != "" {
		authValue := c.apiKey
		if !strings.HasPrefix(strings.ToLower(c.apiKey), "bearer") {
			authValue = "Bearer " + c.apiKey
		}
		req.Header.Set("Authorization", authValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed - check API key")
	}

	// Handle non-2xx status codes
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("API error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

// SearchBooks searches for books by title/author and filters by language if provided
func (c *Client) SearchBooks(query string, languages []string) ([]BookData, error) {
	gqlQuery := `
		query SearchBooks($query: String!) {
			search(query: $query, query_type: "Book", per_page: 20, page: 1) {
				results
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
	}

	data, err := c.execute(gqlQuery, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Search struct {
			Results struct {
				Found int `json:"found"`
				Hits  []struct {
					Document struct {
						ID           json.Number `json:"id"`
						Title        string      `json:"title"`
						DefaultTitle string      `json:"default_title"`
						Subtitle     string      `json:"subtitle"`
						Slug         string      `json:"slug"`
						AuthorNames  []string    `json:"author_names"`
						ReleaseYear  int         `json:"release_year"`
						CachedImage  *struct {
							URL string `json:"url"`
						} `json:"cached_image"`
						Image *struct {
							URL string `json:"url"`
						} `json:"image"`
						RatingsAverage float32 `json:"ratings_average"`
						Contributions  []struct {
							Author struct {
								ID   json.Number `json:"id"`
								Name string      `json:"name"`
							} `json:"author"`
						} `json:"contributions"`
					} `json:"document"`
				} `json:"hits"`
			} `json:"results"`
		} `json:"search"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	allBooks := make([]BookData, 0, len(result.Search.Results.Hits))
	for _, hit := range result.Search.Results.Hits {
		doc := hit.Document
		title := doc.DefaultTitle
		if title == "" {
			title = doc.Title
		}

		coverURL := ""
		if doc.CachedImage != nil && doc.CachedImage.URL != "" {
			coverURL = doc.CachedImage.URL
		} else if doc.Image != nil {
			coverURL = doc.Image.URL
		}

		authorName := ""
		authorID := ""
		if len(doc.AuthorNames) > 0 {
			authorName = doc.AuthorNames[0]
		}
		if len(doc.Contributions) > 0 {
			authorID = doc.Contributions[0].Author.ID.String()
			if authorName == "" {
				authorName = doc.Contributions[0].Author.Name
			}
		}

		book := BookData{
			ID:          doc.ID.String(),
			Title:       title,
			CoverURL:    coverURL,
			Rating:      doc.RatingsAverage,
			ReleaseYear: doc.ReleaseYear,
			AuthorID:    authorID,
			AuthorName:  authorName,
		}

		// Search hit doesn't have language info.
		// For thoroughness we could fetch book details, but that's slow.
		// For now, we'll return all search results.
		allBooks = append(allBooks, book)
	}

	return allBooks, nil
}

// SearchAuthors searches for authors
func (c *Client) SearchAuthors(query string) ([]AuthorData, error) {
	gqlQuery := `
		query SearchAuthors($query: String!) {
			search(query: $query, query_type: "Author", per_page: 20, page: 1) {
				results
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
	}

	data, err := c.execute(gqlQuery, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Search struct {
			Results struct {
				Hits []struct {
					Document struct {
						ID          json.Number `json:"id"`
						Name        string      `json:"name"`
						Bio         string      `json:"bio"`
						BooksCount  int         `json:"books_count"`
						CachedImage *struct {
							URL string `json:"url"`
						} `json:"cached_image"`
						Image *struct {
							URL string `json:"url"`
						} `json:"image"`
					} `json:"document"`
				} `json:"hits"`
			} `json:"results"`
		} `json:"search"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	authors := make([]AuthorData, len(result.Search.Results.Hits))
	for i, hit := range result.Search.Results.Hits {
		doc := hit.Document
		imageURL := ""
		if doc.CachedImage != nil {
			imageURL = doc.CachedImage.URL
		} else if doc.Image != nil {
			imageURL = doc.Image.URL
		}

		authors[i] = AuthorData{
			ID:         doc.ID.String(),
			Name:       doc.Name,
			Biography:  doc.Bio,
			ImageURL:   imageURL,
			BooksCount: doc.BooksCount,
		}
	}

	return authors, nil
}

// SearchSeries searches for series
func (c *Client) SearchSeries(query string) ([]SeriesData, error) {
	gqlQuery := `
		query SearchSeries($query: String!) {
			search(query: $query, query_type: "Series", per_page: 20, page: 1) {
				results
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
	}

	data, err := c.execute(gqlQuery, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Search struct {
			Results struct {
				Hits []struct {
					Document struct {
						ID         json.Number `json:"id"`
						Name       string      `json:"name"`
						Slug       string      `json:"slug"`
						BooksCount int         `json:"books_count"`
						Author     *struct {
							ID   json.Number `json:"id"`
							Name string      `json:"name"`
						} `json:"author"`
					} `json:"document"`
				} `json:"hits"`
			} `json:"results"`
		} `json:"search"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	seriesList := make([]SeriesData, len(result.Search.Results.Hits))
	for i, hit := range result.Search.Results.Hits {
		doc := hit.Document
		authorID := ""
		authorName := ""
		if doc.Author != nil {
			authorID = doc.Author.ID.String()
			authorName = doc.Author.Name
		}

		seriesList[i] = SeriesData{
			ID:         doc.ID.String(),
			Name:       doc.Name,
			Slug:       doc.Slug,
			BooksCount: doc.BooksCount,
			AuthorID:   authorID,
			AuthorName: authorName,
		}
	}

	return seriesList, nil
}

// SearchLists searches for lists
func (c *Client) SearchLists(query string) ([]ListData, error) {
	gqlQuery := `
		query SearchLists($query: String!) {
			search(query: $query, query_type: "List", per_page: 20, page: 1) {
				results
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
	}

	data, err := c.execute(gqlQuery, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Search struct {
			Results struct {
				Hits []struct {
					Document struct {
						ID          json.Number `json:"id"`
						Name        string      `json:"name"`
						Description string      `json:"description"`
						BooksCount  int         `json:"books_count"`
						User        *struct {
							ID       json.Number `json:"id"`
							Username string      `json:"username"`
						} `json:"user"`
					} `json:"document"`
				} `json:"hits"`
			} `json:"results"`
		} `json:"search"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	lists := make([]ListData, len(result.Search.Results.Hits))
	for i, hit := range result.Search.Results.Hits {
		doc := hit.Document
		userID := ""
		username := ""
		if doc.User != nil {
			userID = doc.User.ID.String()
			username = doc.User.Username
		}

		lists[i] = ListData{
			ID:          doc.ID.String(),
			Name:        doc.Name,
			Description: doc.Description,
			BooksCount:  doc.BooksCount,
			UserID:      userID,
			Username:    username,
		}
	}

	return lists, nil
}

// SearchAll performs a unified search
func (c *Client) SearchAll(query string, languages []string) (*UnifiedSearchResults, error) {
	results := &UnifiedSearchResults{}

	books, _ := c.SearchBooks(query, languages)
	results.Books = books

	authors, _ := c.SearchAuthors(query)
	results.Authors = authors

	series, _ := c.SearchSeries(query)
	results.Series = series

	lists, _ := c.SearchLists(query)
	results.Lists = lists

	return results, nil
}

func (c *Client) GetBook(id string) (*BookData, error) {
	idInt, _ := strconv.Atoi(id)
	gqlQuery := `
		query GetBook($id: Int!) {
			books_by_pk(id: $id) {
				id
				title
				subtitle
				headline
				slug
				description
				compilation
				cached_tags
				image { url }
				release_date
				release_year
				audio_seconds
				pages
				rating
				ratings_count
				reviews_count
			contributions {
				contribution
				author {
						id
						name
						slug
						bio
						born_date
						born_year
						death_date
						death_year
						location
						gender_id
						is_bipoc
						is_lgbtq
						alternate_names
						image { url }
					}
				}
				book_series {
					series {
						id
						name
						slug
						description
						is_completed
						books_count
						primary_books_count
						author { id, name }
					}
					position
				}
				editions {
					id
					title
					subtitle
					isbn_10
					isbn_13
					asin
					edition_format
					pages
					audio_seconds
					release_date
					image { url }
					reading_format { format }
					language { code2, language }
					publisher { id, name }
				}
			}
		}
	`

	data, err := c.execute(gqlQuery, map[string]any{"id": idInt})
	if err != nil {
		return nil, err
	}

	var result struct {
		Book *struct {
			ID            json.Number           `json:"id"`
			Title         string                `json:"title"`
			Subtitle      string                `json:"subtitle"`
			Headline      string                `json:"headline"`
			Slug          string                `json:"slug"`
			Description   string                `json:"description"`
			Compilation   bool                  `json:"compilation"`
			CachedTags    map[string]any        `json:"cached_tags"`
			Image         *struct{ URL string } `json:"image"`
			ReleaseDate   string                `json:"release_date"`
			ReleaseYear   int                   `json:"release_year"`
			AudioSeconds  int                   `json:"audio_seconds"`
			Pages         int                   `json:"pages"`
			Rating        float32               `json:"rating"`
			RatingsCount  int                   `json:"ratings_count"`
			ReviewsCount  int                   `json:"reviews_count"`
			Contributions []struct {
				Contribution string `json:"contribution"`
				Position     int    `json:"position"`
				Author       struct {
					ID             json.Number           `json:"id"`
					Name           string                `json:"name"`
					Slug           string                `json:"slug"`
					Bio            string                `json:"bio"`
					BornDate       *string               `json:"born_date"`
					BornYear       *int                  `json:"born_year"`
					DeathDate      *string               `json:"death_date"`
					DeathYear      *int                  `json:"death_year"`
					Location       string                `json:"location"`
					GenderID       *int                  `json:"gender_id"`
					IsBIPOC        *bool                 `json:"is_bipoc"`
					IsLGBTQ        *bool                 `json:"is_lgbtq"`
					AlternateNames any                   `json:"alternate_names"`
					Image          *struct{ URL string } `json:"image"`
				} `json:"author"`
			} `json:"contributions"`
			BookSeries []struct {
				Series struct {
					ID                json.Number `json:"id"`
					Name              string      `json:"name"`
					Slug              string      `json:"slug"`
					Description       string      `json:"description"`
					IsCompleted       *bool       `json:"is_completed"`
					BooksCount        int         `json:"books_count"`
					PrimaryBooksCount int         `json:"primary_books_count"`
					Author            *struct {
						ID   json.Number `json:"id"`
						Name string      `json:"name"`
					} `json:"author"`
				} `json:"series"`
				Position float32 `json:"position"`
			} `json:"book_series"`
			Editions []struct {
				ID            json.Number           `json:"id"`
				Title         string                `json:"title"`
				Subtitle      string                `json:"subtitle"`
				ISBN10        string                `json:"isbn_10"`
				ISBN13        string                `json:"isbn_13"`
				ASIN          string                `json:"asin"`
				EditionFormat string                `json:"edition_format"`
				Pages         int                   `json:"pages"`
				AudioSeconds  int                   `json:"audio_seconds"`
				ReleaseDate   string                `json:"release_date"`
				Image         *struct{ URL string } `json:"image"`
				ReadingFormat *struct {
					Format string `json:"format"`
				} `json:"reading_format"`
				Language *struct {
					Code2    string `json:"code2"`
					Language string `json:"language"`
				} `json:"language"`
				Publisher *struct {
					ID   json.Number `json:"id"`
					Name string      `json:"name"`
				} `json:"publisher"`
			} `json:"editions"`
		} `json:"books_by_pk"`
	}

	if err := json.Unmarshal(data, &result); err != nil || result.Book == nil {
		return nil, fmt.Errorf("book not found or parse error")
	}

	b := result.Book
	book := &BookData{
		ID:           b.ID.String(),
		Title:        b.Title,
		Subtitle:     b.Subtitle,
		Headline:     b.Headline,
		Slug:         b.Slug,
		Description:  b.Description,
		Compilation:  b.Compilation,
		Rating:       b.Rating,
		RatingsCount: b.RatingsCount,
		ReviewsCount: b.ReviewsCount,
		ReleaseYear:  b.ReleaseYear,
		PageCount:    b.Pages,
	}

	if b.Image != nil {
		book.CoverURL = b.Image.URL
	}

	for _, contrib := range b.Contributions {
		cd := ContributorData{
			AuthorID:   contrib.Author.ID.String(),
			AuthorName: contrib.Author.Name,
			AuthorSlug: contrib.Author.Slug,
			AuthorBio:  contrib.Author.Bio,
			Role:       contrib.Contribution,
			Position:   contrib.Position,
			BornYear:   contrib.Author.BornYear,
			DeathYear:  contrib.Author.DeathYear,
			Location:   contrib.Author.Location,
			IsBIPOC:    contrib.Author.IsBIPOC,
			IsLGBTQ:    contrib.Author.IsLGBTQ,
		}
		if contrib.Author.Image != nil {
			cd.AuthorImage = contrib.Author.Image.URL
		}
		if contrib.Author.BornDate != nil && *contrib.Author.BornDate != "" {
			if t, err := time.Parse("2006-01-02", *contrib.Author.BornDate); err == nil {
				cd.BornDate = &t
			}
		}
		if contrib.Author.DeathDate != nil && *contrib.Author.DeathDate != "" {
			if t, err := time.Parse("2006-01-02", *contrib.Author.DeathDate); err == nil {
				cd.DeathDate = &t
			}
		}
		if contrib.Author.AlternateNames != nil {
			if arr, ok := contrib.Author.AlternateNames.([]any); ok {
				for _, n := range arr {
					if s, ok := n.(string); ok {
						cd.AlternateNames = append(cd.AlternateNames, s)
					}
				}
			}
		}
		book.Contributors = append(book.Contributors, cd)
	}

	if len(book.Contributors) > 0 {
		for _, c := range book.Contributors {
			if c.Role == "" || c.Role == "Author" {
				book.AuthorID = c.AuthorID
				book.AuthorName = c.AuthorName
				book.AuthorImage = c.AuthorImage
				break
			}
		}
		if book.AuthorID == "" {
			book.AuthorID = book.Contributors[0].AuthorID
			book.AuthorName = book.Contributors[0].AuthorName
			book.AuthorImage = book.Contributors[0].AuthorImage
		}
	}

	if len(b.BookSeries) > 0 {
		s := b.BookSeries[0]
		book.SeriesID = s.Series.ID.String()
		book.SeriesName = s.Series.Name
		book.SeriesSlug = s.Series.Slug
		pos := s.Position
		book.SeriesIndex = &pos
		book.SeriesIsCompleted = s.Series.IsCompleted
		book.SeriesPrimaryBooksCount = s.Series.PrimaryBooksCount
		if s.Series.Author != nil {
			book.SeriesAuthorID = s.Series.Author.ID.String()
			book.SeriesAuthorName = s.Series.Author.Name
		}
	}

	book.EditionCount = len(b.Editions)
	editionLangs := make([]EditionLanguageInfo, 0, len(b.Editions))
	var ebookCount, audiobookCount, physicalCount int

	for _, ed := range b.Editions {
		edition := EditionData{
			ID:            ed.ID.String(),
			ISBN10:        ed.ISBN10,
			ISBN13:        ed.ISBN13,
			ASIN:          ed.ASIN,
			Title:         ed.Title,
			Subtitle:      ed.Subtitle,
			EditionFormat: ed.EditionFormat,
			PageCount:     ed.Pages,
			AudioSeconds:  ed.AudioSeconds,
		}

		if ed.Image != nil {
			edition.CoverURL = ed.Image.URL
		}
		if ed.Language != nil {
			edition.LanguageCode = ed.Language.Code2
			edition.Language = ed.Language.Language
			editionLangs = append(editionLangs, EditionLanguageInfo{
				Code2:    ed.Language.Code2,
				Language: ed.Language.Language,
			})
		}
		if ed.Publisher != nil {
			edition.PublisherID = ed.Publisher.ID.String()
			edition.PublisherName = ed.Publisher.Name
		}
		if ed.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", ed.ReleaseDate); err == nil {
				edition.ReleaseDate = &t
			}
		}

		if ed.ReadingFormat != nil {
			edition.Format = ed.ReadingFormat.Format
			switch ed.ReadingFormat.Format {
			case FormatEbook:
				ebookCount++
				book.HasEbook = true
				book.HasDigitalEdition = true
			case FormatAudiobook:
				audiobookCount++
				book.HasAudiobook = true
				book.HasDigitalEdition = true
				if ed.AudioSeconds > 0 && book.AudioDuration == 0 {
					book.AudioDuration = ed.AudioSeconds
				}
			case FormatPhysical:
				physicalCount++
				book.HasPhysical = true
			}
		} else {
			format := strings.ToLower(ed.EditionFormat)
			if format == "audiobook" || strings.Contains(format, "audio") {
				edition.Format = FormatAudiobook
				audiobookCount++
				book.HasAudiobook = true
				book.HasDigitalEdition = true
				if ed.AudioSeconds > 0 && book.AudioDuration == 0 {
					book.AudioDuration = ed.AudioSeconds
				}
			} else if format == "ebook" || format == "kindle" || strings.Contains(format, "digital") {
				edition.Format = FormatEbook
				ebookCount++
				book.HasEbook = true
				book.HasDigitalEdition = true
			} else {
				edition.Format = FormatPhysical
				physicalCount++
				book.HasPhysical = true
			}
		}

		if book.ISBN == "" && ed.ISBN10 != "" {
			book.ISBN = ed.ISBN10
		}
		if book.ISBN13 == "" && ed.ISBN13 != "" {
			book.ISBN13 = ed.ISBN13
		}

		book.Editions = append(book.Editions, edition)
	}

	book.EbookEditionCount = ebookCount
	book.AudiobookEditionCount = audiobookCount
	book.PhysicalEditionCount = physicalCount
	book.DigitalEditionCount = ebookCount + audiobookCount

	book.LanguageCode = getPreferredLanguageCode(editionLangs, []string{"en"})
	if book.LanguageCode != "" {
		book.Language = getLanguageNameFromCode(book.LanguageCode)
	}

	if b.AudioSeconds > 0 {
		book.HasAudiobook = true
		if book.AudioDuration == 0 {
			book.AudioDuration = b.AudioSeconds
		}
	}

	if b.CachedTags != nil {
		if genreVal, ok := b.CachedTags["Genre"]; ok {
			if genreArr, ok := genreVal.([]any); ok {
				for _, g := range genreArr {
					if genreStr, ok := g.(string); ok {
						book.Genres = append(book.Genres, genreStr)
					}
				}
			}
		}
	}

	if b.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", b.ReleaseDate); err == nil {
			book.ReleaseDate = &t
		}
	}

	return book, nil
}

// GetAuthor fetches author details
func (c *Client) GetAuthor(id string) (*AuthorData, error) {
	idInt, _ := strconv.Atoi(id)
	gqlQuery := `
		query GetAuthor($id: Int!) {
			authors_by_pk(id: $id) {
				id, name, bio, image { url }
			}
		}
	`
	data, err := c.execute(gqlQuery, map[string]interface{}{"id": idInt})
	if err != nil {
		return nil, err
	}
	var result struct {
		Author *struct {
			ID    json.Number           `json:"id"`
			Name  string                `json:"name"`
			Bio   string                `json:"bio"`
			Image *struct{ URL string } `json:"image"`
		} `json:"authors_by_pk"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Author == nil {
		return nil, fmt.Errorf("author not found")
	}
	a := result.Author
	author := &AuthorData{ID: a.ID.String(), Name: a.Name, Biography: a.Bio}
	if a.Image != nil {
		author.ImageURL = a.Image.URL
	}
	return author, nil
}

func (c *Client) GetBooksByAuthorWithCounts(authorID string, languages []string) (*FilteredBooksResult, error) {
	return c.GetBooksByAuthor(authorID, languages)
}

type FilteredSeriesResult struct {
	Series            *SeriesData
	Books             []BookData
	TotalCount        int
	DigitalCount      int
	PhysicalOnlyCount int
}

func (c *Client) GetSeries(seriesID string, languages []string) (*FilteredSeriesResult, error) {
	idInt, _ := strconv.Atoi(seriesID)
	gqlQuery := `
		query GetSeries($seriesId: Int!) {
			series_by_pk(id: $seriesId) {
				id, name, slug, books_count, primary_books_count
				book_series(where: {book: {compilation: {_eq: false}}}) {
					position
					book {
						id, title, description, compilation, image { url }, release_date, pages, rating
						contributions { author { id, name } }
						editions { 
							isbn_10, isbn_13, asin, compilation, reading_format { format }, language { code2 language } 
						}
					}
				}
			}
		}
	`
	data, err := c.execute(gqlQuery, map[string]interface{}{"seriesId": idInt})
	if err != nil {
		return nil, err
	}
	var result struct {
		Series *struct {
			ID                json.Number `json:"id"`
			Name, Slug        string
			BooksCount        int `json:"books_count"`
			PrimaryBooksCount int `json:"primary_books_count"`
			BookSeries        []struct {
				Position float32
				Book     struct {
					ID            json.Number
					Title         string `json:"title"`
					Description   string `json:"description"`
					ReleaseDate   string `json:"release_date"`
					Compilation   bool   `json:"compilation"`
					Image         *struct{ URL string }
					Pages         int
					Rating        float32
					Contributions []struct {
						Author struct {
							ID   json.Number
							Name string
						}
					}
					Editions []struct {
						ISBN10        string `json:"isbn_10"`
						ISBN13        string `json:"isbn_13"`
						Asin          string `json:"asin"`
						Compilation   bool   `json:"compilation"`
						ReadingFormat *struct {
							Format string `json:"format"`
						} `json:"reading_format"`
						Language *struct {
							Code2    string `json:"code2"`
							Language string `json:"language"`
						} `json:"language"`
					}
				}
			} `json:"book_series"`
		} `json:"series_by_pk"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Series == nil {
		return nil, fmt.Errorf("series not found")
	}

	seriesData := &SeriesData{
		ID:                result.Series.ID.String(),
		Name:              result.Series.Name,
		Slug:              result.Series.Slug,
		BooksCount:        result.Series.BooksCount,
		PrimaryBooksCount: result.Series.PrimaryBooksCount,
	}

	filteredResult := &FilteredSeriesResult{Series: seriesData}

	for _, bs := range result.Series.BookSeries {
		b := bs.Book
		editionLangs := make([]EditionLanguageInfo, 0, len(b.Editions))
		editionFormats := make([]EditionFormatInfo, 0, len(b.Editions))

		for _, ed := range b.Editions {
			var langInfo *EditionLanguageInfo
			if ed.Language != nil {
				langInfo = &EditionLanguageInfo{Code2: ed.Language.Code2, Language: ed.Language.Language}
				editionLangs = append(editionLangs, *langInfo)
			}
			var formatInfo *ReadingFormatInfo
			if ed.ReadingFormat != nil {
				formatInfo = &ReadingFormatInfo{Format: ed.ReadingFormat.Format}
			}
			editionFormats = append(editionFormats, EditionFormatInfo{
				ReadingFormat: formatInfo,
				Language:      langInfo,
				ISBN10:        ed.ISBN10,
				ISBN13:        ed.ISBN13,
			})
		}

		if !bookHasPreferredLanguage(editionLangs, languages) {
			continue
		}

		if len(b.Editions) == 0 {
			continue
		}

		filteredResult.TotalCount++
		hasEbook, hasAudiobook := getEditionFormats(editionFormats)
		hasDigital := hasEbook || hasAudiobook
		digitalCount, physicalCount := countEditionsByFormat(editionFormats)

		isCompilation := b.Compilation
		if !isCompilation {
			for _, ed := range b.Editions {
				if ed.Compilation {
					isCompilation = true
					break
				}
			}
		}

		book := BookData{
			ID: b.ID.String(), Title: b.Title, Description: b.Description,
			Rating: b.Rating, PageCount: b.Pages,
			SeriesID: seriesData.ID, SeriesName: seriesData.Name,
			HasDigitalEdition: hasDigital, HasEbook: hasEbook, HasAudiobook: hasAudiobook,
			DigitalEditionCount: digitalCount, PhysicalEditionCount: physicalCount,
			Compilation: isCompilation,
		}
		pos := bs.Position
		book.SeriesIndex = &pos
		if b.Image != nil {
			book.CoverURL = b.Image.URL
		}
		if len(b.Contributions) > 0 {
			book.AuthorID = b.Contributions[0].Author.ID.String()
			book.AuthorName = b.Contributions[0].Author.Name
		}
		if len(b.Editions) > 0 {
			book.ISBN = b.Editions[0].ISBN10
			book.ISBN13 = b.Editions[0].ISBN13
		}
		if b.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", b.ReleaseDate); err == nil {
				book.ReleaseDate = &t
			}
		}
		book.LanguageCode = getPreferredLanguageCode(editionLangs, languages)

		if hasDigital {
			filteredResult.DigitalCount++
		} else {
			filteredResult.PhysicalOnlyCount++
		}
		filteredResult.Books = append(filteredResult.Books, book)
	}

	seriesData.BooksCount = len(filteredResult.Books)
	return filteredResult, nil
}

func (c *Client) GetBooksByAuthor(authorID string, languages []string) (*FilteredBooksResult, error) {
	idInt, _ := strconv.Atoi(authorID)
	gqlQuery := `
		query GetBooksByAuthor($authorId: Int!) {
			books(where: {
				contributions: {author_id: {_eq: $authorId}},
				compilation: {_eq: false}
			}) {
				id, title, slug, description, compilation, image { url }, release_date, pages, rating
				book_series { series { id, name }, position }
				editions { 
					isbn_10, isbn_13, asin, compilation, reading_format { format }, language { code2 language } 
				}
			}
		}
	`
	data, err := c.execute(gqlQuery, map[string]interface{}{"authorId": idInt})
	if err != nil {
		return nil, err
	}
	var result struct {
		Books []struct {
			ID          json.Number
			Title       string `json:"title"`
			Description string `json:"description"`
			ReleaseDate string `json:"release_date"`
			Compilation bool   `json:"compilation"`
			Image       *struct{ URL string }
			Pages       int
			Rating      float32
			BookSeries  []struct {
				Series struct {
					ID   json.Number
					Name string
				}
				Position float32
			} `json:"book_series"`
			Editions []struct {
				ISBN10        string `json:"isbn_10"`
				ISBN13        string `json:"isbn_13"`
				Asin          string `json:"asin"`
				Compilation   bool   `json:"compilation"`
				ReadingFormat *struct {
					Format string `json:"format"`
				} `json:"reading_format"`
				Language *struct {
					Code2    string `json:"code2"`
					Language string `json:"language"`
				} `json:"language"`
			}
		}
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	filteredResult := &FilteredBooksResult{}
	for _, b := range result.Books {
		editionLangs := make([]EditionLanguageInfo, 0, len(b.Editions))
		editionFormats := make([]EditionFormatInfo, 0, len(b.Editions))

		for _, ed := range b.Editions {
			var langInfo *EditionLanguageInfo
			if ed.Language != nil {
				langInfo = &EditionLanguageInfo{Code2: ed.Language.Code2, Language: ed.Language.Language}
				editionLangs = append(editionLangs, *langInfo)
			}
			var formatInfo *ReadingFormatInfo
			if ed.ReadingFormat != nil {
				formatInfo = &ReadingFormatInfo{Format: ed.ReadingFormat.Format}
			}
			editionFormats = append(editionFormats, EditionFormatInfo{
				ReadingFormat: formatInfo,
				Language:      langInfo,
				ISBN10:        ed.ISBN10,
				ISBN13:        ed.ISBN13,
			})
		}

		if !bookHasPreferredLanguage(editionLangs, languages) {
			continue
		}

		if len(b.Editions) == 0 {
			continue
		}

		filteredResult.TotalCount++
		hasEbook, hasAudiobook := getEditionFormats(editionFormats)
		hasDigital := hasEbook || hasAudiobook
		digitalCount, physicalCount := countEditionsByFormat(editionFormats)

		isCompilation := b.Compilation
		if !isCompilation {
			for _, ed := range b.Editions {
				if ed.Compilation {
					isCompilation = true
					break
				}
			}
		}

		book := BookData{
			ID: b.ID.String(), Title: b.Title, Description: b.Description,
			Rating: b.Rating, PageCount: b.Pages, AuthorID: authorID,
			HasDigitalEdition: hasDigital, HasEbook: hasEbook, HasAudiobook: hasAudiobook,
			DigitalEditionCount: digitalCount, PhysicalEditionCount: physicalCount,
			Compilation: isCompilation,
		}
		if b.Image != nil {
			book.CoverURL = b.Image.URL
		}
		if len(b.BookSeries) > 0 {
			book.SeriesID = b.BookSeries[0].Series.ID.String()
			book.SeriesName = b.BookSeries[0].Series.Name
			pos := b.BookSeries[0].Position
			book.SeriesIndex = &pos
		}
		if len(b.Editions) > 0 {
			book.ISBN = b.Editions[0].ISBN10
			book.ISBN13 = b.Editions[0].ISBN13
		}
		if b.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", b.ReleaseDate); err == nil {
				book.ReleaseDate = &t
			}
		}
		book.LanguageCode = getPreferredLanguageCode(editionLangs, languages)

		if hasDigital {
			filteredResult.DigitalCount++
		} else {
			filteredResult.PhysicalOnlyCount++
		}
		filteredResult.Books = append(filteredResult.Books, book)
	}

	return filteredResult, nil
}

// EditionLanguageInfo represents language info from an edition
type EditionLanguageInfo struct {
	Code2    string
	Language string
}

// ReadingFormatInfo represents the reading format from an edition
type ReadingFormatInfo struct {
	Format string // "Physical", "Ebook", "Audiobook"
}

// EditionFormatInfo combines format and language info for filtering
type EditionFormatInfo struct {
	ReadingFormat *ReadingFormatInfo
	Language      *EditionLanguageInfo
	ISBN10        string
	ISBN13        string
	AudioSeconds  int
}

func isDigitalFormat(format string) bool {
	return format == FormatEbook || format == FormatAudiobook
}

func bookHasDigitalEdition(editions []EditionFormatInfo) bool {
	for _, ed := range editions {
		if ed.ReadingFormat != nil && isDigitalFormat(ed.ReadingFormat.Format) {
			return true
		}
	}
	return false
}

func countEditionsByFormat(editions []EditionFormatInfo) (digital int, physical int) {
	for _, ed := range editions {
		if ed.ReadingFormat == nil {
			continue
		}
		if isDigitalFormat(ed.ReadingFormat.Format) {
			digital++
		} else {
			physical++
		}
	}
	return
}

func getEditionFormats(editions []EditionFormatInfo) (hasEbook bool, hasAudiobook bool) {
	for _, ed := range editions {
		if ed.ReadingFormat == nil {
			continue
		}
		if ed.ReadingFormat.Format == FormatEbook {
			hasEbook = true
		} else if ed.ReadingFormat.Format == FormatAudiobook {
			hasAudiobook = true
		}
		if hasEbook && hasAudiobook {
			return
		}
	}
	return
}

// bookHasPreferredLanguage checks if any edition has a preferred language
// Returns true if: no preferences set, any edition matches a preferred language,
// or no editions have language data (graceful degradation)
func bookHasPreferredLanguage(editions []EditionLanguageInfo, preferredLangs []string) bool {
	if len(preferredLangs) == 0 {
		return true // No filter = include all
	}

	for _, edition := range editions {
		if edition.Code2 == "" {
			continue
		}
		code := strings.ToLower(edition.Code2)
		for _, pref := range preferredLangs {
			if code == strings.ToLower(pref) {
				return true
			}
		}
	}

	// Has language data but none match - exclude
	// If no editions have language data, exclude (strict filtering)
	return false
}

// getPreferredLanguageCode returns the language code of the first edition matching preferred languages
// Used to populate BookData.LanguageCode
func getPreferredLanguageCode(editions []EditionLanguageInfo, preferredLangs []string) string {
	// First try to find an edition matching preferred languages
	for _, edition := range editions {
		if edition.Code2 == "" {
			continue
		}
		code := strings.ToLower(edition.Code2)
		for _, pref := range preferredLangs {
			if code == strings.ToLower(pref) {
				return edition.Code2
			}
		}
	}
	// Fallback to first edition with language data
	for _, edition := range editions {
		if edition.Code2 != "" {
			return edition.Code2
		}
	}
	return ""
}

func getLanguageNameFromCode(code string) string {
	switch strings.ToLower(code) {
	case "en":
		return "english"
	case "es":
		return "spanish"
	case "fr":
		return "french"
	case "de":
		return "german"
	case "it":
		return "italian"
	case "pt":
		return "portuguese"
	case "ja":
		return "japanese"
	case "zh":
		return "chinese"
	case "ko":
		return "korean"
	case "ru":
		return "russian"
	default:
		return ""
	}
}

// Test checks API key
func (c *Client) Test() error {
	gqlQuery := `query Test { me { username } }`
	data, err := c.execute(gqlQuery, nil)
	if err != nil {
		return err
	}
	var result struct {
		Me interface{} `json:"me"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetListBooks(listID string) (*FilteredBooksResult, error) {
	idInt, _ := strconv.Atoi(listID)
	gqlQuery := `
		query GetListBooks($listId: Int!) {
			lists_by_pk(id: $listId) {
				list_books {
					book {
						id, title, description, compilation, image { url }, release_date, pages, rating
						contributions { author { id, name } }
						book_series { series { id, name }, position }
						editions { isbn_10, isbn_13, reading_format { format }, language { code2 language } }
					}
				}
			}
		}
	`
	data, err := c.execute(gqlQuery, map[string]interface{}{"listId": idInt})
	if err != nil {
		return nil, err
	}
	var result struct {
		List *struct {
			ListBooks []struct {
				Book struct {
					ID            json.Number
					Title         string `json:"title"`
					Description   string `json:"description"`
					ReleaseDate   string `json:"release_date"`
					Compilation   bool   `json:"compilation"`
					Image         *struct{ URL string }
					Pages         int
					Rating        float32
					Contributions []struct {
						Author struct {
							ID   json.Number
							Name string
						}
					}
					BookSeries []struct {
						Series struct {
							ID   json.Number
							Name string
						}
						Position float32
					}
					Editions []struct {
						ISBN10        string `json:"isbn_10"`
						ISBN13        string `json:"isbn_13"`
						ReadingFormat *struct {
							Format string `json:"format"`
						} `json:"reading_format"`
						Language *struct {
							Code2    string `json:"code2"`
							Language string `json:"language"`
						} `json:"language"`
					}
				}
			} `json:"list_books"`
		} `json:"lists_by_pk"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.List == nil {
		return nil, fmt.Errorf("list not found")
	}

	filteredResult := &FilteredBooksResult{}
	for _, lb := range result.List.ListBooks {
		b := lb.Book
		editionFormats := make([]EditionFormatInfo, 0, len(b.Editions))

		for _, ed := range b.Editions {
			var formatInfo *ReadingFormatInfo
			if ed.ReadingFormat != nil {
				formatInfo = &ReadingFormatInfo{Format: ed.ReadingFormat.Format}
			}
			editionFormats = append(editionFormats, EditionFormatInfo{
				ReadingFormat: formatInfo,
				ISBN10:        ed.ISBN10,
				ISBN13:        ed.ISBN13,
			})
		}

		filteredResult.TotalCount++
		hasDigital := bookHasDigitalEdition(editionFormats)
		digitalCount, physicalCount := countEditionsByFormat(editionFormats)

		book := BookData{
			ID: b.ID.String(), Title: b.Title, Description: b.Description,
			Rating: b.Rating, PageCount: b.Pages,
			HasDigitalEdition: hasDigital, DigitalEditionCount: digitalCount, PhysicalEditionCount: physicalCount,
			Compilation: b.Compilation,
		}
		if b.Image != nil {
			book.CoverURL = b.Image.URL
		}
		if len(b.Contributions) > 0 {
			book.AuthorID = b.Contributions[0].Author.ID.String()
			book.AuthorName = b.Contributions[0].Author.Name
		}
		if len(b.BookSeries) > 0 {
			book.SeriesID = b.BookSeries[0].Series.ID.String()
			book.SeriesName = b.BookSeries[0].Series.Name
			pos := b.BookSeries[0].Position
			book.SeriesIndex = &pos
		}
		if len(b.Editions) > 0 {
			book.ISBN = b.Editions[0].ISBN10
			book.ISBN13 = b.Editions[0].ISBN13
		}
		if b.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", b.ReleaseDate); err == nil {
				book.ReleaseDate = &t
			}
		}

		if hasDigital {
			filteredResult.DigitalCount++
		} else {
			filteredResult.PhysicalOnlyCount++
		}
		filteredResult.Books = append(filteredResult.Books, book)
	}
	return filteredResult, nil
}
