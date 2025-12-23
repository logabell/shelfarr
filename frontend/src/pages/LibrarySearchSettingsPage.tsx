import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Loader2, Check, X, ExternalLink, AlertTriangle, Eye, EyeOff } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import api, { getSettings, updateSettings } from '@/api/client'

interface HardcoverSettings {
  enabled: boolean
  apiKey: string
  apiUrl: string
  rateLimit: number
  maxDepth: number
  maxTimeout: number
}

interface Settings {
  librarySearchProviders?: {
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

  const { data: settings, isLoading } = useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: getSettings as () => Promise<Settings>,
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
  }, [settings])

  const handleApiKeyChange = (value: string) => {
    setApiKey(value)
    setHasChanges(true)
    setTestStatus('idle')
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

              {/* API Limitations Info */}
              <div className="rounded-lg border border-border bg-secondary/30 p-6">
                <h3 className="font-medium flex items-center gap-2 text-amber-500">
                  <AlertTriangle className="h-4 w-4" />
                  API Limitations
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

