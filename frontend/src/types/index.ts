// API Types for Shelfarr

export type BookStatus = 'missing' | 'downloading' | 'downloaded' | 'unmonitored' | 'unreleased'
export type MediaType = 'ebook' | 'audiobook'

export interface Author {
  id: number
  hardcoverId: string
  name: string
  sortName: string
  biography?: string
  imageUrl: string
  monitored: boolean
  bookCount?: number        // Books in library
  totalBooksCount?: number  // Total books from Hardcover (cached)
  downloadedCount?: number  // Books with files
}

export interface Series {
  id: number
  hardcoverId: string
  name: string
  bookCount?: number        // Books in library
  totalBooksCount?: number  // Total books from Hardcover (cached)
  downloadedCount?: number  // Books with files
}

export interface MediaFile {
  id: number
  filePath: string
  fileName: string
  fileSize: number
  format: string
  mediaType: MediaType
  bitrate?: number
  duration?: number
  editionName?: string
}

export interface Book {
  id: number
  hardcoverId: string
  title: string
  sortTitle: string
  isbn: string
  description: string
  coverUrl: string
  rating: number
  releaseDate?: string
  pageCount: number
  status: BookStatus
  monitored: boolean
  author?: Author
  series?: Series
  seriesIndex?: number
  mediaFiles?: MediaFile[]
  hasEbook: boolean
  hasAudiobook: boolean
  format?: string
}

export interface LibraryResponse {
  books: Book[]
  total: number
  page: number
  pageSize: number
}

export interface LibraryStats {
  totalBooks: number
  monitoredBooks: number
  downloadedBooks: number
  missingBooks: number
  totalAuthors: number
  totalSeries: number
  totalEbooks: number
  totalAudiobooks: number
  totalFileSize: number
}

export interface SearchResult {
  id: string
  title: string
  author: string
  authorId: string
  coverUrl: string
  rating: number
  releaseYear?: number
  isbn?: string
  description?: string
  inLibrary: boolean
}

export interface AuthorSearchResult {
  id: string
  name: string
  imageUrl: string
  booksCount: number
  biography?: string
  inLibrary: boolean
}

export interface SeriesSearchResult {
  id: string
  name: string
  booksCount: number
  authorId?: string
  authorName?: string
  inLibrary: boolean
}

export interface ListSearchResult {
  id: string
  name: string
  description?: string
  booksCount: number
  username?: string
}

export type SearchType = 'all' | 'book' | 'author' | 'series' | 'list'

export interface UnifiedSearchResponse {
  books?: SearchResult[]
  authors?: AuthorSearchResult[]
  series?: SeriesSearchResult[]
  lists?: ListSearchResult[]
}

export interface IndexerSearchResult {
  indexer: string
  title: string
  size: number
  format: string
  seeders?: number
  leechers?: number
  downloadUrl: string
  infoUrl?: string
  publishDate?: string
  quality: string
  freeleech?: boolean
  vip?: boolean
  author?: string
  narrator?: string
  category?: string
  langCode?: string
}

export interface Indexer {
  id: number
  name: string
  type: 'torznab' | 'mam' | 'anna'
  url: string
  apiKey?: string
  cookie?: string
  priority: number
  enabled: boolean
  vipOnly?: boolean
  freeleechOnly?: boolean
}

export interface DownloadClient {
  id: number
  name: string
  type: 'qbittorrent' | 'transmission' | 'deluge' | 'sabnzbd' | 'nzbget' | 'rtorrent' | 'internal'
  url: string
  username?: string
  password?: string
  category: string
  priority: number
  enabled: boolean
  settings?: string // JSON for extra config (seedbox type, path mappings, etc.)
}

export interface User {
  id: number
  username: string
  email: string
  isAdmin: boolean
  canRead: boolean
  canDelete: boolean
}

export interface ReadProgress {
  progress: number
  position: number
  lastReadAt: string
}

export interface SeriesBookEntry {
  index: number
  book?: Book
  hardcoverId?: string
  title: string
  coverUrl?: string
  authorName?: string
  rating?: number
  releaseYear?: number
  compilation?: boolean
  inLibrary: boolean
}

export interface SeriesDetail extends Series {
  description: string
  books: SeriesBookEntry[]
  totalBooks: number
  inLibrary: number
  downloadedCount: number
  missingBooks: number
}

// AuthorBookEntry represents a book by an author (may or may not be in library)
export interface AuthorBookEntry {
  hardcoverId: string
  title: string
  coverUrl?: string
  authorName?: string
  rating: number
  releaseYear?: number
  seriesId?: string
  seriesName?: string
  seriesIndex?: number
  compilation?: boolean
  inLibrary: boolean
  book?: Book
}

export interface AuthorDetail extends Author {
  books: AuthorBookEntry[]
  totalBooks: number
  inLibrary: number
  downloadedCount: number
}

// Legacy interface for backwards compatibility
export interface AuthorWithBooks extends Author {
  books: Book[]
}

export interface QualityProfile {
  id: number
  name: string
  mediaType: MediaType
  formatRanking: string
  minBitrate?: number
}

export interface HardcoverBookResult {
  id: string
  title: string
  description?: string
  coverUrl?: string
  rating: number
  releaseDate?: string
  releaseYear?: number
  pageCount?: number
  isbn?: string
  isbn13?: string
  authorId?: string
  authorName?: string
  authorImage?: string
  seriesId?: string
  seriesName?: string
  seriesIndex?: number
  genres?: string[]
  languageCode?: string
  hasAudiobook: boolean
  hasEbook: boolean
  hasDigitalEdition: boolean
  digitalEditionCount: number
  physicalEditionCount: number
  editionCount?: number
  audioDuration?: number
  inLibrary: boolean
  libraryBook?: Book
}

export interface HardcoverAuthorDetail {
  id: string
  name: string
  biography?: string
  imageUrl?: string
  booksCount: number
  totalBooksCount: number
  digitalBooksCount: number
  physicalOnlyCount: number
  inLibrary: boolean
  books: HardcoverBookResult[]
}

export interface HardcoverSeriesDetail {
  id: string
  name: string
  booksCount: number
  totalBooksCount: number
  digitalBooksCount: number
  physicalOnlyCount: number
  inLibrary: boolean
  books: HardcoverBookResult[]
}

