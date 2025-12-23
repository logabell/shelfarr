import { useState, useMemo } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  User,
  BookOpen,
  Plus,
  Loader2,
  Star,
  Eye,
  AlertCircle,
  CheckCircle2,
  X,
  Book as BookIcon,
  Download,
  EyeOff,
  Clock,
  Library,
  BookMarked
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getHardcoverAuthor, addBook, addAuthor } from '@/api/client'
import { 
  BookSortFilter, 
  sortBooks, 
  filterBooks, 
  getDefaultSortFilterState,
  type SortFilterState 
} from '@/components/library/BookSortFilter'

interface Notification {
  id: string
  type: 'success' | 'error'
  message: string
}

export function HardcoverAuthorPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [addingBooks, setAddingBooks] = useState<Set<string>>(new Set())
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [sortFilterState, setSortFilterState] = useState<SortFilterState>(
    getDefaultSortFilterState(false) // false = not a series view
  )
  const [showPhysical, setShowPhysical] = useState(false)

  const { data: author, isLoading, error } = useQuery({
    queryKey: ['hardcoverAuthor', id, showPhysical],
    queryFn: () => getHardcoverAuthor(id!, showPhysical),
    enabled: !!id,
  })

  const addAuthorMutation = useMutation({
    mutationFn: () => addAuthor(id!, true, false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hardcoverAuthor', id] })
      queryClient.invalidateQueries({ queryKey: ['authors'] })
    },
  })

  const addNotification = (type: 'success' | 'error', message: string) => {
    const notification: Notification = {
      id: Math.random().toString(36).substring(7),
      type,
      message,
    }
    setNotifications(prev => [...prev, notification])
    setTimeout(() => {
      setNotifications(prev => prev.filter(n => n.id !== notification.id))
    }, 3000)
  }

  const removeNotification = (notificationId: string) => {
    setNotifications(prev => prev.filter(n => n.id !== notificationId))
  }

  const addBookMutation = useMutation({
    mutationFn: (bookId: string) => addBook(bookId, true),
    onSuccess: (result, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
      queryClient.invalidateQueries({ queryKey: ['hardcoverAuthor', id] })
      addNotification('success', `Added "${result.title}" to library`)
    },
    onError: (err: Error, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
      const message = err.message.includes('409') || err.message.includes('Conflict')
        ? 'Book is already in library'
        : 'Failed to add book'
      addNotification('error', message)
    },
  })

  const handleAddBook = (bookId: string) => {
    setAddingBooks(prev => new Set(prev).add(bookId))
    addBookMutation.mutate(bookId)
  }

  const handleAddAllBooks = () => {
    if (!author?.books) return
    const booksToAdd = author.books.filter(b => !b.inLibrary)
    booksToAdd.forEach(book => handleAddBook(book.id))
  }

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'downloaded': return 'bg-green-500'
      case 'downloading': return 'bg-sky-500'
      case 'missing': return 'bg-red-500'
      case 'unreleased': return 'bg-purple-500'
      default: return 'bg-neutral-500'
    }
  }

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'downloaded': return <CheckCircle2 className="w-4 h-4 text-green-400" />
      case 'downloading': return <Download className="w-4 h-4 text-sky-400" />
      case 'missing': return <AlertCircle className="w-4 h-4 text-red-400" />
      case 'unreleased': return <Clock className="w-4 h-4 text-purple-400" />
      default: return <AlertCircle className="w-4 h-4 text-neutral-400" />
    }
  }

  // Wrap in useMemo to prevent unnecessary re-renders
  const allBooks = useMemo(() => author?.books || [], [author?.books])
  const booksInLibrary = allBooks.filter(b => b.inLibrary).length
  const totalBooks = allBooks.length
  const missingBooks = totalBooks - booksInLibrary

  // Group books by series
  const seriesGroups = useMemo(() => {
    const seriesMap = new Map<string, { 
      id: string
      name: string
      books: typeof allBooks 
    }>()

    allBooks.forEach(book => {
      if (book.seriesId && book.seriesName) {
        const existing = seriesMap.get(book.seriesId)
        if (existing) {
          existing.books.push(book)
        } else {
          seriesMap.set(book.seriesId, {
            id: book.seriesId,
            name: book.seriesName,
            books: [book],
          })
        }
      }
    })

    // Sort books within each series by series index
    seriesMap.forEach(series => {
      series.books.sort((a, b) => (a.seriesIndex || 0) - (b.seriesIndex || 0))
    })

    return Array.from(seriesMap.values()).sort((a, b) => a.name.localeCompare(b.name))
  }, [allBooks])

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

  if (error || !author) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center max-w-md">
          <User className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h1 className="text-2xl font-bold mb-2">Author Not Found</h1>
          <p className="text-muted-foreground mb-4">
            Unable to load author details from Hardcover.
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
      <Topbar title={author.name} subtitle="Author Preview" />

      {/* Notifications */}
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
        {/* Back Link */}
        <Link
          to="/search"
          className="inline-flex items-center gap-2 text-neutral-400 hover:text-neutral-200 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Search
        </Link>

        {/* Author Header - matching series style */}
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
          <div className="flex items-start gap-6">
            {/* Author Image */}
            <div className="flex-shrink-0 w-32 h-32 rounded-full overflow-hidden bg-gradient-to-br from-amber-900 to-amber-700">
              {author.imageUrl ? (
                <img
                  src={author.imageUrl}
                  alt={author.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  <User className="w-16 h-16 text-amber-400" />
                </div>
              )}
            </div>

            {/* Author Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-neutral-100">{author.name}</h1>
                <div className="flex items-center gap-2">
                  {author.inLibrary ? (
                    <Badge variant="secondary" className="flex items-center gap-1">
                      <Eye className="h-3 w-3" />
                      Following
                    </Badge>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => addAuthorMutation.mutate()}
                      disabled={addAuthorMutation.isPending}
                    >
                      {addAuthorMutation.isPending ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      ) : (
                        <Plus className="h-4 w-4 mr-2" />
                      )}
                      Follow Author
                    </Button>
                  )}
                  {missingBooks > 0 && (
                    <Button
                      onClick={handleAddAllBooks}
                      disabled={addingBooks.size > 0}
                      size="sm"
                    >
                      {addingBooks.size > 0 ? (
                        <Loader2 className="w-4 h-4 animate-spin mr-2" />
                      ) : (
                        <Plus className="w-4 h-4 mr-2" />
                      )}
                      Add All Missing ({missingBooks})
                    </Button>
                  )}
                </div>
              </div>

              {/* Stats */}
              <div className="flex items-center gap-6 mt-4">
                <div className="flex items-center gap-2 text-neutral-400">
                  <BookIcon className="w-4 h-4" />
                  <span>{totalBooks} Books</span>
                </div>
                <div className="flex items-center gap-2 text-sky-400">
                  <BookOpen className="w-4 h-4" />
                  <span>{booksInLibrary} In Library</span>
                </div>
                {missingBooks > 0 && (
                  <div className="flex items-center gap-2 text-amber-400">
                    <Plus className="w-4 h-4" />
                    <span>{missingBooks} Missing</span>
                  </div>
                )}
              </div>

              {/* Progress Bar */}
              <div className="mt-4">
                <div className="flex items-center justify-between text-sm mb-1">
                  <span className="text-neutral-400">Collection Progress</span>
                  <span className="text-neutral-300">
                    {totalBooks > 0 ? Math.round((booksInLibrary / totalBooks) * 100) : 0}%
                  </span>
                </div>
                <div className="h-2 bg-neutral-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-green-600 to-green-400 rounded-full transition-all"
                    style={{ width: `${totalBooks > 0 ? (booksInLibrary / totalBooks) * 100 : 0}%` }}
                  />
                </div>
              </div>

              {/* Biography */}
              {author.biography && (
                <p className="text-neutral-400 mt-4 line-clamp-3">{author.biography}</p>
              )}
            </div>
          </div>
        </div>

        {/* Series Section */}
        {seriesGroups.length > 0 && (
          <div className="mb-6">
            <h2 className="text-xl font-semibold text-neutral-100 mb-4">Series ({seriesGroups.length})</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {seriesGroups.map(series => {
                const inLibraryInSeries = series.books.filter(b => b.inLibrary).length
                const downloadedInSeries = series.books.filter(b => b.libraryBook?.status === 'downloaded').length
                
                return (
                  <Link
                    key={series.id}
                    to={`/hardcover/series/${series.id}`}
                    className="flex items-center gap-4 p-4 rounded-lg bg-neutral-800/50 border border-neutral-700 hover:border-sky-500/50 transition-colors"
                  >
                    <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-amber-900/50 to-amber-700/50 flex items-center justify-center shrink-0">
                      <Library className="w-6 h-6 text-amber-400" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-neutral-200 truncate">{series.name}</h3>
                      <div className="flex items-center gap-3 text-xs mt-1">
                        <span className="text-neutral-400">{series.books.length} books</span>
                        {inLibraryInSeries > 0 && (
                          <span className="text-sky-400">{inLibraryInSeries} in library</span>
                        )}
                        {downloadedInSeries > 0 && (
                          <span className="text-green-400">{downloadedInSeries} downloaded</span>
                        )}
                      </div>
                    </div>
                  </Link>
                )
              })}
            </div>
          </div>
        )}

        {/* Physical-Only Books Notice */}
        {author.physicalOnlyCount > 0 && (
          <div className="bg-neutral-800/30 border border-neutral-700 rounded-lg p-3 flex items-center justify-between">
            <div className="flex items-center gap-3 text-neutral-400">
              <BookMarked className="w-4 h-4 text-amber-500" />
              <span className="text-sm">
                {author.physicalOnlyCount} {author.physicalOnlyCount === 1 ? 'book has' : 'books have'} physical editions only
              </span>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowPhysical(!showPhysical)}
              className="text-neutral-400 hover:text-neutral-200"
            >
              {showPhysical ? (
                <>
                  <EyeOff className="w-4 h-4 mr-2" />
                  Hide Physical-Only
                </>
              ) : (
                <>
                  <Eye className="w-4 h-4 mr-2" />
                  Show Physical-Only
                </>
              )}
            </Button>
          </div>
        )}

        {/* Books Grid */}
        <div>
          <h2 className="text-xl font-semibold text-neutral-100 mb-4">Books</h2>

          {/* Sort/Filter Toolbar */}
          <BookSortFilter
            state={sortFilterState}
            onChange={setSortFilterState}
            showSeriesIndex={false}
            totalCount={allBooks.length}
            filteredCount={processedBooks.length}
            className="mb-4"
          />

          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {processedBooks.map((book, index) => {
              const isAdding = addingBooks.has(book.id)

              if (!book.inLibrary) {
                // Not in library - show preview link and add button
                return (
                  <div
                    key={book.id || `book-${index}`}
                    className="group relative bg-neutral-800/30 border-2 border-dashed border-neutral-700 rounded-xl overflow-hidden hover:border-sky-500/50 transition-colors"
                  >
                    {/* Cover - clickable to preview */}
                    <Link
                      to={`/hardcover/book/${book.id}`}
                      className="block aspect-[2/3] relative"
                    >
                      {book.coverUrl ? (
                        <img
                          src={book.coverUrl}
                          alt={book.title}
                          className="w-full h-full object-cover opacity-60 group-hover:opacity-80 transition-opacity"
                        />
                      ) : (
                        <div className="w-full h-full bg-gradient-to-br from-neutral-700 to-neutral-800 flex items-center justify-center">
                          <BookIcon className="w-12 h-12 text-neutral-600" />
                        </div>
                      )}
                    </Link>

                    {/* Add Button */}
                    <div className="absolute bottom-16 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button
                        size="sm"
                        onClick={(e) => {
                          e.preventDefault()
                          handleAddBook(book.id)
                        }}
                        disabled={isAdding}
                      >
                        {isAdding ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <>
                            <Plus className="w-4 h-4 mr-1" />
                            Add
                          </>
                        )}
                      </Button>
                    </div>

                    {/* Info */}
                    <div className="p-3">
                      <Link
                        to={`/hardcover/book/${book.id}`}
                        className="font-medium text-neutral-400 text-sm line-clamp-2 hover:text-sky-400 transition-colors"
                      >
                        {book.title}
                      </Link>
                      <div className="flex items-center gap-2 mt-1">
                        {book.rating > 0 && (
                          <div className="flex items-center gap-0.5 text-xs text-neutral-500">
                            <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
                            {book.rating.toFixed(1)}
                          </div>
                        )}
                        {book.releaseYear && (
                          <span className="text-xs text-neutral-500">{book.releaseYear}</span>
                        )}
                      </div>
                      <p className="text-xs text-neutral-600 mt-1">Not in library</p>
                    </div>
                  </div>
                )
              }

              // In library - show regular card
              // Route to library book page if we have a library ID, otherwise to Hardcover preview
              const bookLink = book.libraryBook?.id 
                ? `/books/${book.libraryBook.id}` 
                : `/hardcover/book/${book.id}`
              
              return (
                <Link
                  key={book.id || `book-${index}`}
                  to={bookLink}
                  className="group relative bg-neutral-800/50 rounded-xl overflow-hidden hover:ring-2 hover:ring-sky-500/50 transition-all"
                >
                  {/* Cover */}
                  <div className="aspect-[2/3] relative">
                    {book.coverUrl ? (
                      <img
                        src={book.coverUrl}
                        alt={book.title}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-gradient-to-br from-neutral-700 to-neutral-800 flex items-center justify-center">
                        <BookIcon className="w-12 h-12 text-neutral-600" />
                      </div>
                    )}

                    {/* Status Indicator */}
                    <div className={`absolute top-2 right-2 w-3 h-3 rounded-full ${getStatusColor(book.libraryBook?.status)}`} />

                    {/* Monitor Badge */}
                    {book.libraryBook?.monitored === false && (
                      <div className="absolute bottom-2 right-2 bg-neutral-900/80 rounded p-1">
                        <EyeOff className="w-3 h-3 text-neutral-400" />
                      </div>
                    )}
                  </div>

                  {/* Info */}
                  <div className="p-3">
                    <h3 className="font-medium text-neutral-200 text-sm line-clamp-2 group-hover:text-sky-400 transition-colors">
                      {book.title}
                    </h3>
                    <div className="flex items-center gap-2 mt-2">
                      {getStatusIcon(book.libraryBook?.status)}
                      <span className="text-xs text-neutral-400 capitalize">{book.libraryBook?.status || 'unknown'}</span>
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      {book.rating > 0 && (
                        <div className="flex items-center gap-0.5 text-xs text-neutral-400">
                          <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
                          {book.rating.toFixed(1)}
                        </div>
                      )}
                      {book.releaseYear && (
                        <span className="text-xs text-neutral-400">{book.releaseYear}</span>
                      )}
                    </div>
                    {book.seriesName && (
                      <p className="text-xs text-neutral-500 mt-1 line-clamp-1">
                        {book.seriesName} #{book.seriesIndex}
                      </p>
                    )}
                  </div>
                </Link>
              )
            })}
          </div>
          
          {processedBooks.length === 0 && allBooks.length > 0 && (
            <div className="text-center py-12 border border-dashed border-neutral-700 rounded-lg mt-4">
              <BookIcon className="w-12 h-12 mx-auto text-neutral-600 mb-4" />
              <p className="text-neutral-400">No books match your filter</p>
              <button 
                onClick={() => setSortFilterState(getDefaultSortFilterState(false))}
                className="text-sky-400 hover:text-sky-300 text-sm mt-2"
              >
                Clear filters
              </button>
            </div>
          )}
          
          {processedBooks.length === 0 && allBooks.length === 0 && !showPhysical && author.physicalOnlyCount > 0 && (
            <div className="text-center py-12 border border-dashed border-neutral-700 rounded-lg mt-4">
              <BookMarked className="w-12 h-12 mx-auto text-amber-500/50 mb-4" />
              <p className="text-neutral-400">No books with digital editions found</p>
              <p className="text-neutral-500 text-sm mt-1">
                {author.physicalOnlyCount} {author.physicalOnlyCount === 1 ? 'book' : 'books'} with physical editions only
              </p>
              <button 
                onClick={() => setShowPhysical(true)}
                className="text-sky-400 hover:text-sky-300 text-sm mt-3"
              >
                Show physical-only books
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default HardcoverAuthorPage
