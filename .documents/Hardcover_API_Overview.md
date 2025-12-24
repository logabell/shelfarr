# Hardcover.app API - Complete Overview Guide

## Table of Contents
1. [Introduction](#introduction)
2. [API Basics & Authentication](#api-basics--authentication)
3. [GraphQL Console](#graphql-console)
4. [Available Queries](#available-queries)
5. [Available Mutations](#available-mutations)
6. [Search API](#search-api)
7. [Data Types & Fields](#data-types--fields)
8. [Filtering & Ordering](#filtering--ordering)
9. [Rate Limits & Limitations](#rate-limits--limitations)
10. [Error Handling](#error-handling)
11. [Implementation Guide](#implementation-guide)
12. [Code Examples](#code-examples)

---

## Introduction

Hardcover.app is a modern book tracking platform and Goodreads alternative. Their GraphQL API allows developers to:

- Search for books, authors, series, and other content
- Track reading progress and book status
- Manage personal bookshelves and custom lists
- Access book metadata (ratings, reviews, tags, contributors)
- Create reviews and journal entries
- Build custom integrations and tools

The API uses **GraphQL** (powered by Hasura) and is the same API used by the Hardcover website, iOS, and Android apps.

> **Note:** The API is currently in **beta** and subject to change. Features may break without notice.

---

## API Basics & Authentication

### API Endpoint

```
https://api.hardcover.app/v1/graphql
```

### Getting an API Key

1. Log in to your [Hardcover account](https://hardcover.app)
2. Go to **Account Settings**
3. Click on the **Hardcover API** link
4. Copy your API token from the top of the page

### Authentication

All requests require an `Authorization` header with your API token:

```
Authorization: Bearer YOUR_API_TOKEN
```

Or simply:

```
Authorization: YOUR_API_TOKEN
```

### Request Format

All requests are `POST` requests with a JSON body containing your GraphQL query:

```json
{
  "query": "your GraphQL query here",
  "variables": { }
}
```

### Required Headers

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `Authorization` | Your API token |
| `User-Agent` | (Recommended) Description of your script/app |

---

## GraphQL Console

Hardcover provides an interactive GraphQL console for testing queries:

**Console URL:** [https://api.hardcover.app/v1/graphql](https://api.hardcover.app/v1/graphql)

Or via Hasura's public GraphiQL:
```
https://cloud.hasura.io/public/graphiql?endpoint=https://api.hardcover.app/v1/graphql
```

### Using the Console

1. Add your token to the `authorization` header field
2. Click out of the field to load available resources
3. Use the explorer to browse available queries and types
4. Write and test your queries before implementing in code

---

## Available Queries

### User Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `me` | Get current authenticated user's data | Yes |
| `users` | Query public user data | No |
| `users_by_pk(id)` | Get specific user by ID | No |

### Book Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `books` | Query books with filters | No |
| `books_by_pk(id)` | Get specific book by ID | No |
| `search` | Full-text search across content | No |

### Author Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `authors` | Query authors with filters | No |
| `authors_by_pk(id)` | Get specific author by ID | No |

### Series Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `series` | Query book series | No |
| `series_by_pk(id)` | Get specific series by ID | No |

### Edition Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `editions` | Query book editions | No |
| `editions_by_pk(id)` | Get specific edition by ID | No |

### List Queries

| Query | Description | Auth Required |
|-------|-------------|---------------|
| `lists` | Query public and user lists | No |
| `lists_by_pk(id)` | Get specific list by ID | No |

### Other Queries

| Query | Description |
|-------|-------------|
| `publishers` | Query publishers |
| `characters` | Query book characters |
| `prompts` | Query community prompts |

---

## Available Mutations

### User Book Management

| Mutation | Description | Parameters |
|----------|-------------|------------|
| `insert_user_book` | Add book to library | `book_id`, `status_id`, `rating`, etc. |
| `update_user_book` | Update book status/rating | `id`, fields to update |
| `delete_user_book` | Remove book from library | `id` |

### List Management

| Mutation | Description | Parameters |
|----------|-------------|------------|
| `insert_list` | Create a new list | `name`, `description`, etc. |
| `insert_list_book` | Add book to a list | `book_id`, `list_id` |
| `delete_list_book` | Remove book from list | `id` |
| `update_list` | Update list details | `id`, fields to update |

### Review & Journal

| Mutation | Description | Parameters |
|----------|-------------|------------|
| `insert_journal_entry` | Add journal entry | `user_book_id`, `text`, etc. |
| `update_user_book` | Update review/rating | `id`, `review_raw`, `rating` |

---

## Search API

Hardcover uses **Typesense** for search functionality. The search endpoint supports multiple content types.

### Searchable Content Types

| Type | Default Fields | Description |
|------|----------------|-------------|
| `books` | `title`, `isbns`, `series_names`, `author_names`, `alternative_titles` | Book search |
| `authors` | `name`, `name_personal`, `alternate_names`, `series_names`, `books` | Author search |
| `series` | `name` | Series search |
| `characters` | `name`, `author_names`, `books` | Character search |
| `lists` | `name`, `description`, `books`, `username` | List search |
| `publishers` | `name` | Publisher search |
| `users` | `username`, `name` | User search |
| `prompts` | `prompt`, `books`, `lists` | Prompt search |

### Search Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | String | **Required.** Search terms |
| `query_by` | String | Fields to search (comma-separated) |
| `sort_by` | String | Field to sort results |
| `per_page` | Integer | Results per page |
| `page` | Integer | Page number |
| `fields` | String | Fields to search with custom weights |
| `weights` | String | Weights for each field (comma-separated) |

### Search Example

```graphql
query SearchBooks {
  search(
    query: "Brandon Sanderson"
    query_by: "title,author_names"
    per_page: 10
  ) {
    results
  }
}
```

### Book Search Response Fields

| Field | Description |
|-------|-------------|
| `id` | Book ID |
| `title` | Book title |
| `author_names` | Author name(s) |
| `alternative_titles` | Alternative titles |
| `isbns` | ISBN numbers |
| `series_names` | Associated series |
| `activities_count` | Activity count |
| `audio_seconds` | Audiobook duration |
| `compilation` | Is compilation |
| `pages` | Page count |
| `users_count` | Total users with this book |
| `users_read_count` | Users who read this book |
| `rating` | Average rating |
| `ratings_count` | Number of ratings |

---

## Data Types & Fields

### Book Object Fields

```graphql
type Book {
  id: Int!
  title: String!
  slug: String
  description: String
  pages: Int
  release_date: Date
  rating: Float
  ratings_count: Int
  users_count: Int
  users_read_count: Int
  book_status_id: Int
  compilation: Boolean
  
  # Cached fields (faster queries)
  cached_contributors: JSON    # Authors, translators, etc.
  cached_tags: JSON            # User-generated tags
  cached_image: JSON           # Cover image URLs
  
  # Relations
  contributions: [Contribution]
  editions: [Edition]
  book_series: [BookSeries]
  user_books: [UserBook]
}
```

### User Book Object Fields

```graphql
type UserBook {
  id: Int!
  book_id: Int!
  user_id: Int!
  status_id: Int           # 1=Want to Read, 2=Currently Reading, 3=Read, 4=DNF
  rating: Float            # 0-5 scale (supports half stars)
  review_raw: String       # Review text
  has_review: Boolean
  date_added: Timestamp
  started_at: Timestamp
  finished_at: Timestamp
  reviewed_at: Timestamp
  
  # Relations
  book: Book
  user: User
  edition: Edition
}
```

### Reading Status IDs

| Status ID | Status Name | Description |
|-----------|-------------|-------------|
| 1 | Want to Read | Books on your TBR list |
| 2 | Currently Reading | Books in progress |
| 3 | Read | Completed books |
| 4 | Did Not Finish | Abandoned books |
| 5 | (None) | Remove from library |

### Author/Contributor Object

```graphql
type Author {
  id: Int!
  name: String!
  slug: String
  bio: String
  image: String
  
  # Relations
  contributions: [Contribution]
  books: [Book]
}
```

### Series Object

```graphql
type Series {
  id: Int!
  name: String!
  slug: String
  books_count: Int
  
  # Relations
  book_series: [BookSeries]
}
```

### List Object

```graphql
type List {
  id: Int!
  name: String!
  slug: String
  description: String
  created_at: Timestamp
  
  # Relations
  list_books: [ListBook]
  user: User
}
```

### Edition Object

```graphql
type Edition {
  id: Int!
  book_id: Int!
  title: String
  isbn_10: String
  isbn_13: String
  pages: Int
  format: String           # Hardcover, Paperback, eBook, Audiobook
  language_id: Int
  publisher_id: Int
  release_date: Date
  audio_seconds: Int       # For audiobooks
  
  # Relations
  book: Book
  publisher: Publisher
}
```

---

## Filtering & Ordering

Hardcover's API (powered by Hasura) supports powerful filtering using the `where` clause.

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `_eq` | Equals | `{status_id: {_eq: 3}}` |
| `_neq` | Not equals | `{status_id: {_neq: 1}}` |
| `_gt` | Greater than | `{rating: {_gt: 4}}` |
| `_gte` | Greater than or equal | `{pages: {_gte: 200}}` |
| `_lt` | Less than | `{rating: {_lt: 3}}` |
| `_lte` | Less than or equal | `{pages: {_lte: 500}}` |
| `_in` | In array | `{status_id: {_in: [2, 3]}}` |
| `_nin` | Not in array | `{status_id: {_nin: [4]}}` |
| `_is_null` | Is null | `{review_raw: {_is_null: false}}` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `_and` | Logical AND | `{_and: [{status_id: {_eq: 3}}, {has_review: {_eq: true}}]}` |
| `_or` | Logical OR | `{_or: [{status_id: {_eq: 2}}, {status_id: {_eq: 3}}]}` |
| `_not` | Logical NOT | `{_not: {status_id: {_eq: 4}}}` |

### Disabled Operators (2025)

The following operators are **disabled** for security/performance:

- `_like`, `_nlike`
- `_ilike`, `_nilike`
- `_regex`, `_nregex`
- `_iregex`, `_niregex`
- `_similar`, `_nsimilar`

### Ordering Results

```graphql
query {
  me {
    user_books(
      order_by: [
        { date_added: desc },
        { rating: desc_nulls_last }
      ]
    ) {
      book { title }
    }
  }
}
```

**Order directions:**
- `asc` - Ascending
- `desc` - Descending
- `asc_nulls_first` - Ascending with nulls first
- `asc_nulls_last` - Ascending with nulls last
- `desc_nulls_first` - Descending with nulls first
- `desc_nulls_last` - Descending with nulls last

### Pagination

```graphql
query {
  books(
    limit: 20
    offset: 40
    order_by: { users_count: desc }
  ) {
    id
    title
  }
}
```

---

## Rate Limits & Limitations

### Rate Limits

| Limit Type | Value |
|------------|-------|
| Requests per minute | **60** |
| Query timeout | **30 seconds** |
| Query depth (2025) | **3 levels max** |

### Token Limitations

| Limitation | Details |
|------------|---------|
| Token expiry | **1 year** (resets January 1st) |
| Token reset | May be reset without notice during beta |
| Token sharing | **Never share** - can be used to access/delete account |

### Access Restrictions

| Restriction | Details |
|-------------|---------|
| **Data ownership** | You can only access/modify your own data |
| **Browser execution** | Queries must NOT run in browsers |
| **Allowed origins** | Only `localhost` or backend APIs |
| **User data (2025)** | Limited to own data, public data, and followed users |

### Disabled Features

- Like/regex pattern matching operators
- Cross-user private data access
- Browser-side API calls (security risk)

---

## Error Handling

### API Response Codes

| Code | Description | Example Body |
|------|-------------|--------------|
| **200** | Request successful | `{ "data": { ... } }` |
| **401** | Expired or invalid token | `{ "error": "Unable to verify token" }` |
| **403** | Access denied | `{ "error": "Message describing the error" }` |
| **404** | Not found | - |
| **429** | Rate limited | `{ "error": "Throttled" }` |
| **500** | Server error | `{ "error": "An unknown error occurred" }` |

### GraphQL Error Format

```json
{
  "errors": [
    {
      "message": "Error description",
      "extensions": {
        "path": "$.selectionSet.me",
        "code": "validation-failed"
      }
    }
  ],
  "data": null
}
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `Unable to verify token` | Invalid/expired token | Get new token from settings |
| `Throttled` | Rate limit exceeded | Wait and retry with backoff |
| `validation-failed` | Invalid query syntax | Check query in GraphQL console |
| `permission-denied` | Accessing unauthorized data | Only access your own data |

---

## Implementation Guide

### Step 1: Get Your API Token

1. Sign up/log in at [hardcover.app](https://hardcover.app)
2. Navigate to Account Settings → Hardcover API
3. Copy your token

### Step 2: Test in GraphQL Console

Before writing code, test your queries in the console:
- Use the built-in explorer to discover available fields
- Validate your queries work correctly
- Copy working queries to your code

### Step 3: Basic Implementation Structure

```javascript
const HARDCOVER_API = 'https://api.hardcover.app/v1/graphql';
const API_TOKEN = 'your-token-here';

async function hardcoverQuery(query, variables = {}) {
  const response = await fetch(HARDCOVER_API, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': API_TOKEN,
      'User-Agent': 'MyApp/1.0'
    },
    body: JSON.stringify({ query, variables })
  });
  
  const result = await response.json();
  
  if (result.errors) {
    throw new Error(result.errors[0].message);
  }
  
  return result.data;
}
```

### Step 4: Best Practices

| Practice | Recommendation |
|----------|----------------|
| **Caching** | Cache responses to reduce API calls |
| **Rate limiting** | Implement delays between requests |
| **Error handling** | Handle all error codes gracefully |
| **Token security** | Never expose token in client-side code |
| **User-Agent** | Include descriptive User-Agent header |
| **Query optimization** | Only request fields you need |
| **Pagination** | Use `limit` and `offset` for large datasets |

### Step 5: Data Flow Recommendations

1. **Fetch book by ID** when you have it (faster)
2. **Use search** for discovering new books
3. **Use `cached_*` fields** for better performance
4. **Batch operations** where possible

---

## Code Examples

### JavaScript/Fetch - Get User's Read Books

```javascript
const HARDCOVER_API = 'https://api.hardcover.app/v1/graphql';
const API_TOKEN = 'Bearer YOUR_TOKEN_HERE';

async function getReadBooks() {
  const query = `
    query GetReadBooks {
      me {
        user_books(
          where: { status_id: { _eq: 3 } }
          order_by: { finished_at: desc }
          limit: 50
        ) {
          id
          rating
          finished_at
          review_raw
          book {
            id
            title
            slug
            cached_contributors
            cached_image
          }
        }
      }
    }
  `;

  const response = await fetch(HARDCOVER_API, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': API_TOKEN,
    },
    body: JSON.stringify({ query })
  });

  const { data, errors } = await response.json();
  
  if (errors) {
    throw new Error(errors[0].message);
  }

  return data.me[0].user_books;
}

// Usage
getReadBooks().then(books => {
  books.forEach(ub => {
    const author = ub.book.cached_contributors?.[0]?.author?.name || 'Unknown';
    console.log(`${ub.book.title} by ${author} - Rating: ${ub.rating}`);
  });
});
```

### JavaScript - Add Book to "Want to Read"

```javascript
async function addToWantToRead(bookId) {
  const mutation = `
    mutation AddBook($bookId: Int!) {
      insert_user_book(object: {
        book_id: $bookId,
        status_id: 1
      }) {
        id
        status_id
        book {
          title
        }
      }
    }
  `;

  const response = await fetch(HARDCOVER_API, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': API_TOKEN,
    },
    body: JSON.stringify({
      query: mutation,
      variables: { bookId }
    })
  });

  const { data, errors } = await response.json();
  
  if (errors) {
    throw new Error(errors[0].message);
  }

  return data.insert_user_book;
}

// Usage
addToWantToRead(12345).then(result => {
  console.log(`Added "${result.book.title}" to Want to Read!`);
});
```

### JavaScript - Search for Books

```javascript
async function searchBooks(searchTerm, limit = 10) {
  const query = `
    query SearchBooks($term: String!, $limit: Int) {
      books(
        where: { title: { _ilike: $term } }
        limit: $limit
        order_by: { users_count: desc }
      ) {
        id
        title
        slug
        rating
        users_count
        cached_contributors
      }
    }
  `;

  const response = await fetch(HARDCOVER_API, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': API_TOKEN,
    },
    body: JSON.stringify({
      query,
      variables: { 
        term: `%${searchTerm}%`,
        limit 
      }
    })
  });

  const { data, errors } = await response.json();
  
  if (errors) {
    throw new Error(errors[0].message);
  }

  return data.books;
}
```

### Python Example

```python
import requests

HARDCOVER_API = 'https://api.hardcover.app/v1/graphql'
API_TOKEN = 'YOUR_TOKEN_HERE'

def hardcover_query(query: str, variables: dict = None):
    """Execute a GraphQL query against the Hardcover API."""
    headers = {
        'Content-Type': 'application/json',
        'Authorization': API_TOKEN,
        'User-Agent': 'MyPythonApp/1.0'
    }
    
    payload = {'query': query}
    if variables:
        payload['variables'] = variables
    
    response = requests.post(HARDCOVER_API, json=payload, headers=headers)
    response.raise_for_status()
    
    result = response.json()
    
    if 'errors' in result:
        raise Exception(result['errors'][0]['message'])
    
    return result['data']


def get_my_books(status_id: int = None):
    """Get user's books, optionally filtered by status."""
    where_clause = ""
    if status_id:
        where_clause = f"(where: {{status_id: {{_eq: {status_id}}}}})"
    
    query = f"""
    query {{
        me {{
            user_books{where_clause} {{
                id
                rating
                status_id
                date_added
                book {{
                    id
                    title
                    slug
                    cached_contributors
                }}
            }}
        }}
    }}
    """
    
    data = hardcover_query(query)
    return data['me'][0]['user_books']


def add_book_to_library(book_id: int, status_id: int = 1):
    """Add a book to the user's library."""
    mutation = """
    mutation AddBook($bookId: Int!, $statusId: Int!) {
        insert_user_book(object: {
            book_id: $bookId,
            status_id: $statusId
        }) {
            id
            book {
                title
            }
        }
    }
    """
    
    data = hardcover_query(mutation, {
        'bookId': book_id,
        'statusId': status_id
    })
    return data['insert_user_book']


def get_series_books(series_id: int):
    """Get all books in a series."""
    query = """
    query GetSeries($seriesId: Int!) {
        series_by_pk(id: $seriesId) {
            id
            name
            books_count
            book_series(order_by: {position: asc}) {
                position
                book {
                    id
                    title
                    users_read_count
                }
            }
        }
    }
    """
    
    data = hardcover_query(query, {'seriesId': series_id})
    return data['series_by_pk']


def get_user_lists():
    """Get all of the user's custom lists."""
    query = """
    query {
        me {
            lists(order_by: {created_at: desc}) {
                id
                name
                slug
                list_books {
                    book {
                        id
                        title
                    }
                }
            }
        }
    }
    """
    
    data = hardcover_query(query)
    return data['me'][0]['lists']


# Usage examples
if __name__ == '__main__':
    # Get all read books
    read_books = get_my_books(status_id=3)
    for ub in read_books:
        author = ub['book']['cached_contributors'][0]['author']['name'] \
                 if ub['book']['cached_contributors'] else 'Unknown'
        print(f"{ub['book']['title']} by {author}")
    
    # Get books in a series
    discworld = get_series_books(1018)
    print(f"\n{discworld['name']} ({discworld['books_count']} books)")
    for bs in discworld['book_series'][:5]:
        print(f"  {bs['position']}. {bs['book']['title']}")
```

### cURL Examples

```bash
# Get current user info
curl -X POST https://api.hardcover.app/v1/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_TOKEN_HERE" \
  -d '{"query": "{ me { id username } }"}'

# Get user's read books
curl -X POST https://api.hardcover.app/v1/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_TOKEN_HERE" \
  -d '{
    "query": "{ me { user_books(where: {status_id: {_eq: 3}}, limit: 10) { book { title } rating } } }"
  }'

# Add book to Want to Read
curl -X POST https://api.hardcover.app/v1/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_TOKEN_HERE" \
  -d '{
    "query": "mutation { insert_user_book(object: {book_id: 12345, status_id: 1}) { id } }"
  }'

# Get book by ID
curl -X POST https://api.hardcover.app/v1/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_TOKEN_HERE" \
  -d '{
    "query": "{ books_by_pk(id: 12345) { title slug rating cached_contributors } }"
  }'
```

---

## Additional Resources

- **Documentation:** [docs.hardcover.app](https://docs.hardcover.app)
- **GraphQL Console:** [api.hardcover.app/v1/graphql](https://api.hardcover.app/v1/graphql)
- **GitHub Docs Repo:** [github.com/hardcoverapp/hardcover-docs](https://github.com/hardcoverapp/hardcover-docs)
- **Discord Community:** [discord.gg/edGpYN8ym8](https://discord.gg/edGpYN8ym8)
- **Hardcover Website:** [hardcover.app](https://hardcover.app)

---

## Important Notes

1. **Beta Status:** The API is in beta and may change without notice
2. **Token Security:** Never expose your token in client-side code
3. **Data Ownership:** You can only access your own data and public data
4. **Rate Limiting:** Implement proper delays between requests (60/min limit)
5. **Backend Only:** This API should only be used from server-side code
6. **Hasura Based:** Since Hardcover uses Hasura, Hasura documentation can help with advanced queries
7. **2025 Changes:** Query depth limited to 3, data access limited to own/public/followed users
8. **OAuth Coming:** OAuth support for external applications is planned for 2025

---

*Document generated: December 2024*
*API Status: Beta*
*GraphQL Engine: Hasura*
