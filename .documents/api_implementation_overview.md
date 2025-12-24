# Hardcover API Implementation Overview

> **This is a living document.** Update it whenever Hardcover API integration is added, modified, or removed from the codebase.

This document maps where the Hardcover.app API is used throughout the Shelfarr application. For API reference documentation (rate limits, query patterns, schema details), see [hardcover_api_documentation.md](./hardcover_api_documentation.md).

---

## Table of Contents

1. [Backend Implementation](#backend-implementation)
2. [Frontend Implementation](#frontend-implementation)
3. [Data Flow](#data-flow)
4. [Maintenance Instructions](#maintenance-instructions)

---

## Backend Implementation

### Core Client

| File | Description |
|------|-------------|
| `backend/internal/hardcover/client.go` | Core GraphQL client with rate limiting, authentication, and all API methods |

**Client Methods:**
- `SearchBooks(query, languages)` - Search for books
- `SearchAuthors(query)` - Search for authors
- `SearchSeries(query)` - Search for series
- `SearchLists(query)` - Search for user lists
- `SearchAll(query, languages)` - Unified search across all types
- `GetBook(id)` - Fetch detailed book information
- `GetAuthor(id)` - Fetch author details
- `GetBooksByAuthor(authorID, languages)` - Get all books by author
- `GetBooksByAuthorWithCounts(authorID, languages)` - Same with count metadata
- `GetSeries(seriesID, languages)` - Get series with all books
- `GetListBooks(listID)` - Get books from a Hardcover list
- `Test()` - Validate API connection

---

### API Routes

| Route | Method | Handler | File | Purpose |
|-------|--------|---------|------|---------|
| `/api/v1/search/hardcover` | GET | `searchHardcover` | `search.go` | Search books/authors/series/lists |
| `/api/v1/search/hardcover/test` | POST | `testHardcover` | `search.go` | Test API connection |
| `/api/v1/hardcover/book/:id` | GET | `getHardcoverBook` | `hardcover.go` | Get book details before adding |
| `/api/v1/hardcover/book/:id` | POST | `addHardcoverBook` | `hardcover.go` | Add book to library |
| `/api/v1/hardcover/author/:id` | GET | `getHardcoverAuthor` | `hardcover.go` | Get author with books |
| `/api/v1/hardcover/series/:id` | GET | `getHardcoverSeries` | `hardcover.go` | Get series with books |

---

### Handler Files

#### `backend/internal/api/hardcover.go`

Primary handlers for Hardcover detail pages and adding content to library.

| Handler | Client Methods | Purpose |
|---------|----------------|---------|
| `getHardcoverClient()` | - | Creates authenticated client from settings |
| `getHardcoverBook()` | `GetBook` | Fetch book details for preview page |
| `getHardcoverAuthor()` | `GetAuthor`, `GetBooksByAuthor` | Fetch author with filtered books |
| `getHardcoverSeries()` | `GetSeries` | Fetch series with filtered books |
| `addHardcoverBook()` | `GetBook` | Add book to library, create author/series if needed |

**Response Types Defined:**
- `HardcoverBookResponse`
- `HardcoverAuthorResponse`
- `HardcoverSeriesResponse`

---

#### `backend/internal/api/search.go`

Handlers for search functionality and connection testing.

| Handler | Client Methods | Purpose |
|---------|----------------|---------|
| `searchHardcover()` | Routes to type-specific handlers | Main search entry point |
| `searchHardcoverBooks()` | `SearchBooks` | Book search with library status |
| `searchHardcoverAuthors()` | `SearchAuthors` | Author search with library status |
| `searchHardcoverSeries()` | `SearchSeries` | Series search with library status |
| `searchHardcoverLists()` | `SearchLists` | List search |
| `searchHardcoverAll()` | `SearchAll` | Unified search across all types |
| `testHardcover()` | `Test` | Validate API key and connection |

**Response Types Defined:**
- `SearchResult`
- `AuthorSearchResult`
- `SeriesSearchResult`
- `ListSearchResult`
- `UnifiedSearchResponse`

---

#### `backend/internal/api/books.go`

Book management with Hardcover integration.

| Handler | Client Methods | Purpose |
|---------|----------------|---------|
| `addBook()` | `GetBook` | Add book to library from Hardcover ID |

---

#### `backend/internal/api/authors.go`

Author management with Hardcover integration.

| Handler | Client Methods | Purpose |
|---------|----------------|---------|
| `addAuthor()` | `GetAuthor`, `GetBooksByAuthorWithCounts` | Add author, optionally with all books |

---

#### `backend/internal/api/series.go`

Series management with Hardcover integration.

| Handler | Client Methods | Purpose |
|---------|----------------|---------|
| `addSeriesBooks()` | `GetBook` | Bulk add books from series to library |

---

## Frontend Implementation

### API Client

| File | Description |
|------|-------------|
| `frontend/src/api/client.ts` | Centralized API client with all Hardcover-related functions |

**Hardcover Functions (lines 154-257):**
- `searchHardcover(query, type)` - Search with type filter
- `searchHardcoverAuthors(query)` - Author-specific search
- `searchHardcoverSeries(query)` - Series-specific search
- `searchHardcoverLists(query)` - List-specific search
- `searchHardcoverAll(query)` - Unified search
- `testHardcoverConnection()` - Test API connection
- `getHardcoverBook(id)` - Get book details
- `getHardcoverAuthor(id)` - Get author with all books
- `getHardcoverSeries(id)` - Get series with all books
- `addHardcoverBook(id, options)` - Add book to library

**TypeScript Interfaces (lines 185-237):**
- `HardcoverBookDetail`
- `HardcoverAuthorDetail`
- `HardcoverSeriesDetail`

---

### Type Definitions

| File | Types |
|------|-------|
| `frontend/src/types/index.ts` | `SearchResult`, `AuthorSearchResult`, `SeriesSearchResult`, `ListSearchResult`, `UnifiedSearchResponse`, `SearchType` |

---

### Pages

| Page | Functions Used | Purpose |
|------|----------------|---------|
| `SearchPage.tsx` | `searchHardcover`, `searchHardcoverAll` | Search discovery with filtering |
| `HardcoverBookPage.tsx` | `getHardcoverBook`, `addHardcoverBook` | Book preview before adding |
| `HardcoverAuthorPage.tsx` | `getHardcoverAuthor`, `addHardcoverBook` | Author preview, add books |
| `HardcoverSeriesPage.tsx` | `getHardcoverSeries`, `addHardcoverBook` | Series preview, add books |
| `AuthorDetailPage.tsx` | `addHardcoverBook` | Add missing books from author |
| `SeriesDetailPage.tsx` | `addHardcoverBook` | Add missing books from series |
| `ManualImportPage.tsx` | `searchHardcover` | Search for manual import matching |
| `BookDetailPage.tsx` | - | Links to Hardcover author pages |
| `AuthorsPage.tsx` | - | Displays Hardcover book counts |
| `SeriesPage.tsx` | - | Displays Hardcover book counts |

---

### Components

| Component | Functions Used | Purpose |
|-----------|----------------|---------|
| `components/search/AddBookModal.tsx` | `getHardcoverBook`, `addHardcoverBook` | Modal for adding books |
| `components/series/AddSeriesModal.tsx` | - | Add series books (uses hardcover data) |
| `components/library/CatalogBookCard.tsx` | - | Links to Hardcover preview pages |

---

## Data Flow

### Search Flow
```
User enters query
    → SearchPage.tsx calls searchHardcover()
    → client.ts makes GET /api/v1/search/hardcover
    → search.go:searchHardcover() routes by type
    → hardcover/client.go:Search*() calls GraphQL API
    → Results enriched with library status
    → Response returned to frontend
```

### Add Book Flow
```
User clicks "Add to Library"
    → AddBookModal/Page calls addHardcoverBook()
    → client.ts makes POST /api/v1/hardcover/book/:id
    → hardcover.go:addHardcoverBook()
        → Fetches full book data from Hardcover
        → Creates Author if not exists
        → Creates Series if not exists
        → Creates Book record
    → Returns new book ID
```

### Preview Flow
```
User clicks book/author/series from search
    → HardcoverBookPage (etc.) loads
    → Calls getHardcoverBook() (etc.)
    → client.ts makes GET /api/v1/hardcover/book/:id
    → hardcover.go:getHardcoverBook()
        → Fetches from Hardcover API
        → Checks if already in library
    → Displays preview with "Add" option
```

---

## Format Availability

Hardcover's edition format data is incomplete - many books lack proper ebook/audiobook edition metadata even when digital editions exist. To avoid hiding valid books, Shelfarr displays **all books** regardless of detected format availability.

**UI Behavior:**
- All books are shown on author/series pages (no format-based filtering)
- Each book displays format availability badges:
  - ✅ = Format confirmed available
  - ❓ = Format availability unknown
- The `FormatAvailability` component (`frontend/src/components/ui/format-availability.tsx`) handles this display
- Format data is informational only, not used for filtering

**API Changes (December 2025):**
- Removed `includePhysicalOnly` parameter from backend client methods
- Removed `showPhysical` query parameter from frontend API calls
- Backend always returns all books; format flags (`hasEbook`, `hasAudiobook`) preserved for display

---

## Maintenance Instructions

### When to Update This Document

Update this document when you:

1. **Add a new Hardcover API endpoint** - Add to Routes table and Handler Files section
2. **Add a new client method** - Add to Core Client methods list
3. **Create a new page/component using Hardcover** - Add to Pages or Components table
4. **Add new TypeScript types** - Add to Type Definitions section
5. **Modify data flow significantly** - Update Data Flow diagrams
6. **Remove Hardcover functionality** - Remove from relevant sections

### How to Update

1. Identify which section(s) need updating
2. Add/modify the relevant table rows or descriptions
3. Ensure file paths are accurate
4. Update the "Last Updated" timestamp below

### Quick Reference Commands

```bash
# Find all backend files importing hardcover
grep -r "internal/hardcover" backend/internal/api/*.go

# Find all frontend hardcover API calls
grep -r "/hardcover/" frontend/src/**/*.ts frontend/src/**/*.tsx

# Find Hardcover route registrations
grep -n "hardcover" backend/internal/api/server.go
```

---

**Last Updated:** December 24, 2025
