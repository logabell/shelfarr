import { useState, useEffect } from 'react';
import { Plus, Trash2, Settings, Loader2, List, RefreshCw, Clock } from 'lucide-react';
import { apiClient, HardcoverList } from '../api/client';

interface ListFormData {
  name: string;
  hardcoverUrl: string;
  hardcoverId: string;
  enabled: boolean;
  autoAdd: boolean;
  monitor: boolean;
  syncInterval: number;
}

const defaultFormData: ListFormData = {
  name: '',
  hardcoverUrl: '',
  hardcoverId: '',
  enabled: true,
  autoAdd: true,
  monitor: true,
  syncInterval: 6,
};

export default function ListsSettingsPage() {
  const [lists, setLists] = useState<HardcoverList[]>([]);
  const [loading, setLoading] = useState(true);
  const [showDialog, setShowDialog] = useState(false);
  const [editingList, setEditingList] = useState<HardcoverList | null>(null);
  const [formData, setFormData] = useState<ListFormData>(defaultFormData);
  const [syncingList, setSyncingList] = useState<number | null>(null);
  const [syncResult, setSyncResult] = useState<Record<number, { success: boolean; booksAdded?: number }>>({});

  useEffect(() => {
    loadLists();
  }, []);

  const loadLists = async () => {
    try {
      const data = await apiClient.getLists();
      setLists(data);
    } catch (error) {
      console.error('Failed to load lists:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      // Extract list ID from URL if not provided
      let listId = formData.hardcoverId;
      if (!listId && formData.hardcoverUrl) {
        const match = formData.hardcoverUrl.match(/lists\/(\d+)/);
        if (match) {
          listId = match[1];
        }
      }

      const data = { ...formData, hardcoverId: listId };

      if (editingList) {
        await apiClient.updateList(editingList.id, data);
      } else {
        await apiClient.createList(data);
      }
      setShowDialog(false);
      setEditingList(null);
      setFormData(defaultFormData);
      loadLists();
    } catch (error) {
      console.error('Failed to save list:', error);
    }
  };

  const handleEdit = (list: HardcoverList) => {
    setEditingList(list);
    setFormData({
      name: list.name,
      hardcoverUrl: list.hardcoverUrl,
      hardcoverId: list.hardcoverId,
      enabled: list.enabled,
      autoAdd: list.autoAdd,
      monitor: list.monitor,
      syncInterval: list.syncInterval,
    });
    setShowDialog(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this list?')) return;
    try {
      await apiClient.deleteList(id);
      loadLists();
    } catch (error) {
      console.error('Failed to delete list:', error);
    }
  };

  const handleSync = async (id: number) => {
    setSyncingList(id);
    try {
      const result = await apiClient.syncList(id);
      setSyncResult({ ...syncResult, [id]: { success: true, booksAdded: result.booksAdded } });
      loadLists();
    } catch (error) {
      setSyncResult({ ...syncResult, [id]: { success: false } });
    } finally {
      setSyncingList(null);
    }
    setTimeout(() => {
      setSyncResult((prev) => {
        const next = { ...prev };
        delete next[id];
        return next;
      });
    }, 5000);
  };

  const openAddDialog = () => {
    setEditingList(null);
    setFormData(defaultFormData);
    setShowDialog(true);
  };

  const formatLastSync = (lastSynced?: string) => {
    if (!lastSynced) return 'Never';
    const date = new Date(lastSynced);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours < 1) return 'Just now';
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">Import Lists</h1>
          <p className="text-neutral-400 mt-1">Automatically import books from Hardcover.app lists</p>
        </div>
        <button
          onClick={openAddDialog}
          className="flex items-center gap-2 px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add List
        </button>
      </div>

      {/* Lists */}
      {lists.length === 0 ? (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
          <List className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
          <h3 className="text-lg font-medium text-neutral-300 mb-2">No Import Lists</h3>
          <p className="text-neutral-500 mb-4">Add a Hardcover.app list to automatically import books</p>
          <button
            onClick={openAddDialog}
            className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
          >
            Add Your First List
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {lists.map((list) => (
            <div
              key={list.id}
              className={`bg-neutral-800/50 border rounded-xl p-4 ${
                list.enabled ? 'border-neutral-700' : 'border-neutral-700/50 opacity-60'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-gradient-to-br from-amber-600 to-amber-400 rounded-lg flex items-center justify-center">
                    <List className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-neutral-100">{list.name}</h3>
                      {!list.enabled && (
                        <span className="px-2 py-0.5 text-xs bg-neutral-700 text-neutral-400 rounded">
                          Disabled
                        </span>
                      )}
                    </div>
                    <a
                      href={list.hardcoverUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-sky-400 hover:text-sky-300"
                    >
                      {list.hardcoverUrl}
                    </a>
                    <div className="flex items-center gap-4 mt-1 text-xs text-neutral-500">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        Sync: every {list.syncInterval}h
                      </span>
                      <span>Last sync: {formatLastSync(list.lastSyncedAt)}</span>
                      <span className={list.autoAdd ? 'text-green-400' : ''}>
                        {list.autoAdd ? 'Auto-add ON' : 'Auto-add OFF'}
                      </span>
                      <span className={list.monitor ? 'text-sky-400' : ''}>
                        {list.monitor ? 'Monitor ON' : 'Monitor OFF'}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {/* Sync Result */}
                  {syncResult[list.id] && (
                    <span className={`text-xs px-2 py-1 rounded ${
                      syncResult[list.id].success 
                        ? 'bg-green-500/20 text-green-400' 
                        : 'bg-red-500/20 text-red-400'
                    }`}>
                      {syncResult[list.id].success 
                        ? `+${syncResult[list.id].booksAdded || 0} books` 
                        : 'Sync failed'}
                    </span>
                  )}

                  <button
                    onClick={() => handleSync(list.id)}
                    disabled={syncingList === list.id}
                    className="p-2 text-neutral-400 hover:text-sky-400 hover:bg-neutral-700/50 rounded-lg transition-colors disabled:opacity-50"
                    title="Sync now"
                  >
                    {syncingList === list.id ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <RefreshCw className="w-4 h-4" />
                    )}
                  </button>
                  <button
                    onClick={() => handleEdit(list)}
                    className="p-2 text-neutral-400 hover:text-neutral-200 hover:bg-neutral-700/50 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <Settings className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(list.id)}
                    className="p-2 text-neutral-400 hover:text-red-400 hover:bg-neutral-700/50 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add/Edit Dialog */}
      {showDialog && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
          <div className="bg-neutral-800 rounded-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-neutral-700">
              <h2 className="text-xl font-semibold text-neutral-100">
                {editingList ? 'Edit List' : 'Add Import List'}
              </h2>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              {/* Name */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                  placeholder="My Reading List"
                  required
                />
              </div>

              {/* Hardcover URL */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Hardcover List URL</label>
                <input
                  type="url"
                  value={formData.hardcoverUrl}
                  onChange={(e) => setFormData({ ...formData, hardcoverUrl: e.target.value })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                  placeholder="https://hardcover.app/lists/12345"
                  required
                />
                <p className="text-xs text-neutral-500 mt-1">
                  Paste the full URL of your Hardcover list
                </p>
              </div>

              {/* Sync Interval */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">
                  Sync Interval (hours): {formData.syncInterval}
                </label>
                <input
                  type="range"
                  min="1"
                  max="24"
                  value={formData.syncInterval}
                  onChange={(e) => setFormData({ ...formData, syncInterval: parseInt(e.target.value) })}
                  className="w-full accent-sky-500"
                />
                <div className="flex justify-between text-xs text-neutral-500">
                  <span>1h</span>
                  <span>6h</span>
                  <span>12h</span>
                  <span>24h</span>
                </div>
              </div>

              {/* Options */}
              <div className="space-y-3">
                <label className="flex items-center gap-3 text-sm text-neutral-300">
                  <input
                    type="checkbox"
                    checked={formData.autoAdd}
                    onChange={(e) => setFormData({ ...formData, autoAdd: e.target.checked })}
                    className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                  />
                  <div>
                    <span className="font-medium">Auto-add new books</span>
                    <p className="text-xs text-neutral-500">Automatically add new books found in the list</p>
                  </div>
                </label>

                <label className="flex items-center gap-3 text-sm text-neutral-300">
                  <input
                    type="checkbox"
                    checked={formData.monitor}
                    onChange={(e) => setFormData({ ...formData, monitor: e.target.checked })}
                    className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                  />
                  <div>
                    <span className="font-medium">Monitor added books</span>
                    <p className="text-xs text-neutral-500">Set new books as monitored for automatic downloading</p>
                  </div>
                </label>

                <label className="flex items-center gap-3 text-sm text-neutral-300">
                  <input
                    type="checkbox"
                    checked={formData.enabled}
                    onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                    className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                  />
                  <div>
                    <span className="font-medium">Enable list sync</span>
                    <p className="text-xs text-neutral-500">Automatically sync this list on schedule</p>
                  </div>
                </label>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-3 pt-4 border-t border-neutral-700">
                <button
                  type="button"
                  onClick={() => setShowDialog(false)}
                  className="px-4 py-2 text-neutral-400 hover:text-neutral-200 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
                >
                  {editingList ? 'Save Changes' : 'Add List'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

