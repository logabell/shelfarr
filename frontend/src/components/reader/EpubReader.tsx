import { useEffect, useRef, useState, useCallback } from 'react'
import ePub, { Book, Rendition } from 'epubjs'
import { Button } from '@/components/ui/button'
import { 
  ChevronLeft, 
  ChevronRight, 
  Settings, 
  X,
  Sun,
  Moon,
  Coffee,
  Minus,
  Plus,
  List
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface EpubReaderProps {
  url: string
  onClose?: () => void
  onProgressChange?: (progress: number, position: string) => void
  initialProgress?: number
}

type ThemeType = 'light' | 'dark' | 'sepia'

interface TocItem {
  href: string
  label: string
  subitems?: TocItem[]
}

export function EpubReader({ url, onClose, onProgressChange, initialProgress }: EpubReaderProps) {
  const viewerRef = useRef<HTMLDivElement>(null)
  const bookRef = useRef<Book | null>(null)
  const renditionRef = useRef<Rendition | null>(null)
  
  const [isLoading, setIsLoading] = useState(true)
  const [, setCurrentLocation] = useState('')
  const [progress, setProgress] = useState(0)
  const [theme, setTheme] = useState<ThemeType>('dark')
  const [fontSize, setFontSize] = useState(100)
  const [showSettings, setShowSettings] = useState(false)
  const [showToc, setShowToc] = useState(false)
  const [toc, setToc] = useState<TocItem[]>([])
  const [chapterTitle, setChapterTitle] = useState('')

  // Theme configurations
  const themes = {
    light: {
      body: { background: '#ffffff', color: '#1a1a1a' },
    },
    dark: {
      body: { background: '#0f0f0f', color: '#e0e0e0' },
    },
    sepia: {
      body: { background: '#f4ecd8', color: '#5b4636' },
    },
  }

  // Initialize the book
  useEffect(() => {
    if (!viewerRef.current) return

    const book = ePub(url)
    bookRef.current = book

    const rendition = book.renderTo(viewerRef.current, {
      width: '100%',
      height: '100%',
      spread: 'none',
    })

    renditionRef.current = rendition

    // Register themes
    Object.entries(themes).forEach(([name, styles]) => {
      rendition.themes.register(name, styles)
    })
    rendition.themes.select(theme)
    rendition.themes.fontSize(`${fontSize}%`)

    // Load the book
    book.ready.then(() => {
      setIsLoading(false)
      
      // Get table of contents
      book.loaded.navigation.then((nav) => {
        setToc(nav.toc as TocItem[])
      })

      // Go to initial location if provided
      if (initialProgress && initialProgress > 0) {
        const cfi = book.locations.cfiFromPercentage(initialProgress)
        rendition.display(cfi)
      } else {
        rendition.display()
      }
    })

    // Track location changes
    rendition.on('relocated', (location: { start: { cfi: string; percentage: number; displayed: { page: number } } }) => {
      const percent = location.start.percentage
      setProgress(percent)
      setCurrentLocation(location.start.cfi)
      
      if (onProgressChange) {
        onProgressChange(percent, location.start.cfi)
      }

      // Update chapter title
      const chapter = book.navigation?.get(location.start.cfi)
      if (chapter) {
        setChapterTitle(chapter.label)
      }
    })

    // Keyboard navigation
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') {
        rendition.prev()
      } else if (e.key === 'ArrowRight') {
        rendition.next()
      } else if (e.key === 'Escape' && onClose) {
        onClose()
      }
    }
    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      book.destroy()
    }
  }, [url])

  // Update theme
  useEffect(() => {
    if (renditionRef.current) {
      renditionRef.current.themes.select(theme)
    }
  }, [theme])

  // Update font size
  useEffect(() => {
    if (renditionRef.current) {
      renditionRef.current.themes.fontSize(`${fontSize}%`)
    }
  }, [fontSize])

  const handlePrev = useCallback(() => {
    renditionRef.current?.prev()
  }, [])

  const handleNext = useCallback(() => {
    renditionRef.current?.next()
  }, [])

  const handleTocClick = useCallback((href: string) => {
    renditionRef.current?.display(href)
    setShowToc(false)
  }, [])

  const getBackgroundClass = () => {
    switch (theme) {
      case 'light': return 'bg-white'
      case 'dark': return 'bg-[#0f0f0f]'
      case 'sepia': return 'bg-[#f4ecd8]'
    }
  }

  return (
    <div className={cn('fixed inset-0 z-50 flex flex-col', getBackgroundClass())}>
      {/* Header */}
      <header className="flex items-center justify-between p-4 border-b border-border/20">
        <div className="flex items-center gap-4">
          {onClose && (
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-5 w-5" />
            </Button>
          )}
          <Button variant="ghost" size="icon" onClick={() => setShowToc(!showToc)}>
            <List className="h-5 w-5" />
          </Button>
        </div>

        <div className="flex-1 text-center">
          <p className="text-sm font-medium truncate max-w-md mx-auto">
            {chapterTitle || 'Loading...'}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={() => setShowSettings(!showSettings)}>
            <Settings className="h-5 w-5" />
          </Button>
        </div>
      </header>

      {/* Main content area */}
      <div className="flex-1 relative overflow-hidden">
        {/* Loading state */}
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        )}

        {/* Table of Contents Sidebar */}
        {showToc && (
          <div className="absolute left-0 top-0 bottom-0 w-80 bg-card border-r border-border z-10 overflow-y-auto">
            <div className="p-4">
              <h3 className="font-semibold mb-4">Table of Contents</h3>
              <nav className="space-y-1">
                {toc.map((item, idx) => (
                  <button
                    key={idx}
                    onClick={() => handleTocClick(item.href)}
                    className="block w-full text-left px-3 py-2 text-sm rounded-md hover:bg-accent transition-colors"
                  >
                    {item.label}
                  </button>
                ))}
              </nav>
            </div>
          </div>
        )}

        {/* Settings Panel */}
        {showSettings && (
          <div className="absolute right-0 top-0 w-72 bg-card border-l border-border z-10 p-4">
            <h3 className="font-semibold mb-4">Settings</h3>
            
            {/* Theme */}
            <div className="mb-4">
              <label className="text-sm text-muted-foreground mb-2 block">Theme</label>
              <div className="flex gap-2">
                <Button
                  variant={theme === 'light' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTheme('light')}
                >
                  <Sun className="h-4 w-4" />
                </Button>
                <Button
                  variant={theme === 'dark' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTheme('dark')}
                >
                  <Moon className="h-4 w-4" />
                </Button>
                <Button
                  variant={theme === 'sepia' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTheme('sepia')}
                >
                  <Coffee className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Font Size */}
            <div>
              <label className="text-sm text-muted-foreground mb-2 block">
                Font Size: {fontSize}%
              </label>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFontSize(Math.max(50, fontSize - 10))}
                >
                  <Minus className="h-4 w-4" />
                </Button>
                <div className="flex-1 h-2 bg-muted rounded-full">
                  <div
                    className="h-full bg-primary rounded-full transition-all"
                    style={{ width: `${((fontSize - 50) / 150) * 100}%` }}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFontSize(Math.min(200, fontSize + 10))}
                >
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        )}

        {/* Reader container */}
        <div
          ref={viewerRef}
          className={cn(
            'h-full mx-auto transition-all',
            showToc ? 'ml-80' : '',
            showSettings ? 'mr-72' : ''
          )}
          style={{ maxWidth: '800px' }}
        />

        {/* Navigation buttons */}
        <Button
          variant="ghost"
          size="icon"
          className="absolute left-4 top-1/2 -translate-y-1/2 h-12 w-12 rounded-full bg-background/80"
          onClick={handlePrev}
        >
          <ChevronLeft className="h-6 w-6" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="absolute right-4 top-1/2 -translate-y-1/2 h-12 w-12 rounded-full bg-background/80"
          onClick={handleNext}
        >
          <ChevronRight className="h-6 w-6" />
        </Button>
      </div>

      {/* Footer with progress */}
      <footer className="p-4 border-t border-border/20">
        <div className="flex items-center gap-4">
          <span className="text-xs text-muted-foreground">
            {Math.round(progress * 100)}%
          </span>
          <div className="flex-1 h-1 bg-muted rounded-full overflow-hidden">
            <div
              className="h-full bg-primary transition-all duration-300"
              style={{ width: `${progress * 100}%` }}
            />
          </div>
        </div>
      </footer>
    </div>
  )
}

