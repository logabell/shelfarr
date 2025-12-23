import { useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Book,
  User,
  Star,
  Calendar,
  BookOpen,
  Plus,
  Loader2,
  Check,
  Library,
  Hash,
  Headphones,
  Layers
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getHardcoverBook, addHardcoverBook } from '@/api/client'

export function HardcoverBookPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [isAdding, setIsAdding] = useState(false)

  const { data: book, isLoading, error } = useQuery({
    queryKey: ['hardcoverBook', id],
    queryFn: () => getHardcoverBook(id!),
    enabled: !!id,
  })

  const addMutation = useMutation({
    mutationFn: () => addHardcoverBook(id!, { monitored: true }),
    onSuccess: (result) => {
      setIsAdding(false)
      queryClient.invalidateQueries({ queryKey: ['hardcoverBook', id] })
      // Navigate to the library book page
      navigate(`/books/${result.bookId}`)
    },
    onError: () => {
      setIsAdding(false)
    },
  })

  const handleAddToLibrary = () => {
    setIsAdding(true)
    addMutation.mutate()
  }

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !book) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center max-w-md">
          <Book className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h1 className="text-2xl font-bold mb-2">Book Not Found</h1>
          <p className="text-muted-foreground mb-4">
            Unable to load book details from Hardcover.
          </p>
          <Button variant="outline" onClick={() => navigate(-1)}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Book Preview" subtitle="From Hardcover.app" />

      <div className="flex-1 overflow-auto">
        {/* Hero Section */}
        <div className="relative bg-gradient-to-b from-card to-background">
          <div className="max-w-5xl mx-auto px-6 py-8">
            {/* Back button */}
            <button
              onClick={() => navigate(-1)}
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back
            </button>

            <div className="flex gap-8">
              {/* Book Cover */}
              <div className="shrink-0">
                {book.coverUrl ? (
                  <img
                    src={book.coverUrl}
                    alt={book.title}
                    className="w-48 h-72 object-cover rounded-xl shadow-2xl"
                  />
                ) : (
                  <div className="w-48 h-72 bg-gradient-to-br from-primary/20 to-primary/5 rounded-xl flex items-center justify-center shadow-2xl">
                    <Book className="h-20 w-20 text-primary" />
                  </div>
                )}
              </div>

              {/* Book Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h1 className="text-3xl font-bold mb-2">{book.title}</h1>
                    
                    {/* Author Link */}
                    {book.authorName && (
                      <Link
                        to={`/hardcover/author/${book.authorId}`}
                        className="inline-flex items-center gap-2 text-lg text-muted-foreground hover:text-primary transition-colors"
                      >
                        <User className="h-4 w-4" />
                        {book.authorName}
                      </Link>
                    )}
                  </div>

                  {/* Add to Library Button */}
                  <div>
                    {book.inLibrary ? (
                      <Badge variant="secondary" className="flex items-center gap-1 text-sm px-3 py-1.5">
                        <Check className="h-4 w-4" />
                        In Library
                      </Badge>
                    ) : (
                      <Button
                        size="lg"
                        onClick={handleAddToLibrary}
                        disabled={isAdding || addMutation.isPending}
                      >
                        {isAdding || addMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin mr-2" />
                        ) : (
                          <Plus className="h-4 w-4 mr-2" />
                        )}
                        Add to Library
                      </Button>
                    )}
                  </div>
                </div>

                {/* Metadata */}
                <div className="flex flex-wrap items-center gap-4 mt-4 text-sm">
                  {/* Rating */}
                  {book.rating > 0 && (
                    <div className="flex items-center gap-1 text-yellow-500">
                      <Star className="h-4 w-4 fill-current" />
                      <span className="font-medium">{book.rating.toFixed(1)}</span>
                    </div>
                  )}

                  {/* Release Date */}
                  {book.releaseDate && (
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Calendar className="h-4 w-4" />
                      <span>
                        {new Date(book.releaseDate).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                        })}
                      </span>
                    </div>
                  )}

                  {/* Page Count */}
                  {book.pageCount && book.pageCount > 0 && (
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <BookOpen className="h-4 w-4" />
                      <span>{book.pageCount} pages</span>
                    </div>
                  )}

                  {/* ISBN */}
                  {book.isbn13 && (
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Hash className="h-4 w-4" />
                      <span>ISBN: {book.isbn13}</span>
                    </div>
                  )}
                </div>

                {/* Format Availability */}
                {(book.hasAudiobook || book.editionCount && book.editionCount > 1) && (
                  <div className="flex flex-wrap items-center gap-2 mt-4">
                    {book.hasAudiobook && (
                      <Badge variant="secondary" className="flex items-center gap-1">
                        <Headphones className="h-3 w-3" />
                        Audiobook
                        {book.audioDuration && book.audioDuration > 0 && (
                          <span className="text-muted-foreground ml-1">
                            ({Math.floor(book.audioDuration / 3600)}h {Math.floor((book.audioDuration % 3600) / 60)}m)
                          </span>
                        )}
                      </Badge>
                    )}
                    {book.hasEbook && (
                      <Badge variant="secondary" className="flex items-center gap-1">
                        <Book className="h-3 w-3" />
                        Ebook
                      </Badge>
                    )}
                    {book.editionCount && book.editionCount > 1 && (
                      <Badge variant="outline" className="flex items-center gap-1">
                        <Layers className="h-3 w-3" />
                        {book.editionCount} editions
                      </Badge>
                    )}
                  </div>
                )}

                {/* Series Info */}
                {book.seriesName && (
                  <div className="mt-4">
                    <Link
                      to={`/hardcover/series/${book.seriesId}`}
                      className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors"
                    >
                      <Library className="h-4 w-4" />
                      {book.seriesName}
                      {book.seriesIndex && ` #${book.seriesIndex}`}
                    </Link>
                  </div>
                )}

                {/* Genres */}
                {book.genres && book.genres.length > 0 && (
                  <div className="mt-4">
                    <div className="flex flex-wrap gap-2">
                      {book.genres.map((genre, index) => (
                        <Badge key={index} variant="secondary">
                          {genre}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {/* Description */}
                {book.description && (
                  <div className="mt-6">
                    <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide mb-2">
                      Synopsis
                    </h2>
                    <div 
                      className="text-foreground prose prose-sm prose-invert max-w-none"
                      dangerouslySetInnerHTML={{ __html: book.description }}
                    />
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Additional Info Section */}
        <div className="max-w-5xl mx-auto px-6 py-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Book Details Card */}
            <div className="bg-card border rounded-lg p-6">
              <h3 className="font-semibold mb-4">Book Details</h3>
              <dl className="space-y-3 text-sm">
                {book.authorName && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Author</dt>
                    <dd>
                      <Link
                        to={`/hardcover/author/${book.authorId}`}
                        className="text-primary hover:underline"
                      >
                        {book.authorName}
                      </Link>
                    </dd>
                  </div>
                )}
                {book.releaseYear && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Published</dt>
                    <dd>{book.releaseYear}</dd>
                  </div>
                )}
                {book.pageCount && book.pageCount > 0 && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Pages</dt>
                    <dd>{book.pageCount}</dd>
                  </div>
                )}
                {book.isbn && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">ISBN-10</dt>
                    <dd className="font-mono text-xs">{book.isbn}</dd>
                  </div>
                )}
                {book.isbn13 && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">ISBN-13</dt>
                    <dd className="font-mono text-xs">{book.isbn13}</dd>
                  </div>
                )}
              </dl>
            </div>

            {/* Series Info Card */}
            {book.seriesName && (
              <div className="bg-card border rounded-lg p-6">
                <h3 className="font-semibold mb-4">Series</h3>
                <Link
                  to={`/hardcover/series/${book.seriesId}`}
                  className="flex items-center gap-3 p-3 rounded-lg bg-background hover:bg-muted transition-colors"
                >
                  <Library className="h-8 w-8 text-primary" />
                  <div>
                    <p className="font-medium">{book.seriesName}</p>
                    {book.seriesIndex && (
                      <p className="text-sm text-muted-foreground">
                        Book #{book.seriesIndex}
                      </p>
                    )}
                  </div>
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default HardcoverBookPage
