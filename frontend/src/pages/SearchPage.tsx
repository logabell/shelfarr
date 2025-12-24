import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { 
  searchHardcover, 
  searchHardcoverAuthors, 
  searchHardcoverSeries, 
  searchHardcoverLists,
  searchHardcoverAll,
  addAuthor,
  automaticSearch
} from '@/api/client'
import { AddBookModal } from '@/components/search/AddBookModal'
import { 
  Search, 
  Plus, 
  Check, 
  Star, 
  BookOpen, 
  User, 
  Library, 
  ListIcon,
  Filter,
  X,
  ChevronRight,
  CheckCircle2,
  AlertCircle
} from 'lucide-react'
import type { 
  SearchResult, 
  AuthorSearchResult, 
  SeriesSearchResult, 
  ListSearchResult,
  SearchType
} from '@/types'

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const initialQuery = searchParams.get('q') || ''
  const initialType = (searchParams.get('type') as SearchType) || 'all'
  
  const [query, setQuery] = useState(initialQuery)
  const [searchTerm, setSearchTerm] = useState(initialQuery)
  const [searchType, setSearchType] = useState<SearchType>(initialType)
  const [showFilters, setShowFilters] = useState(false)
  
  // Filter state
  const [yearRange, setYearRange] = useState<[number, number]>([1900, new Date().getFullYear()])
  const [minRating, setMinRating] = useState(0)
  
  const [selectedBookId, setSelectedBookId] = useState<string | null>(null)
  const [recentlyAddedAuthors, setRecentlyAddedAuthors] = useState<Set<string>>(new Set())
  const [notifications, setNotifications] = useState<Array<{ id: string; type: 'success' | 'error'; message: string }>>([])
  
  const queryClient = useQueryClient()
  
  const addNotification = (type: 'success' | 'error', message: string) => {
    const id = Math.random().toString(36).substring(7)
    setNotifications(prev => [...prev, { id, type, message }])
    setTimeout(() => {
      setNotifications(prev => prev.filter(n => n.id !== id))
    }, 3000)
  }
  
  const removeNotification = (notificationId: string) => {
    setNotifications(prev => prev.filter(n => n.id !== notificationId))
  }

  // Unified search query
  const { data: unifiedResults, isLoading: isLoadingAll, error: errorAll } = useQuery({
    queryKey: ['search', 'all', searchTerm],
    queryFn: () => searchHardcoverAll(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'all',
    retry: false,
  })

  // Book search query
  const { data: bookResults, isLoading: isLoadingBooks, error: errorBooks } = useQuery({
    queryKey: ['search', 'book', searchTerm],
    queryFn: () => searchHardcover(searchTerm, 'book'),
    enabled: searchTerm.length > 2 && searchType === 'book',
    retry: false,
  })

  // Author search query
  const { data: authorResults, isLoading: isLoadingAuthors, error: errorAuthors } = useQuery({
    queryKey: ['search', 'author', searchTerm],
    queryFn: () => searchHardcoverAuthors(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'author',
    retry: false,
  })

  // Series search query
  const { data: seriesResults, isLoading: isLoadingSeries, error: errorSeries } = useQuery({
    queryKey: ['search', 'series', searchTerm],
    queryFn: () => searchHardcoverSeries(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'series',
    retry: false,
  })

  // List search query
  const { data: listResults, isLoading: isLoadingLists, error: errorLists } = useQuery({
    queryKey: ['search', 'list', searchTerm],
    queryFn: () => searchHardcoverLists(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'list',
    retry: false,
  })

  const isLoading = isLoadingAll || isLoadingBooks || isLoadingAuthors || isLoadingSeries || isLoadingLists
  const searchError = errorAll || errorBooks || errorAuthors || errorSeries || errorLists

  // Filter books based on current filters
  const filteredBooks = useMemo(() => {
    const books = searchType === 'all' ? unifiedResults?.books : bookResults
    if (!books) return []
    
    return books.filter(book => {
      if (book.releaseYear && (book.releaseYear < yearRange[0] || book.releaseYear > yearRange[1])) {
        return false
      }
      if (book.rating < minRating) {
        return false
      }
      return true
    })
  }, [searchType, unifiedResults?.books, bookResults, yearRange, minRating])

  const addAuthorMutation = useMutation({
    mutationFn: (hardcoverId: string) => addAuthor(hardcoverId, true, false),
    onSuccess: (_, hardcoverId) => {
      setRecentlyAddedAuthors(prev => new Set(prev).add(hardcoverId))
      queryClient.invalidateQueries({ queryKey: ['authors'] })
      queryClient.invalidateQueries({ queryKey: ['search'] })
      addNotification('success', 'Author added to library')
    },
    onError: (error: Error) => {
      const message = error.message.includes('409') || error.message.includes('Conflict')
        ? 'Already following this author'
        : error.message || 'Failed to add author'
      addNotification('error', message)
    },
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setSearchTerm(query)
    setSearchParams({ q: query, type: searchType })
  }

  const handleTypeChange = (type: string) => {
    setSearchType(type as SearchType)
    if (searchTerm) {
      setSearchParams({ q: searchTerm, type })
    }
  }

  const handleAddBook = (result: SearchResult) => {
    // Open the AddBookModal instead of directly adding
    setSelectedBookId(result.id)
  }

  const handleBookClick = (result: SearchResult) => {
    // Always open the modal to show details and add options
    setSelectedBookId(result.id)
  }

  const handleAuthorClick = (result: AuthorSearchResult) => {
    navigate(`/hardcover/author/${result.id}`)
  }

  const handleSeriesClick = (result: SeriesSearchResult) => {
    navigate(`/hardcover/series/${result.id}`)
  }

  const handleAddAuthor = (result: AuthorSearchResult) => {
    addAuthorMutation.mutate(result.id)
  }

  const resetFilters = () => {
    setYearRange([1900, new Date().getFullYear()])
    setMinRating(0)
  }

  const hasActiveFilters = yearRange[0] !== 1900 || yearRange[1] !== new Date().getFullYear() || minRating > 0

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Search" />

      {notifications.length > 0 && (
        <div className="fixed top-20 right-4 z-50 flex flex-col gap-2">
          {notifications.map(notification => (
            <div
              key={notification.id}
              className={`flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg ${
                notification.type === 'success'
                  ? 'bg-green-500/90 text-white'
                  : 'bg-red-500/90 text-white'
              }`}
            >
              {notification.type === 'success' ? (
                <CheckCircle2 className="h-5 w-5" />
              ) : (
                <AlertCircle className="h-5 w-5" />
              )}
              <span className="text-sm font-medium">{notification.message}</span>
              <button
                onClick={() => removeNotification(notification.id)}
                className="ml-2 hover:opacity-70"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
      )}
      
      <div className="flex-1 overflow-auto p-6 space-y-6">
        {/* Search Form */}
        <div className="space-y-4">
          <form onSubmit={handleSearch} className="flex gap-2 max-w-3xl">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search for books, authors, series, or lists..."
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="pl-9"
              />
            </div>
            <Button type="submit" disabled={query.length < 3}>
              Search
            </Button>
          </form>

          {/* Category Tabs */}
          <Tabs value={searchType} onValueChange={handleTypeChange} className="w-full max-w-3xl">
            <TabsList className="grid w-full grid-cols-5">
              <TabsTrigger value="all" className="flex items-center gap-1.5">
                <Search className="h-3.5 w-3.5" />
                All
              </TabsTrigger>
              <TabsTrigger value="book" className="flex items-center gap-1.5">
                <BookOpen className="h-3.5 w-3.5" />
                Books
              </TabsTrigger>
              <TabsTrigger value="author" className="flex items-center gap-1.5">
                <User className="h-3.5 w-3.5" />
                Authors
              </TabsTrigger>
              <TabsTrigger value="series" className="flex items-center gap-1.5">
                <Library className="h-3.5 w-3.5" />
                Series
              </TabsTrigger>
              <TabsTrigger value="list" className="flex items-center gap-1.5">
                <ListIcon className="h-3.5 w-3.5" />
                Lists
              </TabsTrigger>
            </TabsList>
          </Tabs>

          {/* Filters - Only for book searches */}
          {(searchType === 'book' || searchType === 'all') && searchTerm && (
            <div className="max-w-3xl">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowFilters(!showFilters)}
                className="flex items-center gap-2"
              >
                <Filter className="h-4 w-4" />
                Filters
                {hasActiveFilters && (
                  <Badge variant="secondary" className="ml-1">Active</Badge>
                )}
              </Button>

              {showFilters && (
                <div className="mt-4 p-4 rounded-lg border border-border bg-card space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-medium">Filter Results</h3>
                    {hasActiveFilters && (
                      <Button variant="ghost" size="sm" onClick={resetFilters}>
                        <X className="h-4 w-4 mr-1" />
                        Reset
                      </Button>
                    )}
                  </div>

                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label>Release Year: {yearRange[0]} - {yearRange[1]}</Label>
                      <Slider
                        value={yearRange}
                        onValueChange={(value) => setYearRange(value as [number, number])}
                        min={1900}
                        max={new Date().getFullYear()}
                        step={1}
                        minStepsBetweenThumbs={1}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label>Minimum Rating: {minRating > 0 ? minRating.toFixed(1) : 'Any'}</Label>
                      <Slider
                        value={[minRating]}
                        onValueChange={(value) => setMinRating(value[0])}
                        min={0}
                        max={5}
                        step={0.5}
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Results */}
        {searchError ? (
          <div className="flex flex-col items-center justify-center py-16">
            <X className="h-16 w-16 text-destructive mb-4" />
            <p className="text-lg text-destructive">Search Error</p>
            <p className="text-sm text-muted-foreground mt-1 max-w-md text-center">
              {searchError instanceof Error ? searchError.message : 'Failed to search. Please check your API key in Settings.'}
            </p>
            <Button variant="outline" className="mt-4" onClick={() => window.location.href = '/settings/library-search'}>
              Go to Settings
            </Button>
          </div>
        ) : isLoading ? (
          <div className="space-y-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex gap-4 p-4 rounded-lg bg-card border border-border">
                <div className="w-20 h-28 skeleton rounded" />
                <div className="flex-1 space-y-2">
                  <div className="h-5 w-3/4 skeleton rounded" />
                  <div className="h-4 w-1/2 skeleton rounded" />
                  <div className="h-3 w-full skeleton rounded" />
                </div>
              </div>
            ))}
          </div>
        ) : searchTerm.length < 3 ? (
          <div className="flex flex-col items-center justify-center py-16">
            <Search className="h-16 w-16 text-muted-foreground mb-4" />
            <p className="text-lg text-muted-foreground">Enter at least 3 characters to search</p>
          </div>
        ) : (
          <div className="space-y-8">
            {/* Unified Results */}
            {searchType === 'all' && unifiedResults && (
              <>
                {/* Books Section */}
                {filteredBooks && filteredBooks.length > 0 && (
                  <section>
                    <div className="flex items-center justify-between mb-4">
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <BookOpen className="h-5 w-5" />
                        Books ({filteredBooks.length})
                      </h2>
                      <Button variant="ghost" size="sm" onClick={() => handleTypeChange('book')}>
                        View All <ChevronRight className="h-4 w-4 ml-1" />
                      </Button>
                    </div>
                    <div className="space-y-3">
                      {filteredBooks.slice(0, 5).map((result) => (
                        <BookResultCard 
                          key={result.id} 
                          result={result} 
                          onAdd={handleAddBook}
                          isAdding={false}
                          onClick={handleBookClick}
                        />
                      ))}
                    </div>
                  </section>
                )}

                {/* Authors Section */}
                {unifiedResults.authors && unifiedResults.authors.length > 0 && (
                  <section>
                    <div className="flex items-center justify-between mb-4">
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <User className="h-5 w-5" />
                        Authors ({unifiedResults.authors.length})
                      </h2>
                      <Button variant="ghost" size="sm" onClick={() => handleTypeChange('author')}>
                        View All <ChevronRight className="h-4 w-4 ml-1" />
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {unifiedResults.authors.slice(0, 4).map((result) => (
                        <AuthorResultCard 
                          key={result.id} 
                          result={result} 
                          onAdd={handleAddAuthor}
                          isAdding={addAuthorMutation.isPending}
                          onClick={handleAuthorClick}
                          recentlyAdded={recentlyAddedAuthors.has(result.id)}
                        />
                      ))}
                    </div>
                  </section>
                )}

                {/* Series Section */}
                {unifiedResults.series && unifiedResults.series.length > 0 && (
                  <section>
                    <div className="flex items-center justify-between mb-4">
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <Library className="h-5 w-5" />
                        Series ({unifiedResults.series.length})
                      </h2>
                      <Button variant="ghost" size="sm" onClick={() => handleTypeChange('series')}>
                        View All <ChevronRight className="h-4 w-4 ml-1" />
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {unifiedResults.series.slice(0, 4).map((result) => (
                        <SeriesResultCard key={result.id} result={result} onClick={handleSeriesClick} />
                      ))}
                    </div>
                  </section>
                )}

                {/* Lists Section */}
                {unifiedResults.lists && unifiedResults.lists.length > 0 && (
                  <section>
                    <div className="flex items-center justify-between mb-4">
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <ListIcon className="h-5 w-5" />
                        Lists ({unifiedResults.lists.length})
                      </h2>
                      <Button variant="ghost" size="sm" onClick={() => handleTypeChange('list')}>
                        View All <ChevronRight className="h-4 w-4 ml-1" />
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {unifiedResults.lists.slice(0, 4).map((result) => (
                        <ListResultCard key={result.id} result={result} />
                      ))}
                    </div>
                  </section>
                )}

                {/* No Results */}
                {(!filteredBooks || filteredBooks.length === 0) && 
                 (!unifiedResults.authors || unifiedResults.authors.length === 0) && 
                 (!unifiedResults.series || unifiedResults.series.length === 0) &&
                 (!unifiedResults.lists || unifiedResults.lists.length === 0) && (
                  <NoResults />
                )}
              </>
            )}

            {/* Book Results */}
            {searchType === 'book' && (
              <>
                {filteredBooks && filteredBooks.length > 0 ? (
                  <div className="space-y-3">
                    {filteredBooks.map((result) => (
                      <BookResultCard 
                        key={result.id} 
                        result={result} 
                        onAdd={handleAddBook}
                        isAdding={false}
                        onClick={handleBookClick}
                      />
                    ))}
                  </div>
                ) : bookResults?.length === 0 ? (
                  <NoResults />
                ) : hasActiveFilters && bookResults && bookResults.length > 0 ? (
                  <div className="flex flex-col items-center justify-center py-16">
                    <Filter className="h-16 w-16 text-muted-foreground mb-4" />
                    <p className="text-lg text-muted-foreground">No results match your filters</p>
                    <Button variant="link" onClick={resetFilters}>Reset filters</Button>
                  </div>
                ) : null}
              </>
            )}

            {/* Author Results */}
            {searchType === 'author' && (
              <>
                {authorResults && authorResults.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {authorResults.map((result) => (
                      <AuthorResultCard 
                        key={result.id} 
                        result={result} 
                        onAdd={handleAddAuthor}
                        isAdding={addAuthorMutation.isPending}
                        onClick={handleAuthorClick}
                        recentlyAdded={recentlyAddedAuthors.has(result.id)}
                      />
                    ))}
                  </div>
                ) : authorResults?.length === 0 ? (
                  <NoResults />
                ) : null}
              </>
            )}

            {/* Series Results */}
            {searchType === 'series' && (
              <>
                {seriesResults && seriesResults.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {seriesResults.map((result) => (
                      <SeriesResultCard key={result.id} result={result} onClick={handleSeriesClick} />
                    ))}
                  </div>
                ) : seriesResults?.length === 0 ? (
                  <NoResults />
                ) : null}
              </>
            )}

            {/* List Results */}
            {searchType === 'list' && (
              <>
                {listResults && listResults.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {listResults.map((result) => (
                      <ListResultCard key={result.id} result={result} />
                    ))}
                  </div>
                ) : listResults?.length === 0 ? (
                  <NoResults />
                ) : null}
              </>
            )}
          </div>
        )}
      </div>

      {/* Add Book Modal */}
      <AddBookModal
        bookId={selectedBookId}
        isOpen={!!selectedBookId}
        onClose={() => setSelectedBookId(null)}
        onSuccess={(bookId, downloadMode, mediaType) => {
          queryClient.invalidateQueries({ queryKey: ['search'] })
          queryClient.invalidateQueries({ queryKey: ['library'] })
          setSelectedBookId(null)
          
          // Handle download based on mode
          if (downloadMode === 'auto') {
            // Trigger automatic download using quality profile
            const searchMediaType = mediaType === 'both' ? 'ebook' : mediaType
            automaticSearch(bookId, searchMediaType).catch(console.error)
          } else if (downloadMode === 'manual') {
            // Navigate to book page for interactive search
            navigate(`/books/${bookId}`)
          }
          // 'none' mode: just close the modal, book is added without download
        }}
      />
    </div>
  )
}

// Book Result Card Component
function BookResultCard({ 
  result, 
  onAdd, 
  isAdding,
  onClick
}: { 
  result: SearchResult
  onAdd: (result: SearchResult) => void
  isAdding: boolean
  onClick?: (result: SearchResult) => void
}) {
  return (
    <div 
      className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors cursor-pointer"
      onClick={() => onClick?.(result)}
    >
      {result.coverUrl ? (
        <img
          src={result.coverUrl}
          alt={result.title}
          className="w-20 h-28 object-cover rounded shadow-md"
        />
      ) : (
        <div className="w-20 h-28 bg-muted rounded flex items-center justify-center">
          <BookOpen className="h-8 w-8 text-muted-foreground" />
        </div>
      )}
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium line-clamp-1">{result.title}</h3>
            <p className="text-sm text-muted-foreground">{result.author}</p>
          </div>
          
          {result.inLibrary ? (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0">
              <Check className="h-3 w-3" />
              In Library
            </Badge>
          ) : (
            <Button
              size="sm"
              onClick={(e) => { e.stopPropagation(); onAdd(result) }}
              disabled={isAdding}
              className="shrink-0"
            >
              <Plus className="h-4 w-4 mr-1" />
              Add
            </Button>
          )}
        </div>
        
        <div className="flex items-center gap-3 mt-2">
          {result.rating > 0 && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <Star className="h-4 w-4 fill-yellow-500 text-yellow-500" />
              {result.rating.toFixed(1)}
            </div>
          )}
          {result.releaseYear && (
            <span className="text-sm text-muted-foreground">
              {result.releaseYear}
            </span>
          )}
          {result.isbn && (
            <span className="text-xs text-muted-foreground">
              ISBN: {result.isbn}
            </span>
          )}
        </div>
        
        {result.description && (
          <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
            {result.description}
          </p>
        )}
      </div>
    </div>
  )
}

function AuthorResultCard({ 
  result, 
  onAdd, 
  isAdding,
  onClick,
  recentlyAdded = false
}: { 
  result: AuthorSearchResult
  onAdd: (result: AuthorSearchResult) => void
  isAdding: boolean
  onClick?: (result: AuthorSearchResult) => void
  recentlyAdded?: boolean
}) {
  const isFollowing = result.inLibrary || recentlyAdded

  return (
    <div 
      className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors cursor-pointer"
      onClick={() => onClick?.(result)}
    >
      {result.imageUrl ? (
        <img
          src={result.imageUrl}
          alt={result.name}
          className="w-16 h-16 object-cover rounded-full"
        />
      ) : (
        <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
          <User className="h-8 w-8 text-muted-foreground" />
        </div>
      )}
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium">{result.name}</h3>
            <p className="text-sm text-muted-foreground">{result.booksCount} books</p>
          </div>
          
          {isFollowing ? (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0 bg-green-600/20 text-green-400 border-green-600">
              <Check className="h-3 w-3" />
              Following
            </Badge>
          ) : (
            <Button
              size="sm"
              onClick={(e) => { e.stopPropagation(); onAdd(result) }}
              disabled={isAdding}
              className="shrink-0"
            >
              <Plus className="h-4 w-4 mr-1" />
              Add
            </Button>
          )}
        </div>
        
        {result.biography && (
          <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
            {result.biography}
          </p>
        )}
      </div>
    </div>
  )
}

// Series Result Card Component
function SeriesResultCard({ 
  result,
  onClick
}: { 
  result: SeriesSearchResult
  onClick?: (result: SeriesSearchResult) => void
}) {
  return (
    <div 
      className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors cursor-pointer"
      onClick={() => onClick?.(result)}
    >
      <div className="w-16 h-16 bg-gradient-to-br from-primary/20 to-primary/5 rounded-lg flex items-center justify-center">
        <Library className="h-8 w-8 text-primary" />
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium">{result.name}</h3>
            {result.authorName && (
              <p className="text-sm text-muted-foreground">by {result.authorName}</p>
            )}
          </div>
          
          {result.inLibrary && (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0">
              <Check className="h-3 w-3" />
              In Library
            </Badge>
          )}
        </div>
        
        <p className="text-sm text-muted-foreground mt-1">
          {result.booksCount} books in series
        </p>
      </div>
    </div>
  )
}

// List Result Card Component
function ListResultCard({ result }: { result: ListSearchResult }) {
  return (
    <div className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors">
      <div className="w-16 h-16 bg-gradient-to-br from-blue-500/20 to-blue-500/5 rounded-lg flex items-center justify-center">
        <ListIcon className="h-8 w-8 text-blue-500" />
      </div>
      
      <div className="flex-1 min-w-0">
        <h3 className="font-medium">{result.name}</h3>
        {result.username && (
          <p className="text-sm text-muted-foreground">by @{result.username}</p>
        )}
        <p className="text-sm text-muted-foreground mt-1">
          {result.booksCount} books
        </p>
        {result.description && (
          <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
            {result.description}
          </p>
        )}
      </div>
    </div>
  )
}

// No Results Component
function NoResults() {
  return (
    <div className="flex flex-col items-center justify-center py-16">
      <BookOpen className="h-16 w-16 text-muted-foreground mb-4" />
      <p className="text-lg text-muted-foreground">No results found</p>
      <p className="text-sm text-muted-foreground mt-1">
        Try a different search term or category
      </p>
    </div>
  )
}

