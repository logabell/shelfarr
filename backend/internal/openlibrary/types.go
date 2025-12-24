package openlibrary

// SearchResponse represents the response from Open Library's search API
type SearchResponse struct {
	NumFound      int         `json:"numFound"`
	Start         int         `json:"start"`
	NumFoundExact bool        `json:"numFoundExact"`
	Docs          []SearchDoc `json:"docs"`
}

// SearchDoc represents a single document in search results
type SearchDoc struct {
	Key                 string   `json:"key"` // Work key like "/works/OL45804W"
	Title               string   `json:"title"`
	AuthorName          []string `json:"author_name,omitempty"`
	AuthorKey           []string `json:"author_key,omitempty"`
	FirstPublishYear    int      `json:"first_publish_year,omitempty"`
	EditionCount        int      `json:"edition_count,omitempty"`
	CoverI              int      `json:"cover_i,omitempty"` // Cover ID for covers API
	ISBN                []string `json:"isbn,omitempty"`
	OCLC                []string `json:"oclc,omitempty"`
	LCCN                []string `json:"lccn,omitempty"`
	HasFulltext         bool     `json:"has_fulltext,omitempty"`
	PublicScanB         bool     `json:"public_scan_b,omitempty"`
	Language            []string `json:"language,omitempty"`
	Subject             []string `json:"subject,omitempty"`
	Publisher           []string `json:"publisher,omitempty"`
	NumberOfPagesMedian int      `json:"number_of_pages_median,omitempty"`
	RatingsAverage      float64  `json:"ratings_average,omitempty"`
	RatingsCount        int      `json:"ratings_count,omitempty"`
	IA                  []string `json:"ia,omitempty"` // Internet Archive identifiers
}

// BookData represents the response from /api/books endpoint with jscmd=data
type BookData struct {
	URL             string               `json:"url"`
	Key             string               `json:"key"`
	Title           string               `json:"title"`
	Subtitle        string               `json:"subtitle,omitempty"`
	Authors         []BookAuthor         `json:"authors,omitempty"`
	Publishers      []BookPublisher      `json:"publishers,omitempty"`
	PublishDate     string               `json:"publish_date,omitempty"`
	NumberOfPages   int                  `json:"number_of_pages,omitempty"`
	Weight          string               `json:"weight,omitempty"`
	Subjects        []BookSubject        `json:"subjects,omitempty"`
	Cover           *BookCover           `json:"cover,omitempty"`
	Identifiers     *BookIdentifiers     `json:"identifiers,omitempty"`
	Classifications *BookClassifications `json:"classifications,omitempty"`
	Excerpts        []BookExcerpt        `json:"excerpts,omitempty"`
	Notes           string               `json:"notes,omitempty"`
}

// BookAuthor represents an author in BookData
type BookAuthor struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// BookPublisher represents a publisher in BookData
type BookPublisher struct {
	Name string `json:"name"`
}

// BookSubject represents a subject in BookData
type BookSubject struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// BookCover represents cover URLs in BookData
type BookCover struct {
	Small  string `json:"small,omitempty"`
	Medium string `json:"medium,omitempty"`
	Large  string `json:"large,omitempty"`
}

// BookIdentifiers contains various book identifiers
type BookIdentifiers struct {
	ISBN10      []string `json:"isbn_10,omitempty"`
	ISBN13      []string `json:"isbn_13,omitempty"`
	OCLC        []string `json:"oclc,omitempty"`
	LCCN        []string `json:"lccn,omitempty"`
	OpenLibrary []string `json:"openlibrary,omitempty"`
}

// BookClassifications contains library classifications
type BookClassifications struct {
	LCClassifications []string `json:"lc_classifications,omitempty"`
	DeweyDecimalClass []string `json:"dewey_decimal_class,omitempty"`
}

// BookExcerpt represents an excerpt from the book
type BookExcerpt struct {
	Text    string `json:"text"`
	Comment string `json:"comment,omitempty"`
}

// Work represents a work from Open Library's works API
type Work struct {
	Key              string       `json:"key"`
	Title            string       `json:"title"`
	Description      interface{}  `json:"description,omitempty"` // Can be string or object with "value"
	Subjects         []string     `json:"subjects,omitempty"`
	SubjectPlaces    []string     `json:"subject_places,omitempty"`
	SubjectTimes     []string     `json:"subject_times,omitempty"`
	SubjectPeople    []string     `json:"subject_people,omitempty"`
	Authors          []WorkAuthor `json:"authors,omitempty"`
	Covers           []int        `json:"covers,omitempty"`
	FirstPublishDate string       `json:"first_publish_date,omitempty"`
	Created          *TypedValue  `json:"created,omitempty"`
	LastModified     *TypedValue  `json:"last_modified,omitempty"`
}

// WorkAuthor represents an author reference in a Work
type WorkAuthor struct {
	Author WorkAuthorRef `json:"author"`
	Type   TypedValue    `json:"type,omitempty"`
}

// WorkAuthorRef is a reference to an author
type WorkAuthorRef struct {
	Key string `json:"key"`
}

// TypedValue represents a typed value with key
type TypedValue struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Edition represents an edition from Open Library
type Edition struct {
	Key            string        `json:"key"`
	Title          string        `json:"title"`
	Subtitle       string        `json:"subtitle,omitempty"`
	Publishers     []string      `json:"publishers,omitempty"`
	PublishDate    string        `json:"publish_date,omitempty"`
	ISBN10         []string      `json:"isbn_10,omitempty"`
	ISBN13         []string      `json:"isbn_13,omitempty"`
	NumberOfPages  int           `json:"number_of_pages,omitempty"`
	Covers         []int         `json:"covers,omitempty"`
	Languages      []LanguageRef `json:"languages,omitempty"`
	Works          []WorkRef     `json:"works,omitempty"`
	PhysicalFormat string        `json:"physical_format,omitempty"`
	Weight         string        `json:"weight,omitempty"`
	Pagination     string        `json:"pagination,omitempty"`
}

// LanguageRef is a reference to a language
type LanguageRef struct {
	Key string `json:"key"`
}

// WorkRef is a reference to a work
type WorkRef struct {
	Key string `json:"key"`
}

// EditionsResponse represents the response from /works/{id}/editions.json
type EditionsResponse struct {
	Links   EditionsLinks `json:"links"`
	Size    int           `json:"size"`
	Entries []Edition     `json:"entries"`
}

// EditionsLinks contains pagination links
type EditionsLinks struct {
	Self string `json:"self"`
	Work string `json:"work"`
	Next string `json:"next,omitempty"`
}

// Author represents an author from Open Library
type Author struct {
	Key            string       `json:"key"`
	Name           string       `json:"name"`
	PersonalName   string       `json:"personal_name,omitempty"`
	AlternateNames []string     `json:"alternate_names,omitempty"`
	Bio            interface{}  `json:"bio,omitempty"` // Can be string or object with "value"
	BirthDate      string       `json:"birth_date,omitempty"`
	DeathDate      string       `json:"death_date,omitempty"`
	Photos         []int        `json:"photos,omitempty"`
	Links          []AuthorLink `json:"links,omitempty"`
	Wikipedia      string       `json:"wikipedia,omitempty"`
}

// AuthorLink represents an external link for an author
type AuthorLink struct {
	URL   string     `json:"url"`
	Title string     `json:"title"`
	Type  TypedValue `json:"type,omitempty"`
}

// AuthorWorksResponse represents the response from /authors/{id}/works.json
type AuthorWorksResponse struct {
	Links   AuthorWorksLinks `json:"links"`
	Size    int              `json:"size"`
	Entries []AuthorWork     `json:"entries"`
}

// AuthorWorksLinks contains links for author works pagination
type AuthorWorksLinks struct {
	Self   string `json:"self"`
	Author string `json:"author"`
	Next   string `json:"next,omitempty"`
}

// AuthorWork represents a work by an author
type AuthorWork struct {
	Key          string       `json:"key"`
	Title        string       `json:"title"`
	EditionCount int          `json:"edition_count,omitempty"`
	CoverID      int          `json:"cover_id,omitempty"`
	Covers       []int        `json:"covers,omitempty"`
	Authors      []WorkAuthor `json:"authors,omitempty"`
}

// AuthorSearchResponse represents author search results
type AuthorSearchResponse struct {
	NumFound int               `json:"numFound"`
	Start    int               `json:"start"`
	Docs     []AuthorSearchDoc `json:"docs"`
}

// AuthorSearchDoc represents an author in search results
type AuthorSearchDoc struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	BirthDate   string   `json:"birth_date,omitempty"`
	DeathDate   string   `json:"death_date,omitempty"`
	TopWork     string   `json:"top_work,omitempty"`
	WorkCount   int      `json:"work_count,omitempty"`
	TopSubjects []string `json:"top_subjects,omitempty"`
}

// RatingsResponse represents the ratings for a work
type RatingsResponse struct {
	Summary RatingsSummary `json:"summary"`
	Counts  RatingsCounts  `json:"counts"`
}

// RatingsSummary contains the summary statistics
type RatingsSummary struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

// RatingsCounts contains counts per rating value
type RatingsCounts struct {
	One   int `json:"1"`
	Two   int `json:"2"`
	Three int `json:"3"`
	Four  int `json:"4"`
	Five  int `json:"5"`
}

// SubjectResponse represents books for a subject
type SubjectResponse struct {
	Key         string             `json:"key"`
	Name        string             `json:"name"`
	SubjectType string             `json:"subject_type"`
	WorkCount   int                `json:"work_count"`
	Works       []SubjectWork      `json:"works"`
	Authors     []SubjectAuthor    `json:"authors,omitempty"`
	Publishers  []SubjectPublisher `json:"publishers,omitempty"`
}

// SubjectWork represents a work under a subject
type SubjectWork struct {
	Key              string              `json:"key"`
	Title            string              `json:"title"`
	EditionCount     int                 `json:"edition_count"`
	CoverID          int                 `json:"cover_id,omitempty"`
	CoverEditionKey  string              `json:"cover_edition_key,omitempty"`
	Authors          []SubjectWorkAuthor `json:"authors,omitempty"`
	FirstPublishYear int                 `json:"first_publish_year,omitempty"`
	HasFulltext      bool                `json:"has_fulltext"`
	PublicScanB      bool                `json:"public_scan_b"`
	IA               string              `json:"ia,omitempty"`
}

// SubjectWorkAuthor is an author in a subject work
type SubjectWorkAuthor struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// SubjectAuthor is an author associated with a subject
type SubjectAuthor struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SubjectPublisher is a publisher associated with a subject
type SubjectPublisher struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// PartnerReadResponse represents available reading options from the Partner/Read API
type PartnerReadResponse struct {
	Items map[string]PartnerReadItem `json:"items"`
}

// PartnerReadItem represents a single item from the Partner/Read API
type PartnerReadItem struct {
	Status      string            `json:"status"` // "full access", "lendable", "checked out", "restricted"
	ItemURL     string            `json:"itemURL,omitempty"`
	Cover       *PartnerReadCover `json:"cover,omitempty"`
	Match       string            `json:"match,omitempty"` // "exact" or "similar"
	PublishDate string            `json:"publishDate,omitempty"`
}

// PartnerReadCover contains cover URLs from Partner API
type PartnerReadCover struct {
	Small  string `json:"small,omitempty"`
	Medium string `json:"medium,omitempty"`
	Large  string `json:"large,omitempty"`
}

// ViewAPIResponse represents the response from /api/books with jscmd=viewapi
type ViewAPIResponse struct {
	BibKey       string `json:"bib_key"`
	InfoURL      string `json:"info_url"`
	Preview      string `json:"preview"` // "noview", "partial", or "full"
	PreviewURL   string `json:"preview_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}
