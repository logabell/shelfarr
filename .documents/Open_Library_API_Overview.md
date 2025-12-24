# Open Library API - Complete Overview Guide

## Table of Contents
1. [Introduction](#introduction)
2. [API Basics & Authentication](#api-basics--authentication)
3. [Available APIs](#available-apis)
4. [Book Search API](#book-search-api)
5. [Books API](#books-api)
6. [Works & Editions API](#works--editions-api)
7. [Authors API](#authors-api)
8. [Subjects API](#subjects-api)
9. [Covers API](#covers-api)
10. [My Books API](#my-books-api)
11. [Partner/Read API](#partnerread-api)
12. [Lists API](#lists-api)
13. [Rate Limits & Limitations](#rate-limits--limitations)
14. [Error Handling](#error-handling)
15. [Implementation Guide](#implementation-guide)
16. [Code Examples](#code-examples)

---

## Introduction

Open Library is a project of the Internet Archive, designed to create a web page for every book ever published. The Open Library API provides free, public access to:

- Over 20 million book records
- Author information and bibliographies
- Book covers and author photos
- Subject classifications
- Reading availability (borrowable/readable books)
- User reading lists and bookshelves

The API is **RESTful** and returns data in **JSON**, **YAML**, and **RDF/XML** formats.

> **Important:** Open Library is a free, community resource. Please use the API responsibly and avoid bulk downloads.

---

## API Basics & Authentication

### Base URLs

| Service | Base URL |
|---------|----------|
| Main API | `https://openlibrary.org` |
| Covers API | `https://covers.openlibrary.org` |
| Search API | `https://openlibrary.org/search.json` |

### Authentication

**Most endpoints require NO authentication** for read operations.

For write operations (editing/adding books), you need:
1. An Open Library account
2. S3 keys (found in account settings)

```bash
# Login example
curl -i -H 'Content-Type: application/json' \
  -d '{"access": "your_access_key", "secret": "your_secret_key"}' \
  https://openlibrary.org/account/login
```

### Required Headers

| Header | Value | Required |
|--------|-------|----------|
| `User-Agent` | `YourAppName/1.0 (contact@email.com)` | **Yes** (for frequent use) |
| `Accept` | `application/json` | Optional |
| `Content-Type` | `application/json` | For POST/PUT requests |

> **Critical:** If making multiple calls per minute, you MUST include a User-Agent header with your app name and contact info, or your application may be blocked.

### Response Formats

Add these extensions to any Open Library URL:
- `.json` - JSON format
- `.yml` - YAML format  
- `.rdf` - RDF/XML format

```
https://openlibrary.org/works/OL45804W.json
https://openlibrary.org/authors/OL23919A.json
https://openlibrary.org/books/OL7353617M.json
```

---

## Available APIs

| API | Purpose | Endpoint Pattern |
|-----|---------|------------------|
| **Search** | Full-text search for books, authors, subjects | `/search.json` |
| **Books** | Fetch book data by ISBN, LCCN, OCLC, OLID | `/api/books` |
| **Works** | Get work-level information | `/works/{OLID}.json` |
| **Editions** | Get edition-level information | `/books/{OLID}.json` |
| **Authors** | Author data and works | `/authors/{OLID}.json` |
| **Subjects** | Books by subject | `/subjects/{subject}.json` |
| **Covers** | Book cover images | `covers.openlibrary.org/b/{key}/{value}-{size}.jpg` |
| **My Books** | User reading lists | `/people/{username}/books/{shelf}.json` |
| **Partner/Read** | Readable/borrowable book info | `/api/volumes/brief/{id-type}/{id}.json` |
| **Lists** | User-created lists | `/people/{username}/lists.json` |
| **Recent Changes** | Track catalog changes | `/recentchanges.json` |

---

## Book Search API

The Search API is the most versatile way to find books.

### Endpoint

```
GET https://openlibrary.org/search.json
```

### Search Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `q` | General search query | `q=the+lord+of+the+rings` |
| `title` | Search by title | `title=harry+potter` |
| `author` | Search by author | `author=tolkien` |
| `subject` | Search by subject | `subject=fantasy` |
| `place` | Search by place | `place=london` |
| `person` | Search by person mentioned | `person=sherlock+holmes` |
| `publisher` | Search by publisher | `publisher=penguin` |
| `isbn` | Search by ISBN | `isbn=9780140328721` |
| `oclc` | Search by OCLC number | `oclc=297222669` |
| `lccn` | Search by LCCN | `lccn=93005405` |

### Pagination & Sorting

| Parameter | Description | Default |
|-----------|-------------|---------|
| `page` | Page number | 1 |
| `limit` | Results per page | 100 |
| `offset` | Starting position | 0 |
| `sort` | Sort order | relevance |

**Sort Options:**
- `new` - Newest first
- `old` - Oldest first
- `rating` - By rating
- `readinglog` - By reading log count
- `want_to_read` - By want-to-read count
- `currently_reading` - By currently-reading count
- `already_read` - By already-read count

### Field Selection

Use the `fields` parameter to limit returned data:

```
/search.json?q=crime&fields=key,title,author_name,first_publish_year
```

### Search Response Fields

| Field | Description |
|-------|-------------|
| `key` | Work key (e.g., `/works/OL27448W`) |
| `title` | Book title |
| `author_name` | Array of author names |
| `author_key` | Array of author OLIDs |
| `first_publish_year` | First publication year |
| `edition_count` | Number of editions |
| `cover_i` | Cover ID for Covers API |
| `isbn` | Array of ISBNs |
| `oclc` | Array of OCLC numbers |
| `lccn` | Array of LCCNs |
| `has_fulltext` | Boolean - readable version available |
| `ia` | Internet Archive identifiers |
| `public_scan_b` | Boolean - public domain scan |
| `language` | Array of language codes |
| `subject` | Array of subjects |
| `publisher` | Array of publishers |
| `number_of_pages_median` | Median page count |
| `ratings_average` | Average rating |
| `ratings_count` | Number of ratings |

### Including Editions

To get edition data with search results:

```
/search.json?q=crime+and+punishment&fields=key,title,author_name,editions
```

Specify edition fields:
```
/search.json?q=crime&fields=key,title,editions,editions.key,editions.title,editions.language
```

### Advanced Search Syntax

```
# Language filter
/search.json?q=sherlock+holmes+language:fre

# Ebook availability
/search.json?q=dickens&has_fulltext=true

# Combine filters
/search.json?title=pride+and+prejudice&author=austen&first_publish_year=1813
```

---

## Books API

The Books API fetches book data using standard identifiers.

### Endpoint

```
GET https://openlibrary.org/api/books
```

### Parameters

| Parameter | Description | Options |
|-----------|-------------|---------|
| `bibkeys` | Comma-separated identifiers | `ISBN:`, `OCLC:`, `LCCN:`, `OLID:` |
| `format` | Response format | `json`, `javascript` |
| `jscmd` | Data detail level | `viewapi`, `data`, `details` |
| `callback` | JSONP callback function | Any function name |

### Identifier Prefixes

| Prefix | Description | Example |
|--------|-------------|---------|
| `ISBN:` | ISBN-10 or ISBN-13 | `ISBN:0451526538` |
| `OCLC:` | OCLC Number | `OCLC:297222669` |
| `LCCN:` | Library of Congress Number | `LCCN:93005405` |
| `OLID:` | Open Library ID | `OLID:OL7353617M` |

### Response Modes (jscmd)

**`jscmd=viewapi`** (default) - Basic info:
```json
{
  "ISBN:0451526538": {
    "bib_key": "ISBN:0451526538",
    "info_url": "https://openlibrary.org/books/OL...",
    "preview": "noview",
    "preview_url": "https://openlibrary.org/books/OL...",
    "thumbnail_url": "https://covers.openlibrary.org/b/id/...-S.jpg"
  }
}
```

**`jscmd=data`** - Detailed data including:
- Publishers, identifiers, classifications
- Cover URLs (small, medium, large)
- Subjects, authors with URLs
- Number of pages, weight
- Table of contents, excerpts

**`jscmd=details`** - Full Open Library record

### Example Request

```bash
curl 'https://openlibrary.org/api/books?bibkeys=ISBN:9780980200447&jscmd=data&format=json'
```

---

## Works & Editions API

### Understanding Works vs Editions

- **Work**: A logical book (e.g., "Pride and Prejudice")
- **Edition**: A specific publication (e.g., Penguin Classics 2003 paperback)

One Work can have many Editions across languages, publishers, and formats.

### Works Endpoint

```
GET https://openlibrary.org/works/{OLID}.json
```

**Response includes:**
- `title` - Work title
- `authors` - Author references
- `description` - Book description
- `subjects` - Subject classifications
- `covers` - Cover IDs
- `first_publish_date` - Original publication date

### Get Work's Editions

```
GET https://openlibrary.org/works/{OLID}/editions.json?limit=10&offset=0
```

### Get Work's Ratings

```
GET https://openlibrary.org/works/{OLID}/ratings.json
```

### Get Work's Bookshelves

```
GET https://openlibrary.org/works/{OLID}/bookshelves.json
```

### Editions Endpoint

```
GET https://openlibrary.org/books/{OLID}.json
```

**Response includes:**
- `title` - Edition title
- `publishers` - Publisher names
- `publish_date` - Publication date
- `isbn_10`, `isbn_13` - ISBNs
- `number_of_pages` - Page count
- `covers` - Cover IDs
- `languages` - Language references
- `works` - Parent work reference

---

## Authors API

### Search Authors

```
GET https://openlibrary.org/search/authors.json?q={query}
```

**Response:**
```json
{
  "numFound": 1,
  "docs": [{
    "key": "OL23919A",
    "name": "J. K. Rowling",
    "birth_date": "31 July 1965",
    "top_work": "Harry Potter and the Philosopher's Stone",
    "work_count": 162,
    "top_subjects": ["Fiction", "Fantasy"]
  }]
}
```

### Get Author Details

```
GET https://openlibrary.org/authors/{OLID}.json
```

### Get Author's Works

```
GET https://openlibrary.org/authors/{OLID}/works.json?limit=50&offset=0
```

### Author Fields

| Field | Description |
|-------|-------------|
| `name` | Author name |
| `personal_name` | Personal/birth name |
| `alternate_names` | Other names used |
| `birth_date` | Birth date |
| `death_date` | Death date |
| `bio` | Biography text |
| `links` | Related URLs |
| `photos` | Photo IDs |

---

## Subjects API

### Endpoint

```
GET https://openlibrary.org/subjects/{subject}.json
```

### Subject Types

| URL Pattern | Type |
|-------------|------|
| `/subjects/{name}` | General subject |
| `/subjects/place:{name}` | Place |
| `/subjects/person:{name}` | Person |
| `/subjects/time:{name}` | Time period |

### Parameters

| Parameter | Description |
|-----------|-------------|
| `details` | Include related data (`true`/`false`) |
| `ebooks` | Only books with ebooks (`true`/`false`) |
| `published_in` | Year range (e.g., `1500-1600`) |
| `limit` | Number of works |
| `offset` | Starting offset |

### Response with `details=true`

Includes:
- `authors` - Prolific authors in subject
- `publishers` - Prominent publishers
- `subjects` - Related subjects
- `places`, `people`, `times` - Related classifications
- `publishing_history` - Publication counts by year

### Example

```bash
curl 'https://openlibrary.org/subjects/love.json?details=true&limit=10'
```

---

## Covers API

### Book Covers

```
https://covers.openlibrary.org/b/{key}/{value}-{size}.jpg
```

| Key | Description | Example |
|-----|-------------|---------|
| `isbn` | ISBN-10 or ISBN-13 | `/b/isbn/0385472579-M.jpg` |
| `oclc` | OCLC Number | `/b/oclc/297222669-M.jpg` |
| `lccn` | LCCN | `/b/lccn/93005405-M.jpg` |
| `olid` | Open Library ID | `/b/olid/OL7353617M-M.jpg` |
| `id` | Cover ID | `/b/id/240726-M.jpg` |

### Author Photos

```
https://covers.openlibrary.org/a/{key}/{value}-{size}.jpg
```

| Key | Description | Example |
|-----|-------------|---------|
| `olid` | Author OLID | `/a/olid/OL229501A-M.jpg` |
| `id` | Photo ID | `/a/id/123456-M.jpg` |

### Image Sizes

| Size | Code | Approximate Dimensions |
|------|------|------------------------|
| Small | `S` | ~75px width |
| Medium | `M` | ~180px width |
| Large | `L` | ~500px width |

### Cover Metadata

Add `.json` to get cover information:
```
https://covers.openlibrary.org/b/id/12547191.json
```

### Rate Limits for Covers

Covers accessed by ISBN, OCLC, LCCN (not Cover ID or OLID) are rate-limited:
- **100 requests per IP per 5 minutes**
- Exceeding returns `403 Forbidden`

---

## My Books API

Access user reading lists (if public or authenticated).

### Reading Log Endpoints

```
GET https://openlibrary.org/people/{username}/books/want-to-read.json
GET https://openlibrary.org/people/{username}/books/currently-reading.json
GET https://openlibrary.org/people/{username}/books/already-read.json
```

### Parameters

| Parameter | Description |
|-----------|-------------|
| `limit` | Results per page |
| `offset` | Starting position |

---

## Partner/Read API

Find readable or borrowable versions of books.

### Single Book Request

```
GET https://openlibrary.org/api/volumes/brief/{id-type}/{id-value}.json
```

**ID Types:** `isbn`, `lccn`, `oclc`, `olid`

### Multiple Books Request

```
GET https://openlibrary.org/api/volumes/brief/json/{bibkeys}
```

Where `bibkeys` is a semicolon-separated list like:
```
isbn:0596156715;lccn:93005405;oclc:297222669
```

### Response Fields

| Field | Values |
|-------|--------|
| `match` | `exact` or `similar` |
| `status` | `full access`, `lendable`, `checked out`, `restricted` |
| `itemURL` | Link to read/borrow |
| `cover` | Cover image URLs |

---

## Lists API

### Get User's Lists

```
GET https://openlibrary.org/people/{username}/lists.json
```

### Get Specific List

```
GET https://openlibrary.org/people/{username}/lists/{list_olid}.json
```

### List Seeds (Contents)

```
GET https://openlibrary.org/people/{username}/lists/{list_olid}/seeds.json
```

---

## Rate Limits & Limitations

### Rate Limiting

| Aspect | Limit |
|--------|-------|
| General API | No strict limit, but must include User-Agent for frequent use |
| Covers (by ISBN/OCLC/LCCN) | **100 requests per IP per 5 minutes** |
| Query results default | 100 items |
| Query results maximum | **1,000 items** |
| Query offset maximum | **10,000** |
| History entries default | 20 items |
| History entries maximum | **1,000 items** |

### Important Restrictions

| Restriction | Details |
|-------------|---------|
| **No bulk downloads** | Use monthly data dumps instead |
| **User-Agent required** | Required for frequent API use (multiple calls/minute) |
| **Rate limiting** | Apps may be blocked without User-Agent |
| **Schema stability** | Search schema not guaranteed stable |
| **Write operations** | Require authentication with S3 keys |

### Bulk Data Access

For large-scale data needs, download monthly dumps:
- **Books dump:** Complete edition records
- **Authors dump:** Complete author records  
- **Covers dump:** Cover images in bulk
- Email: openlibrary@archive.org

---

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| **200** | Success |
| **400** | Bad Request - Invalid parameters |
| **403** | Forbidden - Rate limited or blocked |
| **404** | Not Found - Resource doesn't exist |
| **500** | Internal Server Error |

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| 403 on Covers | Rate limit exceeded | Wait 5 minutes, use Cover ID instead |
| Empty results | Wrong identifier format | Check ISBN/OLID format |
| Missing fields | Data not available | Handle missing fields gracefully |
| Blocked requests | Missing User-Agent | Add User-Agent header |

---

## Implementation Guide

### Step 1: Plan Your Data Needs

1. Determine what data you need (search, specific books, covers)
2. Choose appropriate API endpoint
3. Plan for missing data handling

### Step 2: Set Up Proper Headers

```javascript
const headers = {
  'User-Agent': 'MyBookApp/1.0 (myemail@example.com)',
  'Accept': 'application/json'
};
```

### Step 3: Best Practices

| Practice | Recommendation |
|----------|----------------|
| **Caching** | Cache responses to reduce API load |
| **User-Agent** | Always include with app name and contact |
| **Error handling** | Handle 403, 404, 500 gracefully |
| **Rate limiting** | Implement delays between requests |
| **Field selection** | Only request fields you need |
| **Bulk data** | Use data dumps for large imports |

### Step 4: Data Flow

1. **Search** to find books → get Work/Edition OLIDs
2. **Fetch details** using OLIDs for full data
3. **Get covers** using Cover ID or OLID (not ISBN for high volume)
4. **Cache results** to minimize repeat requests

---

## Code Examples

### JavaScript - Search for Books

```javascript
const USER_AGENT = 'MyBookApp/1.0 (contact@example.com)';

async function searchBooks(query, limit = 10) {
  const url = new URL('https://openlibrary.org/search.json');
  url.searchParams.set('q', query);
  url.searchParams.set('limit', limit);
  url.searchParams.set('fields', 'key,title,author_name,first_publish_year,cover_i,isbn');

  const response = await fetch(url, {
    headers: { 'User-Agent': USER_AGENT }
  });

  if (!response.ok) {
    throw new Error(`Search failed: ${response.status}`);
  }

  const data = await response.json();
  return data.docs;
}

// Usage
searchBooks('lord of the rings').then(books => {
  books.forEach(book => {
    const coverUrl = book.cover_i 
      ? `https://covers.openlibrary.org/b/id/${book.cover_i}-M.jpg`
      : null;
    console.log(`${book.title} (${book.first_publish_year}) - ${book.author_name?.join(', ')}`);
  });
});
```

### JavaScript - Get Book by ISBN

```javascript
async function getBookByISBN(isbn) {
  const url = `https://openlibrary.org/api/books?bibkeys=ISBN:${isbn}&jscmd=data&format=json`;
  
  const response = await fetch(url, {
    headers: { 'User-Agent': USER_AGENT }
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch book: ${response.status}`);
  }

  const data = await response.json();
  return data[`ISBN:${isbn}`];
}

// Usage
getBookByISBN('9780140328721').then(book => {
  if (book) {
    console.log(`Title: ${book.title}`);
    console.log(`Authors: ${book.authors?.map(a => a.name).join(', ')}`);
    console.log(`Pages: ${book.number_of_pages}`);
    console.log(`Cover: ${book.cover?.medium}`);
  }
});
```

### JavaScript - Get Author and Works

```javascript
async function getAuthorWithWorks(authorOLID, worksLimit = 10) {
  // Get author details
  const authorResponse = await fetch(
    `https://openlibrary.org/authors/${authorOLID}.json`,
    { headers: { 'User-Agent': USER_AGENT } }
  );
  const author = await authorResponse.json();

  // Get author's works
  const worksResponse = await fetch(
    `https://openlibrary.org/authors/${authorOLID}/works.json?limit=${worksLimit}`,
    { headers: { 'User-Agent': USER_AGENT } }
  );
  const worksData = await worksResponse.json();

  return {
    name: author.name,
    bio: author.bio,
    birthDate: author.birth_date,
    works: worksData.entries
  };
}

// Usage
getAuthorWithWorks('OL23919A').then(data => {
  console.log(`Author: ${data.name}`);
  console.log(`Works: ${data.works.length}`);
  data.works.forEach(work => console.log(`  - ${work.title}`));
});
```

### Python Example

```python
import requests

USER_AGENT = 'MyBookApp/1.0 (contact@example.com)'
HEADERS = {'User-Agent': USER_AGENT}

def search_books(query: str, limit: int = 10) -> list:
    """Search for books by query string."""
    url = 'https://openlibrary.org/search.json'
    params = {
        'q': query,
        'limit': limit,
        'fields': 'key,title,author_name,first_publish_year,cover_i,isbn'
    }
    
    response = requests.get(url, params=params, headers=HEADERS)
    response.raise_for_status()
    
    return response.json().get('docs', [])


def get_book_by_isbn(isbn: str) -> dict:
    """Get book details by ISBN."""
    url = f'https://openlibrary.org/api/books'
    params = {
        'bibkeys': f'ISBN:{isbn}',
        'jscmd': 'data',
        'format': 'json'
    }
    
    response = requests.get(url, params=params, headers=HEADERS)
    response.raise_for_status()
    
    data = response.json()
    return data.get(f'ISBN:{isbn}')


def get_work_details(work_olid: str) -> dict:
    """Get work details by OLID."""
    url = f'https://openlibrary.org/works/{work_olid}.json'
    
    response = requests.get(url, headers=HEADERS)
    response.raise_for_status()
    
    return response.json()


def get_cover_url(cover_id: int, size: str = 'M') -> str:
    """Generate cover URL from cover ID."""
    return f'https://covers.openlibrary.org/b/id/{cover_id}-{size}.jpg'


def get_subject_books(subject: str, limit: int = 10) -> dict:
    """Get books by subject."""
    url = f'https://openlibrary.org/subjects/{subject}.json'
    params = {'limit': limit}
    
    response = requests.get(url, params=params, headers=HEADERS)
    response.raise_for_status()
    
    return response.json()


# Usage examples
if __name__ == '__main__':
    # Search for books
    results = search_books('python programming', limit=5)
    for book in results:
        print(f"{book['title']} by {', '.join(book.get('author_name', ['Unknown']))}")
    
    # Get book by ISBN
    book = get_book_by_isbn('9780596158064')
    if book:
        print(f"\nBook: {book['title']}")
        print(f"Publisher: {book['publishers'][0]['name']}")
    
    # Get books by subject
    fantasy_books = get_subject_books('fantasy', limit=5)
    print(f"\nFantasy books ({fantasy_books['work_count']} total):")
    for work in fantasy_books['works']:
        print(f"  - {work['title']}")
```

### cURL Examples

```bash
# Search for books
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/search.json?q=harry+potter&limit=5"

# Get book by ISBN
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/api/books?bibkeys=ISBN:9780140328721&jscmd=data&format=json"

# Get work details
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/works/OL82563W.json"

# Get author details
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/authors/OL23919A.json"

# Get author's works
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/authors/OL23919A/works.json?limit=10"

# Get subject books
curl -H "User-Agent: MyApp/1.0 (email@example.com)" \
  "https://openlibrary.org/subjects/science_fiction.json?limit=10"

# Get book cover (small)
curl -o cover.jpg "https://covers.openlibrary.org/b/isbn/9780140328721-S.jpg"
```

---

## Additional Resources

- **Documentation:** [openlibrary.org/developers](https://openlibrary.org/developers)
- **API Reference:** [openlibrary.org/dev/docs/api](https://openlibrary.org/dev/docs/api)
- **GitHub:** [github.com/internetarchive/openlibrary](https://github.com/internetarchive/openlibrary)
- **Bulk Downloads:** [openlibrary.org/developers/dumps](https://openlibrary.org/developers/dumps)
- **Python Client:** [github.com/internetarchive/openlibrary-client](https://github.com/internetarchive/openlibrary-client)

---

## Key Concepts Summary

| Concept | Description |
|---------|-------------|
| **OLID** | Open Library ID (e.g., OL7353617M for editions, OL45804W for works) |
| **Work** | Abstract book entity (one per unique book) |
| **Edition** | Specific publication of a work |
| **Cover ID** | Numeric ID for cover images |
| **Bibkey** | Identifier string like `ISBN:0451526538` |

---

*Document generated: December 2024*
*API Status: Production (Free/Public)*
*Data Source: Internet Archive / Open Library*
