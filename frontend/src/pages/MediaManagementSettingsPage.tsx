import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { 
  Plus, 
  Trash2, 
  Loader2, 
  ArrowLeft, 
  FolderOpen, 
  Book, 
  Headphones,
  HardDrive,
  Info,
  RefreshCw,
  Link2,
  Trash
} from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { DirectoryBrowser } from '@/components/ui/directory-browser'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { 
  getMediaSettings, 
  updateMediaSettings, 
  getRootFolders, 
  addRootFolder, 
  deleteRootFolder,
  type MediaSettings,
  type RootFolder
} from '@/api/client'

// Format bytes to human-readable string
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// Available naming tokens
const NAMING_TOKENS = [
  { token: '{Author}', description: 'Author name' },
  { token: '{Title}', description: 'Book title' },
  { token: '{Series}', description: 'Series name (if any)' },
  { token: '{SeriesIndex}', description: 'Position in series' },
  { token: '{Year}', description: 'Release year' },
  { token: '{Format}', description: 'File format (epub, m4b, etc.)' },
]

export function MediaManagementSettingsPage() {
  const queryClient = useQueryClient()
  const [isAddFolderOpen, setIsAddFolderOpen] = useState(false)
  const [newFolderPath, setNewFolderPath] = useState('')
  const [newFolderType, setNewFolderType] = useState<'ebook' | 'audiobook'>('ebook')
  const [newFolderName, setNewFolderName] = useState('')
  const [localSettings, setLocalSettings] = useState<Partial<MediaSettings>>({})
  const [hasChanges, setHasChanges] = useState(false)

  const { data: settings, isLoading: settingsLoading } = useQuery({
    queryKey: ['mediaSettings'],
    queryFn: getMediaSettings,
  })

  const { data: rootFolders, isLoading: foldersLoading } = useQuery({
    queryKey: ['rootFolders'],
    queryFn: getRootFolders,
  })

  // Initialize local settings when data loads
  useEffect(() => {
    if (settings && Object.keys(localSettings).length === 0) {
      setLocalSettings(settings)
    }
  }, [settings])

  const updateSettingsMutation = useMutation({
    mutationFn: updateMediaSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mediaSettings'] })
      setHasChanges(false)
    },
  })

  const addFolderMutation = useMutation({
    mutationFn: addRootFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rootFolders'] })
      setIsAddFolderOpen(false)
      setNewFolderPath('')
      setNewFolderName('')
    },
  })

  const deleteFolderMutation = useMutation({
    mutationFn: deleteRootFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rootFolders'] })
    },
  })

  const handleSettingChange = <K extends keyof MediaSettings>(key: K, value: MediaSettings[K]) => {
    setLocalSettings(prev => ({ ...prev, [key]: value }))
    setHasChanges(true)
  }

  const handleSaveSettings = () => {
    updateSettingsMutation.mutate(localSettings)
  }

  const handleAddFolder = (e: React.FormEvent) => {
    e.preventDefault()
    addFolderMutation.mutate({
      path: newFolderPath,
      mediaType: newFolderType,
      name: newFolderName || undefined,
    })
  }

  const handleDeleteFolder = (id: number) => {
    if (confirm('Are you sure you want to remove this root folder?')) {
      deleteFolderMutation.mutate(id)
    }
  }

  const isLoading = settingsLoading || foldersLoading

  const ebookFolders = rootFolders?.filter(f => f.mediaType === 'ebook') || []
  const audiobookFolders = rootFolders?.filter(f => f.mediaType === 'audiobook') || []

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Media Management" subtitle="Configure paths, naming, and import settings" />

      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-4xl mx-auto">
          {/* Back link */}
          <Link
            to="/settings"
            className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Settings
          </Link>

          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="space-y-8">
              {/* Root Folders Section */}
              <section>
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h2 className="text-xl font-semibold flex items-center gap-2">
                      <FolderOpen className="h-5 w-5 text-primary" />
                      Root Folders
                    </h2>
                    <p className="text-sm text-muted-foreground mt-1">
                      Configure the directories where your media files are stored
                    </p>
                  </div>
                  <Button onClick={() => setIsAddFolderOpen(true)}>
                    <Plus className="h-4 w-4" />
                    Add Folder
                  </Button>
                </div>

                {/* Ebook Folders */}
                <div className="mb-4">
                  <h3 className="text-sm font-medium flex items-center gap-2 mb-2 text-muted-foreground">
                    <Book className="h-4 w-4" />
                    Ebook Folders
                  </h3>
                  {ebookFolders.length > 0 ? (
                    <div className="space-y-2">
                      {ebookFolders.map((folder) => (
                        <RootFolderCard 
                          key={folder.id} 
                          folder={folder} 
                          onDelete={handleDeleteFolder}
                          isDeleting={deleteFolderMutation.isPending}
                        />
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-6 border border-dashed border-border rounded-lg">
                      <p className="text-sm text-muted-foreground">No ebook folders configured</p>
                    </div>
                  )}
                </div>

                {/* Audiobook Folders */}
                <div>
                  <h3 className="text-sm font-medium flex items-center gap-2 mb-2 text-muted-foreground">
                    <Headphones className="h-4 w-4" />
                    Audiobook Folders
                  </h3>
                  {audiobookFolders.length > 0 ? (
                    <div className="space-y-2">
                      {audiobookFolders.map((folder) => (
                        <RootFolderCard 
                          key={folder.id} 
                          folder={folder} 
                          onDelete={handleDeleteFolder}
                          isDeleting={deleteFolderMutation.isPending}
                        />
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-6 border border-dashed border-border rounded-lg">
                      <p className="text-sm text-muted-foreground">No audiobook folders configured</p>
                    </div>
                  )}
                </div>
              </section>

              {/* File Naming Section */}
              <section>
                <h2 className="text-xl font-semibold flex items-center gap-2 mb-4">
                  <Info className="h-5 w-5 text-primary" />
                  File Naming
                </h2>
                <p className="text-sm text-muted-foreground mb-4">
                  Configure how files and folders are named when importing media
                </p>

                <div className="space-y-4 p-4 rounded-lg bg-card border border-border">
                  {/* Folder Naming */}
                  <div className="space-y-2">
                    <Label htmlFor="folderNaming">Folder Structure</Label>
                    <Input
                      id="folderNaming"
                      value={localSettings.folderNaming || '{Author}/{Series}'}
                      onChange={(e) => handleSettingChange('folderNaming', e.target.value)}
                      placeholder="{Author}/{Series}"
                    />
                  </div>

                  {/* Ebook Naming */}
                  <div className="space-y-2">
                    <Label htmlFor="fileNamingEbook">Ebook File Name</Label>
                    <Input
                      id="fileNamingEbook"
                      value={localSettings.fileNamingEbook || '{Author}/{Title}'}
                      onChange={(e) => handleSettingChange('fileNamingEbook', e.target.value)}
                      placeholder="{Author}/{Title}"
                    />
                  </div>

                  {/* Audiobook Naming */}
                  <div className="space-y-2">
                    <Label htmlFor="fileNamingAudiobook">Audiobook File Name</Label>
                    <Input
                      id="fileNamingAudiobook"
                      value={localSettings.fileNamingAudiobook || '{Author}/{Title}'}
                      onChange={(e) => handleSettingChange('fileNamingAudiobook', e.target.value)}
                      placeholder="{Author}/{Title}"
                    />
                  </div>

                  {/* Available Tokens */}
                  <div className="pt-2">
                    <p className="text-xs text-muted-foreground mb-2">Available tokens:</p>
                    <div className="flex flex-wrap gap-2">
                      {NAMING_TOKENS.map((t) => (
                        <code 
                          key={t.token}
                          className="px-2 py-1 rounded bg-muted text-xs cursor-help"
                          title={t.description}
                        >
                          {t.token}
                        </code>
                      ))}
                    </div>
                  </div>
                </div>
              </section>

              {/* Import Settings Section */}
              <section>
                <h2 className="text-xl font-semibold flex items-center gap-2 mb-4">
                  <RefreshCw className="h-5 w-5 text-primary" />
                  Importing
                </h2>

                <div className="space-y-4 p-4 rounded-lg bg-card border border-border">
                  <div className="flex items-center justify-between">
                    <div>
                      <Label>Rescan After Import</Label>
                      <p className="text-xs text-muted-foreground">Automatically scan library after importing new files</p>
                    </div>
                    <Switch
                      checked={localSettings.rescanAfterImport !== false}
                      onCheckedChange={(checked) => handleSettingChange('rescanAfterImport', checked)}
                    />
                  </div>
                </div>
              </section>

              {/* File Management Section */}
              <section>
                <h2 className="text-xl font-semibold flex items-center gap-2 mb-4">
                  <HardDrive className="h-5 w-5 text-primary" />
                  File Management
                </h2>

                <div className="space-y-4 p-4 rounded-lg bg-card border border-border">
                  {/* Use Hardlinks */}
                  <div className="flex items-center justify-between">
                    <div>
                      <Label className="flex items-center gap-2">
                        <Link2 className="h-4 w-4" />
                        Use Hardlinks
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Use hardlinks instead of copying files (saves disk space, requires same filesystem)
                      </p>
                    </div>
                    <Switch
                      checked={localSettings.useHardlinks || false}
                      onCheckedChange={(checked) => handleSettingChange('useHardlinks', checked)}
                    />
                  </div>

                  {/* Recycle Bin */}
                  <div className="flex items-center justify-between">
                    <div>
                      <Label className="flex items-center gap-2">
                        <Trash className="h-4 w-4" />
                        Recycle Bin
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Move deleted files to recycle bin instead of permanent deletion
                      </p>
                    </div>
                    <Switch
                      checked={localSettings.recycleBinEnabled || false}
                      onCheckedChange={(checked) => handleSettingChange('recycleBinEnabled', checked)}
                    />
                  </div>

                  {/* Recycle Bin Path */}
                  {localSettings.recycleBinEnabled && (
                    <div className="space-y-2 pt-2">
                      <Label htmlFor="recycleBinPath">Recycle Bin Path</Label>
                      <Input
                        id="recycleBinPath"
                        value={localSettings.recycleBinPath || ''}
                        onChange={(e) => handleSettingChange('recycleBinPath', e.target.value)}
                        placeholder="/path/to/recycle-bin"
                      />
                    </div>
                  )}
                </div>
              </section>

              {/* Save Button */}
              {hasChanges && (
                <div className="sticky bottom-6 flex justify-end">
                  <Button 
                    onClick={handleSaveSettings}
                    disabled={updateSettingsMutation.isPending}
                    size="lg"
                    className="shadow-lg"
                  >
                    {updateSettingsMutation.isPending && (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    )}
                    Save Changes
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Add Folder Dialog */}
      <Dialog open={isAddFolderOpen} onOpenChange={setIsAddFolderOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Add Root Folder</DialogTitle>
            <DialogDescription>
              Browse and select a folder for storing your media files
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleAddFolder} className="space-y-4">
            <div className="space-y-2">
              <Label>Folder Path</Label>
              <DirectoryBrowser
                value={newFolderPath}
                onChange={setNewFolderPath}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Media Type</Label>
                <Select
                  value={newFolderType}
                  onValueChange={(value: 'ebook' | 'audiobook') => setNewFolderType(value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ebook">Ebook</SelectItem>
                    <SelectItem value="audiobook">Audiobook</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="name">Display Name (optional)</Label>
                <Input
                  id="name"
                  value={newFolderName}
                  onChange={(e) => setNewFolderName(e.target.value)}
                  placeholder="Main Library"
                />
              </div>
            </div>

            {addFolderMutation.isError && (
              <p className="text-sm text-destructive">
                {(addFolderMutation.error as Error)?.message || 'Failed to add folder'}
              </p>
            )}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setIsAddFolderOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={addFolderMutation.isPending || !newFolderPath}>
                {addFolderMutation.isPending && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                Add Folder
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// Root Folder Card Component
function RootFolderCard({
  folder,
  onDelete,
  isDeleting,
}: {
  folder: RootFolder
  onDelete: (id: number) => void
  isDeleting: boolean
}) {
  const usedSpace = folder.totalSpace - folder.freeSpace
  const usagePercent = folder.totalSpace > 0 ? (usedSpace / folder.totalSpace) * 100 : 0

  return (
    <div className="flex items-center justify-between p-4 rounded-lg bg-card border border-border">
      <div className="flex items-center gap-4 flex-1 min-w-0">
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
          folder.accessible ? 'bg-primary/10' : 'bg-destructive/10'
        }`}>
          {folder.mediaType === 'audiobook' ? (
            <Headphones className={`h-5 w-5 ${folder.accessible ? 'text-primary' : 'text-destructive'}`} />
          ) : (
            <Book className={`h-5 w-5 ${folder.accessible ? 'text-primary' : 'text-destructive'}`} />
          )}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {folder.name && <span className="font-medium">{folder.name}</span>}
            <span className={`text-sm ${folder.name ? 'text-muted-foreground' : 'font-medium'} truncate`}>
              {folder.path}
            </span>
          </div>
          {folder.accessible ? (
            <div className="mt-1">
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{formatBytes(folder.freeSpace)} free of {formatBytes(folder.totalSpace)}</span>
              </div>
              <div className="mt-1 h-1.5 w-48 bg-muted rounded-full overflow-hidden">
                <div 
                  className={`h-full rounded-full transition-all ${
                    usagePercent > 90 ? 'bg-destructive' : usagePercent > 75 ? 'bg-yellow-500' : 'bg-primary'
                  }`}
                  style={{ width: `${usagePercent}%` }}
                />
              </div>
            </div>
          ) : (
            <p className="text-xs text-destructive mt-1">Folder not accessible</p>
          )}
        </div>
      </div>

      <Button
        variant="outline"
        size="sm"
        onClick={() => onDelete(folder.id)}
        disabled={isDeleting}
        className="text-destructive hover:text-destructive shrink-0"
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    </div>
  )
}
