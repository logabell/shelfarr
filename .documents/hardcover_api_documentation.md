# Hardcover.app GraphQL API Documentation

A comprehensive reference for the Hardcover.app GraphQL API, including schema details, limitations, and usage guidelines.

**Official Documentation**: [docs.hardcover.app](https://docs.hardcover.app)  
**Schema Reference**: [GitHub - hardcover-docs/schema.graphql](https://github.com/hardcoverapp/hardcover-docs/blob/main/schema.graphql)

---

## Table of Contents

1. [API Overview](#api-overview)
2. [Authentication](#authentication)
3. [Rate Limits](#rate-limits)
4. [Query Restrictions & Limitations](#query-restrictions--limitations)
5. [Core Types](#core-types)
6. [Search Query](#search-query)
7. [Data Queries](#data-queries)
8. [Input Types](#input-types)
9. [Relationships](#relationships)
10. [Filtering Patterns](#filtering-patterns)
11. [Error Handling](#error-handling)
12. [Testing the API](#testing-the-api)

---

## API Overview

### Endpoint

```
https://api.hardcover.app/v1/graphql
```

### Key Characteristics

- **GraphQL-based**: All interactions use GraphQL queries and mutations
- **Search-first discovery**: The `search()` query is the primary method for finding content; direct text filtering operators are disabled
- **Depth-limited**: Queries have a maximum nesting depth of 3 levels
- **Rate-limited**: 60 requests per minute maximum

---

## Authentication

### API Token

API tokens are obtained from your Hardcover.app account settings page.

**Header Format:**
```
Authorization: Bearer <your_api_token>
```

**Important Notes:**
- Tokens expire annually on January 1st
- Tokens must be kept secure and used server-side only
- Never expose tokens in client-side code or public repositories

### OAuth (Coming Soon)

OAuth support for external applications is planned for future implementation.

---

## Rate Limits

| Limit | Value |
|-------|-------|
| Requests per minute | **60** |
| Request timeout | **30 seconds** |

**Recommendation**: Implement a rate limiter (approximately 1 request per second) to stay comfortably within limits and avoid 429 errors.

---

## Query Restrictions & Limitations

### Disabled Query Operators

The following comparison operators are **disabled** and will cause query errors if used:

| Operator | Description |
|----------|-------------|
| `_like` | SQL LIKE pattern match |
| `_nlike` | NOT LIKE |
| `_ilike` | Case-insensitive LIKE |
| `_nilike` | Case-insensitive NOT LIKE |
| `_regex` | Regular expression match |
| `_nregex` | NOT regex |
| `_iregex` | Case-insensitive regex |
| `_niregex` | Case-insensitive NOT regex |
| `_similar` | SQL SIMILAR TO |
| `_nsimilar` | NOT SIMILAR TO |

**Alternative**: Use the `search()` query function for text-based discovery instead of these operators.

### Query Depth Limit

**Maximum query depth: 3 levels**

Nested relationships beyond 3 levels will fail.

**Valid Example (depth 3):**
```graphql
query {
  books {           # depth 1
    contributions { # depth 2
      author {      # depth 3
        name
      }
    }
  }
}
```

**Invalid Example (depth 4):**
```graphql
query {
  books {              # depth 1
    contributions {    # depth 2
      author {         # depth 3
        contributions { # depth 4 - FAILS
          book { title }
        }
      }
    }
  }
}
```

### Data Access Restrictions

Queries are limited to:
- Your own user data
- Public data
- User data of users you follow

You **cannot** query private data of users you don't follow.

---

## Core Types

### `books`

Represents a book in the Hardcover database.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `title` | `String` | Book title |
| `subtitle` | `String` | Book subtitle |
| `slug` | `String` | URL-friendly identifier |
| `description` | `String` | Book description/summary |
| `headline` | `String` | Short headline |
| `pages` | `Int` | Page count |
| `rating` | `numeric` | Average rating |
| `ratings_count` | `Int!` | Number of ratings |
| `reviews_count` | `Int!` | Number of reviews |
| `release_date` | `date` | Publication date |
| `release_year` | `Int` | Publication year |
| `audio_seconds` | `Int` | Audiobook duration |
| `book_category_id` | `Int!` | Category ID |
| `literary_type_id` | `Int` | Literary type |
| `compilation` | `Boolean!` | Is anthology/collection |
| `locked` | `Boolean!` | Editing locked |
| `state` | `String` | Book state |
| `created_at` | `timestamp!` | Creation timestamp |
| `updated_at` | `timestamptz` | Last update |

**Cached Fields** (denormalized for performance):
- `cached_image`: JSON with image URLs
- `cached_contributors`: JSON array of author info
- `cached_tags`: JSON array of tags
- `cached_featured_series`: JSON with series info

**Relationships**:
- `contributions` → `[contributions!]!` (authors)
- `editions` → `[editions!]!` (book editions)
- `book_series` → `[book_series!]!` (series membership)
- `image` → `images` (cover image)

---

### `authors`

Represents an author.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `name` | `String!` | Author name |
| `name_personal` | `String` | Personal name |
| `slug` | `String` | URL-friendly identifier |
| `bio` | `String` | Biography |
| `title` | `String` | Title/honorific |
| `location` | `String` | Location |
| `born_date` | `date` | Birth date |
| `born_year` | `Int` | Birth year |
| `death_date` | `date` | Death date |
| `death_year` | `Int` | Death year |
| `gender_id` | `Int` | Gender ID |
| `is_bipoc` | `Boolean` | BIPOC author |
| `is_lgbtq` | `Boolean` | LGBTQ+ author |
| `books_count` | `Int!` | Number of books |
| `users_count` | `Int!` | Followers count |
| `locked` | `Boolean!` | Editing locked |
| `state` | `String!` | Author state |

**Cached Fields**:
- `cached_image`: JSON with image URLs
- `alternate_names`: JSON array of aliases

**Relationships**:
- `contributions` → `[contributions!]!` (book contributions)
- `image` → `images` (author photo)
- `canonical` → `authors` (canonical author if duplicate)

---

### `series`

Represents a book series.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `name` | `String!` | Series name |
| `slug` | `String!` | URL-friendly identifier |
| `description` | `String` | Series description |
| `books_count` | `Int!` | Total books in series |
| `primary_books_count` | `Int` | Primary entries count |
| `is_completed` | `Boolean` | Series completed |
| `author_id` | `Int` | Primary author ID |
| `locked` | `Boolean!` | Editing locked |
| `state` | `String!` | Series state |

**Relationships**:
- `book_series` → `[book_series!]!` (books in series with positions)
- `author` → `authors` (primary author)

---

### `editions`

Represents a specific edition of a book.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `book_id` | `Int!` | Parent book ID |
| `title` | `String` | Edition title |
| `subtitle` | `String` | Edition subtitle |
| `isbn_10` | `String` | ISBN-10 |
| `isbn_13` | `String` | ISBN-13 |
| `asin` | `String` | Amazon ASIN |
| `pages` | `Int` | Page count |
| `audio_seconds` | `Int` | Audiobook duration in seconds |
| `release_date` | `date` | Edition release date |
| `edition_format` | `String` | Format string (free-text, e.g., "Hardcover", "Kindle") |
| `reading_format_id` | `Int` | Reading format ID |
| `language_id` | `Int` | Language ID |
| `publisher_id` | `Int` | Publisher ID |

**Relationships**:
- `book` → `books!` (parent book)
- `image` → `images` (cover image)
- `language` → `languages` (edition language)
- `reading_format` → `reading_formats` (format classification)
- `publisher` → `publishers` (publisher info)

---

### `reading_formats`

Represents the reading format classification for editions.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `format` | `String!` | Format type: "Physical", "Ebook", or "Audiobook" |

**Note**: Use `reading_format.format` instead of `edition_format` for reliable format filtering, as it provides consistent enumerated values rather than free-text.

---

### `languages`

Represents a language for edition classification.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `code2` | `String` | ISO 639-1 code (e.g., "en", "es", "fr") |
| `code3` | `String` | ISO 639-2 code (e.g., "eng", "spa", "fra") |
| `language` | `String!` | Full language name (e.g., "English", "Spanish") |

**Usage**: Editions reference languages via `language_id`. Query pattern:
```graphql
editions {
  language { code2 language }
}
```

---

### `lists`

Represents a user-created reading list.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `name` | `String!` | List name |
| `description` | `String` | List description |
| `default_view` | `String!` | Default display view |
| `books_count` | `Int!` | Number of books |
| `likes_count` | `Int!` | Number of likes |
| `followers_count` | `Int` | Number of followers |
| `featured` | `Boolean!` | Is featured |
| `featured_profile` | `Boolean!` | Show on profile |
| `ranked` | `Boolean!` | Is ranked list |
| `imported` | `Boolean!` | Was imported |
| `created_at` | `timestamp` | Creation timestamp |

**Relationships**:
- `list_books` → `[list_books!]!` (books in list)
- `user` → `users` (list owner)
- `followed_lists` → `[followed_lists!]!` (followers)

---

### `users`

Represents a Hardcover user.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `Int!` | Primary identifier |
| `username` | `String` | Username |
| `name` | `String` | Display name |
| `bio` | `String` | Biography |
| `pro` | `Boolean!` | Pro subscriber |
| `librarian` | `Boolean!` | Is librarian |
| `flair` | `String` | User flair |
| `books_count` | `Int!` | Library size |
| `followers_count` | `Int` | Follower count |
| `following_count` | `Int` | Following count |

**Relationships**:
- `user_books` → `[user_books!]!` (user's library)
- `lists` → `[lists!]!` (user's lists)
- `followers` / `following` → `[followed_users!]!`

---

## Search Query

The `search()` query is the **primary method for discovering content**. Direct text filtering with operators like `_ilike` is disabled.

### Syntax

```graphql
search(
  query: String!           # Search text (required)
  query_type: String       # "Book", "Author", "Series", "List", "User"
  page: Int                # Page number (default: 1)
  per_page: Int            # Results per page (default: 20)
  sort: String             # Sort option
  fields: String           # Fields to search
  weights: String          # Field weights
): SearchOutput
```

### SearchOutput

```graphql
type SearchOutput {
  error: String         # Error message if any
  ids: [Int]            # Array of result IDs
  page: Int             # Current page
  per_page: Int         # Results per page
  query: String         # Original query
  query_type: String    # Search type used
  results: jsonb        # Full search results (JSON)
}
```

### Results Structure

The `results` field returns a JSON object with this structure:

```json
{
  "found": 42,
  "hits": [
    {
      "document": {
        "id": 12345,
        "title": "Book Title",
        "default_title": "Book Title",
        "subtitle": "Subtitle",
        "slug": "book-title",
        "author_names": ["Author Name"],
        "release_year": 2023,
        "ratings_average": 4.2,
        "cached_image": { "url": "https://..." },
        "contributions": [
          {
            "author": {
              "id": 789,
              "name": "Author Name"
            }
          }
        ]
      }
    }
  ]
}
```

### Supported `query_type` Values

| Value | Description |
|-------|-------------|
| `Book` | Search for books |
| `Author` | Search for authors |
| `Series` | Search for series |
| `List` | Search for user lists |
| `User` | Search for users |

### Search Examples

**Search for books:**
```graphql
query SearchBooks($query: String!) {
  search(query: $query, query_type: "Book", per_page: 20, page: 1) {
    results
    ids
    error
  }
}
```

**Search for authors:**
```graphql
query SearchAuthors($query: String!) {
  search(query: $query, query_type: "Author", per_page: 20, page: 1) {
    results
    ids
    error
  }
}
```

**Search for series:**
```graphql
query SearchSeries($query: String!) {
  search(query: $query, query_type: "Series", per_page: 20, page: 1) {
    results
    ids
    error
  }
}
```

---

## Data Queries

### Single Entity Queries (by Primary Key)

```graphql
# Get book by ID
books_by_pk(id: Int!): books

# Get author by ID  
authors_by_pk(id: Int!): authors

# Get series by ID
series_by_pk(id: Int!): series

# Get list by ID
lists_by_pk(id: Int!): lists

# Get user by ID
users_by_pk(id: Int!): users

# Get edition by ID
editions_by_pk(id: Int!): editions
```

### Collection Queries

```graphql
# Get books with filtering
books(
  where: books_bool_exp
  order_by: [books_order_by!]
  limit: Int
  offset: Int
): [books!]!

# Get authors with filtering
authors(
  where: authors_bool_exp
  order_by: [authors_order_by!]
  limit: Int
  offset: Int
): [authors!]!

# Get series with filtering
series(
  where: series_bool_exp
  order_by: [series_order_by!]
  limit: Int
  offset: Int
): [series!]!

# Get lists with filtering
lists(
  where: lists_bool_exp
  order_by: [lists_order_by!]
  limit: Int
  offset: Int
): [lists!]!
```

### Current User Query

```graphql
# Get authenticated user data
me: users
```

### Recommended Pattern: Search Then Fetch

Since text-based filtering operators are disabled, the recommended approach is:

1. Use `search()` to find entities by text query
2. Extract IDs from search results
3. Use `_by_pk` queries to fetch full details

**Example:**
```graphql
# Step 1: Search
query {
  search(query: "The Name of the Wind", query_type: "Book", per_page: 5) {
    ids
    results
  }
}

# Step 2: Fetch details for a specific result
query {
  books_by_pk(id: 12345) {
    id
    title
    description
    contributions {
      author { name }
    }
    editions {
      isbn_13
      edition_format
      reading_format { format }
      language { code2 language }
    }
  }
}
```

---

## Input Types

### Book Creation/Update

```graphql
input BookInput {
  book_status_id: Int
  canonical_id: Int
  default_audio_edition_id: Int
  default_cover_edition_id: Int
  default_ebook_edition_id: Int
  default_physical_edition_id: Int
  dto: BookDtoType
  locked: Boolean
  slug: String
  user_added: Boolean
}

input BookDtoType {
  book_category_id: Int
  characters: [CharacterDtoInput]
  collection_book_ids: [Int]
  compilation: Boolean
  description: String
  headline: String
  librarian_tags: [TagsDtoInput]
  literary_type_id: Int
  series: [BookSeriesDtoInput]
  title: String
}

input BookDtoInput {
  asin: String
  audio_seconds: Int
  contributions: [ContributionInputType]
  country_id: Int
  edition_format: String
  isbn_10: String
  isbn_13: String
  language_id: Int
  page_count: Int
  publisher_id: Int
  release_date: date
  subtitle: String
  title: String
}
```

### Author Creation/Update

```graphql
input AuthorInputType {
  alias_id: Int
  bio: String
  born_date: date
  born_year: Int
  death_date: date
  death_year: Int
  gender_id: Int
  id: Int
  image_id: Int
  is_bipoc: Boolean
  is_lgbtq: Boolean
  location: String
  locked: Boolean
  name: String
  name_personal: String
  slug: String
}
```

### List Management

```graphql
input ListInput {
  default_view: String
  description: String
  featured_profile: Boolean
  name: String
  privacy_setting_id: Int
  ranked: Boolean
  url: String
}

input ListBookInput {
  book_id: Int!
  edition_id: Int
  list_id: Int!
  position: Int
}
```

### Series Input

```graphql
input BookSeriesDtoInput {
  details: String
  featured: Boolean
  position: numeric
  series_id: Int
}
```

### Contribution Input

```graphql
input ContributionInputType {
  author_id: Int
  contribution: String    # Role: "Author", "Narrator", "Editor", etc.
  position: Int
}
```

---

## Relationships

### Key Relationship Types

#### `contributions` (Book-Author Connection)

```graphql
type contributions {
  id: Int!
  book_id: Int!
  author_id: Int!
  contribution: String  # Role: "Author", "Narrator", "Editor", etc.
  position: Int
  book: books!
  author: authors!
}
```

#### `book_series` (Book-Series Connection)

```graphql
type book_series {
  id: Int!
  book_id: Int!
  series_id: Int!
  position: numeric    # Book position in series (e.g., 1, 2, 2.5)
  featured: Boolean!
  details: String
  book: books!
  series: series!
}
```

#### `user_books` (User Library)

```graphql
type user_books {
  id: bigint!
  user_id: Int!
  book_id: Int!
  status_id: Int!      # Status: Want to Read, Reading, Read, DNF
  rating: numeric
  review: String
  edition_id: Int
  book: books!
  user: users!
}
```

**Status IDs:**
| ID | Status |
|----|--------|
| 1 | Want to Read |
| 2 | Currently Reading |
| 3 | Read |
| 4 | Did Not Finish (DNF) |

#### `list_books` (List-Book Connection)

```graphql
type list_books {
  id: bigint!
  list_id: Int!
  book_id: Int!
  edition_id: Int
  position: Int
  book: books!
  list: lists!
}
```

#### `followed_users` (User Following)

```graphql
type followed_users {
  id: bigint!
  user_id: Int!
  followed_user_id: Int!
  user: users!
  followed_user: users!
}
```

#### `followed_lists` (List Following)

```graphql
type followed_lists {
  id: bigint!
  user_id: Int!
  list_id: Int!
  user: users!
  list: lists!
}
```

---

## Filtering Patterns

### Language Filtering

Filter books by edition language using the `language` relationship:

```graphql
query GetBookWithLanguages {
  books_by_pk(id: 12345) {
    title
    editions {
      isbn_13
      edition_format
      language {
        code2      # "en", "es", "fr", etc.
        language   # "English", "Spanish", "French", etc.
      }
    }
  }
}
```

**Common Language Codes:**
| Code | Language |
|------|----------|
| `en` | English |
| `es` | Spanish |
| `fr` | French |
| `de` | German |
| `it` | Italian |
| `pt` | Portuguese |
| `ja` | Japanese |
| `zh` | Chinese |

### Format Filtering

Filter books by reading format using the `reading_format` relationship:

```graphql
query GetBookWithFormats {
  books_by_pk(id: 12345) {
    title
    editions {
      isbn_13
      asin
      reading_format {
        format    # "Physical", "Ebook", or "Audiobook"
      }
    }
  }
}
```

**Format Values:**
| Format | Description |
|--------|-------------|
| `Physical` | Print editions (hardcover, paperback, etc.) |
| `Ebook` | Digital reading editions (Kindle, EPUB, etc.) |
| `Audiobook` | Audio editions |

**Why use `reading_format.format` over `edition_format`:**

| Field | Type | Values | Reliability |
|-------|------|--------|-------------|
| `edition_format` | String | Free-text (e.g., "Kindle", "Hardback", "Mass Market") | Inconsistent |
| `reading_format.format` | Enum | "Physical", "Ebook", "Audiobook" | Consistent |

### Combined Edition Query

```graphql
query GetBookEditions {
  books_by_pk(id: 12345) {
    title
    editions {
      id
      isbn_10
      isbn_13
      asin
      pages
      audio_seconds
      release_date
      edition_format
      reading_format { format }
      language { code2 language }
      publisher { name }
      image { url }
    }
  }
}
```

### Depth Limit Workarounds

For queries that would exceed depth 3 (e.g., `series → book_series → book → editions → reading_format`):

1. **Fetch in stages**: Get parent entity first, then fetch child details separately
2. **Use cached fields**: Leverage `cached_image`, `cached_contributors` to avoid joins
3. **Filter client-side**: Fetch more data than needed and filter in your application

**Example: Series with Edition Details**
```graphql
# Query 1: Get series and book IDs
query GetSeriesBooks {
  series_by_pk(id: 123) {
    name
    book_series {
      position
      book {
        id
        title
        cached_image
      }
    }
  }
}

# Query 2: Get edition details for specific books
query GetBookEditions($bookId: Int!) {
  books_by_pk(id: $bookId) {
    editions {
      reading_format { format }
      language { code2 }
    }
  }
}
```

---

## Error Handling

### HTTP Status Codes

| Status | Meaning | Action |
|--------|---------|--------|
| 200 | Success | Process response |
| 401 | Unauthorized | Check/refresh API token |
| 429 | Rate limited | Wait and retry with exponential backoff |
| 400 | Bad request | Check query syntax, look for disabled operators |
| 504 | Gateway timeout | Query too complex; simplify or paginate |

### GraphQL Error Response Format

```json
{
  "errors": [
    {
      "message": "field \"_ilike\" is disabled",
      "extensions": {
        "code": "validation-failed",
        "path": "$.selectionSet.books.args.where.title._ilike"
      }
    }
  ],
  "data": null
}
```

### Common Errors

**Disabled Operator:**
```json
{
  "message": "field \"_ilike\" is disabled"
}
```
*Solution*: Use `search()` query instead.

**Depth Exceeded:**
```json
{
  "message": "max query depth exceeded"
}
```
*Solution*: Reduce nesting or split into multiple queries.

**Invalid Token:**
```json
{
  "message": "Invalid authorization header"
}
```
*Solution*: Verify token format and expiration.

### Token Expiration Handling

Tokens expire annually on January 1st. Monitor for 401 responses and prompt users to refresh their API token from account settings.

---

## Testing the API

### Validate API Connection

```graphql
query TestConnection {
  me {
    id
    username
    name
  }
}
```

### Sample Book Search

```graphql
query SearchBooks {
  search(query: "The Name of the Wind", query_type: "Book", per_page: 5, page: 1) {
    results
    ids
    error
  }
}
```

### Sample Author Search

```graphql
query SearchAuthors {
  search(query: "Patrick Rothfuss", query_type: "Author", per_page: 5, page: 1) {
    results
    ids
    error
  }
}
```

### Fetch Book Details

```graphql
query GetBookDetails($id: Int!) {
  books_by_pk(id: $id) {
    id
    title
    subtitle
    description
    rating
    ratings_count
    release_year
    pages
    cached_image
    contributions {
      contribution
      author {
        id
        name
        slug
      }
    }
    book_series {
      position
      series {
        id
        name
        slug
      }
    }
    editions(limit: 10) {
      id
      isbn_13
      asin
      edition_format
      reading_format { format }
      language { code2 language }
    }
  }
}
```

### Fetch Author with Books

```graphql
query GetAuthorWithBooks($id: Int!) {
  authors_by_pk(id: $id) {
    id
    name
    bio
    books_count
    cached_image
    contributions(limit: 20, order_by: { book: { release_year: desc } }) {
      contribution
      book {
        id
        title
        release_year
        rating
        cached_image
      }
    }
  }
}
```

### Fetch Series with Books

```graphql
query GetSeriesWithBooks($id: Int!) {
  series_by_pk(id: $id) {
    id
    name
    description
    books_count
    is_completed
    author {
      id
      name
    }
    book_series(order_by: { position: asc }) {
      position
      book {
        id
        title
        release_year
        rating
        cached_image
      }
    }
  }
}
```

### Fetch User's Library

```graphql
query GetUserLibrary {
  me {
    user_books(limit: 50, order_by: { updated_at: desc }) {
      status_id
      rating
      book {
        id
        title
        cached_image
        cached_contributors
      }
    }
  }
}
```

---

## Best Practices Summary

1. **Use `search()` for discovery**: Text-based filtering operators are disabled; always use the search query for finding content.

2. **Respect rate limits**: Implement a rate limiter (1 request/second recommended) to avoid 429 errors.

3. **Keep queries shallow**: Maximum depth is 3 levels. Plan your data fetching strategy accordingly.

4. **Use cached fields**: Leverage `cached_image`, `cached_contributors`, `cached_tags` to reduce query complexity.

5. **Search then fetch**: Get IDs via search, then fetch full details with `_by_pk` queries.

6. **Use `reading_format.format`**: For reliable format filtering, use this enumerated field instead of the free-text `edition_format`.

7. **Handle token expiration**: Tokens expire January 1st annually. Implement 401 error handling.

8. **Server-side only**: Never expose API tokens in client-side code.

9. **Paginate large result sets**: Use `limit`/`offset` for collection queries and `page`/`per_page` for search.

10. **Graceful degradation**: Handle cases where editions lack language or format data.

---

## References

- [Hardcover Official Documentation](https://docs.hardcover.app)
- [GraphQL Schema (GitHub)](https://github.com/hardcoverapp/hardcover-docs/blob/main/schema.graphql)
- [Hardcover Discord Community](https://discord.gg/hardcover)
