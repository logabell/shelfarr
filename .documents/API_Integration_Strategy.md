# Shelfarr API Integration Strategy

## Executive Summary

Shelfarr uses a multi-API strategy to gather comprehensive book metadata while respecting rate limits and maximizing data quality. This document defines which API is used for each type of data and how they work together.

**Last Updated:** December 2024

---

## API Priority & Roles

| Priority | API | Role | Rate Limit | Auth Required |
|----------|-----|------|------------|---------------|
| 1 | **Open Library** | Primary search, metadata, covers, authors | Soft limit (User-Agent required) | No |
| 2 | **Google Books** | Ebook detection, purchase links | 1,000/day | API Key |
| 3 | **Hardcover** | Series data, audiobook detection | 60/min | Bearer Token |

### Why This Order?

1. **Open Library as Primary**: No strict rate limit means we can handle heavy search traffic without quota concerns. 20M+ book catalog is comprehensive.

2. **Google Books for Ebook Detection**: The `saleInfo.isEbook` field is the most reliable indicator of ebook availability. However, the 1,000/day limit is too restrictive for primary search use.

3. **Hardcover for Series & Audiobooks**: Hardcover is the ONLY API that provides series data (book positions, series completion status) and reliable audiobook detection via `reading_format`.

---

## Data Source Matrix

### Book Metadata

| Field | Primary Source | Fallback | DB Column |
|-------|---------------|----------|-----------|
| `title` | Open Library | Google Books | `books.title` |
| `description` | Open Library | Google Books | `books.description` |
| `authors` | Open Library | Google Books | via `authors` table |
| `isbn10` | Open Library | Google Books | `books.isbn` |
| `isbn13` | Open Library | Google Books | `books.isbn13` |
| `pageCount` | Open Library | Google Books | `books.page_count` |
| `publishedDate` | Open Library | Google Books | `books.release_date` |
| `language` | Open Library | Google Books | `books.language` |
| `coverUrl` | Open Library | Google Books | `books.cover_url` |
| `rating` | Open Library | Hardcover | `books.rating` |
| `categories` | Open Library | Google Books | `books.genres` (JSON) |

### Ebook/Audiobook Detection

| Field | Source | Cache Duration | DB Column |
|-------|--------|----------------|-----------|
| `isEbook` | Google Books | 24 hours | `books.is_ebook` |
| `hasEpub` | Google Books | 24 hours | `books.has_epub` |
| `hasPdf` | Google Books | 24 hours | `books.has_pdf` |
| `buyLink` | Google Books | 24 hours | `books.buy_link` |
| `isAudiobook` | Hardcover | 24 hours | `books.is_audiobook` |
| `audioDuration` | Hardcover | Forever | `books.audio_duration` |

### Series Data (Hardcover Exclusive)

| Field | Source | DB Location |
|-------|--------|-------------|
| `seriesId` | Hardcover | `series.hardcover_id` |
| `seriesName` | Hardcover | `series.name` |
| `seriesPosition` | Hardcover | `books.series_index` |
| `seriesBooksCount` | Hardcover | `series.books_count` |
| `isSeriesCompleted` | Hardcover | `series.is_completed` |

### Author Data

| Field | Primary Source | Fallback | DB Column |
|-------|---------------|----------|-----------|
| `authorName` | Open Library | Hardcover | `authors.name` |
| `authorBio` | Open Library | Hardcover | `authors.biography` |
| `authorImage` | Open Library | Hardcover | `authors.image_url` |
| `authorBirthDate` | Open Library | Hardcover | `authors.birth_date` |
| `authorBooks` | Open Library | Hardcover | Via relationships |

### Cross-Reference IDs

| ID Field | Purpose | DB Column |
|----------|---------|-----------|
| `openLibraryWorkId` | Link to OL Work | `books.open_library_work_id` |
| `openLibraryEditionId` | Link to OL Edition | `books.open_library_edition_id` |
| `googleVolumeId` | Link to Google Books | `books.google_volume_id` |
| `hardcoverId` | Link to Hardcover | `books.hardcover_id` |

---

## Data Flow Architecture

### 1. Book Search Flow

```
User enters search query
         │
         ▼
┌─────────────────────────────────────┐
│  Open Library Search API            │
│  GET /search.json?q={query}         │
│  • Returns: titles, authors, ISBNs  │
│  • No rate limit concerns           │
└─────────────────────────────────────┘
         │
         ▼
Display search results to user
(Ebook/audiobook status fetched on demand)
```

### 2. Book Detail View Flow

```
User clicks on a book (has ISBN)
         │
         ▼
┌─────────────────────────────────────┐
│  Open Library Books API             │
│  GET /api/books?bibkeys=ISBN:{isbn} │
│  • Returns: full metadata           │
└─────────────────────────────────────┘
         │
         ├──────────────────────────────┐
         ▼                              ▼
┌──────────────────────┐    ┌──────────────────────┐
│  Google Books API    │    │  Hardcover API       │
│  (IF not cached)     │    │  (IF not cached)     │
│  • isEbook           │    │  • Series info       │
│  • EPUB available    │    │  • Audiobook info    │
│  • Buy link          │    │                      │
│  ⚠️ Counts toward    │    │                      │
│     1,000/day quota  │    │                      │
└──────────────────────┘    └──────────────────────┘
         │                              │
         └──────────────┬───────────────┘
                        ▼
              Merge and display
              all book details
```

### 3. Add Book to Library Flow

```
User clicks "Add to Library"
         │
         ▼
┌─────────────────────────────────────┐
│  Check cache for ebook status       │
│  • If cached → use cached data      │
│  • If not → call Google Books API   │
│  • If quota exhausted → warn user   │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Check cache for series/audiobook   │
│  • If cached → use cached data      │
│  • If not → call Hardcover API      │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Store in Database                  │
│  • All metadata fields              │
│  • All cross-reference IDs          │
│  • isEbook, isAudiobook flags       │
│  • Cache timestamps                 │
└─────────────────────────────────────┘
```

### 4. Series Discovery Flow (Hardcover Exclusive)

```
User views series page OR book has series
         │
         ▼
┌─────────────────────────────────────┐
│  Hardcover API                      │
│  series_by_pk(id) + book_series     │
│  • Series name, description         │
│  • All books with positions         │
│  • Completion status                │
└─────────────────────────────────────┘
         │
         ▼
For each book in series:
  - Check if in local library
  - Show "gap" cards for missing books
```

### 5. Author Page Flow

```
User views author page
         │
         ▼
┌─────────────────────────────────────┐
│  Open Library Authors API           │
│  GET /authors/{OLID}.json           │
│  GET /authors/{OLID}/works.json     │
│  • Author bio, image, dates         │
│  • All works by author              │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Hardcover API (supplement)         │
│  • Series associations for books    │
│  • Audiobook availability           │
└─────────────────────────────────────┘
```

---

## Caching Strategy

### Cache Durations

| Data Type | Cache Duration | Storage |
|-----------|---------------|---------|
| Book metadata (from OL) | 7 days | Database |
| Ebook status (from Google) | 24 hours | Database |
| Audiobook status (from HC) | 24 hours | Database |
| Series data (from HC) | 7 days | Database |
| Author data | 7 days | Database |
| Cover images | 30 days | Filesystem |
| Search results | 1 hour | Memory |

### Cache Invalidation

- When user manually refreshes book details
- When metadata sync scheduled task runs
- When book is re-added after deletion

---

## Rate Limit Management

### Google Books (1,000/day)

```go
// Track daily usage in database settings
// Key: "google_books_daily_count"
// Key: "google_books_reset_date"

// Before making request:
if dailyCount >= 1000 {
    return nil, ErrQuotaExhausted
}
```

**When quota exhausted:**
1. Return `isEbook = nil` (unknown status)
2. Allow user to add book with warning
3. Display message: "Google Books API quota reached. Ebook status unavailable until tomorrow."

### Hardcover (60/min)

Already implemented in `backend/internal/hardcover/client.go`:
- Rate limiter: 1 request per second
- Handles 429 responses with backoff

### Open Library (Soft Limit)

```go
// Always include User-Agent
const UserAgent = "Shelfarr/1.0 (https://github.com/shelfarr/shelfarr)"

// For covers: use Cover ID instead of ISBN to avoid rate limits
// ✅ /b/id/12345-M.jpg (no limit)
// ⚠️ /b/isbn/1234567890-M.jpg (100 per 5 min)
```

---

## ISBN Cross-Referencing

ISBN is the universal key that links data across all three APIs.

### Matching Strategy

```
1. Search Open Library → Get ISBNs for each result
2. When user views book details:
   a. Query Google Books by ISBN → Get ebook status
   b. Query Hardcover by ISBN → Get series/audiobook info
3. Store all cross-reference IDs in database
4. Future lookups use stored IDs (faster, no search needed)
```

### Handling Missing ISBNs

Some books (especially older ones) may lack ISBNs:
- Fall back to title + author matching
- Store Open Library Work ID as primary reference
- Mark as "ISBN unavailable" in UI

---

## Database Schema Additions

### Books Table

```sql
-- Cross-reference IDs
open_library_work_id VARCHAR(50)
open_library_edition_id VARCHAR(50)
google_volume_id VARCHAR(50)
-- hardcover_id already exists

-- Ebook/Audiobook detection
is_ebook BOOLEAN          -- From Google Books
is_audiobook BOOLEAN      -- From Hardcover
has_epub BOOLEAN          -- From Google Books
has_pdf BOOLEAN           -- From Google Books
buy_link TEXT             -- From Google Books
audio_duration INTEGER    -- From Hardcover (seconds)

-- Cache timestamps
ebook_checked_at TIMESTAMP
audiobook_checked_at TIMESTAMP
metadata_updated_at TIMESTAMP
```

### Authors Table

```sql
open_library_id VARCHAR(50)
birth_date DATE
death_date DATE
```

### Series Table

```sql
is_completed BOOLEAN
books_count INTEGER
```

---

## Backend Implementation

### Package Structure

```
backend/internal/
├── openlibrary/
│   ├── client.go      # Open Library API client
│   └── types.go       # Response types
├── googlebooks/
│   ├── client.go      # Google Books API client
│   └── types.go       # Response types
├── hardcover/
│   └── client.go      # (existing)
└── metadata/
    └── aggregator.go  # Combines data from all APIs
```

### Aggregator Pattern

```go
type MetadataAggregator struct {
    openLibrary  *openlibrary.Client
    googleBooks  *googlebooks.Client
    hardcover    *hardcover.Client
    db           *gorm.DB
}

func (a *MetadataAggregator) GetBookByISBN(isbn string) (*AggregatedBook, error) {
    // 1. Get base metadata from Open Library
    // 2. Get ebook status from Google Books (if quota available)
    // 3. Get series/audiobook from Hardcover
    // 4. Merge and return
}
```

---

## Settings Configuration

### Required Settings

| Setting Key | Description | Required |
|-------------|-------------|----------|
| `google_books_api_key` | Google Books API key | Yes (for ebook detection) |
| `hardcover_api_key` | Hardcover API token | Yes (for series/audiobooks) |

### Optional Settings

| Setting Key | Description | Default |
|-------------|-------------|---------|
| `preferred_languages` | ISO 639-1 codes | `en` |
| `cache_duration_hours` | Metadata cache duration | `168` (7 days) |

---

## Error Handling

### API Failures

| Scenario | Behavior |
|----------|----------|
| Open Library down | Show error, cannot search |
| Google Books quota exhausted | Continue without ebook status |
| Google Books down | Continue without ebook status |
| Hardcover down | Continue without series/audiobook |
| All APIs down | Show cached data if available |

### User-Facing Messages

| Condition | Message |
|-----------|---------|
| Google quota exhausted | "Ebook availability check unavailable. Resets tomorrow." |
| No ebook found | "Digital edition not found. Physical copy may be available." |
| No ISBN | "ISBN unavailable. Some features may be limited." |

---

## Migration Plan

### Phase 1: Backend Clients
1. Create `openlibrary` package
2. Create `googlebooks` package
3. Create `metadata` aggregator

### Phase 2: Database Schema
1. Add new columns via GORM auto-migration
2. Existing data retains `hardcover_id`

### Phase 3: API Handlers
1. Update search to use Open Library
2. Keep Hardcover endpoints for series/audiobook
3. Add Google Books ebook check endpoint

### Phase 4: Frontend
1. Update types for new fields
2. Update book detail to show ebook/audiobook status
3. Add quota exhausted warning display

---

## Quick Reference

### Which API for What?

| Need | Use | Endpoint |
|------|-----|----------|
| Search books | Open Library | `GET /search.json?q=` |
| Book by ISBN | Open Library | `GET /api/books?bibkeys=ISBN:` |
| Check if ebook | Google Books | `GET /volumes?q=isbn:` |
| Get series info | Hardcover | `series_by_pk(id)` |
| Check audiobook | Hardcover | `editions.reading_format` |
| Author details | Open Library | `GET /authors/{OLID}.json` |
| Author's books | Open Library | `GET /authors/{OLID}/works.json` |
| Cover image | Open Library | `covers.openlibrary.org/b/id/{id}-M.jpg` |

---

## Related Documentation

- `.documents/Open_Library_API_Overview.md` - Full Open Library API reference
- `.documents/Google_Books_API_Overview.md` - Full Google Books API reference
- `.documents/Hardcover_API_Overview.md` - Full Hardcover API reference
- `.documents/Book_APIs_Limitations_Comparison.md` - Rate limits comparison
