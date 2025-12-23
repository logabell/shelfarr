import { useState, useEffect, useRef } from 'react';
import { Plus, Trash2, Settings, CheckCircle2, XCircle, Loader2, Server, RefreshCw, AlertTriangle, HelpCircle, Cloud, ArrowRight } from 'lucide-react';
import { apiClient } from '../api/client';
import type { DownloadClient as DownloadClientType } from '@/types';

type ClientType = 'qbittorrent' | 'transmission' | 'deluge' | 'sabnzbd' | 'nzbget' | 'rapidseedbox-deluge' | 'rapidseedbox-rutorrent';

interface ClientTypeInfo {
  value: ClientType;
  label: string;
  icon: string;
  category: 'torrent' | 'usenet' | 'seedbox';
  defaultPort: number;
  defaultSSL: boolean;
  urlBase?: string;
  helpText?: string;
}

const CLIENT_TYPES: ClientTypeInfo[] = [
  // Torrent Clients
  { value: 'qbittorrent', label: 'qBittorrent', icon: '🌊', category: 'torrent', defaultPort: 8080, defaultSSL: false },
  { value: 'transmission', label: 'Transmission', icon: '⚡', category: 'torrent', defaultPort: 9091, defaultSSL: false },
  { value: 'deluge', label: 'Deluge', icon: '🔥', category: 'torrent', defaultPort: 8112, defaultSSL: false },
  // Usenet Clients
  { value: 'sabnzbd', label: 'SABnzbd', icon: '📰', category: 'usenet', defaultPort: 8085, defaultSSL: false },
  { value: 'nzbget', label: 'NZBGet', icon: '📥', category: 'usenet', defaultPort: 6789, defaultSSL: false },
  // Seedbox Providers
  { 
    value: 'rapidseedbox-deluge', 
    label: 'RapidSeedbox (Deluge)', 
    icon: '☁️', 
    category: 'seedbox', 
    defaultPort: 443, 
    defaultSSL: true,
    helpText: 'Use your Deluge hostname from RapidSeedbox client area. Enable the Label plugin in Deluge before connecting.'
  },
  { 
    value: 'rapidseedbox-rutorrent', 
    label: 'RapidSeedbox (ruTorrent)', 
    icon: '☁️', 
    category: 'seedbox', 
    defaultPort: 443, 
    defaultSSL: true,
    urlBase: '/plugins/rpc/rpc.php',
    helpText: 'Use your ruTorrent hostname from RapidSeedbox client area.'
  },
];

interface RemotePathMapping {
  id?: number;
  clientId: number;
  remotePath: string;
  localPath: string;
}

interface ClientFormData {
  name: string;
  type: ClientType;
  host: string;
  port: number;
  useSsl: boolean;
  urlBase: string;
  username: string;
  password: string;
  category: string;
  priority: number;
  enabled: boolean;
}

const getDefaultFormData = (type?: ClientType): ClientFormData => {
  const clientInfo = CLIENT_TYPES.find(c => c.value === type) || CLIENT_TYPES[0];
  return {
    name: '',
    type: clientInfo.value,
    host: '',
    port: clientInfo.defaultPort,
    useSsl: clientInfo.defaultSSL,
    urlBase: clientInfo.urlBase || '',
    username: '',
    password: '',
    category: 'books',
    priority: 50,
    enabled: true,
  };
};

export default function DownloadClientsSettingsPage() {
  const [clients, setClients] = useState<DownloadClientType[]>([]);
  const [loading, setLoading] = useState(true);
  const [showDialog, setShowDialog] = useState(false);
  const [editingClient, setEditingClient] = useState<DownloadClientType | null>(null);
  const [formData, setFormData] = useState<ClientFormData>(getDefaultFormData());
  const formDataRef = useRef<ClientFormData>(formData);
  
  // Keep ref in sync with state
  useEffect(() => {
    formDataRef.current = formData;
  }, [formData]);
  
  const [testStatus, setTestStatus] = useState<Record<number, 'testing' | 'success' | 'error'>>({});
  
  // Remote path mappings (stored separately)
  const [remotePathMappings, setRemotePathMappings] = useState<RemotePathMapping[]>([]);
  const [newMapping, setNewMapping] = useState<Omit<RemotePathMapping, 'id'>>({ clientId: 0, remotePath: '', localPath: '' });
  
  // Delete confirmation dialog
  const [deleteConfirmClient, setDeleteConfirmClient] = useState<DownloadClientType | null>(null);
  
  // Dialog-specific state
  const [dialogTestStatus, setDialogTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle');
  const [dialogTestMessage, setDialogTestMessage] = useState<string>('');
  const [saving, setSaving] = useState(false);
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    loadClients();
    loadRemotePathMappings();
  }, []);

  const loadClients = async () => {
    try {
      const data = await apiClient.getDownloadClients();
      setClients(data);
    } catch (error) {
      console.error('Failed to load download clients:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadRemotePathMappings = async () => {
    // Load from localStorage for now (can be moved to backend later)
    const stored = localStorage.getItem('remotePathMappings');
    if (stored) {
      try {
        setRemotePathMappings(JSON.parse(stored));
      } catch (e) {
        console.error('Failed to parse remote path mappings:', e);
      }
    }
  };

  const saveRemotePathMappings = (mappings: RemotePathMapping[]) => {
    setRemotePathMappings(mappings);
    localStorage.setItem('remotePathMappings', JSON.stringify(mappings));
  };

  // Build full URL from components
  const buildUrl = (host: string, port: number, useSsl: boolean, urlBase: string): string => {
    const protocol = useSsl ? 'https' : 'http';
    const portStr = (useSsl && port === 443) || (!useSsl && port === 80) ? '' : `:${port}`;
    const base = urlBase.startsWith('/') ? urlBase : (urlBase ? `/${urlBase}` : '');
    return `${protocol}://${host}${portStr}${base}`;
  };

  // Get the underlying client type for API calls
  const getApiClientType = (type: ClientType): string => {
    if (type === 'rapidseedbox-deluge') return 'deluge';
    if (type === 'rapidseedbox-rutorrent') return 'rtorrent';
    return type;
  };

  const handleTestInDialog = async () => {
    if (!formData.name.trim()) {
      setDialogTestStatus('error');
      setDialogTestMessage('Name is required');
      return;
    }
    if (!formData.host.trim()) {
      setDialogTestStatus('error');
      setDialogTestMessage('Host is required');
      return;
    }

    setDialogTestStatus('testing');
    setDialogTestMessage('Testing connection...');

    try {
      const url = buildUrl(formData.host, formData.port, formData.useSsl, formData.urlBase);
      const apiType = getApiClientType(formData.type);
      
      console.log('Testing connection:', { type: apiType, url, username: formData.username, password: '***' });
      
      await apiClient.testDownloadClientConfig({
        type: apiType,
        url: url,
        username: formData.username,
        password: formData.password,
      });
      setDialogTestStatus('success');
      setDialogTestMessage('Connection successful!');
    } catch (error: unknown) {
      console.error('Connection test failed:', error);
      setDialogTestStatus('error');
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { data?: { error?: string }, status?: number } };
        const status = axiosError.response?.status;
        const message = axiosError.response?.data?.error;
        if (status === 401) {
          setDialogTestMessage('Authentication required - please login first or set AUTH_DISABLED=true for development');
        } else {
          setDialogTestMessage(message || `Connection failed (HTTP ${status})`);
        }
      } else if (error instanceof Error) {
        setDialogTestMessage(error.message);
      } else {
        setDialogTestMessage('Connection failed - check host, port, and credentials');
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (dialogTestStatus !== 'success') {
      setDialogTestStatus('error');
      setDialogTestMessage('You must test the connection before saving');
      return;
    }

    // Use ref to get the most current form data
    const currentFormData = formDataRef.current;
    
    console.log('handleSubmit - formData from state:', formData);
    console.log('handleSubmit - formData from ref:', currentFormData);

    setSaving(true);
    try {
      const url = buildUrl(currentFormData.host, currentFormData.port, currentFormData.useSsl, currentFormData.urlBase);
      const apiType = getApiClientType(currentFormData.type) as 'qbittorrent' | 'transmission' | 'deluge' | 'sabnzbd' | 'nzbget' | 'rtorrent' | 'internal';
      const payload = {
        name: currentFormData.name,
        type: apiType,
        url: url,
        username: currentFormData.username,
        password: currentFormData.password,
        category: currentFormData.category,
        priority: currentFormData.priority,
        enabled: currentFormData.enabled,
        settings: JSON.stringify({
          originalType: currentFormData.type,
          host: currentFormData.host,
          port: currentFormData.port,
          useSsl: currentFormData.useSsl,
          urlBase: currentFormData.urlBase,
        }),
      };

      console.log('handleSubmit - payload being sent:', payload);

      if (editingClient) {
        await apiClient.updateDownloadClient(editingClient.id, payload);
      } else {
        await apiClient.createDownloadClient(payload);
      }
      setShowDialog(false);
      setEditingClient(null);
      setFormData(getDefaultFormData());
      resetDialogState();
      loadClients();
    } catch (error) {
      console.error('Failed to save download client:', error);
      setDialogTestStatus('error');
      setDialogTestMessage('Failed to save client');
    } finally {
      setSaving(false);
    }
  };

  const resetDialogState = () => {
    setDialogTestStatus('idle');
    setDialogTestMessage('');
    setShowHelp(false);
  };

  const handleEdit = (client: DownloadClientType) => {
    console.log('handleEdit called with client:', client);
    setEditingClient(client);
    
    // Try to parse stored settings
    let settings: Record<string, unknown> = {};
    try {
      if ((client as { settings?: string }).settings) {
        settings = JSON.parse((client as { settings?: string }).settings || '{}');
      }
    } catch {
      // Parse URL to extract components
      try {
        const url = new URL(client.url);
        settings = {
          host: url.hostname,
          port: parseInt(url.port) || (url.protocol === 'https:' ? 443 : 80),
          useSsl: url.protocol === 'https:',
          urlBase: url.pathname !== '/' ? url.pathname : '',
        };
      } catch {
        settings = { host: client.url, port: 8080, useSsl: false, urlBase: '' };
      }
    }

    setFormData({
      name: client.name,
      type: (settings.originalType as ClientType) || client.type as ClientType,
      host: (settings.host as string) || '',
      port: (settings.port as number) || 8080,
      useSsl: (settings.useSsl as boolean) ?? false,
      urlBase: (settings.urlBase as string) || '',
      username: client.username || '',
      password: client.password || '',
      category: client.category || 'books',
      priority: client.priority,
      enabled: client.enabled,
    });
    setDialogTestStatus('success'); // Already saved = already tested
    setDialogTestMessage('');
    setShowDialog(true);
  };

  const handleDeleteClick = (client: DownloadClientType) => {
    console.log('handleDeleteClick called with client:', client);
    setDeleteConfirmClient(client);
  };

  const handleDeleteConfirm = async () => {
    console.log('handleDeleteConfirm called, deleteConfirmClient:', deleteConfirmClient);
    if (!deleteConfirmClient) return;
    try {
      console.log('Calling deleteDownloadClient with id:', deleteConfirmClient.id);
      await apiClient.deleteDownloadClient(deleteConfirmClient.id);
      // Also remove any path mappings for this client
      const updatedMappings = remotePathMappings.filter(m => m.clientId !== deleteConfirmClient.id);
      saveRemotePathMappings(updatedMappings);
      setDeleteConfirmClient(null);
      loadClients();
    } catch (error) {
      console.error('Failed to delete download client:', error);
    }
  };

  const handleTest = async (id: number) => {
    setTestStatus({ ...testStatus, [id]: 'testing' });
    try {
      await apiClient.testDownloadClient(id);
      setTestStatus({ ...testStatus, [id]: 'success' });
    } catch {
      setTestStatus({ ...testStatus, [id]: 'error' });
    }
    setTimeout(() => {
      setTestStatus((prev) => {
        const next = { ...prev };
        delete next[id];
        return next;
      });
    }, 3000);
  };

  const openAddDialog = () => {
    setEditingClient(null);
    setFormData(getDefaultFormData());
    resetDialogState();
    setShowDialog(true);
  };

  const closeDialog = () => {
    setShowDialog(false);
    setEditingClient(null);
    setFormData(getDefaultFormData());
    resetDialogState();
  };

  const handleTypeChange = (newType: ClientType) => {
    const typeInfo = CLIENT_TYPES.find(c => c.value === newType);
    if (typeInfo) {
      setFormData({
        ...formData,
        type: newType,
        port: typeInfo.defaultPort,
        useSsl: typeInfo.defaultSSL,
        urlBase: typeInfo.urlBase || '',
      });
      // Reset test status when type changes
      if (dialogTestStatus === 'success') {
        setDialogTestStatus('idle');
        setDialogTestMessage('');
      }
    }
  };

  const handleFormChange = (updates: Partial<ClientFormData>) => {
    // Use functional update to avoid stale closure issues
    setFormData(prev => ({ ...prev, ...updates }));
    // Reset test status when connection-related fields change
    if ('host' in updates || 'port' in updates || 'useSsl' in updates || 'urlBase' in updates || 'username' in updates || 'password' in updates) {
      if (dialogTestStatus === 'success') {
        setDialogTestStatus('idle');
        setDialogTestMessage('');
      }
    }
  };

  // Remote Path Mapping handlers
  const addRemotePathMapping = () => {
    if (!newMapping.clientId || !newMapping.remotePath || !newMapping.localPath) {
      return;
    }
    const mapping: RemotePathMapping = {
      id: Date.now(),
      ...newMapping,
    };
    saveRemotePathMappings([...remotePathMappings, mapping]);
    setNewMapping({ clientId: 0, remotePath: '', localPath: '' });
  };

  const removeRemotePathMapping = (id: number) => {
    saveRemotePathMappings(remotePathMappings.filter(m => m.id !== id));
  };

  const currentTypeInfo = CLIENT_TYPES.find(c => c.value === formData.type);
  const isSeedbox = currentTypeInfo?.category === 'seedbox';
  const isUsenet = currentTypeInfo?.category === 'usenet';
  const needsUsername = !['sabnzbd', 'deluge', 'rapidseedbox-deluge'].includes(formData.type);

  // Check if any clients are seedboxes (for showing remote path section)
  const hasSeedboxClients = clients.some(c => {
    try {
      const settings = JSON.parse((c as { settings?: string }).settings || '{}');
      return settings.originalType?.includes('seedbox') || settings.originalType?.includes('rapidseedbox');
    } catch {
      return false;
    }
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">Download Clients</h1>
          <p className="text-neutral-400 mt-1">Configure your torrent, Usenet, and seedbox download clients</p>
        </div>
        <button
          onClick={openAddDialog}
          className="flex items-center gap-2 px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Client
        </button>
      </div>

      {/* Clients List */}
      {clients.length === 0 ? (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
          <Server className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
          <h3 className="text-lg font-medium text-neutral-300 mb-2">No Download Clients</h3>
          <p className="text-neutral-500 mb-4">Add a download client to start downloading books</p>
          <button
            onClick={openAddDialog}
            className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
          >
            Add Your First Client
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {clients.map((client) => {
            // Parse settings to get original type for proper icon/label display
            let isClientSeedbox = false;
            let originalType: string = client.type;
            try {
              const settings = JSON.parse((client as { settings?: string }).settings || '{}');
              isClientSeedbox = settings.originalType?.includes('seedbox') || settings.originalType?.includes('rapidseedbox');
              if (settings.originalType) {
                originalType = settings.originalType;
              }
            } catch {
              isClientSeedbox = false;
            }
            
            // Find the client type info using the original type or fallback to API type
            const clientTypeInfo = CLIENT_TYPES.find((t) => t.value === originalType) || 
                                   CLIENT_TYPES.find((t) => t.value === client.type || getApiClientType(t.value) === client.type);
            
            return (
              <div
                key={client.id}
                className={`bg-neutral-800/50 border rounded-xl p-4 ${
                  client.enabled ? 'border-neutral-700' : 'border-neutral-700/50 opacity-60'
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 bg-neutral-700 rounded-lg flex items-center justify-center text-2xl">
                      {clientTypeInfo?.icon || '📦'}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-medium text-neutral-100">{client.name || 'Unnamed Client'}</h3>
                        {!client.enabled && (
                          <span className="px-2 py-0.5 text-xs bg-neutral-700 text-neutral-400 rounded">
                            Disabled
                          </span>
                        )}
                        {isClientSeedbox && (
                          <span className="px-2 py-0.5 text-xs bg-sky-500/20 text-sky-400 rounded flex items-center gap-1">
                            <Cloud className="w-3 h-3" /> Seedbox
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-neutral-500">
                        {clientTypeInfo?.label || client.type} • {client.url}
                      </p>
                      <p className="text-xs text-neutral-600">
                        Category: {client.category || 'default'} • Priority: {client.priority}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {testStatus[client.id] === 'testing' && (
                      <Loader2 className="w-5 h-5 animate-spin text-sky-400" />
                    )}
                    {testStatus[client.id] === 'success' && (
                      <CheckCircle2 className="w-5 h-5 text-green-400" />
                    )}
                    {testStatus[client.id] === 'error' && (
                      <XCircle className="w-5 h-5 text-red-400" />
                    )}

                    <button
                      onClick={() => handleTest(client.id)}
                      className="p-2 text-neutral-400 hover:text-sky-400 hover:bg-neutral-700/50 rounded-lg transition-colors"
                      title="Test connection"
                    >
                      <RefreshCw className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleEdit(client)}
                      className="p-2 text-neutral-400 hover:text-neutral-200 hover:bg-neutral-700/50 rounded-lg transition-colors"
                      title="Edit"
                    >
                      <Settings className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleDeleteClick(client)}
                      className="p-2 text-neutral-400 hover:text-red-400 hover:bg-neutral-700/50 rounded-lg transition-colors"
                      title="Delete"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Remote Path Mappings Section */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl overflow-hidden">
        <div className="p-4 border-b border-neutral-700 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h2 className="text-lg font-semibold text-neutral-100">Remote Path Mappings</h2>
            <div className="relative group">
              <HelpCircle className="w-4 h-4 text-neutral-500 cursor-help" />
              <div className="absolute left-0 bottom-full mb-2 hidden group-hover:block w-80 p-3 bg-neutral-900 border border-neutral-700 rounded-lg text-xs text-neutral-400 z-10">
                <p className="font-medium text-neutral-300 mb-1">What is Remote Path Mapping?</p>
                <p>When using a seedbox or remote download client, files are downloaded to a path on the remote server. Remote Path Mapping tells Shelfarr how to translate that path to where the files are available locally (after syncing).</p>
                <p className="mt-2 text-sky-400">Required for any seedbox or remote download client.</p>
              </div>
            </div>
          </div>
        </div>

        <div className="p-4 space-y-4">
          {/* Info box for seedbox users */}
          {hasSeedboxClients && remotePathMappings.length === 0 && (
            <div className="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3 flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div className="text-sm">
                <p className="text-amber-400 font-medium">You have seedbox clients configured</p>
                <p className="text-neutral-400 mt-1">Add a remote path mapping to tell Shelfarr where downloaded files appear on your local system after syncing from your seedbox.</p>
              </div>
            </div>
          )}

          {/* Existing mappings */}
          {remotePathMappings.length > 0 && (
            <div className="space-y-2">
              {remotePathMappings.map((mapping) => {
                const client = clients.find(c => c.id === mapping.clientId);
                return (
                  <div key={mapping.id} className="flex items-center gap-3 p-3 bg-neutral-900/50 rounded-lg">
                    <div className="flex-1 grid grid-cols-3 gap-4 items-center">
                      <div>
                        <span className="text-xs text-neutral-500">Client</span>
                        <p className="text-sm text-neutral-200">{client?.name || 'Unknown'}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="flex-1">
                          <span className="text-xs text-neutral-500">Remote Path</span>
                          <p className="text-sm text-neutral-200 font-mono">{mapping.remotePath}</p>
                        </div>
                        <ArrowRight className="w-4 h-4 text-neutral-600 flex-shrink-0" />
                        <div className="flex-1">
                          <span className="text-xs text-neutral-500">Local Path</span>
                          <p className="text-sm text-neutral-200 font-mono">{mapping.localPath}</p>
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() => mapping.id && removeRemotePathMapping(mapping.id)}
                      className="p-2 text-neutral-500 hover:text-red-400 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                );
              })}
            </div>
          )}

          {/* Add new mapping form */}
          <div className="flex items-end gap-3 pt-2">
            <div className="flex-1">
              <label className="block text-xs text-neutral-500 mb-1">Download Client</label>
              <select
                value={newMapping.clientId}
                onChange={(e) => setNewMapping({ ...newMapping, clientId: parseInt(e.target.value) })}
                className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 text-sm focus:outline-none focus:border-sky-500"
              >
                <option value={0}>Select client...</option>
                {clients.map((client) => (
                  <option key={client.id} value={client.id}>{client.name}</option>
                ))}
              </select>
            </div>
            <div className="flex-1">
              <label className="block text-xs text-neutral-500 mb-1">Remote Path</label>
              <input
                type="text"
                value={newMapping.remotePath}
                onChange={(e) => setNewMapping({ ...newMapping, remotePath: e.target.value })}
                placeholder="/home/user/downloads"
                className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 text-sm focus:outline-none focus:border-sky-500 font-mono"
              />
            </div>
            <div className="flex-1">
              <label className="block text-xs text-neutral-500 mb-1">Local Path</label>
              <input
                type="text"
                value={newMapping.localPath}
                onChange={(e) => setNewMapping({ ...newMapping, localPath: e.target.value })}
                placeholder="/mnt/books/downloads"
                className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 text-sm focus:outline-none focus:border-sky-500 font-mono"
              />
            </div>
            <button
              onClick={addRemotePathMapping}
              disabled={!newMapping.clientId || !newMapping.remotePath || !newMapping.localPath}
              className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              Add
            </button>
          </div>
        </div>
      </div>

      {/* Add/Edit Dialog */}
      {showDialog && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
          <div className="bg-neutral-800 rounded-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-neutral-700">
              <h2 className="text-xl font-semibold text-neutral-100">
                {editingClient ? 'Edit Download Client' : 'Add Download Client'}
              </h2>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-5">
              {/* Name */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleFormChange({ name: e.target.value })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                  placeholder="My Seedbox"
                  required
                />
              </div>

              {/* Type - Grouped */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Client Type</label>
                <select
                  value={formData.type}
                  onChange={(e) => handleTypeChange(e.target.value as ClientType)}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                >
                  <optgroup label="🌐 Torrent Clients">
                    {CLIENT_TYPES.filter(c => c.category === 'torrent').map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.icon} {type.label}
                      </option>
                    ))}
                  </optgroup>
                  <optgroup label="📰 Usenet Clients">
                    {CLIENT_TYPES.filter(c => c.category === 'usenet').map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.icon} {type.label}
                      </option>
                    ))}
                  </optgroup>
                  <optgroup label="☁️ Seedbox Providers">
                    {CLIENT_TYPES.filter(c => c.category === 'seedbox').map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.icon} {type.label}
                      </option>
                    ))}
                  </optgroup>
                </select>
              </div>

              {/* Seedbox Help Box */}
              {isSeedbox && (
                <div className="bg-sky-500/10 border border-sky-500/30 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <Cloud className="w-5 h-5 text-sky-400 flex-shrink-0 mt-0.5" />
                    <div className="text-sm">
                      <p className="font-medium text-sky-400 mb-1">Seedbox Configuration</p>
                      <p className="text-neutral-300">{currentTypeInfo?.helpText}</p>
                      <p className="text-neutral-400 mt-2">
                        <strong>Note:</strong> After adding this client, configure Remote Path Mapping below the clients list to map seedbox paths to local paths.
                      </p>
                      <button
                        type="button"
                        onClick={() => setShowHelp(!showHelp)}
                        className="text-sky-400 hover:text-sky-300 mt-2 flex items-center gap-1"
                      >
                        <HelpCircle className="w-4 h-4" />
                        {showHelp ? 'Hide setup guide' : 'Show setup guide'}
                      </button>
                      {showHelp && (
                        <div className="mt-3 p-3 bg-neutral-900/50 rounded-lg text-neutral-400 text-xs space-y-2">
                          <p><strong className="text-neutral-300">1. Connect to Seedbox:</strong> First, add your seedbox credentials here and test the connection.</p>
                          <p><strong className="text-neutral-300">2. Set Up File Sync:</strong> Use Syncthing, rsync, or your seedbox's built-in sync to copy downloaded files to your local storage.</p>
                          <p><strong className="text-neutral-300">3. Add Path Mapping:</strong> After saving, add a Remote Path Mapping in the section below to map the seedbox download path to your local path.</p>
                          <a 
                            href="https://help.rapidseedbox.com/en/articles/6860938-remote-path-mapping-explained"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-sky-400 hover:underline block mt-2"
                          >
                            📖 Read the full Remote Path Mapping guide →
                          </a>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* Host & Port Row */}
              <div className="grid grid-cols-3 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium text-neutral-300 mb-1">Host *</label>
                  <input
                    type="text"
                    value={formData.host}
                    onChange={(e) => handleFormChange({ host: e.target.value })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    placeholder={isSeedbox ? 'xxxxx-dg.swift-XXX.seedbox.vip' : 'localhost'}
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-300 mb-1">Port</label>
                  <input
                    type="number"
                    value={formData.port}
                    onChange={(e) => handleFormChange({ port: parseInt(e.target.value) || 0 })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    min="1"
                    max="65535"
                  />
                </div>
              </div>

              {/* SSL & URL Base Row */}
              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center gap-3 pt-6">
                  <input
                    type="checkbox"
                    id="useSsl"
                    checked={formData.useSsl}
                    onChange={(e) => handleFormChange({ useSsl: e.target.checked })}
                    className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                  />
                  <label htmlFor="useSsl" className="text-sm text-neutral-300">
                    Use SSL (HTTPS)
                  </label>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-300 mb-1">URL Base</label>
                  <input
                    type="text"
                    value={formData.urlBase}
                    onChange={(e) => handleFormChange({ urlBase: e.target.value })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    placeholder={currentTypeInfo?.urlBase || '/'}
                  />
                </div>
              </div>

              {/* Generated URL Preview */}
              <div className="text-xs text-neutral-500">
                URL: <code className="bg-neutral-900 px-2 py-0.5 rounded">{buildUrl(formData.host || 'hostname', formData.port, formData.useSsl, formData.urlBase)}</code>
              </div>

              {/* Username (conditional) */}
              {needsUsername && (
                <div>
                  <label className="block text-sm font-medium text-neutral-300 mb-1">Username</label>
                  <input
                    type="text"
                    value={formData.username}
                    onChange={(e) => handleFormChange({ username: e.target.value })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    placeholder="admin"
                  />
                </div>
              )}

              {/* Password / API Key */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">
                  {isUsenet && formData.type === 'sabnzbd' ? 'API Key' : 
                   formData.type.includes('deluge') ? 'Web UI Password' : 'Password'}
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => handleFormChange({ password: e.target.value })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                  placeholder={formData.type === 'sabnzbd' ? 'Your SABnzbd API key' : '••••••••'}
                />
              </div>

              {/* Category */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Category / Label</label>
                <input
                  type="text"
                  value={formData.category}
                  onChange={(e) => handleFormChange({ category: e.target.value })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                  placeholder="books"
                />
                <p className="text-xs text-neutral-500 mt-1">
                  Downloads will be tagged with this category{isSeedbox ? ' (requires Label plugin enabled in Deluge)' : ''}
                </p>
              </div>

              {/* Priority */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">
                  Priority: {formData.priority}
                </label>
                <input
                  type="range"
                  min="0"
                  max="100"
                  value={formData.priority}
                  onChange={(e) => handleFormChange({ priority: parseInt(e.target.value) })}
                  className="w-full accent-sky-500"
                />
                <p className="text-xs text-neutral-500 mt-1">Higher priority clients are used first</p>
              </div>

              {/* Enabled */}
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => handleFormChange({ enabled: e.target.checked })}
                  className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                />
                <label htmlFor="enabled" className="text-sm text-neutral-300">
                  Enable this download client
                </label>
              </div>

              {/* Test Status Message */}
              {dialogTestMessage && (
                <div className={`flex items-center gap-2 p-3 rounded-lg ${
                  dialogTestStatus === 'success' ? 'bg-green-500/10 border border-green-500/30' :
                  dialogTestStatus === 'error' ? 'bg-red-500/10 border border-red-500/30' :
                  'bg-sky-500/10 border border-sky-500/30'
                }`}>
                  {dialogTestStatus === 'testing' && <Loader2 className="w-4 h-4 animate-spin text-sky-400" />}
                  {dialogTestStatus === 'success' && <CheckCircle2 className="w-4 h-4 text-green-400" />}
                  {dialogTestStatus === 'error' && <AlertTriangle className="w-4 h-4 text-red-400" />}
                  <span className={`text-sm ${
                    dialogTestStatus === 'success' ? 'text-green-400' :
                    dialogTestStatus === 'error' ? 'text-red-400' :
                    'text-sky-400'
                  }`}>
                    {dialogTestMessage}
                  </span>
                </div>
              )}

              {/* Actions */}
              <div className="flex justify-between items-center pt-4 border-t border-neutral-700">
                <button
                  type="button"
                  onClick={handleTestInDialog}
                  disabled={dialogTestStatus === 'testing'}
                  className="flex items-center gap-2 px-4 py-2 bg-neutral-700 text-neutral-200 rounded-lg hover:bg-neutral-600 disabled:opacity-50 transition-colors"
                >
                  {dialogTestStatus === 'testing' ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                  Test Connection
                </button>
                
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={closeDialog}
                    className="px-4 py-2 text-neutral-400 hover:text-neutral-200 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={dialogTestStatus !== 'success' || saving}
                    className="flex items-center gap-2 px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {saving && <Loader2 className="w-4 h-4 animate-spin" />}
                    {editingClient ? 'Save Changes' : 'Add Client'}
                  </button>
                </div>
              </div>

              {dialogTestStatus !== 'success' && dialogTestStatus !== 'testing' && (
                <p className="text-xs text-neutral-500 text-center">
                  You must test the connection before saving
                </p>
              )}
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteConfirmClient && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
          <div className="bg-neutral-800 rounded-xl w-full max-w-md p-6">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-red-500/20 rounded-full flex items-center justify-center">
                <AlertTriangle className="w-5 h-5 text-red-400" />
              </div>
              <h2 className="text-xl font-semibold text-neutral-100">Delete Download Client</h2>
            </div>
            <p className="text-neutral-400 mb-6">
              Are you sure you want to delete <span className="font-medium text-neutral-200">{deleteConfirmClient.name}</span>? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setDeleteConfirmClient(null)}
                className="px-4 py-2 text-neutral-400 hover:text-neutral-200 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-500 transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
