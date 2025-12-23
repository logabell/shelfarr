import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search, Filter, Grid, List, SortAsc } from 'lucide-react'

interface LibraryFiltersProps {
  searchQuery: string
  onSearchChange: (query: string) => void
  statusFilter: string
  onStatusChange: (status: string) => void
  sortBy: string
  onSortChange: (sort: string) => void
  viewMode: 'grid' | 'list'
  onViewModeChange: (mode: 'grid' | 'list') => void
  // Optional additional filters
  monitoredFilter?: string
  onMonitoredChange?: (value: string) => void
}

const statusOptions = [
  { value: '', label: 'All Status' },
  { value: 'missing', label: 'Missing' },
  { value: 'downloaded', label: 'Downloaded' },
  { value: 'downloading', label: 'Downloading' },
  { value: 'unreleased', label: 'Unreleased' },
]

const monitoredOptions = [
  { value: '', label: 'All' },
  { value: 'monitored', label: 'Monitored' },
  { value: 'unmonitored', label: 'Unmonitored' },
]

const sortOptions = [
  { value: 'title', label: 'Title' },
  { value: 'sortTitle', label: 'Sort Title' },
  { value: 'rating', label: 'Rating' },
  { value: 'releaseDate', label: 'Release Date' },
  { value: 'created_at', label: 'Date Added' },
]

export function LibraryFilters({
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusChange,
  sortBy,
  onSortChange,
  viewMode,
  onViewModeChange,
  monitoredFilter,
  onMonitoredChange,
}: LibraryFiltersProps) {
  return (
    <div className="flex flex-wrap items-center gap-4">
      {/* Search */}
      <div className="relative flex-1 min-w-[200px]">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Filter library..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Status Filter */}
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <select
          value={statusFilter}
          onChange={(e) => onStatusChange(e.target.value)}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
        >
          {statusOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {/* Monitored Filter */}
      {onMonitoredChange && (
        <select
          value={monitoredFilter || ''}
          onChange={(e) => onMonitoredChange(e.target.value)}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
        >
          {monitoredOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      )}

      {/* Sort */}
      <div className="flex items-center gap-2">
        <SortAsc className="h-4 w-4 text-muted-foreground" />
        <select
          value={sortBy}
          onChange={(e) => onSortChange(e.target.value)}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
        >
          {sortOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {/* View Mode Toggle */}
      <div className="flex rounded-md border border-input">
        <Button
          variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
          size="icon"
          onClick={() => onViewModeChange('grid')}
          className="rounded-r-none"
        >
          <Grid className="h-4 w-4" />
        </Button>
        <Button
          variant={viewMode === 'list' ? 'secondary' : 'ghost'}
          size="icon"
          onClick={() => onViewModeChange('list')}
          className="rounded-l-none"
        >
          <List className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
