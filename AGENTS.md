# AGENTS.md

## Commands
```bash
# Backend (from backend/)
go run ./cmd/shelfarr          # Run dev server
go test ./...                  # Run all tests
go test ./internal/indexer     # Run single package tests
go test -run TestFunctionName ./internal/api  # Run single test

# Frontend (from frontend/)
npm run dev                    # Dev server (proxies /api to :8080)
npm run build                  # Build (runs tsc + vite)
npm run lint                   # ESLint
```

## Code Style
- **Go**: Standard gofmt, tabs, PascalCase exports, camelCase locals. Error returns as last value; always check errors
- **TypeScript**: ES modules, async/await, explicit types. Path alias `@/` = `frontend/src/`
- **Imports**: stdlib first, external second, internal last (Go); type imports use `import type { X }`
- **Naming**: Go uses PascalCase for exported, camelCase for internal. TS uses camelCase functions, PascalCase types/components
- **API**: All endpoints under `/api/v1/`, handlers in `internal/api/`, client in `frontend/src/api/client.ts`
- **Models**: GORM models in `internal/db/models.go`, TypeScript types in `frontend/src/types/index.ts`
- **Errors**: Go returns `error` as last return value; use `echo.NewHTTPError()` for API errors
- **No type suppression**: Never use `as any`, `@ts-ignore`, `@ts-expect-error`

## Hardcover.app API (Source of Truth)

**Hardcover.app is the primary metadata source for all books, authors, and series.**

### Required Reading Before Modifying Hardcover Code
1. **Schema**: https://github.com/hardcoverapp/hardcover-docs/blob/main/schema.graphql
2. **Context doc**: `.cursor/plans/hardcover_api_context.md`
3. **Client implementation**: `backend/internal/hardcover/client.go`

### API Constraints
- **Rate limit**: 60 req/min (client enforces 1 req/sec via `golang.org/x/time/rate`)
- **Query depth**: Max 3 levels of nesting
- **Disabled operators**: `_like`, `_ilike`, `_regex` - always use `search()` for text matching
- **Auth**: Bearer token; tokens expire annually on January 1st

### Query Patterns
| Operation | Pattern | Example |
|-----------|---------|---------|
| Discovery | `search(query, query_type)` | `search(query: "Sanderson", query_type: "Author")` |
| Detail fetch | `*_by_pk(id)` | `books_by_pk(id: 123)` |
| Language filter | `editions { language { code2 } }` | Filter by ISO 639-1 codes |
| Author's books | `books(where: {contributions: {author_id: {_eq: $id}}})` | Via contributions join |
| Aggregates | `*_aggregate` with `where` | Count books by author |

### Key Relationships (depth-aware)
```
books -> contributions -> author (depth 3 ✓)
books -> editions -> language (depth 3 ✓)
books -> book_series -> series (depth 3 ✓)
series -> book_series -> book -> contributions (depth 4 ✗)
```

### Untapped Schema Features (for future enhancement)
- `books_trending(time_period)` - Trending book discovery
- `series.is_completed` - Series completion status
- `editions.asin` - Amazon/Kindle linking
- `user_books` - Reading status sync (requires user auth)
