import { LibraryStats as Stats } from '@/types'
import { formatFileSize } from '@/lib/utils'
import { Book, BookOpen, Download, HardDrive, Users, Layers } from 'lucide-react'

interface LibraryStatsProps {
  stats: Stats | undefined
  isLoading?: boolean
}

export function LibraryStats({ stats, isLoading }: LibraryStatsProps) {
  const statItems = [
    { label: 'Total Books', value: stats?.totalBooks || 0, icon: Book },
    { label: 'Downloaded', value: stats?.downloadedBooks || 0, icon: Download },
    { label: 'Missing', value: stats?.missingBooks || 0, icon: BookOpen },
    { label: 'Authors', value: stats?.totalAuthors || 0, icon: Users },
    { label: 'Series', value: stats?.totalSeries || 0, icon: Layers },
    { label: 'Library Size', value: formatFileSize(stats?.totalFileSize || 0), icon: HardDrive },
  ]

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
      {statItems.map((item) => (
        <div
          key={item.label}
          className="flex items-center gap-3 rounded-lg border border-border bg-card p-4"
        >
          <div className="rounded-md bg-primary/10 p-2">
            <item.icon className="h-5 w-5 text-primary" />
          </div>
          <div>
            {isLoading ? (
              <div className="h-6 w-12 skeleton rounded" />
            ) : (
              <p className="text-lg font-semibold">{item.value}</p>
            )}
            <p className="text-xs text-muted-foreground">{item.label}</p>
          </div>
        </div>
      ))}
    </div>
  )
}

