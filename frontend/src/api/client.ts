import axios from 'axios'
import type { QueryClient } from '@tanstack/react-query'
import type { 
  Book, 
  LibraryResponse, 
  LibraryStats, 
  Author, 
  Series, 
  SeriesDetail,
  AuthorDetail,
  SearchResult,
  AuthorSearchResult,
  SeriesSearchResult,
  ListSearchResult,
  UnifiedSearchResponse,
  SearchType,
  IndexerSearchResult,
  Indexer,
  DownloadClient,
  User,
  ReadProgress,
  QualityProfile,
  AuthorWithBooks
} from '@/types'

// Re-export types for use in pages
export type { Author, Book, SeriesDetail, AuthorDetail, DownloadClient, AuthorWithBooks }

const API_BASE = import.meta.env.VITE_API_URL || ''

const api = axios.create({
  baseURL: `${API_BASE}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Library endpoints
export const getLibrary = async (params?: {
  page?: number
  pageSize?: number
  status?: string
  authorId?: string
  seriesId?: string
  mediaType?: string
  sortBy?: string
  sortOrder?: string
}): Promise<LibraryResponse> => {
  const { data } = await api.get('/library', { params })
  return data
}

export const getLibraryStats = async (): Promise<LibraryStats> => {
  const { data } = await api.get('/library/stats')
  return data
}

// Book endpoints
export const getBooks = async (params?: { monitored?: boolean; status?: string }): Promise<Book[]> => {
  const { data } = await api.get('/books', { params })
  return data
}

export const getBook = async (id: number): Promise<Book> => {
  const { data } = await api.get(`/books/${id}`)
  return data
}

export const addBook = async (hardcoverId: string, monitored: boolean = true): Promise<Book> => {
  const { data } = await api.post('/books', { hardcoverId, monitored })
  return data
}

export const updateBook = async (id: number, updates: { monitored?: boolean; status?: string }): Promise<Book> => {
  const { data } = await api.put(`/books/${id}`, updates)
  return data
}


export const deleteBook = async (id: number): Promise<{
  message: string;
  bookId: number;
  hardcoverId: string;
  authorId?: number;
  seriesId?: number | null;
}> => {
  const { data } = await api.delete(`/books/${id}`)
  return data
}

export function invalidateAllBookQueries(queryClient: QueryClient) {
  // Invalidate all library-related queries
  queryClient.invalidateQueries({ queryKey: ['library'] });
  queryClient.invalidateQueries({ queryKey: ['library-stats'] });
  queryClient.invalidateQueries({ queryKey: ['libraryStats'] });
  
  // Invalidate all author-related queries
  queryClient.invalidateQueries({ queryKey: ['authors'] });
  queryClient.invalidateQueries({ queryKey: ['author'] });
  queryClient.invalidateQueries({ queryKey: ['hardcover-author'] });
  queryClient.invalidateQueries({ queryKey: ['hardcoverAuthor'] });
  
  // Invalidate all series-related queries
  queryClient.invalidateQueries({ queryKey: ['series'] });
  queryClient.invalidateQueries({ queryKey: ['hardcover-series'] });
  queryClient.invalidateQueries({ queryKey: ['hardcoverSeries'] });
  
  // Invalidate hardcover preview queries
  queryClient.invalidateQueries({ queryKey: ['hardcoverBook'] });
}


// Bulk book operations
export const bulkUpdateBooks = async (
  bookIds: number[], 
  updates: { monitored?: boolean; status?: string }
): Promise<{ updated: number }> => {
  const { data } = await api.put('/books/bulk', { bookIds, ...updates })
  return data
}

export const bulkDeleteBooks = async (
  bookIds: number[], 
  deleteFiles: boolean = false
): Promise<{ deleted: number }> => {
  const { data } = await api.delete('/books/bulk', { 
    data: { bookIds, deleteFiles } 
  })
  return data
}

// Author endpoints
export const getAuthors = async (params?: { monitored?: boolean }): Promise<Author[]> => {
  const { data } = await api.get('/authors', { params })
  return data
}

export const getAuthor = async (id: number): Promise<AuthorDetail> => {
  const { data } = await api.get(`/authors/${id}`)
  return data
}

export const addAuthor = async (hardcoverId: string, monitored: boolean = false, addAllBooks: boolean = false): Promise<Author> => {
  const { data } = await api.post('/authors', { hardcoverId, monitored, addAllBooks })
  return data
}

export const updateAuthor = async (id: number, updates: { monitored: boolean }): Promise<Author> => {
  const { data } = await api.put(`/authors/${id}`, updates)
  return data
}

export const deleteAuthor = async (id: number): Promise<void> => {
  await api.delete(`/authors/${id}`)
}

// Series endpoints
export const getSeries = async (): Promise<Series[]> => {
  const { data } = await api.get('/series')
  return data
}

export const getSeriesDetail = async (id: number): Promise<SeriesDetail> => {
  const { data } = await api.get(`/series/${id}`)
  return data
}

export interface AddSeriesBooksResponse {
  message: string
  addedCount: number
  skippedCount: number
  errors?: string[]
}

export const addSeriesBooks = async (
  seriesId: number, 
  bookIds: string[], 
  monitored: boolean = true
): Promise<AddSeriesBooksResponse> => {
  const { data } = await api.post(`/series/${seriesId}/books`, { bookIds, monitored })
  return data
}

// Search endpoints
export const searchHardcover = async (query: string, type: SearchType = 'book'): Promise<SearchResult[]> => {
  const { data } = await api.get('/search/hardcover', { params: { q: query, type } })
  return data
}

export const searchHardcoverAuthors = async (query: string): Promise<AuthorSearchResult[]> => {
  const { data } = await api.get('/search/hardcover', { params: { q: query, type: 'author' } })
  return data
}

export const searchHardcoverSeries = async (query: string): Promise<SeriesSearchResult[]> => {
  const { data } = await api.get('/search/hardcover', { params: { q: query, type: 'series' } })
  return data
}

export const searchHardcoverLists = async (query: string): Promise<ListSearchResult[]> => {
  const { data } = await api.get('/search/hardcover', { params: { q: query, type: 'list' } })
  return data
}

export const searchHardcoverAll = async (query: string): Promise<UnifiedSearchResponse> => {
  const { data } = await api.get('/search/hardcover', { params: { q: query, type: 'all' } })
  return data
}

export const testHardcoverConnection = async (): Promise<{ message: string }> => {
  const { data } = await api.post('/search/hardcover/test')
  return data
}

// Hardcover detail endpoints
export interface HardcoverBookDetail {
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
  hasAudiobook?: boolean
  hasEbook?: boolean
  hasDigitalEdition?: boolean
  digitalEditionCount?: number
  physicalEditionCount?: number
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
  books: HardcoverBookDetail[]
}

export interface HardcoverSeriesDetail {
  id: string
  name: string
  booksCount: number
  totalBooksCount: number
  digitalBooksCount: number
  physicalOnlyCount: number
  inLibrary: boolean
  books: HardcoverBookDetail[]
}

export const getHardcoverBook = async (id: string): Promise<HardcoverBookDetail> => {
  const { data } = await api.get(`/hardcover/book/${id}`)
  return data
}

export const getHardcoverAuthor = async (id: string): Promise<HardcoverAuthorDetail> => {
  const { data } = await api.get(`/hardcover/author/${id}`)
  return data
}

export const getHardcoverSeries = async (id: string): Promise<HardcoverSeriesDetail> => {
  const { data } = await api.get(`/hardcover/series/${id}`)
  return data
}

export const addHardcoverBook = async (
  id: string, 
  options?: { 
    monitored?: boolean; 
    mediaType?: string;
    forceAuthorId?: number;
    forceSeriesId?: number;
  }
): Promise<{ message: string; bookId: number }> => {
  const { data } = await api.post(`/hardcover/book/${id}`, options)
  return data
}

export const searchIndexers = async (params: { bookId?: number; q?: string; mediaType?: string }): Promise<IndexerSearchResult[]> => {
  const { data } = await api.get('/search/indexers', { params })
  return data
}

// Indexer endpoints
export const getIndexers = async (): Promise<Indexer[]> => {
  const { data } = await api.get('/indexers')
  return data
}

export const addIndexer = async (indexer: Omit<Indexer, 'id'>): Promise<Indexer> => {
  const { data } = await api.post('/indexers', indexer)
  return data
}

export const updateIndexer = async (id: number, indexer: Partial<Indexer>): Promise<Indexer> => {
  const { data } = await api.put(`/indexers/${id}`, indexer)
  return data
}

export const deleteIndexer = async (id: number): Promise<void> => {
  await api.delete(`/indexers/${id}`)
}

export const testIndexer = async (id: number): Promise<{ success: boolean; message: string }> => {
  const { data } = await api.post(`/indexers/${id}/test`)
  return data
}

// Download client endpoints
export const getDownloadClients = async (): Promise<DownloadClient[]> => {
  const { data } = await api.get('/downloadclients')
  return data
}

export const addDownloadClient = async (client: Omit<DownloadClient, 'id'>): Promise<DownloadClient> => {
  const { data } = await api.post('/downloadclients', client)
  return data
}

export const updateDownloadClient = async (id: number, client: Partial<DownloadClient>): Promise<DownloadClient> => {
  const { data } = await api.put(`/downloadclients/${id}`, client)
  return data
}

export const deleteDownloadClient = async (id: number): Promise<void> => {
  await api.delete(`/downloadclients/${id}`)
}

export const testDownloadClient = async (id: number): Promise<{ message: string }> => {
  const { data } = await api.post(`/downloadclients/${id}/test`)
  return data
}

// Test a download client configuration before saving
export const testDownloadClientConfig = async (config: {
  type: string
  url: string
  username?: string
  password?: string
}): Promise<{ message: string }> => {
  const { data } = await api.post('/downloadclients/test', config)
  return data
}

export const createDownloadClient = async (client: Omit<DownloadClient, 'id'>): Promise<DownloadClient> => {
  const { data } = await api.post('/downloadclients', client)
  return data
}

// Media file endpoints
export const streamMediaFile = (id: number): string => {
  return `${API_BASE}/api/v1/mediafiles/${id}/stream`
}

export const sendToKindle = async (id: number): Promise<{ message: string }> => {
  const { data } = await api.post(`/mediafiles/${id}/kindle`)
  return data
}

// User endpoints
export const getUsers = async (): Promise<User[]> => {
  const { data } = await api.get('/users')
  return data
}

export const getCurrentUser = async (): Promise<User> => {
  const { data } = await api.get('/users/me')
  return data
}

export const createUser = async (user: Omit<User, 'id'>): Promise<User> => {
  const { data } = await api.post('/users', user)
  return data
}

export const updateUser = async (id: number, user: Partial<User>): Promise<User> => {
  const { data } = await api.put(`/users/${id}`, user)
  return data
}

// Progress tracking
export const getProgress = async (mediaFileId: number): Promise<ReadProgress> => {
  const { data } = await api.get(`/progress/${mediaFileId}`)
  return data
}

export const updateProgress = async (mediaFileId: number, progress: number, position: number): Promise<void> => {
  await api.put(`/progress/${mediaFileId}`, { progress, position })
}

// Settings
export const getSettings = async (): Promise<Record<string, unknown>> => {
  const { data } = await api.get('/settings')
  return data
}

export const updateSettings = async (settings: Record<string, unknown>): Promise<void> => {
  await api.put('/settings', settings)
}

// Media Management Settings
export interface MediaSettings {
  ebookRootFolder: string
  audiobookRootFolder: string
  fileNamingEbook: string
  fileNamingAudiobook: string
  folderNaming: string
  useHardlinks: boolean
  recycleBinEnabled: boolean
  recycleBinPath: string
  rescanAfterImport: boolean
}

export interface RootFolder {
  id: number
  path: string
  mediaType: 'ebook' | 'audiobook'
  name: string
  freeSpace: number
  totalSpace: number
  accessible: boolean
}

export const getMediaSettings = async (): Promise<MediaSettings> => {
  const { data } = await api.get('/settings/media')
  return data
}

export const updateMediaSettings = async (settings: Partial<MediaSettings>): Promise<void> => {
  await api.put('/settings/media', settings)
}

// General settings
export interface GeneralSettings {
  instanceName: string
  defaultLanguage: string
  preferredLanguages: string[]
  startPage: string
  dateFormat: string
}

export interface LanguageOption {
  code: string
  name: string
}

export const getGeneralSettings = async (): Promise<GeneralSettings> => {
  const { data } = await api.get('/settings/general')
  return data
}

export const updateGeneralSettings = async (settings: Partial<GeneralSettings>): Promise<void> => {
  await api.put('/settings/general', settings)
}

export const getAvailableLanguages = async (): Promise<LanguageOption[]> => {
  const { data } = await api.get('/settings/languages')
  return data
}

export const getNamingPreview = async (template: string): Promise<{ template: string; preview: string }> => {
  const { data } = await api.get('/settings/media/naming-preview', { params: { template } })
  return data
}

// Filesystem browsing for directory selection
export interface DirectoryInfo {
  name: string
  path: string
  hasChildren: boolean
}

export interface BrowseFilesystemResponse {
  currentPath: string
  parent: string
  directories: DirectoryInfo[]
}

export const browseFilesystem = async (path?: string): Promise<BrowseFilesystemResponse> => {
  const { data } = await api.get('/filesystem/browse', { params: { path: path || '/' } })
  return data
}

export const getRootFolders = async (): Promise<RootFolder[]> => {
  const { data } = await api.get('/rootfolders')
  return data
}

export const addRootFolder = async (folder: { path: string; mediaType: string; name?: string }): Promise<RootFolder> => {
  const { data } = await api.post('/rootfolders', folder)
  return data
}

export const deleteRootFolder = async (id: number): Promise<void> => {
  await api.delete(`/rootfolders/${id}`)
}

export const getProfiles = async (): Promise<QualityProfile[]> => {
  const { data } = await api.get('/profiles')
  return data
}

export const getProfile = async (id: number): Promise<QualityProfile> => {
  const { data } = await api.get(`/profiles/${id}`)
  return data
}

export const createProfile = async (profile: Omit<QualityProfile, 'id'>): Promise<QualityProfile> => {
  const { data } = await api.post('/profiles', profile)
  return data
}

export const updateProfile = async (id: number, profile: Partial<Omit<QualityProfile, 'id'>>): Promise<QualityProfile> => {
  const { data } = await api.put(`/profiles/${id}`, profile)
  return data
}

export const deleteProfile = async (id: number): Promise<void> => {
  await api.delete(`/profiles/${id}`)
}

// Import endpoints
export const getPendingImports = async (): Promise<unknown[]> => {
  const { data } = await api.get('/import/pending')
  return data
}

export const manualImport = async (filePath: string, bookId: number, mediaType: string, editionName?: string): Promise<void> => {
  await api.post('/import/manual', { filePath, bookId, mediaType, editionName })
}

// Download endpoints
export interface Download {
  id: number
  bookId: number
  title: string
  status: string
  progress: number
  size: number
  downloaded: number
  addedAt: number
  completedAt?: number
}

export const getDownloads = async (status?: string): Promise<Download[]> => {
  const { data } = await api.get('/downloads', { params: { status } })
  return data
}

export const getDownload = async (id: number): Promise<Download> => {
  const { data } = await api.get(`/downloads/${id}`)
  return data
}

export const triggerDownload = async (params: {
  bookId: number
  indexer: string
  downloadUrl: string
  title: string
  size: number
  format: string
  mediaType: string
}): Promise<Download> => {
  const { data } = await api.post('/downloads', params)
  return data
}

export const deleteDownload = async (id: number): Promise<void> => {
  await api.delete(`/downloads/${id}`)
}

export const automaticSearch = async (bookId: number, mediaType?: string): Promise<{
  message: string
  downloadId: number
  title: string
  indexer: string
  size: number
  format: string
}> => {
  const { data } = await api.post(`/books/${bookId}/search`, null, { params: { mediaType } })
  return data
}

// Activity/History endpoints
export interface ActivityEvent {
  id: number
  type: string
  title: string
  message: string
  bookId?: number
  bookTitle?: string
  status: string
  timestamp: string
}

export const getActivity = async (limit?: number): Promise<ActivityEvent[]> => {
  const { data } = await api.get('/activity', { params: { limit } })
  return data
}

export const getActivityHistory = async (page?: number, pageSize?: number): Promise<{
  activities: ActivityEvent[]
  total: number
  page: number
  pageSize: number
}> => {
  const { data } = await api.get('/activity/history', { params: { page, pageSize } })
  return data
}

// Wanted endpoints
export interface WantedBook {
  id: number
  title: string
  authorId: number
  authorName: string
  coverUrl: string
  status: string
  seriesName?: string
  seriesIndex?: number
  monitored: boolean
  hasEbook: boolean
  hasAudiobook: boolean
}

export const getWanted = async (page?: number, pageSize?: number): Promise<{
  books: WantedBook[]
  total: number
  page: number
  pageSize: number
}> => {
  const { data } = await api.get('/wanted', { params: { page, pageSize } })
  return data
}

export const getWantedMissing = async (page?: number, pageSize?: number, mediaType?: string): Promise<{
  books: WantedBook[]
  total: number
  page: number
  pageSize: number
}> => {
  const { data } = await api.get('/wanted/missing', { params: { page, pageSize, mediaType } })
  return data
}

export const getWantedCutoff = async (page?: number, pageSize?: number): Promise<{
  books: WantedBook[]
  total: number
  page: number
  pageSize: number
}> => {
  const { data } = await api.get('/wanted/cutoff', { params: { page, pageSize } })
  return data
}

// System endpoints
export interface SystemStatus {
  version: string
  startTime: string
  uptime: string
  os: string
  arch: string
  goVersion: string
  database: {
    type: string
    path: string
    size: number
    status: string
  }
  disk: {
    booksPath: { path: string; exists: boolean; writable: boolean; usedBytes?: number }
    audiobooksPath: { path: string; exists: boolean; writable: boolean; usedBytes?: number }
    downloadsPath: { path: string; exists: boolean; writable: boolean; usedBytes?: number }
  }
  clients: {
    indexers: number
    downloadClients: number
    webSockets: number
  }
  library: {
    totalBooks: number
    monitoredBooks: number
    totalAuthors: number
    totalSeries: number
    totalMediaFiles: number
    totalSize: number
  }
}

export const getSystemStatus = async (): Promise<SystemStatus> => {
  const { data } = await api.get('/system/status')
  return data
}

export interface TaskInfo {
  name: string
  interval: string
  lastRun: string
  nextRun: string
  running: boolean
  enabled: boolean
  lastStatus: string
}

export const getSystemTasks = async (): Promise<TaskInfo[]> => {
  const { data } = await api.get('/system/tasks')
  return data
}

export const runSystemTask = async (name: string): Promise<{ message: string }> => {
  const { data } = await api.post(`/system/tasks/${name}/run`)
  return data
}

export const getSystemLogs = async (limit?: number): Promise<Array<{
  timestamp: string
  level: string
  message: string
}>> => {
  const { data } = await api.get('/system/logs', { params: { limit } })
  return data
}

// Notification endpoints
export interface Notification {
  id: number
  name: string
  type: string
  enabled: boolean
  webhookUrl?: string
  discordWebhook?: string
  telegramBotToken?: string
  telegramChatId?: string
  emailTo?: string
  onGrab: boolean
  onDownload: boolean
  onUpgrade: boolean
  onImport: boolean
  onDelete: boolean
  onHealthIssue: boolean
}

export const getNotifications = async (): Promise<Notification[]> => {
  const { data } = await api.get('/notifications')
  return data
}

export const createNotification = async (notification: Omit<Notification, 'id'>): Promise<Notification> => {
  const { data } = await api.post('/notifications', notification)
  return data
}

export const updateNotification = async (id: number, notification: Partial<Notification>): Promise<Notification> => {
  const { data } = await api.put(`/notifications/${id}`, notification)
  return data
}

export const deleteNotification = async (id: number): Promise<void> => {
  await api.delete(`/notifications/${id}`)
}

export const testNotification = async (id: number): Promise<{ message: string }> => {
  const { data } = await api.post(`/notifications/${id}/test`)
  return data
}

// List endpoints
export interface HardcoverList {
  id: number
  name: string
  hardcoverUrl: string
  hardcoverId: string
  enabled: boolean
  autoAdd: boolean
  monitor: boolean
  syncInterval: number
  qualityProfile?: number
  lastSyncedAt?: string
}

export const getLists = async (): Promise<HardcoverList[]> => {
  const { data } = await api.get('/lists')
  return data
}

export const createList = async (list: Omit<HardcoverList, 'id'>): Promise<HardcoverList> => {
  const { data } = await api.post('/lists', list)
  return data
}

export const updateList = async (id: number, list: Partial<HardcoverList>): Promise<HardcoverList> => {
  const { data } = await api.put(`/lists/${id}`, list)
  return data
}

export const deleteList = async (id: number): Promise<void> => {
  await api.delete(`/lists/${id}`)
}

export const syncList = async (id: number): Promise<{ message: string; booksAdded: number }> => {
  const { data } = await api.post(`/lists/${id}/sync`)
  return data
}

// Export apiClient as an object for convenient access
export const apiClient = {
  // Library
  getLibrary,
  getLibraryStats,
  // Books
  getBooks,
  getBook,
  addBook,
  updateBook,
  deleteBook,
  // Authors
  getAuthors,
  getAuthor,
  addAuthor,
  updateAuthor,
  deleteAuthor,
  // Series
  getSeries,
  getSeriesDetail,
  addSeriesBooks,
  // Search
  searchHardcover,
  searchHardcoverAuthors,
  searchHardcoverSeries,
  searchHardcoverLists,
  searchHardcoverAll,
  testHardcoverConnection,
  searchIndexers,
  // Hardcover detail
  getHardcoverBook,
  getHardcoverAuthor,
  getHardcoverSeries,
  addHardcoverBook,
  // Indexers
  getIndexers,
  addIndexer,
  updateIndexer,
  deleteIndexer,
  testIndexer,
  // Download clients
  getDownloadClients,
  addDownloadClient,
  createDownloadClient,
  updateDownloadClient,
  deleteDownloadClient,
  testDownloadClient,
  testDownloadClientConfig,
  // Media
  streamMediaFile,
  sendToKindle,
  // Users
  getUsers,
  getCurrentUser,
  createUser,
  updateUser,
  // Progress
  getProgress,
  updateProgress,
  // Settings
  getSettings,
  updateSettings,
  getMediaSettings,
  updateMediaSettings,
  getNamingPreview,
  getRootFolders,
  addRootFolder,
  deleteRootFolder,
  getProfiles,
  getProfile,
  createProfile,
  updateProfile,
  deleteProfile,
  // Imports
  getPendingImports,
  manualImport,
  // Downloads
  getDownloads,
  getDownload,
  triggerDownload,
  deleteDownload,
  automaticSearch,
  // Activity
  getActivity,
  getActivityHistory,
  // Wanted
  getWanted,
  getWantedMissing,
  getWantedCutoff,
  // System
  getSystemStatus,
  getSystemTasks,
  runSystemTask,
  getSystemLogs,
  // Notifications
  getNotifications,
  createNotification,
  updateNotification,
  deleteNotification,
  testNotification,
  // Lists
  getLists,
  createList,
  updateList,
  deleteList,
  syncList,
}

export default api

