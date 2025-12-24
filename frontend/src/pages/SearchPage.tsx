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
  searchHardcoverSeries, 
  searchHardcoverLists,
  searchOpenLibrary,
  addOpenLibraryBook,
  addOpenLibraryAuthor,
  automaticSearch,
} from '@/api/client'
import { AddBookModal } from '@/components/search/AddBookModal'
import { 
  Search, 
  Plus, 
  Check, 
  BookOpen, 
  User, 
  Library, 
  ListIcon,
  Filter,
  X,
  Book
} from 'lucide-react'
import type { 
  SeriesSearchResult, 
  ListSearchResult,
  SearchType,
  OpenLibrarySearchResult,
  OpenLibraryAuthorSearchResult
} from '@/types'

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const initialQuery = searchParams.get('q') || ''
  const initialType = (searchParams.get('type') as SearchType) || 'book'
  
  const [query, setQuery] = useState(initialQuery)
  const [searchTerm, setSearchTerm] = useState(initialQuery)
  const [searchType, setSearchType] = useState<SearchType>(initialType)
  const [showFilters, setShowFilters] = useState(false)
  
  // Filter state
  const [yearRange, setYearRange] = useState<[number, number]>([1900, new Date().getFullYear()])
  const [minRating, setMinRating] = useState(0)
  
  // Book detail modal state
  const [selectedBookId, setSelectedBookId] = useState<string | null>(null)
  
  const queryClient = useQueryClient()

  // Book search query (Open Library)
  const { data: bookResults, isLoading: isLoadingBooks, error: errorBooks } = useQuery({
    queryKey: ['search', 'book', searchTerm],
    queryFn: () => searchOpenLibrary(searchTerm, 20, 'book'),
    enabled: searchTerm.length > 2 && searchType === 'book',
    retry: false,
  })

  // Author search query (Open Library)
  const { data: authorResults, isLoading: isLoadingAuthors, error: errorAuthors } = useQuery({
    queryKey: ['search', 'author', searchTerm],
    queryFn: () => searchOpenLibrary(searchTerm, 20, 'author'),
    enabled: searchTerm.length > 2 && searchType === 'author',
    retry: false,
  })

  // Series search query (Hardcover)
  const { data: seriesResults, isLoading: isLoadingSeries, error: errorSeries } = useQuery({
    queryKey: ['search', 'series', searchTerm],
    queryFn: () => searchHardcoverSeries(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'series',
    retry: false,
  })

  // List search query (Hardcover)
  const { data: listResults, isLoading: isLoadingLists, error: errorLists } = useQuery({
    queryKey: ['search', 'list', searchTerm],
    queryFn: () => searchHardcoverLists(searchTerm),
    enabled: searchTerm.length > 2 && searchType === 'list',
    retry: false,
  })

  const isLoading = isLoadingBooks || isLoadingAuthors || isLoadingSeries || isLoadingLists
  const searchError = errorBooks || errorAuthors || errorSeries || errorLists

  const filteredBooks = useMemo(() => {
    const books = bookResults?.results
    if (!books) return []
    
    return books.filter(book => {
      if (book.firstPublishYear && (book.firstPublishYear < yearRange[0] || book.firstPublishYear > yearRange[1])) {
        return false
      }
      if (book.rating && book.rating < minRating) {
        return false
      }
      return true
    })
  }, [bookResults, yearRange, minRating])

  const addOpenLibraryBookMutation = useMutation({
    mutationFn: (openLibraryWorkId: string) => addOpenLibraryBook(openLibraryWorkId, true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['library'] })
      queryClient.invalidateQueries({ queryKey: ['search'] })
    },
  })

  const addOpenLibraryAuthorMutation = useMutation({
    mutationFn: (openLibraryId: string) => addOpenLibraryAuthor(openLibraryId, true, false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['authors'] })
      queryClient.invalidateQueries({ queryKey: ['search'] })
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

  const handleSeriesClick = (result: SeriesSearchResult) => {
    navigate(`/hardcover/series/${result.id}`)
  }

  const resetFilters = () => {
    setYearRange([1900, new Date().getFullYear()])
    setMinRating(0)
  }

  const hasActiveFilters = yearRange[0] !== 1900 || yearRange[1] !== new Date().getFullYear() || minRating > 0

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Search" />
      
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
          <Tabs value={searchType} onValueChange={handleTypeChange} className="w-full max-w-4xl">
            <TabsList className="grid w-full grid-cols-4">
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
            {/* Book Results */}
            {searchType === 'book' && (
              <>
                {filteredBooks && filteredBooks.length > 0 ? (
                  <div className="space-y-3">
                    {filteredBooks.map((result) => (
                      <OpenLibraryResultCard 
                        key={result.key} 
                        result={result} 
                        onAdd={(workId) => addOpenLibraryBookMutation.mutate(workId)}
                        isAdding={addOpenLibraryBookMutation.isPending}
                      />
                    ))}
                  </div>
                ) : bookResults?.results.length === 0 ? (
                  <NoResults />
                ) : hasActiveFilters && bookResults && bookResults.results.length > 0 ? (
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
                {authorResults?.authorResults && authorResults.authorResults.length > 0 ? (
                  <div className="space-y-3">
                    {authorResults.authorResults.map((result) => (
                      <OpenLibraryAuthorResultCard 
                        key={result.key} 
                        result={result} 
                        onAdd={(authorId) => addOpenLibraryAuthorMutation.mutate(authorId)}
                        isAdding={addOpenLibraryAuthorMutation.isPending}
                      />
                    ))}
                  </div>
                ) : authorResults?.authorResults?.length === 0 ? (
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

            {/* Open Library Results - Removed as it is now the default */}
            
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

function OpenLibraryResultCard({ 
  result, 
  onAdd,
  isAdding 
}: { 
  result: OpenLibrarySearchResult 
  onAdd: (workId: string) => void
  isAdding: boolean
}) {
  return (
    <div className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors">
      {result.coverUrl ? (
        <img
          src={result.coverUrl}
          alt={result.title}
          className="w-20 h-28 object-cover rounded shadow-md"
        />
      ) : (
        <div className="w-20 h-28 bg-muted rounded flex items-center justify-center">
          <Book className="h-8 w-8 text-muted-foreground" />
        </div>
      )}
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium line-clamp-1">{result.title}</h3>
            {result.authors && result.authors.length > 0 && (
              <p className="text-sm text-muted-foreground">{result.authors.join(', ')}</p>
            )}
          </div>
          
          {result.inLibrary ? (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0">
              <Check className="h-3 w-3" />
              In Library
            </Badge>
          ) : (
            <Button
              size="sm"
              onClick={() => onAdd(result.key)}
              disabled={isAdding}
              className="shrink-0"
            >
              <Plus className="h-4 w-4 mr-1" />
              {isAdding ? 'Adding...' : 'Add'}
            </Button>
          )}
        </div>
        
        <div className="flex items-center gap-3 mt-2 flex-wrap">
          {result.firstPublishYear && (
            <span className="text-sm text-muted-foreground">
              {result.firstPublishYear}
            </span>
          )}
          {result.isbn13 && (
            <span className="text-xs text-muted-foreground">
              ISBN: {result.isbn13}
            </span>
          )}
          {result.language && (
            <span className="text-xs text-muted-foreground uppercase">
              {result.language}
            </span>
          )}
          {result.isEbook && (
            <Badge variant="outline" className="border-green-500 text-green-500 text-xs">
              Ebook
            </Badge>
          )}
          {result.hasEpub && (
            <Badge variant="outline" className="border-blue-500 text-blue-500 text-xs">
              EPUB
            </Badge>
          )}
          {result.hasFulltext && (
            <Badge variant="outline" className="text-xs">
              Full Text
            </Badge>
          )}
        </div>
        
        {result.subjects && result.subjects.length > 0 && (
          <p className="mt-2 text-xs text-muted-foreground line-clamp-1">
            {result.subjects.slice(0, 3).join(' • ')}
          </p>
        )}
      </div>
    </div>
  )
}

function OpenLibraryAuthorResultCard({ 
  result, 
  onAdd,
  isAdding 
}: { 
  result: OpenLibraryAuthorSearchResult 
  onAdd: (authorId: string) => void
  isAdding: boolean
}) {
  const navigate = useNavigate()

  const handleClick = () => {
    if (result.inLibrary) {
      navigate(`/authors?search=${encodeURIComponent(result.name)}`)
    }
  }

  return (
    <div 
      className={`flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors ${result.inLibrary ? 'cursor-pointer' : ''}`}
      onClick={result.inLibrary ? handleClick : undefined}
    >
      {result.imageUrl ? (
        <img
          src={result.imageUrl}
          alt={result.name}
          className="w-20 h-20 object-cover rounded-full shadow-md"
        />
      ) : (
        <div className="w-20 h-20 bg-muted rounded-full flex items-center justify-center">
          <User className="h-8 w-8 text-muted-foreground" />
        </div>
      )}
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium line-clamp-1">{result.name}</h3>
            {result.topWork && (
              <p className="text-sm text-muted-foreground">Known for: {result.topWork}</p>
            )}
          </div>
          
          {result.inLibrary ? (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0">
              <Check className="h-3 w-3" />
              In Library
            </Badge>
          ) : (
            <Button
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onAdd(result.key)
              }}
              disabled={isAdding}
              className="shrink-0"
            >
              <Plus className="h-4 w-4 mr-1" />
              {isAdding ? 'Adding...' : 'Add'}
            </Button>
          )}
        </div>
        
        <div className="flex items-center gap-3 mt-2 flex-wrap">
          {result.workCount && (
            <span className="text-sm text-muted-foreground">
              {result.workCount} works
            </span>
          )}
          {result.birthDate && (
            <span className="text-xs text-muted-foreground">
              Born: {result.birthDate}
            </span>
          )}
          {result.deathDate && (
            <span className="text-xs text-muted-foreground">
              Died: {result.deathDate}
            </span>
          )}
        </div>
        
        {result.topSubjects && result.topSubjects.length > 0 && (
          <p className="mt-2 text-xs text-muted-foreground line-clamp-1">
            {result.topSubjects.slice(0, 3).join(' • ')}
          </p>
        )}
      </div>
    </div>
  )
}

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

