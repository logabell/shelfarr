# Shelfarr

A self-hosted media automation and consumption suite for Ebooks and Audiobooks.

Shelfarr unifies the "PVR" aspects of the *Arr suite (monitoring, searching, downloading, and upgrading media) with media consumption features (streaming, reading, and listening via the web).

## Features

- 📚 **Library Management** - Track and organize your ebook and audiobook collection
- 🔍 **Smart Search** - Waterfall search across multiple indexers (MAM, Torznab, Anna's Archive)
- 📥 **Automated Downloads** - Integration with qBittorrent, Transmission, SABnzbd, and more
- 📖 **Integrated Reader** - Built-in epub.js reader with theme and font controls
- 🎧 **Audio Player** - HTML5 audio player with speed control and chapter navigation
- 🔄 **Format Conversion** - Automatic conversion via Calibre (ebook-convert)
- 🎵 **Audio Binding** - Combine MP3s into M4B via FFmpeg
- 📱 **PWA Support** - Mobile-friendly with "Add to Home Screen"
- 👥 **Multi-User** - User accounts with permissions and progress sync
- 📧 **Send to Kindle** - Email books directly to your Kindle device

## Tech Stack

- **Backend**: Go (Golang) with Echo framework
- **Frontend**: React (TypeScript) with Tailwind CSS
- **Database**: SQLite (WAL mode) or PostgreSQL
- **Tools**: Calibre, FFmpeg (bundled in Docker)

## Quick Start

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/shelfarr/shelfarr.git
cd shelfarr

# Create your docker-compose.yml
cp docker/docker-compose.yml .

# Edit paths in docker-compose.yml
# - /path/to/books:/books
# - /path/to/audiobooks:/audiobooks
# - /path/to/downloads:/downloads

# Start Shelfarr
docker-compose up -d
```

Access Shelfarr at `http://localhost:8080`

### Development Setup

```bash
# Backend
cd backend
go mod download
go run ./cmd/shelfarr

# Frontend (in another terminal)
cd frontend
npm install
npm run dev
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SHELFARR_LISTEN_ADDR` | `:8080` | HTTP server address |
| `SHELFARR_CONFIG_PATH` | `/config` | Config and database directory |
| `SHELFARR_BOOKS_PATH` | `/books` | Ebook library root |
| `SHELFARR_AUDIOBOOKS_PATH` | `/audiobooks` | Audiobook library root |
| `SHELFARR_DOWNLOADS_PATH` | `/downloads` | Download staging area |
| `HARDCOVER_API_URL` | `https://api.hardcover.app/v1/graphql` | Hardcover API endpoint |

### Volume Mounts

| Path | Purpose |
|------|---------|
| `/config` | Database, logs, settings |
| `/books` | Ebook library |
| `/audiobooks` | Audiobook library |
| `/downloads` | Download staging |

## Indexers

### Supported

- **MyAnonaMouse (MAM)** - Native JSON API with cookie auth
- **Torznab** - Generic support via Prowlarr/Jackett
- **Anna's Archive** - Direct link scraper

### Search Logic

1. Search by ISBN (high accuracy)
2. Search by Author + Title
3. Search by Title (fuzzy matching)

## Quality Profiles

### Ebooks
Format ranking: `EPUB > AZW3 > MOBI > PDF`

### Audiobooks
- Format ranking: `M4B > MP3`
- Minimum bitrate: 64kbps (configurable)

## Status Indicators

| Color | Status |
|-------|--------|
| 🟢 Green | Downloaded & Imported |
| 🔵 Blue | Downloaded (Unmonitored) |
| 🟠 Orange | Downloading / Queued |
| 🔴 Red | Missing (Monitored) |
| 🟣 Purple | Unreleased |
| ⚫ Black | Unmonitored / Missing |

## API

The API follows REST conventions at `/api/v1/`:

- `GET /library` - Library grid with pagination
- `GET /books/:id` - Book details
- `POST /books` - Add book from Hardcover
- `GET /search/hardcover` - Search Hardcover.app
- `GET /search/indexers` - Search configured indexers
- `POST /import/manual` - Manual file import

## License

MIT License - See [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by:
- [Radarr](https://radarr.video/)
- [Sonarr](https://sonarr.tv/)
- [Readarr](https://readarr.com/)
- [Hardcover.app](https://hardcover.app/)

