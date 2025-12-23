import { Book } from '@/types'
import { cn, getStatusColor } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { BookOpen, Star } from 'lucide-react'
import { Link } from 'react-router-dom'

interface BookCardProps {
  book: Book
  selected?: boolean
  onSelect?: (book: Book) => void
  selectionMode?: boolean
}

export function BookCard({ book, selected, onSelect, selectionMode }: BookCardProps) {
  const handleClick = (e: React.MouseEvent) => {
    if (selectionMode && onSelect) {
      e.preventDefault()
      onSelect(book)
    }
  }

  return (
    <Link
      to={`/books/${book.id}`}
      className={cn(
        'book-card group relative block overflow-hidden rounded-lg bg-card shadow-lg transition-all',
        'hover:ring-2 hover:ring-primary/50',
        selected && 'ring-2 ring-primary'
      )}
      onClick={handleClick}
    >
      {/* Cover Image */}
      <div className="aspect-[2/3] overflow-hidden">
        {book.coverUrl ? (
          <img
            src={book.coverUrl}
            alt={book.title}
            className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-muted">
            <BookOpen className="h-12 w-12 text-muted-foreground" />
          </div>
        )}
      </div>

      {/* Selection Checkbox (when in selection mode) */}
      {selectionMode && (
        <div className="absolute left-2 top-2 z-10">
          <div
            className={cn(
              'h-5 w-5 rounded border-2 flex items-center justify-center transition-colors',
              selected
                ? 'bg-primary border-primary text-primary-foreground'
                : 'bg-background/80 border-muted-foreground'
            )}
          >
            {selected && (
              <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            )}
          </div>
        </div>
      )}

      {/* Badges */}
      {!selectionMode && (
        <div className="absolute left-2 top-2 flex flex-col gap-1">
          {/* Format Badge */}
          {book.format && (
            <Badge variant="secondary" className="text-[10px] uppercase">
              {book.format}
            </Badge>
          )}
          
          {/* Audiobook indicator */}
          {book.hasAudiobook && (
            <Badge variant="outline" className="bg-background/80 text-[10px]">
              🎧
            </Badge>
          )}
        </div>
      )}

      {/* Rating Badge */}
      {book.rating > 0 && (
        <div className="absolute right-2 top-2">
          <Badge variant="secondary" className="flex items-center gap-1 bg-background/80">
            <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
            <span className="text-[10px]">{book.rating.toFixed(1)}</span>
          </Badge>
        </div>
      )}

      {/* Title & Author */}
      <div className="p-3">
        <h3 className="line-clamp-2 text-sm font-medium leading-tight group-hover:text-primary">
          {book.title}
        </h3>
        {book.author && (
          <p className="mt-1 text-xs text-muted-foreground">
            {book.author.name}
          </p>
        )}
        {book.series && (
          <p className="mt-0.5 text-[10px] text-muted-foreground">
            {book.series.name} #{book.seriesIndex}
          </p>
        )}
      </div>

      {/* Status Bar */}
      <div className={cn('status-bar', getStatusColor(book.status))} />
    </Link>
  )
}

// Skeleton version for loading states
export function BookCardSkeleton() {
  return (
    <div className="overflow-hidden rounded-lg bg-card shadow-lg">
      <div className="aspect-[2/3] skeleton" />
      <div className="p-3 space-y-2">
        <div className="h-4 skeleton rounded w-3/4" />
        <div className="h-3 skeleton rounded w-1/2" />
      </div>
      <div className="h-1 skeleton" />
    </div>
  )
}
