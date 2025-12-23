# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shelfarr is a self-hosted media automation and consumption suite for ebooks and audiobooks. It combines:
- PVR functionality (monitoring, searching, downloading, upgrading media)
- Media consumption (streaming, reading, listening via the web)
- Integration with Hardcover.app for metadata
- Support for multiple indexers (MyAnonaMouse, Torznab, Anna's Archive)
- Download client integration (qBittorrent, Transmission, SABnzbd, etc.)

## Architecture

### Monorepo Structure
- **backend/** - Go backend with Echo framework
- **frontend/** - React TypeScript frontend with Vite
- **docker/** - Docker Compose configurations

### Backend (Go)
- **Entry point**: `backend/cmd/shelfarr/main.go`
- **Framework**: Echo v4 (HTTP server)
- **Database**: GORM with SQLite (supports PostgreSQL)
- **Architecture**: Clean architecture with internal packages

Key internal packages:
- `internal/api/` - HTTP handlers and routes (server.go sets up all routes)
- `internal/db/` - Database models and initialization (models.go contains all entity definitions)
- `internal/auth/` - JWT authentication and SSO support
- `internal/hardcover/` - Hardcover.app API client for book metadata
- `internal/indexer/` - Search indexer implementations (MAM, Torznab, Anna's Archive)
- `internal/downloader/` - Download client integrations
- `internal/media/` - Media file handling and format conversion
- `internal/kindle/` - Kindle email delivery
- `internal/realtime/` - WebSocket hub for real-time updates
- `internal/scheduler/` - Background job scheduling
- `internal/config/` - Configuration management

### Frontend (React + TypeScript)
- **Build tool**: Vite
- **UI Framework**: React 19 with Tailwind CSS
- **Component library**: Radix UI primitives
- **Data fetching**: TanStack Query
- **Routing**: React Router v7
- **Path alias**: `@/` maps to `frontend/src/`

Key directories:
- `src/api/client.ts` - Central API client with all backend endpoints
- `src/pages/` - Route pages (one per view)
- `src/components/ui/` - shadcn/ui components
- `src/components/layout/` - Navigation, header, etc.
- `src/components/library/` - Library-specific components
- `src/components/reader/` - EPUB.js reader integration
- `src/components/player/` - Audio player components
- `src/types/` - TypeScript type definitions

## Development Commands

### Backend Development
```bash
cd backend
go mod download          # Install dependencies
go run ./cmd/shelfarr    # Run development server (listens on :8080)
go build -o shelfarr ./cmd/shelfarr  # Build binary
go test ./...            # Run tests
```

### Frontend Development
```bash
cd frontend
npm install              # Install dependencies
npm run dev              # Run Vite dev server (proxies /api to :8080)
npm run build            # Build for production
npm run lint             # Run ESLint
npm run preview          # Preview production build
```

### Docker Development
```bash
# From project root
docker compose -f docker/docker-compose.dev.yml up  # Run with hot reload
docker compose -f docker/docker-compose.yml up      # Run production build
```

## Key Technical Details

### API Communication
- Frontend uses Axios client in `src/api/client.ts`
- All endpoints are under `/api/v1/`
- Vite dev server proxies `/api` to `http://localhost:8080`
- Backend serves static frontend from `public/` directory in production

### Authentication Flow
1. JWT-based authentication (tokens in Authorization header)
2. Optional SSO via header-based authentication
3. Auth middleware applied to protected routes in `internal/api/server.go`
4. Default admin user created on first run

### Database Models
All models defined in `backend/internal/db/models.go`:
- **Core entities**: Book, Author, Series, MediaFile
- **User management**: User, ReadProgress
- **Infrastructure**: Indexer, DownloadClient, QualityProfile, Notification
- **Operations**: Download, Setting, HardcoverList

### Book Status Flow
Books follow this status progression:
1. `missing` - Monitored but no files
2. `downloading` - Active download in progress
3. `downloaded` - File imported successfully
4. `unmonitored` - Not actively monitored
5. `unreleased` - Future release date

### Media Types
Two primary types throughout the codebase:
- `ebook` - EPUB, AZW3, MOBI, PDF
- `audiobook` - M4B, MP3

### Search Flow
1. User searches Hardcover.app (metadata source)
2. Book added to library (creates Book + Author + Series if needed)
3. Automatic/Manual search across configured indexers
4. Results sent to configured download client
5. Completed downloads imported to library
6. Media files linked to books

### Real-time Updates
- WebSocket hub in `internal/realtime/`
- Endpoint: `/ws`
- Used for download progress, import status, library updates

### Configuration
Environment variables (see README.md for full list):
- `SHELFARR_LISTEN_ADDR` - Server address (default `:8080`)
- `SHELFARR_CONFIG_PATH` - Database and config directory
- `SHELFARR_BOOKS_PATH` - Ebook library path
- `SHELFARR_AUDIOBOOKS_PATH` - Audiobook library path
- `HARDCOVER_API_URL` - Metadata API endpoint

### Hardcover.app API Limitations

**Critical: The Hardcover API is still in beta and may change without notice.**

Rate Limits & Timeouts:
- **60 requests per minute** - Rate limiting is enforced
- **30 second query timeout** - Long queries will fail
- **Maximum depth of 3** - GraphQL queries cannot nest deeper than 3 levels

Data Access Restrictions:
- Can only access: your own user data, public data, and data of users you follow
- Must run from backend (server-side) - **never from browser**
- Only for offline/local use - cannot be used from public websites

Disabled Query Operators:
- `_like`, `_nlike`, `_ilike` - LIKE pattern matching disabled
- `_regex`, `_nregex`, `_iregex`, `_niregex` - Regex matching disabled
- `_similar`, `_nsimilar` - Similar matching disabled

Token & Security:
- API tokens **expire after 1 year**, and reset on January 1st
- Tokens may be reset without notice during beta
- **Never share your token** - it provides full account access
- Include a `user-agent` header describing your script

GraphQL ID Types:
- Book, Author, Series IDs are `Int!` type - must convert string IDs to integers
- Use `strconv.Atoi()` when passing IDs from URL parameters to GraphQL queries

## Common Patterns

### Adding a New API Endpoint
1. Add handler function in appropriate `internal/api/*.go` file
2. Register route in `internal/api/server.go` setupRoutes()
3. Add TypeScript function in `frontend/src/api/client.ts`
4. Add types to `frontend/src/types/` if needed
5. Use in frontend pages via TanStack Query

### Adding a New Database Model
1. Define struct in `internal/db/models.go`
2. Add to migrations in `internal/db/database.go` Migrate()
3. Create corresponding TypeScript type in `frontend/src/types/`
4. Add API endpoints if needed

### Frontend Component Structure
- Use shadcn/ui components from `components/ui/`
- Page components in `src/pages/` are route-level
- Shared components organized by feature (library, reader, player)
- Use TanStack Query for data fetching with `useQuery` and `useMutation`

## Testing Notes
- Backend uses standard Go testing (`*_test.go` files)
- Run specific package: `go test github.com/shelfarr/shelfarr/internal/indexer`
- Frontend uses built-in Vite testing capabilities
