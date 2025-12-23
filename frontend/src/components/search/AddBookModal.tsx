import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import {
  BookOpen,
  Headphones,
  Plus,
  Check,
  Loader2,
  Star,
  Search,
  Zap,
  X
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { getHardcoverBook, addHardcoverBook } from '@/api/client'

export type MediaTypeOption = 'ebook' | 'audiobook' | 'both'
export type DownloadMode = 'auto' | 'manual' | 'none'

interface AddBookModalProps {
  bookId: string | null
  isOpen: boolean
  onClose: () => void
  onSuccess: (bookId: number, downloadMode: DownloadMode, mediaType: MediaTypeOption) => void
}

export function AddBookModal({ bookId, isOpen, onClose, onSuccess }: AddBookModalProps) {
  const [mediaType, setMediaType] = useState<MediaTypeOption>('ebook')
  const [downloadMode, setDownloadMode] = useState<DownloadMode>('auto')

  const { data: book, isLoading, error } = useQuery({
    queryKey: ['hardcoverBook', bookId],
    queryFn: () => getHardcoverBook(bookId!),
    enabled: !!bookId && isOpen,
  })

  const addMutation = useMutation({
    mutationFn: () => addHardcoverBook(bookId!, { monitored: true, mediaType }),
    onSuccess: (result) => {
      onSuccess(result.bookId, downloadMode, mediaType)
    },
    onError: (error) => {
      console.error('Failed to add book:', error)
    },
  })

  const handleAdd = () => {
    addMutation.mutate()
  }

  const resetAndClose = () => {
    setMediaType('ebook')
    setDownloadMode('auto')
    onClose()
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && resetAndClose()}>
      <DialogContent className="max-w-lg">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="text-center py-12 px-4">
            <X className="h-12 w-12 text-destructive mx-auto mb-4" />
            <p className="text-destructive font-medium mb-2">Failed to load book details</p>
            <p className="text-sm text-muted-foreground mb-4">
              {(() => {
                if (error instanceof Error) {
                  // Extract error message from Axios error response
                  const axiosError = error as any
                  if (axiosError?.response?.data?.error) {
                    return axiosError.response.data.error
                  }
                  return error.message
                }
                return 'Please check your Hardcover API configuration in Settings'
              })()}
            </p>
            <div className="flex flex-col gap-2 items-center">
              <Button variant="outline" onClick={resetAndClose}>
                Close
              </Button>
              <Button 
                variant="link" 
                size="sm"
                onClick={() => window.location.href = '/settings'}
              >
                Go to Settings
              </Button>
            </div>
          </div>
        ) : book ? (
          <>
            <DialogHeader>
              <DialogTitle>Add to Library</DialogTitle>
              <DialogDescription>
                Configure how you want to add this book
              </DialogDescription>
            </DialogHeader>

            {/* Book Info */}
            <div className="flex gap-4 p-4 rounded-lg bg-secondary/30 border border-border">
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
              <div className="flex-1 min-w-0">
                <h3 className="font-medium line-clamp-2">{book.title}</h3>
                {book.authorName && (
                  <p className="text-sm text-muted-foreground mt-1">
                    by {book.authorName}
                  </p>
                )}
                <div className="flex items-center gap-2 mt-2">
                  {book.rating > 0 && (
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Star className="h-3 w-3 fill-yellow-500 text-yellow-500" />
                      {book.rating.toFixed(1)}
                    </div>
                  )}
                  {book.releaseYear && (
                    <span className="text-xs text-muted-foreground">{book.releaseYear}</span>
                  )}
                </div>
              </div>
            </div>

            {book.inLibrary ? (
              <div className="py-4 text-center">
                <Badge variant="secondary" className="flex items-center gap-2 justify-center mx-auto">
                  <Check className="h-4 w-4" />
                  Already in Library
                </Badge>
                <p className="text-sm text-muted-foreground mt-2">
                  This book is already in your library
                </p>
              </div>
            ) : (
              <>
                {/* Media Type Selection */}
                <div className="space-y-3">
                  <Label>Media Type</Label>
                  <div className="grid grid-cols-3 gap-2">
                    <MediaTypeButton
                      icon={BookOpen}
                      label="Ebook"
                      selected={mediaType === 'ebook'}
                      onClick={() => setMediaType('ebook')}
                    />
                    <MediaTypeButton
                      icon={Headphones}
                      label="Audiobook"
                      selected={mediaType === 'audiobook'}
                      onClick={() => setMediaType('audiobook')}
                    />
                    <MediaTypeButton
                      icon={Plus}
                      label="Both"
                      selected={mediaType === 'both'}
                      onClick={() => setMediaType('both')}
                    />
                  </div>
                </div>

                {/* Download Mode Selection */}
                <div className="space-y-3">
                  <Label>Download Mode</Label>
                  <div className="space-y-2">
                    <DownloadModeOption
                      icon={Zap}
                      title="Automatic Search"
                      description="Find and download the best match based on your quality profile"
                      selected={downloadMode === 'auto'}
                      onClick={() => setDownloadMode('auto')}
                    />
                    <DownloadModeOption
                      icon={Search}
                      title="Interactive Search"
                      description="Browse available torrents and choose which to download"
                      selected={downloadMode === 'manual'}
                      onClick={() => setDownloadMode('manual')}
                    />
                    <DownloadModeOption
                      icon={BookOpen}
                      title="Add Only"
                      description="Just add to library without searching for downloads"
                      selected={downloadMode === 'none'}
                      onClick={() => setDownloadMode('none')}
                    />
                  </div>
                </div>

                {addMutation.isError && (
                  <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3">
                    <p className="text-sm text-destructive">
                      {addMutation.error instanceof Error 
                        ? addMutation.error.message 
                        : 'Failed to add book to library'}
                    </p>
                  </div>
                )}
                <DialogFooter>
                  <Button variant="outline" onClick={resetAndClose}>
                    Cancel
                  </Button>
                  <Button onClick={handleAdd} disabled={addMutation.isPending}>
                    {addMutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                    <Plus className="h-4 w-4" />
                    Add to Library
                  </Button>
                </DialogFooter>
              </>
            )}
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

// Media Type Button Component
function MediaTypeButton({
  icon: Icon,
  label,
  selected,
  onClick
}: {
  icon: React.ElementType
  label: string
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex flex-col items-center gap-2 p-4 rounded-lg border-2 transition-all ${
        selected
          ? 'border-primary bg-primary/10 text-primary'
          : 'border-border hover:border-primary/50 hover:bg-muted'
      }`}
    >
      <Icon className="h-6 w-6" />
      <span className="text-sm font-medium">{label}</span>
    </button>
  )
}

// Download Mode Option Component
function DownloadModeOption({
  icon: Icon,
  title,
  description,
  selected,
  onClick
}: {
  icon: React.ElementType
  title: string
  description: string
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full flex items-start gap-3 p-3 rounded-lg border-2 text-left transition-all ${
        selected
          ? 'border-primary bg-primary/10'
          : 'border-border hover:border-primary/50 hover:bg-muted'
      }`}
    >
      <div className={`p-2 rounded-lg ${selected ? 'bg-primary/20' : 'bg-muted'}`}>
        <Icon className={`h-4 w-4 ${selected ? 'text-primary' : 'text-muted-foreground'}`} />
      </div>
      <div>
        <p className={`font-medium ${selected ? 'text-primary' : ''}`}>{title}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      {selected && (
        <div className="ml-auto">
          <Check className="h-5 w-5 text-primary" />
        </div>
      )}
    </button>
  )
}

export default AddBookModal
