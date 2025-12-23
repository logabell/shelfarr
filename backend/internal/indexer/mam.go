package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MAMIndexer implements the MyAnonaMouse indexer
type MAMIndexer struct {
	name          string
	cookie        string
	vipOnly       bool
	freeleechOnly bool
	httpClient    *http.Client
}

// MAM category constants
const (
	MAMCategoryAudiobooks = 13
	MAMCategoryEbooks     = 14
	MAMCategoryMusicology = 15
	MAMCategoryRadio      = 16
)

// Helper functions to handle interface{} types from MAM API
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case string:
		i, _ := strconv.Atoi(val)
		return i
	case float64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	default:
		return 0
	}
}

// NewMAMIndexer creates a new MAM indexer
func NewMAMIndexer(name, cookie string, vipOnly, freeleechOnly bool) *MAMIndexer {
	return &MAMIndexer{
		name:          name,
		cookie:        cookie,
		vipOnly:       vipOnly,
		freeleechOnly: freeleechOnly,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *MAMIndexer) Name() string {
	return m.name
}

func (m *MAMIndexer) Type() string {
	return "mam"
}

// normalizeCookie ensures the cookie has the proper format
// Users from MAM Security page get just the token value, so we prepend "mam_id=" if needed
func (m *MAMIndexer) normalizeCookie() string {
	cookie := strings.TrimSpace(m.cookie)
	if cookie == "" {
		return cookie
	}
	// If the cookie doesn't contain an equals sign, assume it's just the token value
	if !strings.Contains(cookie, "=") {
		cookie = "mam_id=" + cookie
	}
	return cookie
}

// mamSearchRequest represents the JSON request body for MAM search API
type mamSearchRequest struct {
	Tor       mamTorParams `json:"tor"`
	Thumbnail string       `json:"thumbnail,omitempty"`
	DlLink    string       `json:"dlLink,omitempty"`
}

type mamTorParams struct {
	Text                  string   `json:"text"`
	SrchIn                []string `json:"srchIn,omitempty"`
	SearchType            string   `json:"searchType"`
	SearchIn              string   `json:"searchIn,omitempty"`
	Cat                   []string `json:"cat,omitempty"`
	MainCat               []int    `json:"main_cat,omitempty"`
	BrowseFlagsHideVsShow string   `json:"browseFlagsHideVsShow,omitempty"`
	StartDate             string   `json:"startDate,omitempty"`
	EndDate               string   `json:"endDate,omitempty"`
	Hash                  string   `json:"hash,omitempty"`
	SortType              string   `json:"sortType"`
	StartNumber           string   `json:"startNumber"`
}

// mamSearchResponse represents the JSON response from MAM search API
type mamSearchResponse struct {
	Error      string           `json:"error,omitempty"`
	Total      int              `json:"total"`
	TotalFound int              `json:"total_found"`
	Data       []mamTorrentData `json:"data"`
}

type mamTorrentData struct {
	ID                interface{} `json:"id"` // Can be string or number
	Title             string      `json:"title"`
	Name              string      `json:"name"` // Alternative title field
	AuthorInfo        string      `json:"author_info"`
	NarratorInfo      string      `json:"narrator_info"`
	SeriesInfo        string      `json:"series_info"`
	Category          interface{} `json:"category"` // Can be string or number
	Catname           string      `json:"catname"`
	MainCat           interface{} `json:"main_cat"` // Can be string or number
	Size              interface{} `json:"size"`     // Size as string or number
	NumFiles          interface{} `json:"numfiles"`
	Seeders           interface{} `json:"seeders"`
	Leechers          interface{} `json:"leechers"`
	TimesCompleted    interface{} `json:"times_completed"`
	VIP               interface{} `json:"vip"`    // Can be "0"/"1" or 0/1
	Free              interface{} `json:"free"`   // Can be "0"/"1" or 0/1
	FlVIP             interface{} `json:"fl_vip"` // Can be "0"/"1" or 0/1
	Added             string      `json:"added"`
	Tags              string      `json:"tags"`
	Filetype          string      `json:"filetype"`
	LangCode          interface{} `json:"lang_code"` // Can be string or number
	Language          interface{} `json:"language"`  // Can be string or number
	Dl                string      `json:"dl"`        // Download hash
	MySnatched        interface{} `json:"my_snatched"`
	PersonalFreeleech interface{} `json:"personal_freeleech"`
	Bookmarked        interface{} `json:"bookmarked"`
	Description       string      `json:"description,omitempty"`
}

func (m *MAMIndexer) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	baseURL := "https://www.myanonamouse.net/tor/js/loadSearchJSONbasic.php"

	// Build search terms - prioritize Author+Title, fall back to ISBN
	searchTerms := query.Title
	if query.Author != "" {
		searchTerms = query.Author + " " + searchTerms
	}
	// If no title/author, try ISBN
	if strings.TrimSpace(searchTerms) == "" && query.ISBN != "" {
		searchTerms = query.ISBN
	}

	// Build search request
	searchReq := mamSearchRequest{
		Tor: mamTorParams{
			Text:        strings.TrimSpace(searchTerms),
			SearchType:  m.getSearchType(),
			SortType:    "default",
			StartNumber: "0",
			SrchIn:      []string{"title", "author", "narrator", "series"},
			SearchIn:    "torrents",
			Cat:         []string{"0"}, // All subcategories
		},
		DlLink: "true", // Request download hash
	}

	// Set main category based on media type
	if query.MediaType == "audiobook" {
		searchReq.Tor.MainCat = []int{MAMCategoryAudiobooks}
	} else if query.MediaType == "ebook" {
		searchReq.Tor.MainCat = []int{MAMCategoryEbooks}
	}
	// If not specified, search both

	jsonBody, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", m.normalizeCookie())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Shelfarr/1.0")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MAM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("MAM authentication failed - check cookie")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mamResp mamSearchResponse
	if err := json.Unmarshal(body, &mamResp); err != nil {
		return nil, fmt.Errorf("failed to parse MAM response: %w", err)
	}

	if mamResp.Error != "" {
		return nil, fmt.Errorf("MAM API error: %s", mamResp.Error)
	}

	// Convert to SearchResult
	results := make([]SearchResult, 0, len(mamResp.Data))
	for _, item := range mamResp.Data {
		// Parse numeric values from interface{} (can be string or number)
		size := toInt64(item.Size)
		seeders := toInt(item.Seeders)
		leechers := toInt(item.Leechers)

		// Determine if freeleech/VIP (can be "1" or 1)
		isFree := toString(item.Free) == "1" || toString(item.FlVIP) == "1"
		isVIP := toString(item.VIP) == "1" || toString(item.FlVIP) == "1"

		// Apply VIP/Freeleech filters
		if m.vipOnly && !isVIP {
			continue
		}
		if m.freeleechOnly && !isFree {
			continue
		}

		// Use title or name field
		title := item.Title
		if title == "" {
			title = item.Name
		}

		// Build download URL using hash
		downloadURL := ""
		if item.Dl != "" {
			downloadURL = fmt.Sprintf("https://www.myanonamouse.net/tor/download.php/%s", item.Dl)
		}

		result := SearchResult{
			Title:       title,
			Size:        size,
			Seeders:     seeders,
			Leechers:    leechers,
			DownloadURL: downloadURL,
			InfoURL:     fmt.Sprintf("https://www.myanonamouse.net/t/%s", toString(item.ID)),
			PublishDate: item.Added,
			Freeleech:   isFree,
			VIP:         isVIP,
			Indexer:     m.name,
			LangCode:    toString(item.LangCode),
		}

		// Detect format from filetype or title
		if item.Filetype != "" {
			result.Format = detectFormatFromFiletype(item.Filetype)
		} else {
			result.Format = detectFormat(title)
		}

		// Parse author info for display
		result.Author = parseAuthorInfo(item.AuthorInfo)
		result.Narrator = parseAuthorInfo(item.NarratorInfo)
		result.Category = item.Catname

		// Parse series info
		seriesName, seriesIdx := parseSeriesInfo(item.SeriesInfo)
		result.SeriesName = seriesName
		result.SeriesIndex = seriesIdx

		// Parse bitrate from tags (for audiobooks)
		if query.MediaType == "audiobook" || strings.Contains(strings.ToLower(item.Catname), "audio") {
			result.Bitrate = parseBitrateFromTags(item.Tags)
			result.Duration = parseDurationFromTitle(title)
		}

		results = append(results, result)
	}

	return results, nil
}

func (m *MAMIndexer) getSearchType() string {
	if m.freeleechOnly && m.vipOnly {
		return "fl-VIP"
	}
	if m.freeleechOnly {
		return "fl"
	}
	if m.vipOnly {
		return "VIP"
	}
	return "all"
}

func (m *MAMIndexer) Test(ctx context.Context) error {
	// Test by making a simple search request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.myanonamouse.net/jsonLoad.php", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", m.normalizeCookie())
	req.Header.Set("User-Agent", "Shelfarr/1.0")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirect to login page (indicates auth failure)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed - invalid or expired cookie")
	}

	// Check if we got a valid JSON response (authenticated users get JSON)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// If the response contains HTML or login redirect, auth failed
	if bytes.Contains(body, []byte("<html")) || bytes.Contains(body, []byte("Login")) {
		return fmt.Errorf("authentication failed - cookie may be expired")
	}

	return nil
}

func (m *MAMIndexer) Download(ctx context.Context, result SearchResult) (string, error) {
	// For MAM, the download URL should already contain the hash
	if result.DownloadURL != "" && strings.Contains(result.DownloadURL, "download.php") {
		return result.DownloadURL, nil
	}

	// If we need to fetch the download link, extract torrent ID and get it
	// This would require another API call to get the torrent page
	return result.DownloadURL, nil
}

// parseAuthorInfo extracts author names from MAM's JSON format
// Input format: {"8234": "Kerrelyn Sparks", "123": "Another Author"}
func parseAuthorInfo(info string) string {
	if info == "" || info == "null" {
		return ""
	}

	var authors map[string]string
	if err := json.Unmarshal([]byte(info), &authors); err != nil {
		return ""
	}

	names := make([]string, 0, len(authors))
	for _, name := range authors {
		names = append(names, name)
	}

	return strings.Join(names, ", ")
}

// detectFormatFromFiletype extracts format from MAM's filetype field
// Input format: "m4a mp3" or "epub pdf"
func detectFormatFromFiletype(filetype string) string {
	filetype = strings.ToLower(filetype)

	// Prioritize certain formats
	if strings.Contains(filetype, "m4b") {
		return "M4B"
	}
	if strings.Contains(filetype, "m4a") {
		return "M4A"
	}
	if strings.Contains(filetype, "mp3") {
		return "MP3"
	}
	if strings.Contains(filetype, "epub") {
		return "EPUB"
	}
	if strings.Contains(filetype, "azw3") {
		return "AZW3"
	}
	if strings.Contains(filetype, "mobi") {
		return "MOBI"
	}
	if strings.Contains(filetype, "pdf") {
		return "PDF"
	}

	return strings.ToUpper(strings.Split(filetype, " ")[0])
}

// detectFormat tries to detect the format from the title
func detectFormat(title string) string {
	title = strings.ToLower(title)

	formatPatterns := []struct {
		pattern string
		format  string
	}{
		{`\bm4b\b`, "M4B"},
		{`\bm4a\b`, "M4A"},
		{`\bmp3\b`, "MP3"},
		{`\bflac\b`, "FLAC"},
		{`\bepub\b`, "EPUB"},
		{`\bazw3?\b`, "AZW3"},
		{`\bmobi\b`, "MOBI"},
		{`\bpdf\b`, "PDF"},
		{`\bcbz\b`, "CBZ"},
		{`\bcbr\b`, "CBR"},
	}

	for _, fp := range formatPatterns {
		if matched, _ := regexp.MatchString(fp.pattern, title); matched {
			return fp.format
		}
	}

	return "Unknown"
}

// parseSeriesInfo extracts series name and index from MAM's JSON format
// Input format: {"67": ["Love at Stake", "01-16, 13.5"]}
func parseSeriesInfo(info string) (name string, index string) {
	if info == "" || info == "null" || info == "{}" {
		return "", ""
	}

	// Parse the JSON map
	var series map[string]interface{}
	if err := json.Unmarshal([]byte(info), &series); err != nil {
		return "", ""
	}

	// Get the first series (usually there's only one)
	for _, v := range series {
		switch val := v.(type) {
		case []interface{}:
			if len(val) >= 1 {
				if n, ok := val[0].(string); ok {
					name = n
				}
			}
			if len(val) >= 2 {
				if idx, ok := val[1].(string); ok {
					index = idx
				}
			}
			return name, index
		case string:
			name = val
			return name, ""
		}
	}

	return "", ""
}

// parseBitrateFromTags extracts bitrate from MAM tags field
// Tags can contain values like "64kbps", "128kbps", "256kbps", "64-128 Kbps"
func parseBitrateFromTags(tags string) int {
	if tags == "" {
		return 0
	}

	tags = strings.ToLower(tags)

	// Look for patterns like "128kbps", "64 kbps", "64–128 kbps"
	patterns := []string{
		`(\d+)\s*kbps`,  // Simple: 128kbps
		`(\d+)\s*kb/s`,  // Alternative: 128kb/s
		`\b(\d+)\s*k\b`, // Short: 128k (must be word boundary)
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(tags, -1)

		if len(matches) > 0 {
			// If multiple matches (like "64-128"), take the higher one
			maxBitrate := 0
			for _, match := range matches {
				if len(match) > 1 {
					if bitrate, err := strconv.Atoi(match[1]); err == nil {
						if bitrate > maxBitrate {
							maxBitrate = bitrate
						}
					}
				}
			}
			if maxBitrate > 0 {
				return maxBitrate
			}
		}
	}

	return 0
}

// parseDurationFromTitle attempts to extract duration from title
// Looks for patterns like "12h30m", "10 hrs", "10 hours", "(12:30:45)"
func parseDurationFromTitle(title string) int {
	title = strings.ToLower(title)

	// Pattern: "Xh Ym" or "XhYm"
	hm := regexp.MustCompile(`(\d+)\s*h\s*(\d+)\s*m`)
	if matches := hm.FindStringSubmatch(title); len(matches) == 3 {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		return hours*3600 + minutes*60
	}

	// Pattern: "X hours Y minutes" or "X hrs"
	hrs := regexp.MustCompile(`(\d+)\s*(?:hours?|hrs?)`)
	mins := regexp.MustCompile(`(\d+)\s*(?:minutes?|mins?)`)

	totalSeconds := 0
	if matches := hrs.FindStringSubmatch(title); len(matches) == 2 {
		hours, _ := strconv.Atoi(matches[1])
		totalSeconds += hours * 3600
	}
	if matches := mins.FindStringSubmatch(title); len(matches) == 2 {
		minutes, _ := strconv.Atoi(matches[1])
		totalSeconds += minutes * 60
	}
	if totalSeconds > 0 {
		return totalSeconds
	}

	// Pattern: "(HH:MM:SS)" or "HH:MM:SS"
	timestamp := regexp.MustCompile(`(\d{1,2}):(\d{2}):(\d{2})`)
	if matches := timestamp.FindStringSubmatch(title); len(matches) == 4 {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		seconds, _ := strconv.Atoi(matches[3])
		return hours*3600 + minutes*60 + seconds
	}

	return 0
}
