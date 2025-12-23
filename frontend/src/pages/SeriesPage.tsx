import { useQuery } from '@tanstack/react-query'
import { Topbar } from '@/components/layout/Topbar'
import { getSeries } from '@/api/client'
import { Link } from 'react-router-dom'
import { Layers, ChevronRight, BookOpen, CheckCircle2, Library } from 'lucide-react'

export function SeriesPage() {
  const { data: seriesList, isLoading, refetch } = useQuery({
    queryKey: ['series'],
    queryFn: getSeries,
  })

  return (
    <div className="flex flex-col h-full">
      <Topbar 
        title="Series" 
        subtitle={`${seriesList?.length || 0} series`}
        onRefresh={() => refetch()}
        isRefreshing={isLoading}
      />
      
      <div className="flex-1 overflow-auto p-6">
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="h-24 skeleton rounded-lg" />
            ))}
          </div>
        ) : seriesList?.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16">
            <Layers className="h-16 w-16 text-muted-foreground mb-4" />
            <p className="text-lg text-muted-foreground">No series found</p>
            <p className="text-sm text-muted-foreground mt-1">
              Add books that are part of a series to see them here
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {seriesList?.map((series) => {
              const totalBooks = series.totalBooksCount || series.bookCount || 0
              const inLibrary = series.bookCount || 0
              const downloaded = series.downloadedCount || 0
              const progressPercent = totalBooks > 0 ? (downloaded / totalBooks) * 100 : 0
              const hasHardcoverData = (series.totalBooksCount || 0) > 0

              return (
                <Link
                  key={series.id}
                  to={`/series/${series.id}`}
                  className="flex items-center justify-between p-4 rounded-lg bg-card border border-border hover:border-primary transition-colors"
                >
                  <div className="flex items-center gap-4 flex-1 min-w-0">
                    <div className="rounded-md bg-primary/10 p-2 shrink-0">
                      <Layers className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium truncate">{series.name}</h3>
                      
                      {/* Stats Row */}
                      <div className="flex items-center gap-4 mt-1 text-sm">
                        {hasHardcoverData ? (
                          <>
                            <span className="flex items-center gap-1 text-muted-foreground">
                              <BookOpen className="h-3.5 w-3.5" />
                              {totalBooks} books
                            </span>
                            <span className="flex items-center gap-1 text-sky-400">
                              <Library className="h-3.5 w-3.5" />
                              {inLibrary} in library
                            </span>
                            <span className="flex items-center gap-1 text-green-400">
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              {downloaded} downloaded
                            </span>
                          </>
                        ) : (
                          <span className="text-muted-foreground">
                            {inLibrary} book{inLibrary !== 1 ? 's' : ''} in library
                          </span>
                        )}
                      </div>

                      {/* Progress Bar (only show if we have Hardcover data) */}
                      {hasHardcoverData && (
                        <div className="mt-2 max-w-xs">
                          <div className="h-1.5 bg-neutral-700 rounded-full overflow-hidden">
                            <div 
                              className="h-full bg-gradient-to-r from-green-600 to-green-400 rounded-full transition-all"
                              style={{ width: `${progressPercent}%` }}
                            />
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                  
                  {/* Right side: progress percentage or chevron */}
                  <div className="flex items-center gap-3 shrink-0 ml-4">
                    {hasHardcoverData && (
                      <span className="text-sm font-medium text-neutral-400">
                        {Math.round(progressPercent)}%
                      </span>
                    )}
                    <ChevronRight className="h-5 w-5 text-muted-foreground" />
                  </div>
                </Link>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

