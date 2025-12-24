import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { User, Mic, Edit, PenTool, Languages, Users } from 'lucide-react'
import { getBookContributors } from '@/api/client'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Contributor, ContributorRole } from '@/types'

interface ContributorsListProps {
  bookId: number
}

const roleIcons: Record<ContributorRole, React.ReactNode> = {
  Author: <User className="h-3 w-3" />,
  Narrator: <Mic className="h-3 w-3" />,
  Editor: <Edit className="h-3 w-3" />,
  Illustrator: <PenTool className="h-3 w-3" />,
  Translator: <Languages className="h-3 w-3" />,
  Contributor: <Users className="h-3 w-3" />,
}

export function ContributorsList({ bookId }: ContributorsListProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['contributors', bookId],
    queryFn: () => getBookContributors(bookId),
  })

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center gap-3 p-3 rounded-lg border border-border bg-card">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="space-y-2">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-16" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (!data?.contributors || data.contributors.length === 0) {
    return <div className="text-sm text-muted-foreground italic">No contributors found.</div>
  }

  const grouped = data.contributors.reduce((acc, contributor) => {
    const role = contributor.role
    if (!acc[role]) acc[role] = []
    acc[role].push(contributor)
    return acc
  }, {} as Record<ContributorRole, Contributor[]>)

  const roleOrder: ContributorRole[] = ['Author', 'Narrator', 'Editor', 'Illustrator', 'Translator', 'Contributor']
  const sortedRoles = roleOrder.filter(role => grouped[role] && grouped[role].length > 0)

  Object.keys(grouped).forEach((role) => {
    if (!roleOrder.includes(role as ContributorRole)) {
      sortedRoles.push(role as ContributorRole)
    }
  })

  return (
    <div className="space-y-6">
      {sortedRoles.map((role) => (
        <div key={role}>
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-2 uppercase tracking-wider">
            {roleIcons[role]}
            {role}s
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {grouped[role].map((contributor) => (
              <div key={contributor.id} className="flex items-center gap-3 p-3 rounded-lg border border-border bg-card hover:border-primary/50 transition-colors">
                {contributor.authorImage ? (
                  <img
                    src={contributor.authorImage}
                    alt={contributor.authorName}
                    className="h-10 w-10 rounded-full object-cover bg-muted"
                  />
                ) : (
                  <div className="h-10 w-10 rounded-full bg-muted flex items-center justify-center shrink-0">
                    <User className="h-5 w-5 text-muted-foreground" />
                  </div>
                )}
                <div className="min-w-0">
                  <Link
                    to={contributor.authorId ? `/authors/${contributor.authorId}` : '#'}
                    className={`font-medium block truncate ${contributor.authorId ? 'hover:text-primary hover:underline' : 'pointer-events-none'}`}
                    title={contributor.authorName}
                  >
                    {contributor.authorName}
                  </Link>
                  <div className="flex mt-1">
                    <Badge variant="outline" className="text-[10px] h-5 px-1.5 font-normal">
                      {contributor.role}
                    </Badge>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
