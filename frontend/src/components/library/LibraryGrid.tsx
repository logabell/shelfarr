import { Book } from '@/types'
import { BookCard, BookCardSkeleton } from './BookCard'

interface LibraryGridProps {
  books: Book[]
  isLoading?: boolean
  selectedBooks?: Set<number>
  selectionMode?: boolean
  onSelectBook?: (book: Book) => void
}

export function LibraryGrid({ 
  books, 
  isLoading,
  selectedBooks,
  selectionMode,
  onSelectBook
}: LibraryGridProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
        {Array.from({ length: 24 }).map((_, i) => (
          <BookCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (books.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16">
        <p className="text-lg text-muted-foreground">No books in your library</p>
        <p className="text-sm text-muted-foreground mt-1">
          Add books to start building your collection
        </p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
      {books.map((book) => (
        <BookCard
          key={book.id}
          book={book}
          selected={selectedBooks?.has(book.id)}
          selectionMode={selectionMode}
          onSelect={onSelectBook}
        />
      ))}
    </div>
  )
}
