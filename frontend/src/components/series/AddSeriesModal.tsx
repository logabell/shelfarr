import { useState, useMemo } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  BookOpen,
  Plus,
  Check,
  Loader2,
  X,
  CheckSquare,
  Square
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { addSeriesBooks } from '@/api/client'
import type { SeriesBookEntry } from '@/types'

interface AddSeriesModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (addedCount: number) => void
  seriesId: number
  seriesName: string
  books: SeriesBookEntry[]
}

export function AddSeriesModal({ 
  isOpen, 
  onClose, 
  onSuccess, 
  seriesId, 
  seriesName,
  books 
}: AddSeriesModalProps) {
  const queryClient = useQueryClient()
  const [selectedBookIds, setSelectedBookIds] = useState<Set<string>>(new Set())

  const availableBooks = useMemo(() => {
    return books.filter(b => !b.inLibrary && b.hardcoverId)
  }, [books])

  const addMutation = useMutation({
    mutationFn: () => addSeriesBooks(seriesId, Array.from(selectedBookIds), true),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['series', String(seriesId)] })
      onSuccess(result.addedCount)
      resetAndClose()
    },
  })

  const toggleBook = (hardcoverId: string) => {
    setSelectedBookIds(prev => {
      const next = new Set(prev)
      if (next.has(hardcoverId)) {
        next.delete(hardcoverId)
      } else {
        next.add(hardcoverId)
      }
      return next
    })
  }

  const selectAll = () => {
    setSelectedBookIds(new Set(availableBooks.map(b => b.hardcoverId!)))
  }

  const deselectAll = () => {
    setSelectedBookIds(new Set())
  }

  const resetAndClose = () => {
    setSelectedBookIds(new Set())
    onClose()
  }

  const allSelected = availableBooks.length > 0 && 
    availableBooks.every(b => selectedBookIds.has(b.hardcoverId!))

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && resetAndClose()}>
      <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Add Books from Series</DialogTitle>
          <DialogDescription>
            Select which books from "{seriesName}" to add to your library
          </DialogDescription>
        </DialogHeader>

        {availableBooks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <Check className="h-12 w-12 text-green-500 mb-4" />
            <p className="text-lg font-medium">All books already in library</p>
            <p className="text-sm text-muted-foreground mt-1">
              Every book in this series has been added
            </p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between py-2 border-b">
              <span className="text-sm text-muted-foreground">
                {selectedBookIds.size} of {availableBooks.length} selected
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={allSelected ? deselectAll : selectAll}
              >
                {allSelected ? (
                  <>
                    <Square className="h-4 w-4 mr-2" />
                    Deselect All
                  </>
                ) : (
                  <>
                    <CheckSquare className="h-4 w-4 mr-2" />
                    Select All
                  </>
                )}
              </Button>
            </div>

            <div className="flex-1 overflow-y-auto py-2">
              <div className="grid grid-cols-1 gap-2">
                {availableBooks.map((book) => (
                  <button
                    key={book.hardcoverId}
                    type="button"
                    onClick={() => toggleBook(book.hardcoverId!)}
                    className={`flex items-center gap-3 p-3 rounded-lg border-2 text-left transition-all ${
                      selectedBookIds.has(book.hardcoverId!)
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50 hover:bg-muted'
                    }`}
                  >
                    {book.coverUrl ? (
                      <img
                        src={book.coverUrl}
                        alt={book.title}
                        className="w-10 h-14 object-cover rounded shadow-sm shrink-0"
                      />
                    ) : (
                      <div className="w-10 h-14 bg-muted rounded flex items-center justify-center shrink-0">
                        <BookOpen className="h-4 w-4 text-muted-foreground" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        {book.index > 0 && (
                          <span className="text-xs font-medium text-muted-foreground shrink-0">
                            #{book.index}
                          </span>
                        )}
                        <span className="font-medium truncate">{book.title}</span>
                      </div>
                      {book.authorName && (
                        <p className="text-sm text-muted-foreground truncate">
                          {book.authorName}
                        </p>
                      )}
                    </div>
                    <div className={`shrink-0 ${
                      selectedBookIds.has(book.hardcoverId!)
                        ? 'text-primary'
                        : 'text-muted-foreground'
                    }`}>
                      {selectedBookIds.has(book.hardcoverId!) ? (
                        <CheckSquare className="h-5 w-5" />
                      ) : (
                        <Square className="h-5 w-5" />
                      )}
                    </div>
                  </button>
                ))}
              </div>
            </div>

            {addMutation.isError && (
              <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3">
                <div className="flex items-center gap-2 text-destructive">
                  <X className="h-4 w-4" />
                  <p className="text-sm">
                    {addMutation.error instanceof Error 
                      ? addMutation.error.message 
                      : 'Failed to add books to library'}
                  </p>
                </div>
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={resetAndClose}>
                Cancel
              </Button>
              <Button 
                onClick={() => addMutation.mutate()} 
                disabled={addMutation.isPending || selectedBookIds.size === 0}
              >
                {addMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <Plus className="h-4 w-4 mr-2" />
                )}
                Add {selectedBookIds.size} Book{selectedBookIds.size !== 1 ? 's' : ''}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

export default AddSeriesModal
