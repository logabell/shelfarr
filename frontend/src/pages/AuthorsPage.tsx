import { useQuery } from '@tanstack/react-query'
import { Topbar } from '@/components/layout/Topbar'
import { getAuthors } from '@/api/client'
import { Link } from 'react-router-dom'
import { User, ChevronRight, BookOpen, CheckCircle2, Library, Eye } from 'lucide-react'

export function AuthorsPage() {
  const { data: authors, isLoading, refetch } = useQuery({
    queryKey: ['authors'],
    queryFn: () => getAuthors(),
  })

  return (
    <div className="flex flex-col h-full">
      <Topbar 
        title="Authors" 
        subtitle={`${authors?.length || 0} authors`}
        onRefresh={() => refetch()}
        isRefreshing={isLoading}
      />
      
      <div className="flex-1 overflow-auto p-6">
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {Array.from({ length: 12 }).map((_, i) => (
              <div key={i} className="h-28 skeleton rounded-lg" />
            ))}
          </div>
        ) : authors?.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16">
            <User className="h-16 w-16 text-muted-foreground mb-4" />
            <p className="text-lg text-muted-foreground">No authors found</p>
            <p className="text-sm text-muted-foreground mt-1">
              Add books to start tracking authors
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {authors?.map((author) => {
              const totalBooks = author.totalBooksCount || author.bookCount || 0
              const inLibrary = author.bookCount || 0
              const downloaded = author.downloadedCount || 0
              const progressPercent = totalBooks > 0 ? (downloaded / totalBooks) * 100 : 0
              const hasHardcoverData = (author.totalBooksCount || 0) > 0

              return (
                <Link
                  key={author.id}
                  to={`/authors/${author.id}`}
                  className="flex items-center gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary transition-colors"
                >
                  {author.imageUrl ? (
                    <img
                      src={author.imageUrl}
                      alt={author.name}
                      className="h-16 w-16 rounded-full object-cover shrink-0"
                    />
                  ) : (
                    <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center shrink-0">
                      <User className="h-8 w-8 text-muted-foreground" />
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium truncate">{author.name}</h3>
                      {author.monitored && (
                        <Eye className="h-3.5 w-3.5 text-primary shrink-0" />
                      )}
                    </div>
                    
                    {/* Stats Row */}
                    <div className="flex items-center gap-3 mt-1 text-xs">
                      {hasHardcoverData ? (
                        <>
                          <span className="flex items-center gap-1 text-muted-foreground">
                            <BookOpen className="h-3 w-3" />
                            {totalBooks}
                          </span>
                          <span className="flex items-center gap-1 text-sky-400">
                            <Library className="h-3 w-3" />
                            {inLibrary}
                          </span>
                          <span className="flex items-center gap-1 text-green-400">
                            <CheckCircle2 className="h-3 w-3" />
                            {downloaded}
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
                      <div className="mt-2">
                        <div className="h-1 bg-neutral-700 rounded-full overflow-hidden">
                          <div 
                            className="h-full bg-gradient-to-r from-green-600 to-green-400 rounded-full transition-all"
                            style={{ width: `${progressPercent}%` }}
                          />
                        </div>
                      </div>
                    )}
                  </div>
                  
                  {/* Right side */}
                  <div className="flex items-center gap-2 shrink-0">
                    {hasHardcoverData && (
                      <span className="text-xs font-medium text-neutral-400">
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

