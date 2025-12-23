import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { 
  Search, 
  Trash2, 
  Eye, 
  EyeOff, 
  X, 
  CheckSquare,
  Square,
  Loader2
} from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface LibraryToolbarProps {
  selectedCount: number
  totalCount: number
  onSelectAll: () => void
  onClearSelection: () => void
  onSearchSelected: () => void
  onRemoveSelected: (deleteFiles: boolean) => void
  onSetMonitored: (monitored: boolean) => void
  isLoading?: boolean
}

export function LibraryToolbar({
  selectedCount,
  totalCount,
  onSelectAll,
  onClearSelection,
  onSearchSelected,
  onRemoveSelected,
  onSetMonitored,
  isLoading,
}: LibraryToolbarProps) {
  const [showRemoveDialog, setShowRemoveDialog] = useState(false)
  const [deleteFiles, setDeleteFiles] = useState(false)

  const allSelected = selectedCount === totalCount && totalCount > 0

  const handleRemove = () => {
    onRemoveSelected(deleteFiles)
    setShowRemoveDialog(false)
    setDeleteFiles(false)
  }

  if (selectedCount === 0) {
    return null
  }

  return (
    <>
      <div className="flex items-center justify-between p-4 bg-primary/10 border border-primary/20 rounded-lg mb-4">
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium">
            {selectedCount} of {totalCount} selected
          </span>
          
          <Button
            variant="ghost"
            size="sm"
            onClick={allSelected ? onClearSelection : onSelectAll}
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

        <div className="flex items-center gap-2">
          {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
          
          <Button
            variant="secondary"
            size="sm"
            onClick={onSearchSelected}
            disabled={isLoading}
          >
            <Search className="h-4 w-4 mr-2" />
            Search Selected
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onSetMonitored(true)}
            disabled={isLoading}
          >
            <Eye className="h-4 w-4 mr-2" />
            Set Monitored
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onSetMonitored(false)}
            disabled={isLoading}
          >
            <EyeOff className="h-4 w-4 mr-2" />
            Set Unmonitored
          </Button>

          <Button
            variant="destructive"
            size="sm"
            onClick={() => setShowRemoveDialog(true)}
            disabled={isLoading}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Remove
          </Button>

          <Button
            variant="ghost"
            size="icon"
            onClick={onClearSelection}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <Dialog open={showRemoveDialog} onOpenChange={setShowRemoveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove {selectedCount} book(s)?</DialogTitle>
            <DialogDescription>
              This will remove the selected books from your library.
            </DialogDescription>
          </DialogHeader>
          
          <div className="py-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={deleteFiles}
                onChange={(e) => setDeleteFiles(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <span className="font-medium text-sm">Also delete files from disk</span>
                <p className="text-xs text-muted-foreground">
                  This will permanently delete any downloaded files
                </p>
              </div>
            </label>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowRemoveDialog(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleRemove}
            >
              Remove {deleteFiles ? '& Delete Files' : ''}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
