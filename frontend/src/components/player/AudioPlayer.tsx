import { useRef, useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { 
  Play, 
  Pause, 
  SkipBack, 
  SkipForward, 
  Volume2, 
  VolumeX,
  X,
  List,
  Rewind,
  FastForward
} from 'lucide-react'
import { cn, formatDuration } from '@/lib/utils'

interface Chapter {
  title: string
  startTime: number
  endTime: number
}

interface AudioPlayerProps {
  url: string
  title: string
  author?: string
  coverUrl?: string
  chapters?: Chapter[]
  onClose?: () => void
  onProgressChange?: (progress: number, position: number) => void
  initialPosition?: number
}

const PLAYBACK_RATES = [0.5, 0.75, 1, 1.25, 1.5, 2, 3]

export function AudioPlayer({ 
  url, 
  title, 
  author, 
  coverUrl, 
  chapters = [],
  onClose, 
  onProgressChange,
  initialPosition = 0 
}: AudioPlayerProps) {
  const audioRef = useRef<HTMLAudioElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(initialPosition)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [playbackRate, setPlaybackRate] = useState(1)
  const [showChapters, setShowChapters] = useState(false)
  const [currentChapter, setCurrentChapter] = useState<Chapter | null>(null)

  // Initialize audio
  useEffect(() => {
    const audio = audioRef.current
    if (!audio) return

    const handleLoadedMetadata = () => {
      setDuration(audio.duration)
      if (initialPosition > 0) {
        audio.currentTime = initialPosition
      }
    }

    const handleTimeUpdate = () => {
      setCurrentTime(audio.currentTime)
      
      // Find current chapter
      if (chapters.length > 0) {
        const chapter = chapters.find(
          (c) => audio.currentTime >= c.startTime && audio.currentTime < c.endTime
        )
        setCurrentChapter(chapter || null)
      }

      // Report progress
      if (onProgressChange && audio.duration > 0) {
        onProgressChange(audio.currentTime / audio.duration, audio.currentTime)
      }
    }

    const handleEnded = () => {
      setIsPlaying(false)
    }

    audio.addEventListener('loadedmetadata', handleLoadedMetadata)
    audio.addEventListener('timeupdate', handleTimeUpdate)
    audio.addEventListener('ended', handleEnded)

    return () => {
      audio.removeEventListener('loadedmetadata', handleLoadedMetadata)
      audio.removeEventListener('timeupdate', handleTimeUpdate)
      audio.removeEventListener('ended', handleEnded)
    }
  }, [chapters, initialPosition, onProgressChange])

  // Update playback rate
  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.playbackRate = playbackRate
    }
  }, [playbackRate])

  // Update volume
  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.volume = isMuted ? 0 : volume
    }
  }, [volume, isMuted])

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement) return

      switch (e.key) {
        case ' ':
          e.preventDefault()
          togglePlay()
          break
        case 'ArrowLeft':
          skip(-10)
          break
        case 'ArrowRight':
          skip(10)
          break
        case 'ArrowUp':
          setVolume((v) => Math.min(1, v + 0.1))
          break
        case 'ArrowDown':
          setVolume((v) => Math.max(0, v - 0.1))
          break
        case 'm':
          setIsMuted((m) => !m)
          break
        case 'Escape':
          if (onClose) onClose()
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  const togglePlay = useCallback(() => {
    if (!audioRef.current) return
    
    if (isPlaying) {
      audioRef.current.pause()
    } else {
      audioRef.current.play()
    }
    setIsPlaying(!isPlaying)
  }, [isPlaying])

  const skip = useCallback((seconds: number) => {
    if (!audioRef.current) return
    audioRef.current.currentTime = Math.max(0, Math.min(duration, currentTime + seconds))
  }, [currentTime, duration])

  const seekTo = useCallback((time: number) => {
    if (!audioRef.current) return
    audioRef.current.currentTime = time
  }, [])

  const goToChapter = useCallback((chapter: Chapter) => {
    seekTo(chapter.startTime)
    setShowChapters(false)
  }, [seekTo])

  const cyclePlaybackRate = useCallback(() => {
    const currentIndex = PLAYBACK_RATES.indexOf(playbackRate)
    const nextIndex = (currentIndex + 1) % PLAYBACK_RATES.length
    setPlaybackRate(PLAYBACK_RATES[nextIndex])
  }, [playbackRate])

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-background">
      {/* Hidden audio element */}
      <audio ref={audioRef} src={url} preload="metadata" />

      {/* Header */}
      <header className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex items-center gap-4">
          {onClose && (
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-5 w-5" />
            </Button>
          )}
          {chapters.length > 0 && (
            <Button variant="ghost" size="icon" onClick={() => setShowChapters(!showChapters)}>
              <List className="h-5 w-5" />
            </Button>
          )}
        </div>

        <div className="flex-1 text-center">
          <h1 className="font-semibold truncate max-w-md mx-auto">{title}</h1>
          {author && <p className="text-sm text-muted-foreground">{author}</p>}
        </div>

        <div className="w-20" /> {/* Spacer for alignment */}
      </header>

      {/* Main content */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="flex flex-col items-center max-w-md w-full">
          {/* Cover art */}
          <div className="w-64 h-64 mb-8 rounded-lg overflow-hidden shadow-2xl">
            {coverUrl ? (
              <img src={coverUrl} alt={title} className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full bg-muted flex items-center justify-center">
                <Volume2 className="h-16 w-16 text-muted-foreground" />
              </div>
            )}
          </div>

          {/* Chapter info */}
          {currentChapter && (
            <p className="text-sm text-muted-foreground mb-4 text-center">
              {currentChapter.title}
            </p>
          )}

          {/* Progress bar */}
          <div className="w-full mb-4">
            <input
              type="range"
              min={0}
              max={duration || 100}
              value={currentTime}
              onChange={(e) => seekTo(Number(e.target.value))}
              className="w-full h-2 bg-muted rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:bg-primary [&::-webkit-slider-thumb]:rounded-full"
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{formatDuration(Math.floor(currentTime))}</span>
              <span>{formatDuration(Math.floor(duration))}</span>
            </div>
          </div>

          {/* Controls */}
          <div className="flex items-center gap-4 mb-8">
            <Button variant="ghost" size="icon" onClick={() => skip(-30)}>
              <Rewind className="h-5 w-5" />
            </Button>
            <Button variant="ghost" size="icon" onClick={() => skip(-10)}>
              <SkipBack className="h-5 w-5" />
            </Button>
            <Button
              size="icon"
              className="h-16 w-16 rounded-full"
              onClick={togglePlay}
            >
              {isPlaying ? (
                <Pause className="h-8 w-8" />
              ) : (
                <Play className="h-8 w-8 ml-1" />
              )}
            </Button>
            <Button variant="ghost" size="icon" onClick={() => skip(10)}>
              <SkipForward className="h-5 w-5" />
            </Button>
            <Button variant="ghost" size="icon" onClick={() => skip(30)}>
              <FastForward className="h-5 w-5" />
            </Button>
          </div>

          {/* Secondary controls */}
          <div className="flex items-center gap-6">
            {/* Playback rate */}
            <Button
              variant="outline"
              size="sm"
              onClick={cyclePlaybackRate}
              className="min-w-[60px]"
            >
              {playbackRate}x
            </Button>

            {/* Volume */}
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsMuted(!isMuted)}
              >
                {isMuted || volume === 0 ? (
                  <VolumeX className="h-5 w-5" />
                ) : (
                  <Volume2 className="h-5 w-5" />
                )}
              </Button>
              <input
                type="range"
                min={0}
                max={1}
                step={0.1}
                value={isMuted ? 0 : volume}
                onChange={(e) => {
                  setVolume(Number(e.target.value))
                  setIsMuted(false)
                }}
                className="w-24 h-1 bg-muted rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:bg-primary [&::-webkit-slider-thumb]:rounded-full"
              />
            </div>
          </div>
        </div>
      </div>

      {/* Chapters sidebar */}
      {showChapters && chapters.length > 0 && (
        <div className="absolute right-0 top-16 bottom-0 w-80 bg-card border-l border-border overflow-y-auto">
          <div className="p-4">
            <h3 className="font-semibold mb-4">Chapters</h3>
            <nav className="space-y-1">
              {chapters.map((chapter, idx) => (
                <button
                  key={idx}
                  onClick={() => goToChapter(chapter)}
                  className={cn(
                    'block w-full text-left px-3 py-2 text-sm rounded-md transition-colors',
                    currentChapter?.title === chapter.title
                      ? 'bg-primary text-primary-foreground'
                      : 'hover:bg-accent'
                  )}
                >
                  <span className="block truncate">{chapter.title}</span>
                  <span className="text-xs text-muted-foreground">
                    {formatDuration(Math.floor(chapter.startTime))}
                  </span>
                </button>
              ))}
            </nav>
          </div>
        </div>
      )}
    </div>
  )
}

