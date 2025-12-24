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
  Layers,
  Trash2
} from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { FormatAvailability } from '@/components/ui/format-availability'
import { getHardcoverBook, addHardcoverBook, deleteBook, invalidateAllBookQueries } from '@/api/client'

export function HardcoverBookPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [isAdding, setIsAdding] = useState(false)

  const decodedId = id ? decodeURIComponent(id) : ''

  const { data: book, isLoading, error } = useQuery({
    queryKey: ['hardcoverBook', decodedId],
    queryFn: () => getHardcoverBook(decodedId),
    enabled: !!decodedId,
    retry: 1,
  })

  const addMutation = useMutation({
    mutationFn: () => addHardcoverBook(id!, { monitored: true }),
    onSuccess: (result) => {
      setIsAdding(false)
      invalidateAllBookQueries(queryClient)
      // Navigate to the library book page
      navigate(`/books/${result.bookId}`)
    },
    onError: (error: Error) => {
      setIsAdding(false)
      if (error.message.includes('409') || error.message.includes('Conflict') || error.message.includes('already in library')) {
        invalidateAllBookQueries(queryClient)
      }
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (bookId: number) => deleteBook(bookId),
    onSuccess: () => {
      invalidateAllBookQueries(queryClient)
      queryClient.invalidateQueries({ queryKey: ['hardcoverBook', id] })
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

                  {/* Add/Delete Button */}
                  <div>
                    {book.inLibrary ? (
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="flex items-center gap-1 text-sm px-3 py-1.5">
                          <Check className="h-4 w-4" />
                          In Library
                        </Badge>
                        {book.libraryBook?.id && (
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => deleteMutation.mutate(book.libraryBook!.id)}
                            disabled={deleteMutation.isPending}
                          >
                            {deleteMutation.isPending ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <Trash2 className="h-4 w-4" />
                            )}
                          </Button>
                        )}
                      </div>
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
                <div className="flex flex-wrap items-center gap-3 mt-4">
                  <FormatAvailability
                    hasEbook={book.hasEbook}
                    hasAudiobook={book.hasAudiobook}
                  />
                  {book.hasAudiobook && book.audioDuration && book.audioDuration > 0 && (
                    <span className="text-sm text-muted-foreground">
                      ({Math.floor(book.audioDuration / 3600)}h {Math.floor((book.audioDuration % 3600) / 60)}m)
                    </span>
                  )}
                  {book.editionCount && book.editionCount > 1 && (
                    <Badge variant="outline" className="flex items-center gap-1">
                      <Layers className="h-3 w-3" />
                      {book.editionCount} editions
                    </Badge>
                  )}
                </div>

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
        <div className="max-w-5xl mx-auto px-6 py-8 space-y-8">
          
          {book.contributors && book.contributors.length > 0 && (
            <section>
              <h2 className="text-xl font-semibold mb-4">Contributors</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {book.contributors.map((contributor: any, i: number) => (
                  <div key={i} className="flex items-center gap-3 p-3 rounded-lg border border-border bg-card">
                    {contributor.authorImage ? (
                      <img
                        src={contributor.authorImage}
                        alt={contributor.authorName}
                        className="h-10 w-10 rounded-full object-cover bg-muted"
                      />
                    ) : (
                      <div className="h-10 w-10 rounded-full bg-muted flex items-center justify-center shrink-0">
                        <User className="h-5 w-5 text-muted-foreground" />
                      </div>
                    )}
                    <div>
                      <div className="font-medium">{contributor.authorName}</div>
                      <Badge variant="outline" className="text-[10px] h-5 px-1.5 font-normal mt-1">
                        {contributor.role}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </section>
          )}

          {book.editions && book.editions.length > 0 && (
            <section>
              <h2 className="text-xl font-semibold mb-4">Editions</h2>
              <div className="rounded-lg border border-border bg-card overflow-hidden shadow-sm">
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-border bg-muted/30">
                        <th className="h-10 px-4 text-left font-medium text-muted-foreground w-16">Cover</th>
                        <th className="h-10 px-4 text-left font-medium text-muted-foreground">Format</th>
                        <th className="h-10 px-4 text-left font-medium text-muted-foreground">Publisher</th>
                        <th className="h-10 px-4 text-left font-medium text-muted-foreground">Length</th>
                        <th className="h-10 px-4 text-left font-medium text-muted-foreground">Release Date</th>
                      </tr>
                    </thead>
                    <tbody>
                      {book.editions.slice(0, 10).map((edition: any, i: number) => (
                        <tr key={i} className="border-b border-border last:border-0">
                          <td className="p-2">
                            <div className="h-12 w-8 bg-muted rounded overflow-hidden">
                              {edition.coverUrl && (
                                <img src={edition.coverUrl} alt="" className="h-full w-full object-cover" />
                              )}
                            </div>
                          </td>
                          <td className="p-4 align-middle">
                            <div className="font-medium">{edition.editionFormat || edition.format}</div>
                            <div className="text-xs text-muted-foreground mt-0.5">
                              {edition.isbn13 || edition.isbn10 || edition.asin || '-'}
                            </div>
                          </td>
                          <td className="p-4 align-middle">
                            <span className="truncate max-w-[200px]" title={edition.publisherName}>
                              {edition.publisherName || '-'}
                            </span>
                          </td>
                          <td className="p-4 align-middle">
                            {edition.audioSeconds ? (
                              <span>{Math.floor(edition.audioSeconds / 3600)}h {Math.floor((edition.audioSeconds % 3600) / 60)}m</span>
                            ) : edition.pageCount ? (
                              <span>{edition.pageCount} pages</span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </td>
                          <td className="p-4 align-middle">
                            {edition.releaseDate ? new Date(edition.releaseDate).toLocaleDateString() : '-'}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </section>
          )}

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
