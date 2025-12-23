import { useState, useEffect } from 'react';
import { Plus, Trash2, Settings, CheckCircle2, XCircle, Loader2, Bell, RefreshCw } from 'lucide-react';
import { apiClient, Notification } from '../api/client';

type NotificationType = 'webhook' | 'discord' | 'telegram';

const NOTIFICATION_TYPES: { value: NotificationType; label: string; icon: string }[] = [
  { value: 'webhook', label: 'Webhook', icon: '🔗' },
  { value: 'discord', label: 'Discord', icon: '💬' },
  { value: 'telegram', label: 'Telegram', icon: '📱' },
];

interface NotificationFormData {
  name: string;
  type: NotificationType;
  enabled: boolean;
  webhookUrl: string;
  discordWebhook: string;
  telegramBotToken: string;
  telegramChatId: string;
  onGrab: boolean;
  onDownload: boolean;
  onUpgrade: boolean;
  onImport: boolean;
  onDelete: boolean;
  onHealthIssue: boolean;
}

const defaultFormData: NotificationFormData = {
  name: '',
  type: 'webhook',
  enabled: true,
  webhookUrl: '',
  discordWebhook: '',
  telegramBotToken: '',
  telegramChatId: '',
  onGrab: false,
  onDownload: true,
  onUpgrade: false,
  onImport: true,
  onDelete: false,
  onHealthIssue: true,
};

export default function NotificationsSettingsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [showDialog, setShowDialog] = useState(false);
  const [editingNotification, setEditingNotification] = useState<Notification | null>(null);
  const [formData, setFormData] = useState<NotificationFormData>(defaultFormData);
  const [testStatus, setTestStatus] = useState<Record<number, 'testing' | 'success' | 'error'>>({});

  useEffect(() => {
    loadNotifications();
  }, []);

  const loadNotifications = async () => {
    try {
      const data = await apiClient.getNotifications();
      setNotifications(data);
    } catch (error) {
      console.error('Failed to load notifications:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingNotification) {
        await apiClient.updateNotification(editingNotification.id, formData);
      } else {
        await apiClient.createNotification(formData);
      }
      setShowDialog(false);
      setEditingNotification(null);
      setFormData(defaultFormData);
      loadNotifications();
    } catch (error) {
      console.error('Failed to save notification:', error);
    }
  };

  const handleEdit = (notification: Notification) => {
    setEditingNotification(notification);
    setFormData({
      name: notification.name,
      type: notification.type as NotificationType,
      enabled: notification.enabled,
      webhookUrl: notification.webhookUrl || '',
      discordWebhook: notification.discordWebhook || '',
      telegramBotToken: notification.telegramBotToken || '',
      telegramChatId: notification.telegramChatId || '',
      onGrab: notification.onGrab,
      onDownload: notification.onDownload,
      onUpgrade: notification.onUpgrade,
      onImport: notification.onImport,
      onDelete: notification.onDelete,
      onHealthIssue: notification.onHealthIssue,
    });
    setShowDialog(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this notification?')) return;
    try {
      await apiClient.deleteNotification(id);
      loadNotifications();
    } catch (error) {
      console.error('Failed to delete notification:', error);
    }
  };

  const handleTest = async (id: number) => {
    setTestStatus({ ...testStatus, [id]: 'testing' });
    try {
      await apiClient.testNotification(id);
      setTestStatus({ ...testStatus, [id]: 'success' });
    } catch (error) {
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
    setEditingNotification(null);
    setFormData(defaultFormData);
    setShowDialog(true);
  };

  const getEnabledTriggers = (notification: Notification): string[] => {
    const triggers: string[] = [];
    if (notification.onGrab) triggers.push('Grab');
    if (notification.onDownload) triggers.push('Download');
    if (notification.onUpgrade) triggers.push('Upgrade');
    if (notification.onImport) triggers.push('Import');
    if (notification.onDelete) triggers.push('Delete');
    if (notification.onHealthIssue) triggers.push('Health');
    return triggers;
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
          <h1 className="text-2xl font-bold text-neutral-100">Notifications</h1>
          <p className="text-neutral-400 mt-1">Configure alerts for downloads, imports, and system events</p>
        </div>
        <button
          onClick={openAddDialog}
          className="flex items-center gap-2 px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Notification
        </button>
      </div>

      {/* Notifications List */}
      {notifications.length === 0 ? (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
          <Bell className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
          <h3 className="text-lg font-medium text-neutral-300 mb-2">No Notifications</h3>
          <p className="text-neutral-500 mb-4">Add a notification to receive alerts</p>
          <button
            onClick={openAddDialog}
            className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
          >
            Add Your First Notification
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {notifications.map((notification) => (
            <div
              key={notification.id}
              className={`bg-neutral-800/50 border rounded-xl p-4 ${
                notification.enabled ? 'border-neutral-700' : 'border-neutral-700/50 opacity-60'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-neutral-700 rounded-lg flex items-center justify-center text-2xl">
                    {NOTIFICATION_TYPES.find((t) => t.value === notification.type)?.icon || '🔔'}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-neutral-100">{notification.name}</h3>
                      {!notification.enabled && (
                        <span className="px-2 py-0.5 text-xs bg-neutral-700 text-neutral-400 rounded">
                          Disabled
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-neutral-500">
                      {NOTIFICATION_TYPES.find((t) => t.value === notification.type)?.label || notification.type}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      {getEnabledTriggers(notification).map((trigger) => (
                        <span
                          key={trigger}
                          className="text-xs bg-neutral-700 text-neutral-300 px-2 py-0.5 rounded"
                        >
                          {trigger}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {/* Test Status */}
                  {testStatus[notification.id] === 'testing' && (
                    <Loader2 className="w-5 h-5 animate-spin text-sky-400" />
                  )}
                  {testStatus[notification.id] === 'success' && (
                    <CheckCircle2 className="w-5 h-5 text-green-400" />
                  )}
                  {testStatus[notification.id] === 'error' && (
                    <XCircle className="w-5 h-5 text-red-400" />
                  )}

                  <button
                    onClick={() => handleTest(notification.id)}
                    className="p-2 text-neutral-400 hover:text-sky-400 hover:bg-neutral-700/50 rounded-lg transition-colors"
                    title="Test notification"
                  >
                    <RefreshCw className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleEdit(notification)}
                    className="p-2 text-neutral-400 hover:text-neutral-200 hover:bg-neutral-700/50 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <Settings className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(notification.id)}
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
                {editingNotification ? 'Edit Notification' : 'Add Notification'}
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
                  placeholder="My Notification"
                  required
                />
              </div>

              {/* Type */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-1">Type</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value as NotificationType })}
                  className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                >
                  {NOTIFICATION_TYPES.map((type) => (
                    <option key={type.value} value={type.value}>
                      {type.icon} {type.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Type-specific fields */}
              {formData.type === 'webhook' && (
                <div>
                  <label className="block text-sm font-medium text-neutral-300 mb-1">Webhook URL</label>
                  <input
                    type="url"
                    value={formData.webhookUrl}
                    onChange={(e) => setFormData({ ...formData, webhookUrl: e.target.value })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    placeholder="https://example.com/webhook"
                    required
                  />
                </div>
              )}

              {formData.type === 'discord' && (
                <div>
                  <label className="block text-sm font-medium text-neutral-300 mb-1">Discord Webhook URL</label>
                  <input
                    type="url"
                    value={formData.discordWebhook}
                    onChange={(e) => setFormData({ ...formData, discordWebhook: e.target.value })}
                    className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                    placeholder="https://discord.com/api/webhooks/..."
                    required
                  />
                </div>
              )}

              {formData.type === 'telegram' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-neutral-300 mb-1">Bot Token</label>
                    <input
                      type="text"
                      value={formData.telegramBotToken}
                      onChange={(e) => setFormData({ ...formData, telegramBotToken: e.target.value })}
                      className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                      placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-300 mb-1">Chat ID</label>
                    <input
                      type="text"
                      value={formData.telegramChatId}
                      onChange={(e) => setFormData({ ...formData, telegramChatId: e.target.value })}
                      className="w-full px-3 py-2 bg-neutral-900 border border-neutral-700 rounded-lg text-neutral-100 focus:outline-none focus:border-sky-500"
                      placeholder="-1001234567890"
                      required
                    />
                  </div>
                </>
              )}

              {/* Triggers */}
              <div>
                <label className="block text-sm font-medium text-neutral-300 mb-2">Notification Triggers</label>
                <div className="grid grid-cols-2 gap-2">
                  {[
                    { key: 'onGrab', label: 'On Grab' },
                    { key: 'onDownload', label: 'On Download' },
                    { key: 'onUpgrade', label: 'On Upgrade' },
                    { key: 'onImport', label: 'On Import' },
                    { key: 'onDelete', label: 'On Delete' },
                    { key: 'onHealthIssue', label: 'On Health Issue' },
                  ].map(({ key, label }) => (
                    <label key={key} className="flex items-center gap-2 text-sm text-neutral-300">
                      <input
                        type="checkbox"
                        checked={formData[key as keyof NotificationFormData] as boolean}
                        onChange={(e) => setFormData({ ...formData, [key]: e.target.checked })}
                        className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                      />
                      {label}
                    </label>
                  ))}
                </div>
              </div>

              {/* Enabled */}
              <div className="flex items-center gap-3 pt-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  className="w-4 h-4 rounded border-neutral-600 bg-neutral-900 text-sky-500 focus:ring-sky-500"
                />
                <label htmlFor="enabled" className="text-sm text-neutral-300">
                  Enable this notification
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
                  {editingNotification ? 'Save Changes' : 'Add Notification'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

