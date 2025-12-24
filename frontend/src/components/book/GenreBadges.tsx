import { Badge } from '@/components/ui/badge'

interface GenreBadgesProps {
  genres?: string[]
}

export function GenreBadges({ genres }: GenreBadgesProps) {
  if (!genres || genres.length === 0) return null

  const displayGenres = genres.slice(0, 5)
  const remainingCount = genres.length - 5

  return (
    <div className="flex flex-wrap gap-2">
      {displayGenres.map((genre) => (
        <Badge key={genre} variant="secondary" className="hover:bg-secondary/80">
          {genre}
        </Badge>
      ))}
      {remainingCount > 0 && (
        <Badge variant="outline" className="text-muted-foreground">
          +{remainingCount} more
        </Badge>
      )}
    </div>
  )
}
