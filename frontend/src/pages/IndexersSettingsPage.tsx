import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, Zap, Check, X, Loader2, ArrowLeft, HelpCircle, ExternalLink, AlertTriangle } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
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
  Collapsible,
  CollapsibleContent,
} from '@/components/ui/collapsible'
import { getIndexers, addIndexer, updateIndexer, deleteIndexer, testIndexer } from '@/api/client'
import type { Indexer } from '@/types'

type IndexerType = 'torznab' | 'mam' | 'anna'

interface IndexerFormData {
  name: string
  type: IndexerType
  url: string
  apiKey: string
  cookie: string
  priority: number
  enabled: boolean
  vipOnly: boolean
  freeleechOnly: boolean
}

const defaultFormData: IndexerFormData = {
  name: '',
  type: 'mam',
  url: '',
  apiKey: '',
  cookie: '',
  priority: 0,
  enabled: true,
  vipOnly: false,
  freeleechOnly: false,
}

const indexerTypeInfo: Record<IndexerType, { name: string; description: string; fields: string[] }> = {
  mam: {
    name: 'MyAnonamouse',
    description: 'Native MAM integration with cookie authentication',
    fields: ['cookie', 'vipOnly', 'freeleechOnly'],
  },
  torznab: {
    name: 'Torznab',
    description: 'Generic Torznab API (Prowlarr, Jackett)',
    fields: ['url', 'apiKey'],
  },
  anna: {
    name: "Anna's Archive",
    description: 'Web scraper for direct downloads',
    fields: [],
  },
}

export function IndexersSettingsPage() {
  const queryClient = useQueryClient()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingIndexer, setEditingIndexer] = useState<Indexer | null>(null)
  const [formData, setFormData] = useState<IndexerFormData>(defaultFormData)
  const [testResults, setTestResults] = useState<Record<number, { success: boolean; message: string } | null>>({})
  const [showMamHelp, setShowMamHelp] = useState(false)

  const { data: indexers, isLoading } = useQuery({
    queryKey: ['indexers'],
    queryFn: getIndexers,
  })

  const addMutation = useMutation({
    mutationFn: addIndexer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['indexers'] })
      closeDialog()
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Indexer> }) => updateIndexer(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['indexers'] })
      closeDialog()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteIndexer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['indexers'] })
    },
  })

  const testMutation = useMutation({
    mutationFn: testIndexer,
    onSuccess: (result, id) => {
      setTestResults(prev => ({ ...prev, [id]: result }))
    },
    onError: (error, id) => {
      setTestResults(prev => ({ ...prev, [id]: { success: false, message: String(error) } }))
    },
  })

  const openAddDialog = () => {
    setEditingIndexer(null)
    setFormData(defaultFormData)
    setIsDialogOpen(true)
  }

  const openEditDialog = (indexer: Indexer) => {
    setEditingIndexer(indexer)
    setFormData({
      name: indexer.name,
      type: indexer.type,
      url: indexer.url,
      apiKey: '',
      cookie: '',
      priority: indexer.priority,
      enabled: indexer.enabled,
      vipOnly: indexer.vipOnly || false,
      freeleechOnly: indexer.freeleechOnly || false,
    })
    setIsDialogOpen(true)
  }

  const closeDialog = () => {
    setIsDialogOpen(false)
    setEditingIndexer(null)
    setFormData(defaultFormData)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    const data = {
      name: formData.name,
      type: formData.type,
      url: formData.type === 'mam' ? 'https://www.myanonamouse.net' : formData.url,
      apiKey: formData.apiKey,
      cookie: formData.cookie,
      priority: formData.priority,
      enabled: formData.enabled,
      vipOnly: formData.vipOnly,
      freeleechOnly: formData.freeleechOnly,
    }

    if (editingIndexer) {
      updateMutation.mutate({ id: editingIndexer.id, data })
    } else {
      addMutation.mutate(data)
    }
  }

  const handleTest = (id: number) => {
    setTestResults(prev => ({ ...prev, [id]: null }))
    testMutation.mutate(id)
  }

  const handleDelete = (id: number) => {
    if (confirm('Are you sure you want to delete this indexer?')) {
      deleteMutation.mutate(id)
    }
  }

  const typeFields = indexerTypeInfo[formData.type].fields

  return (
    <div className="flex flex-col h-full">
      <Topbar title="Indexers" subtitle="Configure search providers" />

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
              <h2 className="text-xl font-semibold">Configured Indexers</h2>
              <p className="text-sm text-muted-foreground mt-1">
                Indexers are used to search for books and audiobooks across various sources
              </p>
            </div>
            <Button onClick={openAddDialog}>
              <Plus className="h-4 w-4" />
              Add Indexer
            </Button>
          </div>

          {/* Indexers List */}
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : indexers && indexers.length > 0 ? (
            <div className="space-y-3">
              {indexers.map((indexer) => (
                <div
                  key={indexer.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-card border border-border"
                >
                  <div className="flex items-center gap-4">
                    <div
                      className={`w-2 h-8 rounded-full ${
                        indexer.enabled ? 'bg-status-downloaded' : 'bg-muted'
                      }`}
                    />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{indexer.name}</span>
                        <span className="px-2 py-0.5 text-xs rounded-full bg-secondary text-secondary-foreground">
                          {indexerTypeInfo[indexer.type].name}
                        </span>
                        {!indexer.enabled && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-destructive/20 text-destructive">
                            Disabled
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-muted-foreground mt-1">
                        Priority: {indexer.priority}
                        {indexer.vipOnly && ' • VIP Only'}
                        {indexer.freeleechOnly && ' • Freeleech Only'}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {/* Test result indicator */}
                    {testResults[indexer.id] && (
                      <div
                        className={`flex items-center gap-1 px-2 py-1 rounded text-xs ${
                          testResults[indexer.id]?.success
                            ? 'bg-status-downloaded/20 text-status-downloaded'
                            : 'bg-destructive/20 text-destructive'
                        }`}
                      >
                        {testResults[indexer.id]?.success ? (
                          <Check className="h-3 w-3" />
                        ) : (
                          <X className="h-3 w-3" />
                        )}
                        {testResults[indexer.id]?.success ? 'Connected' : 'Failed'}
                      </div>
                    )}

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleTest(indexer.id)}
                      disabled={testMutation.isPending && testMutation.variables === indexer.id}
                    >
                      {testMutation.isPending && testMutation.variables === indexer.id ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Zap className="h-4 w-4" />
                      )}
                      Test
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => openEditDialog(indexer)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDelete(indexer.id)}
                      className="text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-12 border border-dashed border-border rounded-lg">
              <div className="text-muted-foreground mb-4">No indexers configured yet</div>
              <Button onClick={openAddDialog}>
                <Plus className="h-4 w-4" />
                Add Your First Indexer
              </Button>
            </div>
          )}

          {/* Info Cards */}
          <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
            {Object.entries(indexerTypeInfo).map(([type, info]) => (
              <div key={type} className="p-4 rounded-lg bg-secondary/50 border border-border">
                <h4 className="font-medium">{info.name}</h4>
                <p className="text-sm text-muted-foreground mt-1">{info.description}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Add/Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editingIndexer ? 'Edit Indexer' : 'Add Indexer'}</DialogTitle>
            <DialogDescription>
              {editingIndexer
                ? 'Update the indexer configuration'
                : 'Configure a new indexer to search for content'}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Indexer"
                required
              />
            </div>

            {/* Type */}
            <div className="space-y-2">
              <Label>Type</Label>
              <Select
                value={formData.type}
                onValueChange={(value: IndexerType) => setFormData({ ...formData, type: value })}
                disabled={!!editingIndexer}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="mam">MyAnonamouse (MAM)</SelectItem>
                  <SelectItem value="torznab">Torznab (Prowlarr/Jackett)</SelectItem>
                  <SelectItem value="anna">Anna's Archive</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* URL (for torznab) */}
            {typeFields.includes('url') && (
              <div className="space-y-2">
                <Label htmlFor="url">URL</Label>
                <Input
                  id="url"
                  type="url"
                  value={formData.url}
                  onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                  placeholder="https://indexer.example.com/api"
                  required={formData.type === 'torznab'}
                />
              </div>
            )}

            {/* API Key (for torznab) */}
            {typeFields.includes('apiKey') && (
              <div className="space-y-2">
                <Label htmlFor="apiKey">API Key</Label>
                <Input
                  id="apiKey"
                  type="password"
                  value={formData.apiKey}
                  onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
                  placeholder={editingIndexer ? '••••••••' : 'Your API key'}
                />
                {editingIndexer && (
                  <p className="text-xs text-muted-foreground">Leave blank to keep existing key</p>
                )}
              </div>
            )}

            {/* Cookie (for MAM) */}
            {typeFields.includes('cookie') && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="cookie">Session Cookie</Label>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowMamHelp(!showMamHelp)}
                    className="h-6 px-2 text-xs text-muted-foreground hover:text-foreground"
                  >
                    <HelpCircle className="h-3.5 w-3.5 mr-1" />
                    How to get cookie
                  </Button>
                </div>
                <Input
                  id="cookie"
                  type="password"
                  value={formData.cookie}
                  onChange={(e) => setFormData({ ...formData, cookie: e.target.value })}
                  placeholder={editingIndexer ? '••••••••' : 'Paste your session token here'}
                />
                <p className="text-xs text-muted-foreground">
                  Just paste the token value - we'll add the <code className="px-1 py-0.5 rounded bg-muted">mam_id=</code> prefix automatically
                </p>
                
                {/* MAM Authentication Help Section */}
                <Collapsible open={showMamHelp} onOpenChange={setShowMamHelp}>
                  <CollapsibleContent className="mt-3">
                    <div className="rounded-lg border border-border bg-secondary/30 p-4 space-y-4 text-sm">
                      <div className="flex items-start gap-2">
                        <HelpCircle className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                        <div>
                          <h4 className="font-semibold text-foreground">How to Get Your MAM Session Token</h4>
                          <p className="text-muted-foreground mt-1">
                            MAM uses session tokens for API access. Follow these steps:
                          </p>
                        </div>
                      </div>
                      
                      <div className="space-y-3 pl-7">
                        <div className="flex gap-2">
                          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground text-xs font-medium shrink-0">1</span>
                          <p className="text-muted-foreground">
                            Log in to{' '}
                            <a 
                              href="https://www.myanonamouse.net" 
                              target="_blank" 
                              rel="noopener noreferrer"
                              className="text-primary hover:underline inline-flex items-center gap-1"
                            >
                              MyAnonamouse.net
                              <ExternalLink className="h-3 w-3" />
                            </a>
                          </p>
                        </div>
                        
                        <div className="flex gap-2">
                          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground text-xs font-medium shrink-0">2</span>
                          <p className="text-muted-foreground">
                            Go to your profile, then navigate to{' '}
                            <a 
                              href="https://www.myanonamouse.net/preferences/security" 
                              target="_blank" 
                              rel="noopener noreferrer"
                              className="text-primary hover:underline inline-flex items-center gap-1"
                            >
                              Preferences → Security
                              <ExternalLink className="h-3 w-3" />
                            </a>
                          </p>
                        </div>
                        
                        <div className="flex gap-2">
                          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground text-xs font-medium shrink-0">3</span>
                          <p className="text-muted-foreground">
                            Click <strong>"Add Session"</strong> to create a new session token
                          </p>
                        </div>
                        
                        <div className="flex gap-2">
                          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground text-xs font-medium shrink-0">4</span>
                          <p className="text-muted-foreground">
                            Enter your <strong>server's public IP address</strong> (where Shelfarr is hosted)
                          </p>
                        </div>
                        
                        <div className="flex gap-2">
                          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground text-xs font-medium shrink-0">5</span>
                          <p className="text-muted-foreground">
                            Copy the generated <strong>token value</strong> and paste it above
                          </p>
                        </div>
                      </div>
                      
                      <div className="rounded-md bg-muted/50 p-3 pl-7">
                        <p className="text-xs text-muted-foreground font-mono break-all">
                          <strong>Example token:</strong><br />
                          a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
                        </p>
                      </div>
                      
                      <div className="flex items-start gap-2 rounded-md bg-yellow-500/10 border border-yellow-500/20 p-3">
                        <AlertTriangle className="h-4 w-4 text-yellow-500 shrink-0 mt-0.5" />
                        <div className="text-xs">
                          <p className="font-medium text-yellow-600 dark:text-yellow-400">Important Notes:</p>
                          <ul className="mt-1 space-y-1 text-muted-foreground">
                            <li>• The session must be created with your server's <strong>public IP address</strong></li>
                            <li>• Session tokens can be revoked anytime from the Security page</li>
                            <li>• Keep your token private - it provides API access to your account</li>
                          </ul>
                        </div>
                      </div>
                      
                      <div className="pt-2 pl-7">
                        <a 
                          href="https://www.myanonamouse.net/preferences/security"
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs text-primary hover:underline inline-flex items-center gap-1"
                        >
                          <ExternalLink className="h-3 w-3" />
                          Go to MAM Security Settings
                        </a>
                      </div>
                    </div>
                  </CollapsibleContent>
                </Collapsible>
              </div>
            )}

            {/* Priority */}
            <div className="space-y-2">
              <Label htmlFor="priority">Priority</Label>
              <Input
                id="priority"
                type="number"
                min="0"
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 0 })}
              />
              <p className="text-xs text-muted-foreground">
                Lower numbers are searched first
              </p>
            </div>

            {/* Enabled */}
            <div className="flex items-center justify-between">
              <Label htmlFor="enabled">Enabled</Label>
              <Switch
                id="enabled"
                checked={formData.enabled}
                onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
              />
            </div>

            {/* MAM-specific options */}
            {typeFields.includes('vipOnly') && (
              <>
                <div className="flex items-center justify-between">
                  <div>
                    <Label htmlFor="vipOnly">VIP Only</Label>
                    <p className="text-xs text-muted-foreground">Only return VIP torrents</p>
                  </div>
                  <Switch
                    id="vipOnly"
                    checked={formData.vipOnly}
                    onCheckedChange={(checked) => setFormData({ ...formData, vipOnly: checked })}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <Label htmlFor="freeleechOnly">Freeleech Only</Label>
                    <p className="text-xs text-muted-foreground">Only return freeleech torrents</p>
                  </div>
                  <Switch
                    id="freeleechOnly"
                    checked={formData.freeleechOnly}
                    onCheckedChange={(checked) => setFormData({ ...formData, freeleechOnly: checked })}
                  />
                </div>
              </>
            )}

            <DialogFooter className="mt-6">
              <Button type="button" variant="outline" onClick={closeDialog}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={addMutation.isPending || updateMutation.isPending}
              >
                {(addMutation.isPending || updateMutation.isPending) && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                {editingIndexer ? 'Save Changes' : 'Add Indexer'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

