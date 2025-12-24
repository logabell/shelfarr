import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Loader2, Check, X, ExternalLink, AlertTriangle, Eye, EyeOff, BookOpen, Database } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import api, { getSettings, updateSettings, getGoogleBooksQuota } from '@/api/client'

interface HardcoverSettings {
  enabled: boolean
  apiKey: string
  apiUrl: string
  rateLimit: number
  maxDepth: number
  maxTimeout: number
}

interface GoogleBooksSettings {
  enabled: boolean
  apiKey: string
  dailyQuota: number
}

interface Settings {
  librarySearchProviders?: {
    openLibrary?: { enabled: boolean }
    googleBooks?: GoogleBooksSettings
    hardcover?: HardcoverSettings
  }
}

export default function LibrarySearchSettingsPage() {
  const queryClient = useQueryClient()
  const [apiKey, setApiKey] = useState('')
  const [showApiKey, setShowApiKey] = useState(false)
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [testMessage, setTestMessage] = useState('')
  const [hasChanges, setHasChanges] = useState(false)

  const [googleApiKey, setGoogleApiKey] = useState('')
  const [showGoogleApiKey, setShowGoogleApiKey] = useState(false)
  const [googleTestStatus, setGoogleTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [googleTestMessage, setGoogleTestMessage] = useState('')
  const [googleHasChanges, setGoogleHasChanges] = useState(false)

  const { data: settings, isLoading } = useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: getSettings as () => Promise<Settings>,
  })

  const { data: googleQuota } = useQuery({
    queryKey: ['googleBooksQuota'],
    queryFn: getGoogleBooksQuota,
    refetchInterval: 60000,
  })

  const saveMutation = useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      setHasChanges(false)
    },
  })

  useEffect(() => {
    if (settings?.librarySearchProviders?.hardcover?.apiKey) {
      setApiKey(settings.librarySearchProviders.hardcover.apiKey)
    }
    if (settings?.librarySearchProviders?.googleBooks?.apiKey) {
      setGoogleApiKey(settings.librarySearchProviders.googleBooks.apiKey)
    }
  }, [settings])

  const handleApiKeyChange = (value: string) => {
    setApiKey(value)
    setHasChanges(true)
    setTestStatus('idle')
  }

  const handleGoogleApiKeyChange = (value: string) => {
    setGoogleApiKey(value)
    setGoogleHasChanges(true)
    setGoogleTestStatus('idle')
  }

  const handleSave = async () => {
    await saveMutation.mutateAsync({
      librarySearchProviders: {
        hardcover: {
          apiKey: apiKey,
        },
      },
    })
  }

  const handleGoogleSave = async () => {
    await saveMutation.mutateAsync({
      librarySearchProviders: {
        googleBooks: {
          apiKey: googleApiKey,
        },
      },
    })
    setGoogleHasChanges(false)
    queryClient.invalidateQueries({ queryKey: ['googleBooksQuota'] })
  }

  const handleGoogleTest = async () => {
    if (!googleApiKey) {
      setGoogleTestStatus('error')
      setGoogleTestMessage('Please enter an API key first')
      return
    }

    setGoogleTestStatus('testing')
    setGoogleTestMessage('')

    try {
      await saveMutation.mutateAsync({
        librarySearchProviders: {
          googleBooks: {
            apiKey: googleApiKey,
          },
        },
      })

      const { data } = await api.post('/search/googlebooks/test')
      
      setGoogleTestStatus('success')
      setGoogleTestMessage(`Connection successful! Quota: ${data.quotaRemaining}/${1000} remaining`)
      setGoogleHasChanges(false)
      queryClient.invalidateQueries({ queryKey: ['googleBooksQuota'] })
    } catch (error: unknown) {
      setGoogleTestStatus('error')
      const axiosError = error as { response?: { data?: { error?: string } }, message?: string }
      const errorMessage = axiosError.response?.data?.error || axiosError.message || 'Connection failed'
      setGoogleTestMessage(errorMessage)
    }
  }

  const handleTest = async () => {
    if (!apiKey) {
      setTestStatus('error')
      setTestMessage('Please enter an API key first')
      return
    }

    setTestStatus('testing')
    setTestMessage('')

    try {
      // Save the key first
      await saveMutation.mutateAsync({
        librarySearchProviders: {
          hardcover: {
            apiKey: apiKey,
          },
        },
      })

      // Test the API connection using the dedicated test endpoint
      // This runs "query { me { username } }" to validate the API key
      const { data } = await api.post('/search/hardcover/test')
      
      setTestStatus('success')
      setTestMessage(data.message || 'Connection successful! API key is valid.')
      setHasChanges(false)
    } catch (error: unknown) {
      setTestStatus('error')
      // Handle axios error response
      const axiosError = error as { response?: { data?: { error?: string } }, message?: string }
      const errorMessage = axiosError.response?.data?.error || axiosError.message || 'Connection failed'
      setTestMessage(errorMessage)
    }
  }

  const hardcoverSettings = settings?.librarySearchProviders?.hardcover

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Library Search Providers" subtitle="Configure book metadata sources" />

      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-3xl mx-auto">
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
              {/* Hardcover.app Section */}
              <div className="rounded-lg border border-border bg-card">
                <div className="p-6 border-b border-border">
                  <div className="flex items-start justify-between">
                    <div>
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        Hardcover.app
                        {testStatus === 'success' && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-status-downloaded/20 text-status-downloaded">
                            Connected
                          </span>
                        )}
                        {hardcoverSettings?.apiKey && testStatus !== 'success' && testStatus !== 'error' && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
                            Not Tested
                          </span>
                        )}
                      </h2>
                      <p className="text-sm text-muted-foreground mt-1">
                        Book metadata and search provider
                      </p>
                    </div>
                    <a
                      href="https://hardcover.app/account/api"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      Get API Key
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  </div>
                </div>

                <div className="p-6 space-y-6">
                  {/* API Key Input */}
                  <div className="space-y-2">
                    <Label htmlFor="apiKey">API Key</Label>
                    <div className="flex gap-2">
                      <div className="relative flex-1">
                        <Input
                          id="apiKey"
                          type={showApiKey ? 'text' : 'password'}
                          value={apiKey}
                          onChange={(e) => handleApiKeyChange(e.target.value)}
                          placeholder="Enter your Hardcover.app API key (starts with 'bearer')"
                          className="pr-10"
                        />
                        <button
                          type="button"
                          onClick={() => setShowApiKey(!showApiKey)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        >
                          {showApiKey ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                      <Button
                        variant="outline"
                        onClick={handleTest}
                        disabled={testStatus === 'testing' || !apiKey}
                      >
                        {testStatus === 'testing' ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : testStatus === 'success' ? (
                          <Check className="h-4 w-4 text-status-downloaded" />
                        ) : testStatus === 'error' ? (
                          <X className="h-4 w-4 text-destructive" />
                        ) : null}
                        Test
                      </Button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Get your API key from{' '}
                      <a
                        href="https://hardcover.app/account/api"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline"
                      >
                        hardcover.app/account/api
                      </a>
                    </p>
                    {testMessage && (
                      <p
                        className={`text-sm ${
                          testStatus === 'success' ? 'text-status-downloaded' : 'text-destructive'
                        }`}
                      >
                        {testMessage}
                      </p>
                    )}
                  </div>

                  {/* Save Button */}
                  <div className="flex justify-end">
                    <Button
                      onClick={handleSave}
                      disabled={saveMutation.isPending || !hasChanges}
                    >
                      {saveMutation.isPending && (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )}
                      Save Changes
                    </Button>
                  </div>
                </div>
              </div>

              {/* Open Library Section */}
              <div className="rounded-lg border border-border bg-card">
                <div className="p-6 border-b border-border">
                  <div className="flex items-start justify-between">
                    <div>
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <Database className="h-5 w-5 text-primary" />
                        Open Library
                        <span className="px-2 py-0.5 text-xs rounded-full bg-status-downloaded/20 text-status-downloaded">
                          Always Enabled
                        </span>
                      </h2>
                      <p className="text-sm text-muted-foreground mt-1">
                        Free, open-source book database with no API key required
                      </p>
                    </div>
                    <a
                      href="https://openlibrary.org/developers/api"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      API Docs
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  </div>
                </div>

                <div className="p-6">
                  <div className="space-y-4 text-sm text-muted-foreground">
                    <p>
                      Open Library is a free, open-source book catalog that provides:
                    </p>
                    <ul className="space-y-2 ml-4">
                      <li className="flex items-start gap-2">
                        <span className="text-foreground">•</span>
                        <span>Book metadata (titles, authors, ISBNs, covers)</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-foreground">•</span>
                        <span>Author information and bibliographies</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-foreground">•</span>
                        <span>Series and subject classification</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-foreground">•</span>
                        <span>High-resolution book covers</span>
                      </li>
                    </ul>
                    <p className="text-xs italic">
                      No configuration needed. Open Library is used automatically as a supplementary metadata source.
                    </p>
                  </div>
                </div>
              </div>

              {/* Google Books Section */}
              <div className="rounded-lg border border-border bg-card">
                <div className="p-6 border-b border-border">
                  <div className="flex items-start justify-between">
                    <div>
                      <h2 className="text-lg font-semibold flex items-center gap-2">
                        <BookOpen className="h-5 w-5 text-primary" />
                        Google Books
                        {googleTestStatus === 'success' && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-status-downloaded/20 text-status-downloaded">
                            Connected
                          </span>
                        )}
                        {googleApiKey && googleTestStatus !== 'success' && googleTestStatus !== 'error' && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
                            Not Tested
                          </span>
                        )}
                        {!googleApiKey && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-amber-500/20 text-amber-500">
                            Optional
                          </span>
                        )}
                      </h2>
                      <p className="text-sm text-muted-foreground mt-1">
                        Ebook availability detection (EPUB/PDF status)
                      </p>
                    </div>
                    <a
                      href="https://console.cloud.google.com/apis/library/books.googleapis.com"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      Get API Key
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  </div>
                </div>

                <div className="p-6 space-y-6">
                  {googleQuota && (
                    <div className="rounded-lg bg-secondary/50 p-4">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">Daily Quota Usage</span>
                        <span className="text-sm text-muted-foreground">
                          {googleQuota.quotaUsed} / {googleQuota.quotaLimit} requests
                        </span>
                      </div>
                      <div className="h-2 bg-secondary rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all ${
                            googleQuota.quotaUsed / googleQuota.quotaLimit > 0.9
                              ? 'bg-destructive'
                              : googleQuota.quotaUsed / googleQuota.quotaLimit > 0.7
                              ? 'bg-amber-500'
                              : 'bg-status-downloaded'
                          }`}
                          style={{
                            width: `${Math.min(100, (googleQuota.quotaUsed / googleQuota.quotaLimit) * 100)}%`,
                          }}
                        />
                      </div>
                      <p className="text-xs text-muted-foreground mt-2">
                        {googleQuota.remaining} requests remaining today. Resets at midnight UTC.
                      </p>
                    </div>
                  )}

                  {/* API Key Input */}
                  <div className="space-y-2">
                    <Label htmlFor="googleApiKey">API Key</Label>
                    <div className="flex gap-2">
                      <div className="relative flex-1">
                        <Input
                          id="googleApiKey"
                          type={showGoogleApiKey ? 'text' : 'password'}
                          value={googleApiKey}
                          onChange={(e) => handleGoogleApiKeyChange(e.target.value)}
                          placeholder="Enter your Google Books API key"
                          className="pr-10"
                        />
                        <button
                          type="button"
                          onClick={() => setShowGoogleApiKey(!showGoogleApiKey)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        >
                          {showGoogleApiKey ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                      <Button
                        variant="outline"
                        onClick={handleGoogleTest}
                        disabled={googleTestStatus === 'testing' || !googleApiKey}
                      >
                        {googleTestStatus === 'testing' ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : googleTestStatus === 'success' ? (
                          <Check className="h-4 w-4 text-status-downloaded" />
                        ) : googleTestStatus === 'error' ? (
                          <X className="h-4 w-4 text-destructive" />
                        ) : null}
                        Test
                      </Button>
                    </div>
                    {googleTestMessage && (
                      <p
                        className={`text-sm ${
                          googleTestStatus === 'success' ? 'text-status-downloaded' : 'text-destructive'
                        }`}
                      >
                        {googleTestMessage}
                      </p>
                    )}
                  </div>

                  {/* What it provides */}
                  <div className="text-sm text-muted-foreground space-y-2">
                    <p className="font-medium text-foreground">What Google Books provides:</p>
                    <ul className="space-y-1 ml-4">
                      <li>• Detection of ebook availability (EPUB, PDF formats)</li>
                      <li>• Preview availability status</li>
                      <li>• Supplementary book descriptions</li>
                    </ul>
                  </div>

                  {/* Save Button */}
                  <div className="flex justify-end">
                    <Button
                      onClick={handleGoogleSave}
                      disabled={saveMutation.isPending || !googleHasChanges}
                    >
                      {saveMutation.isPending && (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )}
                      Save Changes
                    </Button>
                  </div>
                </div>
              </div>

              {/* Google Books Setup Instructions */}
              <div className="rounded-lg border border-border bg-card p-6">
                <h3 className="font-medium">How to Get a Google Books API Key</h3>
                <ol className="mt-4 space-y-3 text-sm text-muted-foreground list-decimal list-inside">
                  <li>
                    Go to the{' '}
                    <a
                      href="https://console.cloud.google.com/"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      Google Cloud Console
                    </a>
                  </li>
                  <li>Create a new project or select an existing one</li>
                  <li>
                    Enable the{' '}
                    <a
                      href="https://console.cloud.google.com/apis/library/books.googleapis.com"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      Books API
                    </a>
                  </li>
                  <li>Go to "Credentials" and create an API key</li>
                  <li>Optionally restrict the key to only the Books API</li>
                  <li>Copy and paste the key above</li>
                </ol>
                <p className="mt-4 text-xs text-muted-foreground">
                  Free tier: 1,000 requests/day. No billing required for basic usage.
                </p>
              </div>

              {/* Hardcover API Limitations Info */}
              <div className="rounded-lg border border-border bg-secondary/30 p-6">
                <h3 className="font-medium flex items-center gap-2 text-amber-500">
                  <AlertTriangle className="h-4 w-4" />
                  Hardcover API Limitations
                </h3>
                <ul className="mt-4 space-y-2 text-sm text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-foreground">•</span>
                    <span>API tokens automatically expire after 1 year, and reset on January 1st</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-foreground">•</span>
                    <span>Rate-limited to 60 requests per minute (handled automatically)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-foreground">•</span>
                    <span>Queries have a max timeout of 30 seconds</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-foreground">•</span>
                    <span>Queries are limited to a maximum depth of 3</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-foreground">•</span>
                    <span>Access is restricted to public data and your own user data</span>
                  </li>
                </ul>
              </div>

              {/* How to Get API Key */}
              <div className="rounded-lg border border-border bg-card p-6">
                <h3 className="font-medium">How to Get Your API Key</h3>
                <ol className="mt-4 space-y-3 text-sm text-muted-foreground list-decimal list-inside">
                  <li>
                    Create an account at{' '}
                    <a
                      href="https://hardcover.app"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      hardcover.app
                    </a>
                  </li>
                  <li>
                    Go to{' '}
                    <a
                      href="https://hardcover.app/account/api"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      Account &gt; API Settings
                    </a>
                  </li>
                  <li>Copy your API key (it starts with "bearer")</li>
                  <li>Paste the key in the field above and save</li>
                </ol>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

