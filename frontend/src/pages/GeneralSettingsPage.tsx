import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Loader2, Globe, Settings, Calendar, Home, Check, X, RefreshCw, Database } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { getGeneralSettings, updateGeneralSettings, getAvailableLanguages, refreshAllMetadata } from '@/api/client'

interface GeneralSettings {
  instanceName: string
  defaultLanguage: string
  preferredLanguages: string[]
  startPage: string
  dateFormat: string
}

interface LanguageOption {
  code: string
  name: string
}

const START_PAGE_OPTIONS = [
  { value: 'library', label: 'Library' },
  { value: 'series', label: 'Series' },
  { value: 'authors', label: 'Authors' },
  { value: 'activity', label: 'Activity' },
  { value: 'wanted', label: 'Wanted' },
]

const DATE_FORMAT_OPTIONS = [
  { value: 'MMMM d, yyyy', label: 'January 1, 2024' },
  { value: 'MMM d, yyyy', label: 'Jan 1, 2024' },
  { value: 'MM/dd/yyyy', label: '01/01/2024' },
  { value: 'dd/MM/yyyy', label: '01/01/2024 (DD/MM)' },
  { value: 'yyyy-MM-dd', label: '2024-01-01 (ISO)' },
]

export function GeneralSettingsPage() {
  const queryClient = useQueryClient()
  const [localSettings, setLocalSettings] = useState<Partial<GeneralSettings>>({})
  const [hasChanges, setHasChanges] = useState(false)

  const { data: settings, isLoading: settingsLoading } = useQuery({
    queryKey: ['generalSettings'],
    queryFn: getGeneralSettings,
  })

  const { data: languages, isLoading: languagesLoading } = useQuery({
    queryKey: ['availableLanguages'],
    queryFn: getAvailableLanguages,
  })

  // Initialize local settings when data loads
  useEffect(() => {
    if (settings && Object.keys(localSettings).length === 0) {
      setLocalSettings(settings)
    }
  }, [settings])

  const updateSettingsMutation = useMutation({
    mutationFn: updateGeneralSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['generalSettings'] })
      setHasChanges(false)
    },
  })

  const [refreshResult, setRefreshResult] = useState<{ refreshed: number; failed: number; errors?: string[] } | null>(null)

  const refreshMetadataMutation = useMutation({
    mutationFn: refreshAllMetadata,
    onSuccess: (result) => {
      setRefreshResult({ refreshed: result.refreshed, failed: result.failed, errors: result.errors })
      queryClient.invalidateQueries({ queryKey: ['library'] })
      queryClient.invalidateQueries({ queryKey: ['books'] })
    },
  })

  const handleChange = <K extends keyof GeneralSettings>(key: K, value: GeneralSettings[K]) => {
    setLocalSettings(prev => ({ ...prev, [key]: value }))
    setHasChanges(true)
  }

  const handleLanguageToggle = (langCode: string) => {
    const current = localSettings.preferredLanguages || []
    let updated: string[]
    
    if (current.includes(langCode)) {
      // Don't allow removing the last language
      if (current.length <= 1) return
      updated = current.filter(l => l !== langCode)
    } else {
      updated = [...current, langCode]
    }
    
    handleChange('preferredLanguages', updated)
  }

  const handleSave = () => {
    updateSettingsMutation.mutate(localSettings)
  }

  const handleCancel = () => {
    if (settings) {
      setLocalSettings(settings)
      setHasChanges(false)
    }
  }

  if (settingsLoading || languagesLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Topbar title="General Settings" />

      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-3xl mx-auto space-y-8">
          {/* Back Link */}
          <Link
            to="/settings"
            className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Settings
          </Link>

          {/* Instance Settings */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Settings className="h-5 w-5 text-primary" />
              <h2>Instance</h2>
            </div>

            <div className="bg-card border rounded-lg p-6 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="instanceName">Instance Name</Label>
                <Input
                  id="instanceName"
                  value={localSettings.instanceName || ''}
                  onChange={(e) => handleChange('instanceName', e.target.value)}
                  placeholder="Bookarr"
                />
                <p className="text-xs text-muted-foreground">
                  The name displayed in the browser tab and UI
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="startPage">Start Page</Label>
                <Select
                  value={localSettings.startPage || 'library'}
                  onValueChange={(value) => handleChange('startPage', value)}
                >
                  <SelectTrigger id="startPage">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {START_PAGE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        <div className="flex items-center gap-2">
                          <Home className="h-4 w-4" />
                          {option.label}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  The page to display when opening the app
                </p>
              </div>
            </div>
          </section>

          {/* Language Settings */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Globe className="h-5 w-5 text-primary" />
              <h2>Language Preferences</h2>
            </div>

            <div className="bg-card border rounded-lg p-6 space-y-6">
              <div className="space-y-2">
                <Label>Default Language</Label>
                <Select
                  value={localSettings.defaultLanguage || 'en'}
                  onValueChange={(value) => handleChange('defaultLanguage', value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {languages?.map((lang: LanguageOption) => (
                      <SelectItem key={lang.code} value={lang.code}>
                        {lang.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Primary language for browsing books and series
                </p>
              </div>

              <div className="space-y-3">
                <Label>Preferred Languages</Label>
                <p className="text-sm text-muted-foreground">
                  Select all languages you want to see when browsing books, authors, and series. 
                  Books in other languages will be filtered out.
                </p>
                
                <div className="flex flex-wrap gap-2 pt-2">
                  {languages?.map((lang: LanguageOption) => {
                    const isSelected = (localSettings.preferredLanguages || []).includes(lang.code)
                    return (
                      <Badge
                        key={lang.code}
                        variant={isSelected ? 'default' : 'outline'}
                        className="cursor-pointer select-none transition-colors"
                        onClick={() => handleLanguageToggle(lang.code)}
                      >
                        {isSelected && <Check className="h-3 w-3 mr-1" />}
                        {lang.name}
                      </Badge>
                    )
                  })}
                </div>
              </div>
            </div>
          </section>

          {/* Date Format Settings */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Calendar className="h-5 w-5 text-primary" />
              <h2>Display</h2>
            </div>

            <div className="bg-card border rounded-lg p-6 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="dateFormat">Date Format</Label>
                <Select
                  value={localSettings.dateFormat || 'MMMM d, yyyy'}
                  onValueChange={(value) => handleChange('dateFormat', value)}
                >
                  <SelectTrigger id="dateFormat">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DATE_FORMAT_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  How dates are displayed throughout the application
                </p>
              </div>
            </div>
          </section>

          {/* Metadata Maintenance */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Database className="h-5 w-5 text-primary" />
              <h2>Metadata</h2>
            </div>

            <div className="bg-card border rounded-lg p-6 space-y-4">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Refresh All Metadata</Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      Re-fetch metadata for all books from Hardcover.app including editions, contributors, and genres.
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setRefreshResult(null)
                      refreshMetadataMutation.mutate(undefined)
                    }}
                    disabled={refreshMetadataMutation.isPending}
                  >
                    {refreshMetadataMutation.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                      <RefreshCw className="h-4 w-4 mr-2" />
                    )}
                    {refreshMetadataMutation.isPending ? 'Refreshing...' : 'Refresh Metadata'}
                  </Button>
                </div>
                
                {refreshResult && (
                  <div className="mt-3 p-3 rounded-lg bg-muted text-sm space-y-2">
                    <div>
                      <span className="text-green-500 font-medium">{refreshResult.refreshed}</span> books refreshed
                      {refreshResult.failed > 0 && (
                        <span className="ml-2 text-destructive">({refreshResult.failed} failed)</span>
                      )}
                    </div>
                    {refreshResult.errors && refreshResult.errors.length > 0 && (
                      <div className="text-xs text-destructive space-y-1 max-h-32 overflow-y-auto">
                        {refreshResult.errors.slice(0, 5).map((err, i) => (
                          <div key={i}>{err}</div>
                        ))}
                        {refreshResult.errors.length > 5 && (
                          <div className="text-muted-foreground">...and {refreshResult.errors.length - 5} more</div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </section>

          {/* Save/Cancel Buttons */}
          {hasChanges && (
            <div className="flex items-center justify-end gap-4 py-4 border-t">
              <Button
                variant="outline"
                onClick={handleCancel}
                disabled={updateSettingsMutation.isPending}
              >
                <X className="h-4 w-4 mr-2" />
                Cancel
              </Button>
              <Button
                onClick={handleSave}
                disabled={updateSettingsMutation.isPending}
              >
                {updateSettingsMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <Check className="h-4 w-4 mr-2" />
                )}
                Save Changes
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default GeneralSettingsPage
