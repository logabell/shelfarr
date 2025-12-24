import { Link } from 'react-router-dom'
import { 
  Book as BookIcon, 
  Plus, 
  Loader2, 
  CheckCircle2, 
  AlertCircle, 
  Download, 
  Clock,
  EyeOff,
  Star,
  Layers,
  CircleDashed,
  Library,
  Trash2
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FormatAvailability } from '@/components/ui/format-availability'
import { cn } from '@/lib/utils'
import type { Book, BookStatus } from '@/types'

interface CatalogBookCardProps {
  // Book identification
  hardcoverId: string
  title: string
  coverUrl?: string
  rating?: number
  releaseYear?: number
  compilation?: boolean
  
  // Series info (optional)
  seriesIndex?: number
  seriesName?: string
  
  // Author info (optional, for series view)
  authorName?: string
  
  // Edition info (optional)
  hasAudiobook?: boolean
  hasEbook?: boolean
  editionCount?: number
  
  // Library status
  inLibrary: boolean
  libraryBook?: Book
  
  // Actions
  onAdd?: (hardcoverId: string) => void
  isAdding?: boolean
  onDelete?: (bookId: number) => void
  isDeleting?: boolean
}

const getStatusColor = (status?: BookStatus | 'not_in_library') => {
  switch (status) {
    case 'downloaded': return 'bg-green-500'
    case 'downloading': return 'bg-sky-500'
    case 'missing': return 'bg-red-500'
    case 'unreleased': return 'bg-purple-500'
    case 'not_in_library': return 'bg-neutral-600'
    default: return 'bg-neutral-500'
  }
}

const getStatusIcon = (status?: BookStatus | 'not_in_library') => {
  switch (status) {
    case 'downloaded': return <CheckCircle2 className="w-4 h-4 text-green-400" />
    case 'downloading': return <Download className="w-4 h-4 text-sky-400" />
    case 'missing': return <AlertCircle className="w-4 h-4 text-red-400" />
    case 'unreleased': return <Clock className="w-4 h-4 text-purple-400" />
    case 'not_in_library': return <CircleDashed className="w-4 h-4 text-neutral-500" />
    default: return <AlertCircle className="w-4 h-4 text-neutral-400" />
  }
}

const getStatusLabel = (status?: BookStatus | 'not_in_library') => {
  switch (status) {
    case 'downloaded': return 'Downloaded'
    case 'downloading': return 'Downloading'
    case 'missing': return 'Missing'
    case 'unreleased': return 'Unreleased'
    case 'not_in_library': return 'Not in library'
    default: return 'Unknown'
  }
}

export function CatalogBookCard({
  hardcoverId,
  title,
  coverUrl,
  rating,
  releaseYear,
  compilation,
  seriesIndex,
  seriesName,
  authorName,
  hasAudiobook,
  hasEbook,
  editionCount,
  inLibrary,
  libraryBook,
  onAdd,
  isAdding = false,
  onDelete,
  isDeleting = false,
}: CatalogBookCardProps) {
  // Determine status - check if unreleased based on release year
  const isUnreleased = releaseYear && releaseYear > new Date().getFullYear()
  const displayStatus: BookStatus | 'not_in_library' = inLibrary 
    ? (libraryBook?.status || 'missing')
    : (isUnreleased ? 'unreleased' : 'not_in_library')
  
  const FormatBadges = () => (
    <div className="flex items-center gap-1">
      {compilation && (
        <div className="bg-amber-600/80 rounded px-1 py-0.5 flex items-center gap-0.5" title="Collection/Box Set">
          <Library className="w-3 h-3 text-white" />
          <span className="text-[10px] text-white font-medium">Collection</span>
        </div>
      )}
      <div className="bg-black/60 rounded px-1 py-0.5">
        <FormatAvailability
          hasEbook={hasEbook}
          hasAudiobook={hasAudiobook}
          compact
        />
      </div>
      {editionCount && editionCount > 1 && (
        <div className="bg-sky-500/80 rounded px-1 py-0.5 flex items-center gap-0.5" title={`${editionCount} editions`}>
          <Layers className="w-3 h-3 text-white" />
          <span className="text-[10px] text-white font-medium">{editionCount}</span>
        </div>
      )}
    </div>
  )

  if (!inLibrary) {
    // Not in library - show preview card with add button
    return (
      <div className="group relative bg-neutral-800/30 border-2 border-dashed border-neutral-700 rounded-xl overflow-hidden hover:border-sky-500/50 transition-colors">
        {/* Cover - clickable to preview */}
        <Link
          to={`/hardcover/book/${hardcoverId}`}
          className="block aspect-[2/3] relative"
        >
          {coverUrl ? (
            <img
              src={coverUrl}
              alt={title}
              className="w-full h-full object-cover opacity-60 group-hover:opacity-80 transition-opacity"
            />
          ) : (
            <div className="w-full h-full bg-gradient-to-br from-neutral-700 to-neutral-800 flex items-center justify-center">
              <BookIcon className="w-12 h-12 text-neutral-600" />
            </div>
          )}

          {/* Series Number Badge */}
          {seriesIndex != null && seriesIndex > 0 && (
            <div className="absolute top-2 left-2 bg-black/70 text-white text-xs font-bold px-2 py-1 rounded">
              #{seriesIndex}
            </div>
          )}

          {/* Status Indicator */}
          <div className={cn('absolute top-2 right-2 w-3 h-3 rounded-full', getStatusColor(displayStatus))} />
          
          {/* Format Badges */}
          <div className="absolute bottom-2 left-2">
            <FormatBadges />
          </div>
        </Link>

        {/* Add Button - positioned at bottom right */}
        {onAdd && (
          <div className="absolute bottom-16 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              size="sm"
              onClick={(e) => {
                e.preventDefault()
                onAdd(hardcoverId)
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
        )}

        {/* Info */}
        <div className="p-3">
          <Link
            to={`/hardcover/book/${hardcoverId}`}
            className="font-medium text-neutral-400 text-sm line-clamp-2 hover:text-sky-400 transition-colors"
          >
            {title}
          </Link>
          {authorName && (
            <p className="text-xs text-neutral-500 mt-1">{authorName}</p>
          )}
          <div className="flex items-center gap-2 mt-2">
            {getStatusIcon(displayStatus)}
            <span className="text-xs text-neutral-500">{getStatusLabel(displayStatus)}</span>
          </div>
          <div className="flex items-center gap-2 mt-1">
            {rating != null && rating > 0 && (
              <div className="flex items-center gap-0.5 text-xs text-neutral-500">
                <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
                {rating.toFixed(1)}
              </div>
            )}
            {releaseYear && (
              <span className="text-xs text-neutral-500">{releaseYear}</span>
            )}
          </div>
        </div>
      </div>
    )
  }

  // In library - show full card with status
  const bookId = libraryBook?.id
  const status = libraryBook?.status
  const monitored = libraryBook?.monitored

  return (
    <Link
      to={bookId ? `/books/${bookId}` : `/hardcover/book/${hardcoverId}`}
      className="group relative bg-neutral-800/50 rounded-xl overflow-hidden hover:ring-2 hover:ring-sky-500/50 transition-all"
    >
      {/* Cover */}
      <div className="aspect-[2/3] relative">
        {coverUrl ? (
          <img
            src={coverUrl}
            alt={title}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-neutral-700 to-neutral-800 flex items-center justify-center">
            <BookIcon className="w-12 h-12 text-neutral-600" />
          </div>
        )}

        {/* Series Number Badge */}
        {seriesIndex != null && seriesIndex > 0 && (
          <div className="absolute top-2 left-2 bg-black/70 text-white text-xs font-bold px-2 py-1 rounded">
            #{seriesIndex}
          </div>
        )}

        {/* Status Indicator */}
        <div className={cn('absolute top-2 right-2 w-3 h-3 rounded-full', getStatusColor(status))} />

        {/* Format Badges */}
        <div className="absolute bottom-2 left-2">
          <FormatBadges />
        </div>

        {/* Monitor Badge */}
        {monitored === false && (
          <div className="absolute bottom-2 right-2 bg-neutral-900/80 rounded p-1">
            <EyeOff className="w-3 h-3 text-neutral-400" />
          </div>
        )}

        {/* Delete Button */}
        {onDelete && bookId && (
          <div className="absolute bottom-16 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              size="sm"
              variant="destructive"
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
                onDelete(bookId)
              }}
              disabled={isDeleting}
            >
              {isDeleting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Trash2 className="w-4 h-4" />
              )}
            </Button>
          </div>
        )}
      </div>

      {/* Info */}
      <div className="p-3">
        <h3 className="font-medium text-neutral-200 text-sm line-clamp-2 group-hover:text-sky-400 transition-colors">
          {libraryBook?.title || title}
        </h3>
        {authorName && (
          <p className="text-xs text-neutral-500 mt-1">{authorName}</p>
        )}
        {seriesName && !authorName && (
          <p className="text-xs text-neutral-500 mt-1 line-clamp-1">
            {seriesName} {seriesIndex ? `#${seriesIndex}` : ''}
          </p>
        )}
        <div className="flex items-center gap-2 mt-2">
          {getStatusIcon(status)}
          <span className="text-xs text-neutral-400 capitalize">{status || 'unknown'}</span>
        </div>
        <div className="flex items-center gap-2 mt-1">
          {rating != null && rating > 0 && (
            <div className="flex items-center gap-0.5 text-xs text-neutral-400">
              <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
              {rating.toFixed(1)}
            </div>
          )}
          {releaseYear && (
            <span className="text-xs text-neutral-400">{releaseYear}</span>
          )}
        </div>
      </div>
    </Link>
  )
}

// Export status helpers for use in other components
export { getStatusColor, getStatusIcon }
