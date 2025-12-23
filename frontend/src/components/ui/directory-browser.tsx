import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { 
  Folder, 
  FolderOpen, 
  ChevronRight,
  ArrowUp,
  Loader2,
  Home
} from 'lucide-react'
import { Button } from './button'
import { Input } from './input'
import { browseFilesystem, type DirectoryInfo } from '@/api/client'
import { cn } from '@/lib/utils'

interface DirectoryBrowserProps {
  value: string
  onChange: (path: string) => void
  className?: string
}

export function DirectoryBrowser({ value, onChange, className }: DirectoryBrowserProps) {
  const [currentPath, setCurrentPath] = useState(value || '/')
  const [isTyping, setIsTyping] = useState(false)
  const [inputValue, setInputValue] = useState(value || '/')

  // Fetch directories for current path
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['filesystem', currentPath],
    queryFn: () => browseFilesystem(currentPath),
    staleTime: 30000, // Cache for 30 seconds
  })

  // Sync value with input
  useEffect(() => {
    if (!isTyping) {
      setInputValue(value || '/')
      setCurrentPath(value || '/')
    }
  }, [value, isTyping])

  const handleDirectoryClick = (dir: DirectoryInfo) => {
    setCurrentPath(dir.path)
    setInputValue(dir.path)
    onChange(dir.path)
  }

  const handleNavigateUp = () => {
    if (data?.parent) {
      setCurrentPath(data.parent)
      setInputValue(data.parent)
      onChange(data.parent)
    }
  }

  const handleNavigateRoot = () => {
    setCurrentPath('/')
    setInputValue('/')
    onChange('/')
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value)
    setIsTyping(true)
  }

  const handleInputBlur = () => {
    setIsTyping(false)
    if (inputValue !== currentPath) {
      setCurrentPath(inputValue)
      onChange(inputValue)
    }
  }

  const handleInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      setIsTyping(false)
      if (inputValue !== currentPath) {
        setCurrentPath(inputValue)
        onChange(inputValue)
      }
    }
  }

  return (
    <div className={cn("border border-border rounded-lg overflow-hidden", className)}>
      {/* Path input and navigation */}
      <div className="flex items-center gap-1 p-2 bg-muted/50 border-b border-border">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleNavigateRoot}
          className="h-8 w-8 p-0"
          title="Go to root"
        >
          <Home className="h-4 w-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleNavigateUp}
          disabled={!data?.parent}
          className="h-8 w-8 p-0"
          title="Go up one level"
        >
          <ArrowUp className="h-4 w-4" />
        </Button>
        <Input
          value={inputValue}
          onChange={handleInputChange}
          onBlur={handleInputBlur}
          onKeyDown={handleInputKeyDown}
          className="flex-1 h-8 font-mono text-sm"
          placeholder="/path/to/folder"
        />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => refetch()}
          className="h-8 w-8 p-0"
          title="Refresh"
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>

      {/* Directory listing */}
      <div className="h-64 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="flex items-center justify-center h-full p-4 text-center">
            <p className="text-sm text-muted-foreground">
              Cannot access this path. Please enter a valid directory path.
            </p>
          </div>
        ) : data?.directories.length === 0 ? (
          <div className="flex items-center justify-center h-full p-4 text-center">
            <p className="text-sm text-muted-foreground">
              No subdirectories found in this location
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {data?.directories.map((dir) => (
              <DirectoryItem
                key={dir.path}
                directory={dir}
                isSelected={dir.path === value}
                onClick={() => handleDirectoryClick(dir)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Selected path display */}
      <div className="p-2 bg-muted/30 border-t border-border">
        <p className="text-xs text-muted-foreground">
          Selected: <span className="font-mono text-foreground">{value || 'None'}</span>
        </p>
      </div>
    </div>
  )
}

function DirectoryItem({ 
  directory, 
  isSelected, 
  onClick 
}: { 
  directory: DirectoryInfo
  isSelected: boolean
  onClick: () => void 
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-muted/50 transition-colors",
        isSelected && "bg-primary/10 text-primary"
      )}
    >
      {directory.hasChildren ? (
        <FolderOpen className="h-4 w-4 text-yellow-500 shrink-0" />
      ) : (
        <Folder className="h-4 w-4 text-yellow-500 shrink-0" />
      )}
      <span className="text-sm truncate flex-1">{directory.name}</span>
      {directory.hasChildren && (
        <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
      )}
    </button>
  )
}

export default DirectoryBrowser
