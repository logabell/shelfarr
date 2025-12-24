import { useState, useMemo } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Library,
  BookOpen,
  Plus,
  Check,
  Loader2,
  Star,
  Book as BookIcon
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getHardcoverSeries, addHardcoverBook, invalidateAllBookQueries, type HardcoverBookDetail, type HardcoverSeriesDetail, type Book } from '@/api/client'
import { CatalogBookCard } from '@/components/library/CatalogBookCard'
import { 
  BookSortFilter, 
  sortBooks, 
  filterBooks, 
  getDefaultSortFilterState,
  type SortFilterState 
} from '@/components/library/BookSortFilter'

export function HardcoverSeriesPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [addingBooks, setAddingBooks] = useState<Set<string>>(new Set())
  const [sortFilterState, setSortFilterState] = useState<SortFilterState>(
    getDefaultSortFilterState(true) // true = for series (default sort by series index)
  )

  const { data: series, isLoading, error } = useQuery({
    queryKey: ['hardcoverSeries', id],
    queryFn: () => getHardcoverSeries(id!),
    enabled: !!id,
  })

  const addBookMutation = useMutation({
    mutationFn: (bookId: string) => addHardcoverBook(bookId, { monitored: true }),
    onSuccess: (response, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })

      queryClient.setQueryData<HardcoverSeriesDetail>(['hardcoverSeries', id], (old) => {
        if (!old) return old
        return {
          ...old,
          books: old.books.map((b) => 
            b.id === bookId 
              ? { 
                  ...b, 
                  inLibrary: true, 
                  libraryBook: { 
                    id: response.bookId, 
                    hardcoverId: bookId,
                    title: b.title,
                    coverUrl: b.coverUrl || '',
                    rating: b.rating,
                    description: b.description || '',
                    pageCount: b.pageCount || 0,
                    status: 'missing',
                    monitored: true,
                    hasEbook: b.hasEbook,
                    hasAudiobook: b.hasAudiobook,
                    isbn: b.isbn || ''
                  } as unknown as Book
                } 
              : b
          )
        }
      })

      invalidateAllBookQueries(queryClient)
    },
    onError: (error: Error, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
      if (error.message.includes('409') || error.message.includes('Conflict') || error.message.includes('already in library')) {
        invalidateAllBookQueries(queryClient)
      }
    },
  })

  const handleAddBook = (bookId: string) => {
    setAddingBooks(prev => new Set(prev).add(bookId))
    addBookMutation.mutate(bookId)
  }

  const handleAddMissingBooks = () => {
    if (!series?.books) return
    const booksToAdd = series.books.filter(b => !b.inLibrary)
    booksToAdd.forEach(book => handleAddBook(book.id))
  }

  // Wrap in useMemo to prevent unnecessary re-renders
  const allBooks = useMemo(() => series?.books || [], [series?.books])
  const booksInLibrary = allBooks.filter(b => b.inLibrary).length
  const totalBooks = allBooks.length

  // Apply sorting and filtering
  const processedBooks = useMemo(() => {
    const filtered = filterBooks(allBooks, sortFilterState.filterStatus)
    return sortBooks(filtered, sortFilterState.sortField, sortFilterState.sortOrder)
  }, [allBooks, sortFilterState])

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !series) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center max-w-md">
          <h1 className="text-2xl font-bold mb-2">Failed to Load Series</h1>
          <p className="text-muted-foreground mb-4">
            {error instanceof Error 
              ? error.message 
              : 'Please check your Hardcover API configuration in Settings'}
          </p>
          <Link to="/search" className="text-primary hover:underline">
            Return to Search
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Topbar title={series.name} subtitle="Series Preview" />

      <div className="flex-1 overflow-auto">
        {/* Hero Section */}
        <div className="relative bg-gradient-to-b from-card to-background">
          <div className="max-w-6xl mx-auto px-6 py-8">
            <Link
              to="/search"
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Search
            </Link>

            <div className="flex gap-8">
              {/* Series Icon */}
              <div className="shrink-0">
                <div className="w-40 h-40 bg-gradient-to-br from-primary/20 to-primary/5 rounded-xl flex items-center justify-center shadow-2xl">
                  <Library className="h-16 w-16 text-primary" />
                </div>
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h1 className="text-3xl font-bold mb-2">{series.name}</h1>
                    <p className="text-muted-foreground">
                      {series.booksCount} books
                      {booksInLibrary > 0 && (
                        <span className="text-primary"> • {booksInLibrary} in library</span>
                      )}
                    </p>
                  </div>

                  {series.inLibrary && (
                    <Badge variant="secondary" className="flex items-center gap-1">
                      <Check className="h-3 w-3" />
                      In Library
                    </Badge>
                  )}
                </div>

                {/* Progress Bar */}
                {totalBooks > 0 && (
                  <div className="mt-4">
                    <div className="flex justify-between text-sm text-muted-foreground mb-1">
                      <span>Collection Progress</span>
                      <span>{booksInLibrary} / {totalBooks}</span>
                    </div>
                    <div className="h-2 w-full max-w-md bg-muted rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-primary rounded-full transition-all"
                        style={{ width: `${(booksInLibrary / totalBooks) * 100}%` }}
                      />
                    </div>
                  </div>
                )}

                {/* Actions */}
                <div className="flex gap-2 mt-6">
                  {booksInLibrary < totalBooks && (
                    <Button 
                      onClick={handleAddMissingBooks}
                      disabled={addingBooks.size > 0}
                    >
                      {addingBooks.size > 0 ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Plus className="h-4 w-4" />
                      )}
                      Add Missing Books ({totalBooks - booksInLibrary})
                    </Button>
                  )}
                  {booksInLibrary === totalBooks && totalBooks > 0 && (
                    <Badge variant="secondary" className="text-base px-4 py-2">
                      <Check className="h-4 w-4 mr-2" />
                      Complete Collection
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Books Section */}
        <div className="max-w-6xl mx-auto px-6 py-8">
          <h2 className="text-xl font-semibold mb-4">Books in Series</h2>

          {/* Sort/Filter Toolbar */}
          <BookSortFilter
            state={sortFilterState}
            onChange={setSortFilterState}
            showSeriesIndex={true}
            showViewToggle={true}
            totalCount={allBooks.length}
            filteredCount={processedBooks.length}
            className="mb-4"
          />
          
          {processedBooks.length > 0 ? (
            sortFilterState.viewMode === 'grid' ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
                {processedBooks.map((book) => (
                  <CatalogBookCard
                    key={book.id}
                    hardcoverId={book.id}
                    title={book.title}
                    coverUrl={book.coverUrl}
                    rating={book.rating}
                    seriesIndex={book.seriesIndex}
                    authorName={book.authorName}
                    inLibrary={book.inLibrary}
                    libraryBook={book.libraryBook}
                    onAdd={handleAddBook}
                    isAdding={addingBooks.has(book.id)}
                  />
                ))}
              </div>
            ) : (
              <div className="space-y-3">
                {processedBooks.map((book) => (
                  <SeriesBookCard
                    key={book.id}
                    book={book}
                    onAdd={handleAddBook}
                    isAdding={addingBooks.has(book.id)}
                    onAuthorClick={(authorId) => navigate(`/hardcover/author/${authorId}`)}
                  />
                ))}
              </div>
            )
          ) : allBooks.length > 0 ? (
            <div className="text-center py-12 border border-dashed border-border rounded-lg">
              <BookIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No books match your filter</p>
              <button 
                onClick={() => setSortFilterState(getDefaultSortFilterState(true))}
                className="text-primary hover:underline text-sm mt-2"
              >
                Clear filters
              </button>
            </div>
          ) : (
            <div className="text-center py-12 border border-dashed border-border rounded-lg">
              <BookIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No books found in this series</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// Series Book Card Component
function SeriesBookCard({
  book,
  onAdd,
  isAdding,
  onAuthorClick
}: {
  book: HardcoverBookDetail
  onAdd: (id: string) => void
  isAdding: boolean
  onAuthorClick?: (authorId: string) => void
}) {
  return (
    <div className="flex gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary/50 transition-colors">
      {/* Series Index */}
      <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary font-bold text-lg shrink-0">
        {book.seriesIndex || '?'}
      </div>

      {/* Cover */}
      {book.coverUrl ? (
        <img
          src={book.coverUrl}
          alt={book.title}
          className="w-16 h-24 object-cover rounded shadow-md shrink-0"
        />
      ) : (
        <div className="w-16 h-24 bg-muted rounded flex items-center justify-center shrink-0">
          <BookOpen className="h-6 w-6 text-muted-foreground" />
        </div>
      )}

      {/* Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h3 className="font-medium line-clamp-1">{book.title}</h3>
            {book.authorName && (
              <p className="text-sm text-muted-foreground">
                by{' '}
                <button
                  className="text-primary hover:underline"
                  onClick={(e) => {
                    e.stopPropagation()
                    if (book.authorId) {
                      onAuthorClick?.(book.authorId)
                    }
                  }}
                >
                  {book.authorName}
                </button>
              </p>
            )}
          </div>

          {book.inLibrary ? (
            <Badge variant="secondary" className="flex items-center gap-1 shrink-0">
              <Check className="h-3 w-3" />
              In Library
            </Badge>
          ) : (
            <Button
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onAdd(book.id)
              }}
              disabled={isAdding}
              className="shrink-0"
            >
              {isAdding ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Plus className="h-4 w-4" />
              )}
              Add
            </Button>
          )}
        </div>

        <div className="flex items-center gap-3 mt-2">
          {book.rating > 0 && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <Star className="h-4 w-4 fill-yellow-500 text-yellow-500" />
              {book.rating.toFixed(1)}
            </div>
          )}
          {book.releaseYear && (
            <span className="text-sm text-muted-foreground">
              {book.releaseYear}
            </span>
          )}
          {book.pageCount && book.pageCount > 0 && (
            <span className="text-sm text-muted-foreground">
              {book.pageCount} pages
            </span>
          )}
        </div>

        {book.description && (
          <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
            {book.description}
          </p>
        )}
      </div>
    </div>
  )
}

export default HardcoverSeriesPage
