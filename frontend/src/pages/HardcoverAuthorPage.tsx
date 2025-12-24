import { useState, useMemo } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  User,
  BookOpen,
  Plus,
  Loader2,
  Eye,
  AlertCircle,
  CheckCircle2,
  X,
  Book as BookIcon,
  Library
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getHardcoverAuthor, addHardcoverBook, addAuthor, deleteBook, invalidateAllBookQueries, type HardcoverAuthorDetail, type Book } from '@/api/client'
import { CatalogBookCard } from '@/components/library/CatalogBookCard'
import { 
  BookSortFilter, 
  sortBooks, 
  filterBooks, 
  getDefaultSortFilterState,
  type SortFilterState 
} from '@/components/library/BookSortFilter'

interface Notification {
  id: string
  type: 'success' | 'error' | 'info'
  message: string
}

export function HardcoverAuthorPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [addingBooks, setAddingBooks] = useState<Set<string>>(new Set())
  const [deletingBooks, setDeletingBooks] = useState<Set<number>>(new Set())
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [sortFilterState, setSortFilterState] = useState<SortFilterState>(
    getDefaultSortFilterState(false)
  )

  const { data: author, isLoading, error } = useQuery({
    queryKey: ['hardcoverAuthor', id],
    queryFn: () => getHardcoverAuthor(id!),
    enabled: !!id,
  })

  const addAuthorMutation = useMutation({
    mutationFn: () => addAuthor(id!, true, false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hardcoverAuthor', id] })
      queryClient.invalidateQueries({ queryKey: ['authors'] })
      addNotification('success', `Now following ${author?.name || 'author'}`)
    },
    onError: (error: Error) => {
      const message = error.message.includes('409') || error.message.includes('Conflict')
        ? 'Already following this author'
        : error.message || 'Failed to follow author'
      addNotification('error', message)
    },
  })

  const addNotification = (type: 'success' | 'error' | 'info', message: string) => {
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
    mutationFn: (bookId: string) => addHardcoverBook(bookId, { monitored: true }),
    onSuccess: (response, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })

      queryClient.setQueryData<HardcoverAuthorDetail>(['hardcoverAuthor', id], (old) => {
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
      const bookTitle = author?.books.find(b => b.id === bookId)?.title || 'Book'
      addNotification('success', `Added "${bookTitle}" to library`)
    },
    onError: (err: Error, bookId) => {
      setAddingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
      if (err.message.includes('409') || err.message.includes('Conflict') || err.message.includes('already in library')) {
        invalidateAllBookQueries(queryClient)
        addNotification('info', 'Book is already in your library')
      } else {
        addNotification('error', 'Failed to add book')
      }
    },
  })

  const deleteBookMutation = useMutation({
    mutationFn: (bookId: number) => deleteBook(bookId),
    onSuccess: (_, bookId) => {
      setDeletingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })

      queryClient.setQueryData<HardcoverAuthorDetail>(['hardcoverAuthor', id], (old) => {
        if (!old) return old
        return {
          ...old,
          books: old.books.map((b) => 
            b.libraryBook?.id === bookId 
              ? { ...b, inLibrary: false, libraryBook: undefined } 
              : b
          )
        }
      })

      invalidateAllBookQueries(queryClient)
      addNotification('success', 'Book removed from library')
    },
    onError: (error: Error, bookId) => {
      setDeletingBooks(prev => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
      
      if (error.message.includes('404')) {
         queryClient.setQueryData<HardcoverAuthorDetail>(['hardcoverAuthor', id], (old) => {
            if (!old) return old
            return {
              ...old,
              books: old.books.map((b) => 
                b.libraryBook?.id === bookId 
                  ? { ...b, inLibrary: false, libraryBook: undefined } 
                  : b
              )
            }
          })
          addNotification('success', 'Book removed from library')
          return
      }
      addNotification('error', error.message || 'Failed to remove book')
    },
  })



  const handleAddBook = (bookId: string) => {
    setAddingBooks(prev => new Set(prev).add(bookId))
    addBookMutation.mutate(bookId)
  }

  const handleDeleteBook = (bookId: number) => {
    setDeletingBooks(prev => new Set(prev).add(bookId))
    deleteBookMutation.mutate(bookId)
  }

  const handleAddAllBooks = () => {
    if (!author?.books) return
    const booksToAdd = author.books.filter(b => !b.inLibrary)
    booksToAdd.forEach(book => handleAddBook(book.id))
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
                  : notification.type === 'info'
                  ? 'bg-blue-500/90 text-white'
                  : 'bg-red-500/90 text-white'
              }`}
            >
              {notification.type === 'success' ? (
                <CheckCircle2 className="h-5 w-5" />
              ) : notification.type === 'info' ? (
                <AlertCircle className="h-5 w-5" />
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
            {processedBooks.map((book, index) => (
              <CatalogBookCard
                key={book.id || `book-${index}`}
                hardcoverId={book.id}
                title={book.title}
                coverUrl={book.coverUrl}
                rating={book.rating}
                releaseYear={book.releaseYear}
                seriesIndex={book.seriesIndex}
                seriesName={book.seriesName}
                hasAudiobook={book.hasAudiobook}
                hasEbook={book.hasEbook}
                editionCount={book.editionCount}
                inLibrary={book.inLibrary}
                libraryBook={book.libraryBook}
                onAdd={handleAddBook}
                isAdding={addingBooks.has(book.id)}
                onDelete={handleDeleteBook}
                isDeleting={book.libraryBook?.id ? deletingBooks.has(book.libraryBook.id) : false}
              />
            ))}
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
        </div>
      </div>
    </div>
  )
}

export default HardcoverAuthorPage
