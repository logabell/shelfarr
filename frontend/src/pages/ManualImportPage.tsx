import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { 
  getPendingImports, 
  searchHardcover, 
  manualImport 
} from '@/api/client'
import { 
  File, 
  Search, 
  Import, 
  Eye, 
  BookOpen,
  Headphones,
  GripVertical,
  ChevronRight,
  Check,
  X
} from 'lucide-react'
import { cn, formatFileSize } from '@/lib/utils'
import type { SearchResult } from '@/types'

interface PendingFile {
  path: string
  name: string
  size: number
  extension: string
  modified: string
}

export function ManualImportPage() {
  const [selectedFile, setSelectedFile] = useState<PendingFile | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedBook, setSelectedBook] = useState<SearchResult | null>(null)
  const [mediaType, setMediaType] = useState<'ebook' | 'audiobook'>('ebook')
  const queryClient = useQueryClient()

  // Fetch pending imports
  const { data: pendingFiles = [], isLoading: filesLoading, refetch } = useQuery({
    queryKey: ['pending-imports'],
    queryFn: getPendingImports,
  })

  // Search Hardcover
  const { data: searchResults = [], isLoading: searchLoading } = useQuery({
    queryKey: ['search', searchQuery],
    queryFn: () => searchHardcover(searchQuery),
    enabled: searchQuery.length > 2,
  })

  // Import mutation
  const importMutation = useMutation({
    mutationFn: () => {
      if (!selectedFile || !selectedBook) throw new Error('Missing selection')
      return manualImport(selectedFile.path, parseInt(selectedBook.id), mediaType)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-imports'] })
      queryClient.invalidateQueries({ queryKey: ['library'] })
      setSelectedFile(null)
      setSelectedBook(null)
      setSearchQuery('')
    },
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    // Search is automatic via React Query
  }

  const detectMediaType = (file: PendingFile): 'ebook' | 'audiobook' => {
    const audioExtensions = ['mp3', 'm4b', 'm4a', 'flac', 'ogg', 'aac']
    return audioExtensions.includes(file.extension.toLowerCase()) ? 'audiobook' : 'ebook'
  }

  const handleFileSelect = (file: PendingFile) => {
    setSelectedFile(file)
    setMediaType(detectMediaType(file))
    // Pre-fill search with filename (cleaned up)
    const cleanName = file.name
      .replace(/\.[^.]+$/, '') // Remove extension
      .replace(/[_-]/g, ' ') // Replace underscores/dashes with spaces
      .replace(/\s+/g, ' ') // Normalize spaces
      .trim()
    setSearchQuery(cleanName)
  }

  const canImport = selectedFile && selectedBook

  return (
    <div className="flex flex-col h-full">
      <Topbar 
        title="Manual Import" 
        subtitle="Map downloaded files to books"
        onRefresh={() => refetch()}
        isRefreshing={filesLoading}
      />
      
      <div className="flex-1 flex overflow-hidden">
        {/* Left Pane - Pending Files */}
        <div className="w-1/2 border-r border-border flex flex-col">
          <div className="p-4 border-b border-border bg-card">
            <h2 className="font-semibold flex items-center gap-2">
              <File className="h-4 w-4" />
              Unimported Files
            </h2>
            <p className="text-sm text-muted-foreground mt-1">
              Files in your downloads folder
            </p>
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            {filesLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="h-16 skeleton rounded-lg" />
                ))}
              </div>
            ) : (pendingFiles as PendingFile[]).length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <Import className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">No files pending import</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Download some books to see them here
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {(pendingFiles as PendingFile[]).map((file, idx) => (
                  <button
                    key={idx}
                    onClick={() => handleFileSelect(file)}
                    className={cn(
                      'w-full flex items-center gap-3 p-3 rounded-lg border transition-colors text-left',
                      selectedFile?.path === file.path
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50'
                    )}
                  >
                    <div className="rounded-md bg-muted p-2">
                      {detectMediaType(file) === 'audiobook' ? (
                        <Headphones className="h-5 w-5 text-muted-foreground" />
                      ) : (
                        <BookOpen className="h-5 w-5 text-muted-foreground" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{file.name}</p>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>{formatFileSize(file.size)}</span>
                        <span>•</span>
                        <Badge variant="outline" className="text-[10px]">
                          {file.extension.toUpperCase()}
                        </Badge>
                      </div>
                    </div>
                    <Button variant="ghost" size="icon" className="flex-shrink-0">
                      <Eye className="h-4 w-4" />
                    </Button>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Drag indicator */}
        <div className="flex items-center justify-center w-8 bg-muted/50">
          <GripVertical className="h-6 w-6 text-muted-foreground" />
        </div>

        {/* Right Pane - Book Search */}
        <div className="w-1/2 flex flex-col">
          <div className="p-4 border-b border-border bg-card">
            <h2 className="font-semibold flex items-center gap-2">
              <Search className="h-4 w-4" />
              Search Hardcover.app
            </h2>
            <p className="text-sm text-muted-foreground mt-1">
              Find the book to map this file to
            </p>
          </div>

          <div className="p-4 border-b border-border">
            <form onSubmit={handleSearch} className="flex gap-2">
              <Input
                type="search"
                placeholder="Search for a book..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="flex-1"
              />
            </form>
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            {searchLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="h-20 skeleton rounded-lg" />
                ))}
              </div>
            ) : searchResults.length === 0 && searchQuery.length > 2 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <BookOpen className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">No results found</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Try a different search term
                </p>
              </div>
            ) : searchQuery.length <= 2 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <Search className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">Search for a book</p>
                <p className="text-sm text-muted-foreground mt-1">
                  {selectedFile 
                    ? 'Search filled from filename - modify as needed'
                    : 'Select a file first, or search manually'}
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {searchResults.map((result) => (
                  <button
                    key={result.id}
                    onClick={() => setSelectedBook(result)}
                    className={cn(
                      'w-full flex items-center gap-3 p-3 rounded-lg border transition-colors text-left',
                      selectedBook?.id === result.id
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50'
                    )}
                  >
                    {result.coverUrl ? (
                      <img
                        src={result.coverUrl}
                        alt={result.title}
                        className="w-12 h-16 object-cover rounded"
                      />
                    ) : (
                      <div className="w-12 h-16 bg-muted rounded flex items-center justify-center">
                        <BookOpen className="h-6 w-6 text-muted-foreground" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{result.title}</p>
                      <p className="text-sm text-muted-foreground truncate">
                        {result.author}
                      </p>
                      {result.releaseYear && (
                        <p className="text-xs text-muted-foreground">
                          {result.releaseYear}
                        </p>
                      )}
                    </div>
                    {result.inLibrary ? (
                      <Badge variant="secondary" className="flex-shrink-0">
                        <Check className="h-3 w-3 mr-1" />
                        In Library
                      </Badge>
                    ) : (
                      <ChevronRight className="h-5 w-5 text-muted-foreground flex-shrink-0" />
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Import Action */}
          {canImport && (
            <div className="p-4 border-t border-border bg-card">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <p className="text-sm font-medium">Ready to import</p>
                  <p className="text-xs text-muted-foreground">
                    {selectedFile.name} → {selectedBook.title}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="outline">
                    {mediaType === 'audiobook' ? '🎧 Audiobook' : '📖 Ebook'}
                  </Badge>
                </div>
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => {
                    setSelectedFile(null)
                    setSelectedBook(null)
                  }}
                >
                  <X className="h-4 w-4 mr-2" />
                  Cancel
                </Button>
                <Button
                  className="flex-1"
                  onClick={() => importMutation.mutate()}
                  disabled={importMutation.isPending}
                >
                  <Import className="h-4 w-4 mr-2" />
                  Import
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

