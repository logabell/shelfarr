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

| Document | Purpose |
|----------|---------|
| `.documents/hardcover_api_documentation.md` | **API reference** - Rate limits, query patterns, schema, error handling |
| `.documents/api_implementation_overview.md` | **Implementation guide** - Where API is used in codebase (living document) |
| `backend/internal/hardcover/client.go` | **Client implementation** - Go client with all API methods |
| [Official Schema](https://github.com/hardcoverapp/hardcover-docs/blob/main/schema.graphql) | **GraphQL schema** - Authoritative type definitions |

### Maintenance Requirements

**When modifying Hardcover integration, you MUST:**

1. **Consult** `.documents/hardcover_api_documentation.md` for API constraints and patterns
2. **Update** `.documents/api_implementation_overview.md` when adding/removing/modifying:
   - API routes or handlers
   - Client methods
   - Frontend pages or components that use Hardcover data
3. **Verify** changes respect rate limits (60 req/min) and query depth (max 3 levels)
