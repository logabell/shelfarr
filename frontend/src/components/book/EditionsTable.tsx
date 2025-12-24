import { useQuery } from '@tanstack/react-query'
import { Book, Calendar, Layers, Clock, Building2 } from 'lucide-react'
import { getBookEditions } from '@/api/client'
import { Skeleton } from '@/components/ui/skeleton'

interface EditionsTableProps {
  bookId: number
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours} hr ${minutes} min`
  return `${minutes} min`
}

function formatDate(dateString?: string): string {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

export function EditionsTable({ bookId }: EditionsTableProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['editions', bookId],
    queryFn: () => getBookEditions(bookId),
  })

  if (isLoading) {
    return (
      <div className="space-y-2">
        <div className="rounded-lg border border-border bg-card overflow-hidden">
          <div className="p-4 space-y-4">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        </div>
      </div>
    )
  }

  if (!data?.editions || data.editions.length === 0) {
    return <div className="text-sm text-muted-foreground italic">No editions found.</div>
  }

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/30">
              <th className="h-10 px-4 text-left font-medium text-muted-foreground w-16">Cover</th>
              <th className="h-10 px-4 text-left font-medium text-muted-foreground">Format</th>
              <th className="h-10 px-4 text-left font-medium text-muted-foreground">Publisher</th>
              <th className="h-10 px-4 text-left font-medium text-muted-foreground">Length</th>
              <th className="h-10 px-4 text-left font-medium text-muted-foreground">Release Date</th>
            </tr>
          </thead>
          <tbody>
            {data.editions.map((edition) => (
              <tr key={edition.id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                <td className="p-2">
                  <div className="h-12 w-8 bg-muted rounded overflow-hidden shadow-sm">
                    {edition.coverUrl ? (
                      <img src={edition.coverUrl} alt="" className="h-full w-full object-cover" loading="lazy" />
                    ) : (
                      <div className="h-full w-full flex items-center justify-center">
                        <Book className="h-4 w-4 text-muted-foreground/50" />
                      </div>
                    )}
                  </div>
                </td>
                <td className="p-4 align-middle">
                  <div className="font-medium">{edition.editionFormat || edition.format}</div>
                  <div className="text-xs text-muted-foreground mt-0.5">
                    {edition.isbn13 || edition.isbn10 || edition.asin || '-'}
                  </div>
                </td>
                <td className="p-4 align-middle">
                  <div className="flex items-center gap-2">
                    <Building2 className="h-3 w-3 text-muted-foreground" />
                    <span className="truncate max-w-[200px]" title={edition.publisherName}>
                      {edition.publisherName || '-'}
                    </span>
                  </div>
                </td>
                <td className="p-4 align-middle">
                  {edition.audioSeconds ? (
                    <div className="flex items-center gap-2">
                      <Clock className="h-3 w-3 text-muted-foreground" />
                      <span>{formatDuration(edition.audioSeconds)}</span>
                    </div>
                  ) : edition.pageCount ? (
                    <div className="flex items-center gap-2">
                      <Layers className="h-3 w-3 text-muted-foreground" />
                      <span>{edition.pageCount} pages</span>
                    </div>
                  ) : (
                    <span className="text-muted-foreground">-</span>
                  )}
                </td>
                <td className="p-4 align-middle">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-3 w-3 text-muted-foreground" />
                    <span>{formatDate(edition.releaseDate)}</span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
