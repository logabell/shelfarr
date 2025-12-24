import { Book, Headphones, Check, HelpCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

interface FormatAvailabilityProps {
  hasEbook?: boolean
  hasAudiobook?: boolean
  compact?: boolean
  className?: string
}

export function FormatAvailability({
  hasEbook,
  hasAudiobook,
  compact = false,
  className,
}: FormatAvailabilityProps) {
  const iconSize = compact ? 'h-3 w-3' : 'h-4 w-4'
  const textSize = compact ? 'text-xs' : 'text-sm'
  const gap = compact ? 'gap-0.5' : 'gap-1'
  const itemGap = compact ? 'gap-2' : 'gap-3'

  return (
    <div className={cn('flex items-center', itemGap, className)}>
      <div className={cn('flex items-center', gap)}>
        <Book className={cn(iconSize, 'text-muted-foreground')} />
        {!compact && <span className={cn(textSize, 'text-muted-foreground')}>Ebook</span>}
        {hasEbook ? (
          <Check className={cn(iconSize, 'text-green-500')} />
        ) : (
          <HelpCircle className={cn(iconSize, 'text-muted-foreground/50')} />
        )}
      </div>

      <div className={cn('flex items-center', gap)}>
        <Headphones className={cn(iconSize, 'text-muted-foreground')} />
        {!compact && <span className={cn(textSize, 'text-muted-foreground')}>Audiobook</span>}
        {hasAudiobook ? (
          <Check className={cn(iconSize, 'text-green-500')} />
        ) : (
          <HelpCircle className={cn(iconSize, 'text-muted-foreground/50')} />
        )}
      </div>
    </div>
  )
}
