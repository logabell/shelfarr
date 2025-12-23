import { useState, useMemo, useEffect, useRef } from 'react'
import { useParams, Link, useSearchParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Book,
  Download,
  Play,
  Star,
  Calendar,
  BookOpen,
  Headphones,
  Search,
  Loader2,
  ExternalLink,
  Check,
  Trash2,
  Send,
  Filter,
  SortDesc,
  X
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
} from '@/components/ui/collapsible'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { getBook, searchIndexers, updateBook, triggerDownload, deleteBook } from '@/api/client'
import type { IndexerSearchResult } from '@/types'

// Sort options
type SortOption = 'seeders-desc' | 'seeders-asc' | 'size-desc' | 'size-asc' | 'quality-desc'

// Filter state
interface FilterState {
  minSeeders: number
  freeleechOnly: boolean
  vipOnly: boolean
  formats: string[]
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDate(dateString?: string): string {
  if (!dateString) return 'Unknown'
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

const statusColors: Record<string, string> = {
  downloaded: 'bg-status-downloaded',
  missing: 'bg-status-missing',
  downloading: 'bg-status-downloading',
  unmonitored: 'bg-status-unmonitored',
  unreleased: 'bg-status-unreleased',
}

const defaultFilters: FilterState = {
  minSeeders: 0,
  freeleechOnly: false,
  vipOnly: false,
  formats: []
}

// Available formats for filtering
const EBOOK_FORMATS = ['epub', 'azw3', 'mobi', 'pdf', 'cbz', 'cbr']
const AUDIOBOOK_FORMATS = ['m4b', 'mp3', 'flac', 'm4a']

export function BookDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchMediaType, setSearchMediaType] = useState<'ebook' | 'audiobook'>('ebook')
  const [isSearching, setIsSearching] = useState(false)
  const hasAutoSearched = useRef(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  
  // Sort and filter state
  const [sortOption, setSortOption] = useState<SortOption>('seeders-desc')
  const [filters, setFilters] = useState<FilterState>(defaultFilters)
  const [showFilters, setShowFilters] = useState(false)

  const { data: book, isLoading: bookLoading } = useQuery({
    queryKey: ['book', id],
    queryFn: () => getBook(Number(id)),
    enabled: !!id,
  })

  const { data: searchResults, refetch: refetchSearch } = useQuery({
    queryKey: ['indexerSearch', id, searchMediaType],
    queryFn: () => searchIndexers({ bookId: Number(id), mediaType: searchMediaType }),
    enabled: false,
  })

  // Get available formats based on media type
  const availableFormats = searchMediaType === 'audiobook' ? AUDIOBOOK_FORMATS : EBOOK_FORMATS

  // Filtered and sorted results
  const processedResults = useMemo(() => {
    if (!searchResults) return []

    let results = [...searchResults]

    // Apply filters
    if (filters.minSeeders > 0) {
      results = results.filter(r => (r.seeders || 0) >= filters.minSeeders)
    }
    if (filters.freeleechOnly) {
      results = results.filter(r => r.freeleech)
    }
    if (filters.vipOnly) {
      results = results.filter(r => r.vip)
    }
    if (filters.formats.length > 0) {
      results = results.filter(r => 
        filters.formats.some(f => r.format?.toLowerCase().includes(f.toLowerCase()))
      )
    }

    // Apply sorting
    switch (sortOption) {
      case 'seeders-desc':
        results.sort((a, b) => (b.seeders || 0) - (a.seeders || 0))
        break
      case 'seeders-asc':
        results.sort((a, b) => (a.seeders || 0) - (b.seeders || 0))
        break
      case 'size-desc':
        results.sort((a, b) => (b.size || 0) - (a.size || 0))
        break
      case 'size-asc':
        results.sort((a, b) => (a.size || 0) - (b.size || 0))
        break
      case 'quality-desc':
        // Sort by quality score if available
        results.sort((a, b) => {
          const scoreA = a.quality?.includes('Excellent') ? 3 : a.quality?.includes('Good') ? 2 : 1
          const scoreB = b.quality?.includes('Excellent') ? 3 : b.quality?.includes('Good') ? 2 : 1
          return scoreB - scoreA
        })
        break
    }

    return results
  }, [searchResults, sortOption, filters])

  // Check if any filters are active
  const hasActiveFilters = filters.minSeeders > 0 || 
    filters.freeleechOnly || 
    filters.vipOnly || 
    filters.formats.length > 0

  const resetFilters = () => {
    setFilters(defaultFilters)
  }

  const toggleFormat = (format: string) => {
    setFilters(prev => ({
      ...prev,
      formats: prev.formats.includes(format)
        ? prev.formats.filter(f => f !== format)
        : [...prev.formats, format]
    }))
  }

  const updateMutation = useMutation({
    mutationFn: (updates: { monitored?: boolean; status?: string }) =>
      updateBook(Number(id), updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['book', id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteBook(Number(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['library'] })
      navigate('/')
    },
  })

  const handleDeleteBook = () => {
    setShowDeleteDialog(false)
    deleteMutation.mutate()
  }

  const handleSearch = async () => {
    setIsSearching(true)
    await refetchSearch()
    setIsSearching(false)
  }

  // Auto-trigger search when navigating with ?search=true
  useEffect(() => {
    if (searchParams.get('search') === 'true' && book && !hasAutoSearched.current) {
      hasAutoSearched.current = true
      handleSearch()
    }
  }, [searchParams, book])

  const handleToggleMonitored = () => {
    if (book) {
      updateMutation.mutate({ monitored: !book.monitored })
    }
  }

  if (bookLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!book) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold">Book Not Found</h1>
          <Link to="/" className="text-primary hover:underline mt-4 block">
            Return to Library
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Topbar title={book.title} />

      <div className="flex-1 overflow-auto">
        {/* Hero Section */}
        <div className="relative bg-gradient-to-b from-card to-background">
          <div className="absolute inset-0 overflow-hidden">
            {book.coverUrl && (
              <img
                src={book.coverUrl}
                alt=""
                className="w-full h-full object-cover opacity-10 blur-3xl scale-110"
              />
            )}
          </div>

          <div className="relative max-w-6xl mx-auto px-6 py-8">
            <Link
              to="/"
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Library
            </Link>

            <div className="flex gap-8">
              {/* Cover */}
              <div className="shrink-0">
                <div className="relative w-48 aspect-[2/3] rounded-lg overflow-hidden shadow-2xl">
                  {book.coverUrl ? (
                    <img
                      src={book.coverUrl}
                      alt={book.title}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full bg-muted flex items-center justify-center">
                      <Book className="h-16 w-16 text-muted-foreground" />
                    </div>
                  )}
                  <div
                    className={`absolute bottom-0 left-0 right-0 h-1 ${
                      statusColors[book.status] || 'bg-muted'
                    }`}
                  />
                </div>
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h1 className="text-3xl font-bold mb-2">{book.title}</h1>
                    {book.author && (
                      <div className="flex items-center gap-2">
                        <Link
                          to={`/authors/${book.author.id}`}
                          className="text-lg text-muted-foreground hover:text-primary transition-colors"
                        >
                          {book.author.name}
                        </Link>
                        {book.author.hardcoverId && (
                          <Link
                            to={`/hardcover/author/${book.author.hardcoverId}`}
                            className="text-xs text-muted-foreground hover:text-primary"
                            title="View more author details from Hardcover"
                          >
                            <ExternalLink className="h-3 w-3" />
                          </Link>
                        )}
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant={book.monitored ? 'default' : 'outline'}
                      onClick={handleToggleMonitored}
                      disabled={updateMutation.isPending}
                    >
                      {book.monitored ? (
                        <>
                          <Check className="h-4 w-4" />
                          Monitored
                        </>
                      ) : (
                        'Monitor'
                      )}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => setShowDeleteDialog(true)}
                      disabled={deleteMutation.isPending}
                      className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
                    >
                      {deleteMutation.isPending ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                {/* Metadata */}
                <div className="flex flex-wrap gap-4 mt-4">
                  {book.rating > 0 && (
                    <div className="flex items-center gap-1 text-sm">
                      <Star className="h-4 w-4 text-yellow-500 fill-yellow-500" />
                      <span>{book.rating.toFixed(1)}</span>
                    </div>
                  )}

                  {book.releaseDate && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <Calendar className="h-4 w-4" />
                      <span>{formatDate(book.releaseDate)}</span>
                    </div>
                  )}

                  {book.pageCount > 0 && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <BookOpen className="h-4 w-4" />
                      <span>{book.pageCount} pages</span>
                    </div>
                  )}

                  {book.series && (
                    <Link
                      to={`/series/${book.series.id}`}
                      className="flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      #{book.seriesIndex} in {book.series.name}
                    </Link>
                  )}
                </div>

                {/* Format Badges */}
                <div className="flex gap-2 mt-4">
                  {book.hasEbook && (
                    <Badge variant="secondary">
                      <Book className="h-3 w-3 mr-1" />
                      Ebook
                    </Badge>
                  )}
                  {book.hasAudiobook && (
                    <Badge variant="secondary">
                      <Headphones className="h-3 w-3 mr-1" />
                      Audiobook
                    </Badge>
                  )}
                  <Badge
                    variant="outline"
                    className={`${statusColors[book.status]} bg-opacity-20`}
                  >
                    {book.status}
                  </Badge>
                </div>

                {/* Description */}
                {book.description && (
                  <p className="mt-4 text-sm text-muted-foreground line-clamp-4">
                    {book.description}
                  </p>
                )}

                {/* Action Buttons */}
                <div className="flex gap-2 mt-6">
                  {book.hasEbook && book.mediaFiles?.some((f) => f.mediaType === 'ebook') && (
                    <Button>
                      <BookOpen className="h-4 w-4" />
                      Read
                    </Button>
                  )}
                  {book.hasAudiobook && book.mediaFiles?.some((f) => f.mediaType === 'audiobook') && (
                    <Button variant="secondary">
                      <Play className="h-4 w-4" />
                      Listen
                    </Button>
                  )}
                  {book.hasEbook && (
                    <Button variant="outline">
                      <Send className="h-4 w-4" />
                      Send to Kindle
                    </Button>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Content Sections */}
        <div className="max-w-6xl mx-auto px-6 py-8 space-y-8">
          {/* Media Files */}
          {book.mediaFiles && book.mediaFiles.length > 0 && (
            <section>
              <h2 className="text-xl font-semibold mb-4">Files</h2>
              <div className="space-y-2">
                {book.mediaFiles.map((file) => (
                  <div
                    key={file.id}
                    className="flex items-center justify-between p-4 rounded-lg bg-card border border-border"
                  >
                    <div className="flex items-center gap-3">
                      {file.mediaType === 'audiobook' ? (
                        <Headphones className="h-5 w-5 text-primary" />
                      ) : (
                        <Book className="h-5 w-5 text-primary" />
                      )}
                      <div>
                        <div className="font-medium">{file.fileName}</div>
                        <div className="text-sm text-muted-foreground">
                          {file.format.toUpperCase()} • {formatBytes(file.fileSize)}
                          {file.editionName && ` • ${file.editionName}`}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button variant="ghost" size="sm">
                        <ExternalLink className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" className="text-destructive">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* Search Section */}
          <section>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Search for Downloads</h2>
              <div className="flex items-center gap-2">
                <Select
                  value={searchMediaType}
                  onValueChange={(v) => setSearchMediaType(v as 'ebook' | 'audiobook')}
                >
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ebook">Ebook</SelectItem>
                    <SelectItem value="audiobook">Audiobook</SelectItem>
                  </SelectContent>
                </Select>
                <Button onClick={handleSearch} disabled={isSearching}>
                  {isSearching ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Search className="h-4 w-4" />
                  )}
                  Search
                </Button>
              </div>
            </div>

            {/* Sort and Filter Controls */}
            {searchResults && searchResults.length > 0 && (
              <div className="mb-4 space-y-3">
                {/* Top bar with sort and filter toggle */}
                <div className="flex items-center justify-between gap-4">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">
                      {processedResults.length} of {searchResults.length} results
                    </span>
                    {hasActiveFilters && (
                      <Button variant="ghost" size="sm" onClick={resetFilters} className="h-7 text-xs">
                        <X className="h-3 w-3 mr-1" />
                        Clear filters
                      </Button>
                    )}
                  </div>
                  
                  <div className="flex items-center gap-2">
                    {/* Sort Select */}
                    <Select value={sortOption} onValueChange={(v) => setSortOption(v as SortOption)}>
                      <SelectTrigger className="w-40 h-9">
                        <SortDesc className="h-4 w-4 mr-2" />
                        <SelectValue placeholder="Sort by..." />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="seeders-desc">Seeders (High → Low)</SelectItem>
                        <SelectItem value="seeders-asc">Seeders (Low → High)</SelectItem>
                        <SelectItem value="size-desc">Size (Largest)</SelectItem>
                        <SelectItem value="size-asc">Size (Smallest)</SelectItem>
                        <SelectItem value="quality-desc">Quality Score</SelectItem>
                      </SelectContent>
                    </Select>

                    {/* Filter Toggle */}
                    <Button
                      variant={showFilters ? 'secondary' : 'outline'}
                      size="sm"
                      onClick={() => setShowFilters(!showFilters)}
                      className="h-9"
                    >
                      <Filter className="h-4 w-4 mr-1" />
                      Filters
                      {hasActiveFilters && (
                        <Badge variant="secondary" className="ml-2 h-5 px-1.5">
                          {filters.formats.length + (filters.freeleechOnly ? 1 : 0) + (filters.vipOnly ? 1 : 0) + (filters.minSeeders > 0 ? 1 : 0)}
                        </Badge>
                      )}
                    </Button>
                  </div>
                </div>

                {/* Filter Panel */}
                <Collapsible open={showFilters} onOpenChange={setShowFilters}>
                  <CollapsibleContent>
                    <div className="p-4 rounded-lg border border-border bg-card space-y-4">
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {/* Format Filters */}
                        <div className="space-y-2">
                          <Label className="text-sm font-medium">Formats</Label>
                          <div className="flex flex-wrap gap-2">
                            {availableFormats.map(format => (
                              <Button
                                key={format}
                                size="sm"
                                variant={filters.formats.includes(format) ? 'default' : 'outline'}
                                onClick={() => toggleFormat(format)}
                                className="h-7 text-xs uppercase"
                              >
                                {format}
                              </Button>
                            ))}
                          </div>
                        </div>

                        {/* Min Seeders */}
                        <div className="space-y-2">
                          <Label className="text-sm font-medium">Minimum Seeders</Label>
                          <Input
                            type="number"
                            min={0}
                            value={filters.minSeeders || ''}
                            onChange={(e) => setFilters(prev => ({ 
                              ...prev, 
                              minSeeders: parseInt(e.target.value) || 0 
                            }))}
                            placeholder="0"
                            className="h-9 w-24"
                          />
                        </div>

                        {/* Toggle Filters */}
                        <div className="space-y-3">
                          <div className="flex items-center justify-between">
                            <Label className="text-sm">Freeleech Only</Label>
                            <Switch
                              checked={filters.freeleechOnly}
                              onCheckedChange={(v) => setFilters(prev => ({ ...prev, freeleechOnly: v }))}
                            />
                          </div>
                          <div className="flex items-center justify-between">
                            <Label className="text-sm">VIP Only</Label>
                            <Switch
                              checked={filters.vipOnly}
                              onCheckedChange={(v) => setFilters(prev => ({ ...prev, vipOnly: v }))}
                            />
                          </div>
                        </div>
                      </div>
                    </div>
                  </CollapsibleContent>
                </Collapsible>
              </div>
            )}

            {/* Results */}
            {searchResults && searchResults.length > 0 ? (
              processedResults.length > 0 ? (
                <div className="space-y-2">
                  {processedResults.map((result, index) => (
                    <SearchResultCard 
                      key={index} 
                      result={result} 
                      bookId={book.id} 
                      mediaType={searchMediaType}
                      onDownloadStarted={() => {
                        queryClient.invalidateQueries({ queryKey: ['book', id] })
                      }}
                    />
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground border border-dashed border-border rounded-lg">
                  <Filter className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No results match your filters</p>
                  <Button variant="link" onClick={resetFilters} className="mt-2">
                    Clear filters
                  </Button>
                </div>
              )
            ) : searchResults && searchResults.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground border border-dashed border-border rounded-lg">
                No results found. Try searching a different media type.
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground border border-dashed border-border rounded-lg">
                Click Search to find available downloads from your configured indexers.
              </div>
            )}
          </section>
        </div>
      </div>

      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Book</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove "{book.title}" from your library? 
              This will not delete any downloaded files.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteBook}>
              Remove
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function SearchResultCard({ 
  result, 
  bookId,
  mediaType,
  onDownloadStarted
}: { 
  result: IndexerSearchResult
  bookId: number
  mediaType: 'ebook' | 'audiobook'
  onDownloadStarted?: () => void
}) {
  const [downloading, setDownloading] = useState(false)
  const [downloadSuccess, setDownloadSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleDownload = async () => {
    setDownloading(true)
    setError(null)
    try {
      await triggerDownload({
        bookId,
        indexer: result.indexer,
        downloadUrl: result.downloadUrl,
        title: result.title,
        size: result.size,
        format: result.format || '',
        mediaType
      })
      setDownloadSuccess(true)
      onDownloadStarted?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Download failed')
    } finally {
      setDownloading(false)
    }
  }

  return (
    <div className="flex items-center justify-between p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium truncate">{result.title}</span>
          {result.freeleech && (
            <Badge variant="secondary" className="bg-green-500/20 text-green-500">
              FL
            </Badge>
          )}
          {result.vip && (
            <Badge variant="secondary" className="bg-yellow-500/20 text-yellow-500">
              VIP
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-4 mt-1 text-sm text-muted-foreground">
          <span>{result.indexer}</span>
          <span>{result.format}</span>
          <span>{formatBytes(result.size)}</span>
          {result.seeders !== undefined && (
            <span className="text-green-500">{result.seeders} seeders</span>
          )}
          {result.langCode && (
            <span className="text-blue-400" title="Language">{result.langCode}</span>
          )}
          {result.author && <span>{result.author}</span>}
          {result.narrator && <span>Narrated by {result.narrator}</span>}
        </div>
      </div>

      <div className="flex items-center gap-2 ml-4">
        <Badge
          variant="outline"
          className={
            result.quality.includes('Excellent')
              ? 'border-green-500 text-green-500'
              : result.quality.includes('Good')
              ? 'border-blue-500 text-blue-500'
              : ''
          }
        >
          {result.quality}
        </Badge>
        <Button 
          size="sm" 
          onClick={handleDownload} 
          disabled={downloading || downloadSuccess}
          variant={downloadSuccess ? 'secondary' : error ? 'destructive' : 'default'}
          title={error || undefined}
        >
          {downloading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : downloadSuccess ? (
            <Check className="h-4 w-4" />
          ) : (
            <Download className="h-4 w-4" />
          )}
        </Button>
        {result.infoUrl && (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => window.open(result.infoUrl, '_blank')}
          >
            <ExternalLink className="h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  )
}

