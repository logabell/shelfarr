package googlebooks

type VolumesResponse struct {
	Kind       string   `json:"kind"`
	TotalItems int      `json:"totalItems"`
	Items      []Volume `json:"items,omitempty"`
}

type Volume struct {
	Kind       string     `json:"kind"`
	ID         string     `json:"id"`
	Etag       string     `json:"etag"`
	SelfLink   string     `json:"selfLink"`
	VolumeInfo VolumeInfo `json:"volumeInfo"`
	SaleInfo   SaleInfo   `json:"saleInfo"`
	AccessInfo AccessInfo `json:"accessInfo"`
}

type VolumeInfo struct {
	Title               string               `json:"title"`
	Subtitle            string               `json:"subtitle,omitempty"`
	Authors             []string             `json:"authors,omitempty"`
	Publisher           string               `json:"publisher,omitempty"`
	PublishedDate       string               `json:"publishedDate,omitempty"`
	Description         string               `json:"description,omitempty"`
	IndustryIdentifiers []IndustryIdentifier `json:"industryIdentifiers,omitempty"`
	PageCount           int                  `json:"pageCount,omitempty"`
	Categories          []string             `json:"categories,omitempty"`
	AverageRating       float32              `json:"averageRating,omitempty"`
	RatingsCount        int                  `json:"ratingsCount,omitempty"`
	MaturityRating      string               `json:"maturityRating,omitempty"`
	ImageLinks          *ImageLinks          `json:"imageLinks,omitempty"`
	Language            string               `json:"language,omitempty"`
	PreviewLink         string               `json:"previewLink,omitempty"`
	InfoLink            string               `json:"infoLink,omitempty"`
	CanonicalVolumeLink string               `json:"canonicalVolumeLink,omitempty"`
}

type IndustryIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type ImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail,omitempty"`
	Thumbnail      string `json:"thumbnail,omitempty"`
	Small          string `json:"small,omitempty"`
	Medium         string `json:"medium,omitempty"`
	Large          string `json:"large,omitempty"`
	ExtraLarge     string `json:"extraLarge,omitempty"`
}

type SaleInfo struct {
	Country     string  `json:"country"`
	Saleability string  `json:"saleability"`
	IsEbook     bool    `json:"isEbook"`
	ListPrice   *Price  `json:"listPrice,omitempty"`
	RetailPrice *Price  `json:"retailPrice,omitempty"`
	BuyLink     string  `json:"buyLink,omitempty"`
	Offers      []Offer `json:"offers,omitempty"`
}

type Price struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}

type Offer struct {
	FinskyOfferType int    `json:"finskyOfferType"`
	ListPrice       *Price `json:"listPrice,omitempty"`
	RetailPrice     *Price `json:"retailPrice,omitempty"`
}

type AccessInfo struct {
	Country                string     `json:"country"`
	Viewability            string     `json:"viewability"`
	Embeddable             bool       `json:"embeddable"`
	PublicDomain           bool       `json:"publicDomain"`
	TextToSpeechPermission string     `json:"textToSpeechPermission"`
	Epub                   FormatInfo `json:"epub"`
	Pdf                    FormatInfo `json:"pdf"`
	WebReaderLink          string     `json:"webReaderLink,omitempty"`
	AccessViewStatus       string     `json:"accessViewStatus"`
	QuoteSharingAllowed    bool       `json:"quoteSharingAllowed"`
}

type FormatInfo struct {
	IsAvailable  bool   `json:"isAvailable"`
	AcsTokenLink string `json:"acsTokenLink,omitempty"`
	DownloadLink string `json:"downloadLink,omitempty"`
}

type EbookInfo struct {
	VolumeID    string
	IsEbook     bool
	HasEpub     bool
	HasPdf      bool
	BuyLink     string
	Title       string
	Authors     []string
	ISBN10      string
	ISBN13      string
	CoverURL    string
	Description string
	PageCount   int
	Language    string
	Categories  []string
	Rating      float32
}
