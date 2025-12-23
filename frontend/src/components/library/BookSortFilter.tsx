import { useState } from 'react'
import { 
  ArrowUpDown, 
  Filter, 
  ArrowUp, 
  ArrowDown,
  LayoutGrid,
  List,
  X
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'

export type SortField = 'title' | 'releaseDate' | 'rating' | 'seriesIndex'
export type SortOrder = 'asc' | 'desc'
export type FilterStatus = 'all' | 'inLibrary' | 'notInLibrary' | 'downloaded' | 'missing'
export type ViewMode = 'grid' | 'list'

export interface SortFilterState {
  sortField: SortField
  sortOrder: SortOrder
  filterStatus: FilterStatus
  viewMode: ViewMode
  hideCompilations: boolean
}

interface BookSortFilterProps {
  state: SortFilterState
  onChange: (state: SortFilterState) => void
  showSeriesIndex?: boolean
  showViewToggle?: boolean
  totalCount?: number
  filteredCount?: number
  className?: string
}

const SORT_OPTIONS: { value: SortField; label: string }[] = [
  { value: 'title', label: 'Title' },
  { value: 'releaseDate', label: 'Release Date' },
  { value: 'rating', label: 'Rating' },
  { value: 'seriesIndex', label: 'Series Order' },
]

const FILTER_OPTIONS: { value: FilterStatus; label: string }[] = [
  { value: 'all', label: 'All Books' },
  { value: 'inLibrary', label: 'In Library' },
  { value: 'notInLibrary', label: 'Not in Library' },
  { value: 'downloaded', label: 'Downloaded' },
  { value: 'missing', label: 'Missing' },
]

export function BookSortFilter({
  state,
  onChange,
  showSeriesIndex = false,
  showViewToggle = true,
  totalCount,
  filteredCount,
  className,
}: BookSortFilterProps) {
  const [showFilters, setShowFilters] = useState(false)
  
  const sortOptions = showSeriesIndex 
    ? SORT_OPTIONS 
    : SORT_OPTIONS.filter(o => o.value !== 'seriesIndex')

  const handleSortFieldChange = (field: SortField) => {
    onChange({ ...state, sortField: field })
  }

  const handleSortOrderToggle = () => {
    onChange({ ...state, sortOrder: state.sortOrder === 'asc' ? 'desc' : 'asc' })
  }

  const handleFilterChange = (status: FilterStatus) => {
    onChange({ ...state, filterStatus: status })
  }

  const handleViewModeToggle = () => {
    onChange({ ...state, viewMode: state.viewMode === 'grid' ? 'list' : 'grid' })
  }

  const handleResetFilters = () => {
    onChange({
      sortField: showSeriesIndex ? 'seriesIndex' : 'title',
      sortOrder: 'asc',
      filterStatus: 'all',
      viewMode: 'grid',
      hideCompilations: true,
    })
  }

  const handleHideCompilationsToggle = () => {
    onChange({ ...state, hideCompilations: !state.hideCompilations })
  }

  const hasActiveFilters = state.filterStatus !== 'all' || !state.hideCompilations
  const activeFilterCount = (state.filterStatus !== 'all' ? 1 : 0) + (!state.hideCompilations ? 1 : 0)
  const isFiltered = filteredCount !== undefined && totalCount !== undefined && filteredCount !== totalCount

  return (
    <div className={cn('flex flex-wrap items-center gap-3', className)}>
      {/* Sort Controls */}
      <div className="flex items-center gap-2">
        <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
        <Select value={state.sortField} onValueChange={(v) => handleSortFieldChange(v as SortField)}>
          <SelectTrigger className="w-[140px] h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {sortOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="sm"
          onClick={handleSortOrderToggle}
          className="h-8 w-8 p-0"
        >
          {state.sortOrder === 'asc' ? (
            <ArrowUp className="h-4 w-4" />
          ) : (
            <ArrowDown className="h-4 w-4" />
          )}
        </Button>
      </div>

      {/* Filter Controls */}
      <div className="flex items-center gap-2">
        <Button
          variant={showFilters ? 'secondary' : 'outline'}
          size="sm"
          onClick={() => setShowFilters(!showFilters)}
          className="h-8"
        >
          <Filter className="h-4 w-4 mr-1" />
          Filter
          {activeFilterCount > 0 && (
            <Badge variant="default" className="ml-1 h-5 px-1.5 text-xs">
              {activeFilterCount}
            </Badge>
          )}
        </Button>

        {showFilters && (
          <>
            <Select value={state.filterStatus} onValueChange={(v) => handleFilterChange(v as FilterStatus)}>
              <SelectTrigger className="w-[140px] h-8 text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {FILTER_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div className="flex items-center gap-2">
              <Switch
                id="hide-collections"
                checked={state.hideCompilations}
                onCheckedChange={handleHideCompilationsToggle}
              />
              <Label htmlFor="hide-collections" className="text-sm text-muted-foreground cursor-pointer">
                Hide Collections
              </Label>
            </div>
          </>
        )}

        {hasActiveFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={handleResetFilters}
            className="h-8 px-2 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4 mr-1" />
            Clear
          </Button>
        )}
      </div>

      {/* View Toggle */}
      {showViewToggle && (
        <div className="flex items-center gap-1 ml-auto">
          <Button
            variant={state.viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => state.viewMode !== 'grid' && handleViewModeToggle()}
            className="h-8 w-8 p-0"
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
          <Button
            variant={state.viewMode === 'list' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => state.viewMode !== 'list' && handleViewModeToggle()}
            className="h-8 w-8 p-0"
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Count Display */}
      {totalCount !== undefined && (
        <span className="text-sm text-muted-foreground ml-2">
          {isFiltered ? `${filteredCount} of ${totalCount}` : `${totalCount}`} books
        </span>
      )}
    </div>
  )
}

// Helper function to sort books
export function sortBooks<T extends { 
  title?: string
  releaseYear?: number
  releaseDate?: string
  rating?: number
  seriesIndex?: number | null
  index?: number | null
}>(
  books: T[],
  sortField: SortField,
  sortOrder: SortOrder
): T[] {
  return [...books].sort((a, b) => {
    let comparison = 0
    
    switch (sortField) {
      case 'title':
        comparison = (a.title || '').localeCompare(b.title || '')
        break
      case 'releaseDate':
        const aYear = a.releaseYear || (a.releaseDate ? new Date(a.releaseDate).getFullYear() : 0)
        const bYear = b.releaseYear || (b.releaseDate ? new Date(b.releaseDate).getFullYear() : 0)
        comparison = aYear - bYear
        break
      case 'rating':
        comparison = (a.rating || 0) - (b.rating || 0)
        break
      case 'seriesIndex':
        const aIndex = a.seriesIndex ?? a.index ?? 999
        const bIndex = b.seriesIndex ?? b.index ?? 999
        comparison = aIndex - bIndex
        break
    }
    
    return sortOrder === 'asc' ? comparison : -comparison
  })
}

// Helper function to filter books
export function filterBooks<T extends {
  inLibrary?: boolean
  libraryBook?: { status?: string } | null
  book?: { status?: string } | null
  compilation?: boolean
}>(
  books: T[],
  filterStatus: FilterStatus,
  hideCompilations: boolean = true
): T[] {
  return books.filter((book) => {
    if (hideCompilations && book.compilation) {
      return false
    }

    if (filterStatus === 'all') return true
    
    const status = book.libraryBook?.status || book.book?.status
    
    switch (filterStatus) {
      case 'inLibrary':
        return book.inLibrary === true
      case 'notInLibrary':
        return book.inLibrary !== true
      case 'downloaded':
        return status === 'downloaded'
      case 'missing':
        return book.inLibrary === true && status === 'missing'
      default:
        return true
    }
  })
}

// Default state factory
export function getDefaultSortFilterState(forSeries: boolean = false): SortFilterState {
  return {
    sortField: forSeries ? 'seriesIndex' : 'title',
    sortOrder: 'asc',
    filterStatus: 'all',
    viewMode: 'grid',
    hideCompilations: true,
  }
}

export default BookSortFilter
