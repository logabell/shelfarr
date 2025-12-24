import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '@/components/layout/Topbar'
import { LibraryGrid } from '@/components/library/LibraryGrid'
import { LibraryStats } from '@/components/library/LibraryStats'
import { LibraryFilters } from '@/components/library/LibraryFilters'
import { LibraryToolbar } from '@/components/library/LibraryToolbar'
import { getLibrary, getLibraryStats, bulkUpdateBooks, bulkDeleteBooks, invalidateAllBookQueries } from '@/api/client'
import { Button } from '@/components/ui/button'
import { CheckSquare } from 'lucide-react'
import type { Book } from '@/types'

export function LibraryPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sortBy, setSortBy] = useState('title')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [selectionMode, setSelectionMode] = useState(false)
  const [selectedBooks, setSelectedBooks] = useState<Set<number>>(new Set())
  const [monitoredFilter, setMonitoredFilter] = useState('')

  // Fetch library data
  const { data: libraryData, isLoading: libraryLoading, refetch } = useQuery({
    queryKey: ['library', statusFilter, sortBy],
    queryFn: () => getLibrary({ 
      status: statusFilter || undefined,
      sortBy,
      pageSize: 100 
    }),
  })

  // Fetch stats
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['library-stats'],
    queryFn: getLibraryStats,
  })

  // Bulk update mutation
  const bulkUpdateMutation = useMutation({
    mutationFn: ({ bookIds, monitored }: { bookIds: number[], monitored: boolean }) =>
      bulkUpdateBooks(bookIds, { monitored }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['library'] })
      setSelectedBooks(new Set())
      setSelectionMode(false)
    },
  })

  // Bulk delete mutation
  const bulkDeleteMutation = useMutation({
    mutationFn: ({ bookIds, deleteFiles }: { bookIds: number[], deleteFiles: boolean }) =>
      bulkDeleteBooks(bookIds, deleteFiles),
    onSuccess: () => {
      invalidateAllBookQueries(queryClient)
      setSelectedBooks(new Set())
      setSelectionMode(false)
    },
  })

  // Client-side filtering for search and monitored status
  const filteredBooks = useMemo(() => {
    if (!libraryData?.books) return []
    
    let books = libraryData.books

    // Filter by monitored status
    if (monitoredFilter === 'monitored') {
      books = books.filter((book: Book) => book.monitored)
    } else if (monitoredFilter === 'unmonitored') {
      books = books.filter((book: Book) => !book.monitored)
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      books = books.filter((book: Book) => 
        book.title.toLowerCase().includes(query) ||
        book.author?.name.toLowerCase().includes(query) ||
        book.series?.name.toLowerCase().includes(query)
      )
    }

    return books
  }, [libraryData?.books, searchQuery, monitoredFilter])

  const handleSelectBook = (book: Book) => {
    setSelectedBooks(prev => {
      const next = new Set(prev)
      if (next.has(book.id)) {
        next.delete(book.id)
      } else {
        next.add(book.id)
      }
      return next
    })
  }

  const handleToggleSelectionMode = () => {
    if (selectionMode) {
      setSelectedBooks(new Set())
    }
    setSelectionMode(!selectionMode)
  }

  const handleSelectAll = () => {
    if (selectedBooks.size === filteredBooks.length) {
      setSelectedBooks(new Set())
    } else {
      setSelectedBooks(new Set(filteredBooks.map(b => b.id)))
    }
  }

  const handleClearSelection = () => {
    setSelectedBooks(new Set())
    setSelectionMode(false)
  }

  const handleSearchSelected = () => {
    // Navigate to first selected book with search=true, or could batch search
    const firstBookId = Array.from(selectedBooks)[0]
    if (firstBookId) {
      navigate(`/books/${firstBookId}?search=true`)
    }
  }

  const handleRemoveSelected = (deleteFiles: boolean) => {
    const bookIds = Array.from(selectedBooks)
    bulkDeleteMutation.mutate({ bookIds, deleteFiles })
  }

  const handleSetMonitored = (monitored: boolean) => {
    const bookIds = Array.from(selectedBooks)
    bulkUpdateMutation.mutate({ bookIds, monitored })
  }

  const isLoading = bulkUpdateMutation.isPending || bulkDeleteMutation.isPending

  return (
    <div className="flex flex-col h-full">
      <Topbar 
        title="Library" 
        subtitle={`${libraryData?.total || 0} books`}
        onRefresh={() => refetch()}
        isRefreshing={libraryLoading}
      />
      
      <div className="flex-1 overflow-auto p-6 space-y-6">
        {/* Stats */}
        <LibraryStats stats={stats} isLoading={statsLoading} />

        {/* Filters */}
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <LibraryFilters
              searchQuery={searchQuery}
              onSearchChange={setSearchQuery}
              statusFilter={statusFilter}
              onStatusChange={setStatusFilter}
              sortBy={sortBy}
              onSortChange={setSortBy}
              viewMode={viewMode}
              onViewModeChange={setViewMode}
              monitoredFilter={monitoredFilter}
              onMonitoredChange={setMonitoredFilter}
            />
          </div>
          <Button
            variant={selectionMode ? 'secondary' : 'outline'}
            size="sm"
            onClick={handleToggleSelectionMode}
          >
            <CheckSquare className="h-4 w-4 mr-2" />
            {selectionMode ? 'Exit Select' : 'Select'}
          </Button>
        </div>

        {/* Bulk Action Toolbar */}
        {selectionMode && (
          <LibraryToolbar
            selectedCount={selectedBooks.size}
            totalCount={filteredBooks.length}
            onSelectAll={handleSelectAll}
            onClearSelection={handleClearSelection}
            onSearchSelected={handleSearchSelected}
            onRemoveSelected={handleRemoveSelected}
            onSetMonitored={handleSetMonitored}
            isLoading={isLoading}
          />
        )}

        {/* Grid */}
        <LibraryGrid
          books={filteredBooks}
          isLoading={libraryLoading}
          selectedBooks={selectedBooks}
          selectionMode={selectionMode}
          onSelectBook={handleSelectBook}
        />
      </div>
    </div>
  )
}
