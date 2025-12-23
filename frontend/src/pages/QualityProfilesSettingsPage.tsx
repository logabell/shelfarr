import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, GripVertical, Loader2, ArrowLeft, Book, Headphones } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Slider } from '@/components/ui/slider'
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
import { getProfiles, createProfile, updateProfile, deleteProfile } from '@/api/client'
import type { QualityProfile, MediaType } from '@/types'

// Available formats by media type
const EBOOK_FORMATS = ['epub', 'azw3', 'mobi', 'pdf', 'cbz', 'cbr']
const AUDIOBOOK_FORMATS = ['m4b', 'mp3', 'flac', 'm4a']

interface ProfileFormData {
  name: string
  mediaType: MediaType
  formatRanking: string[]
  minBitrate: number
}

const defaultFormData: ProfileFormData = {
  name: '',
  mediaType: 'ebook',
  formatRanking: [...EBOOK_FORMATS],
  minBitrate: 64,
}

export function QualityProfilesSettingsPage() {
  const queryClient = useQueryClient()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingProfile, setEditingProfile] = useState<QualityProfile | null>(null)
  const [formData, setFormData] = useState<ProfileFormData>(defaultFormData)
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const { data: profiles, isLoading } = useQuery({
    queryKey: ['profiles'],
    queryFn: getProfiles,
  })

  const createMutation = useMutation({
    mutationFn: createProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      closeDialog()
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Omit<QualityProfile, 'id'>> }) => updateProfile(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      closeDialog()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
    },
  })

  const openAddDialog = () => {
    setEditingProfile(null)
    setFormData(defaultFormData)
    setIsDialogOpen(true)
  }

  const openEditDialog = (profile: QualityProfile) => {
    setEditingProfile(profile)
    const formats = profile.formatRanking ? profile.formatRanking.split(',').map(f => f.trim().toLowerCase()) : []
    const availableFormats = profile.mediaType === 'audiobook' ? AUDIOBOOK_FORMATS : EBOOK_FORMATS
    // Add any missing formats to the end
    const missingFormats = availableFormats.filter(f => !formats.includes(f))
    setFormData({
      name: profile.name,
      mediaType: profile.mediaType,
      formatRanking: [...formats, ...missingFormats],
      minBitrate: profile.minBitrate || 64,
    })
    setIsDialogOpen(true)
  }

  const closeDialog = () => {
    setIsDialogOpen(false)
    setEditingProfile(null)
    setFormData(defaultFormData)
  }

  const handleMediaTypeChange = (type: MediaType) => {
    const formats = type === 'audiobook' ? [...AUDIOBOOK_FORMATS] : [...EBOOK_FORMATS]
    setFormData({ ...formData, mediaType: type, formatRanking: formats })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newRanking = [...formData.formatRanking]
    const draggedItem = newRanking[draggedIndex]
    newRanking.splice(draggedIndex, 1)
    newRanking.splice(index, 0, draggedItem)
    
    setFormData({ ...formData, formatRanking: newRanking })
    setDraggedIndex(index)
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    const data = {
      name: formData.name,
      mediaType: formData.mediaType,
      formatRanking: formData.formatRanking.join(','),
      minBitrate: formData.mediaType === 'audiobook' ? formData.minBitrate : 0,
    }

    if (editingProfile) {
      updateMutation.mutate({ id: editingProfile.id, data })
    } else {
      createMutation.mutate(data)
    }
  }

  const handleDelete = (id: number) => {
    if (confirm('Are you sure you want to delete this profile?')) {
      deleteMutation.mutate(id)
    }
  }

  const ebookProfiles = profiles?.filter(p => p.mediaType === 'ebook') || []
  const audiobookProfiles = profiles?.filter(p => p.mediaType === 'audiobook') || []

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Quality Profiles" subtitle="Configure format preferences" />

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

          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold">Quality Profiles</h2>
              <p className="text-sm text-muted-foreground mt-1">
                Define format preferences for automatic downloads. Formats are prioritized from top to bottom.
              </p>
            </div>
            <Button onClick={openAddDialog}>
              <Plus className="h-4 w-4" />
              Add Profile
            </Button>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="space-y-8">
              {/* Ebook Profiles */}
              <section>
                <h3 className="text-lg font-medium flex items-center gap-2 mb-4">
                  <Book className="h-5 w-5 text-primary" />
                  Ebook Profiles
                </h3>
                {ebookProfiles.length > 0 ? (
                  <div className="space-y-3">
                    {ebookProfiles.map((profile) => (
                      <ProfileCard 
                        key={profile.id} 
                        profile={profile} 
                        onEdit={openEditDialog}
                        onDelete={handleDelete}
                        isDeleting={deleteMutation.isPending}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 border border-dashed border-border rounded-lg">
                    <p className="text-muted-foreground">No ebook profiles configured</p>
                  </div>
                )}
              </section>

              {/* Audiobook Profiles */}
              <section>
                <h3 className="text-lg font-medium flex items-center gap-2 mb-4">
                  <Headphones className="h-5 w-5 text-primary" />
                  Audiobook Profiles
                </h3>
                {audiobookProfiles.length > 0 ? (
                  <div className="space-y-3">
                    {audiobookProfiles.map((profile) => (
                      <ProfileCard 
                        key={profile.id} 
                        profile={profile} 
                        onEdit={openEditDialog}
                        onDelete={handleDelete}
                        isDeleting={deleteMutation.isPending}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 border border-dashed border-border rounded-lg">
                    <p className="text-muted-foreground">No audiobook profiles configured</p>
                  </div>
                )}
              </section>
            </div>
          )}
        </div>
      </div>

      {/* Add/Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editingProfile ? 'Edit Profile' : 'Add Profile'}</DialogTitle>
            <DialogDescription>
              {editingProfile
                ? 'Update the quality profile settings'
                : 'Create a new quality profile to define format preferences'}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Profile Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., High Quality Ebooks"
                required
              />
            </div>

            {/* Media Type */}
            <div className="space-y-2">
              <Label>Media Type</Label>
              <Select
                value={formData.mediaType}
                onValueChange={(value: MediaType) => handleMediaTypeChange(value)}
                disabled={!!editingProfile}
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

            {/* Format Priority */}
            <div className="space-y-2">
              <Label>Format Priority (drag to reorder)</Label>
              <p className="text-xs text-muted-foreground">
                Formats at the top are preferred. Downloads will try to match the highest priority format first.
              </p>
              <div className="mt-2 space-y-1 rounded-lg border border-border p-2 bg-secondary/30">
                {formData.formatRanking.map((format, index) => (
                  <div
                    key={format}
                    draggable
                    onDragStart={() => handleDragStart(index)}
                    onDragOver={(e) => handleDragOver(e, index)}
                    onDragEnd={handleDragEnd}
                    className={`flex items-center gap-2 p-2 rounded-md bg-background border border-border cursor-move transition-colors ${
                      draggedIndex === index ? 'opacity-50 bg-primary/10 border-primary' : 'hover:bg-muted'
                    }`}
                  >
                    <GripVertical className="h-4 w-4 text-muted-foreground" />
                    <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary/10 text-primary text-xs font-medium">
                      {index + 1}
                    </span>
                    <span className="font-mono text-sm uppercase">{format}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Min Bitrate (for audiobooks) */}
            {formData.mediaType === 'audiobook' && (
              <div className="space-y-2">
                <Label>Minimum Bitrate: {formData.minBitrate} kbps</Label>
                <p className="text-xs text-muted-foreground">
                  Reject audiobooks below this bitrate
                </p>
                <Slider
                  value={[formData.minBitrate]}
                  onValueChange={([value]) => setFormData({ ...formData, minBitrate: value })}
                  min={32}
                  max={320}
                  step={16}
                />
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>32 kbps</span>
                  <span>320 kbps</span>
                </div>
              </div>
            )}

            <DialogFooter className="mt-6">
              <Button type="button" variant="outline" onClick={closeDialog}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {(createMutation.isPending || updateMutation.isPending) && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                {editingProfile ? 'Save Changes' : 'Create Profile'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// Profile Card Component
function ProfileCard({
  profile,
  onEdit,
  onDelete,
  isDeleting,
}: {
  profile: QualityProfile
  onEdit: (profile: QualityProfile) => void
  onDelete: (id: number) => void
  isDeleting: boolean
}) {
  const formats = profile.formatRanking ? profile.formatRanking.split(',').map(f => f.trim().toUpperCase()) : []

  return (
    <div className="flex items-center justify-between p-4 rounded-lg bg-card border border-border">
      <div className="flex items-center gap-4">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          {profile.mediaType === 'audiobook' ? (
            <Headphones className="h-5 w-5 text-primary" />
          ) : (
            <Book className="h-5 w-5 text-primary" />
          )}
        </div>
        <div>
          <div className="font-medium">{profile.name}</div>
          <div className="flex items-center gap-2 mt-1">
            <div className="flex items-center gap-1">
              {formats.slice(0, 4).map((format, i) => (
                <span
                  key={format}
                  className={`px-1.5 py-0.5 text-xs rounded font-mono ${
                    i === 0 
                      ? 'bg-primary/20 text-primary' 
                      : 'bg-secondary text-secondary-foreground'
                  }`}
                >
                  {format}
                </span>
              ))}
              {formats.length > 4 && (
                <span className="text-xs text-muted-foreground">+{formats.length - 4} more</span>
              )}
            </div>
            {profile.mediaType === 'audiobook' && profile.minBitrate && (
              <span className="text-xs text-muted-foreground">
                • Min {profile.minBitrate} kbps
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={() => onEdit(profile)}>
          <Pencil className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onDelete(profile.id)}
          disabled={isDeleting}
          className="text-destructive hover:text-destructive"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
