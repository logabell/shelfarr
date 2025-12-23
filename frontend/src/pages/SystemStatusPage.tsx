import { useState, useEffect } from 'react';
import { 
  Server, 
  Database, 
  HardDrive, 
  Cpu, 
  Clock, 
  CheckCircle2, 
  XCircle, 
  Loader2, 
  RefreshCw,
  Play,
  Calendar,
  Wifi,
  Activity,
  FileArchive
} from 'lucide-react';
import { apiClient, SystemStatus, TaskInfo } from '../api/client';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function SystemStatusPage() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [runningTask, setRunningTask] = useState<string | null>(null);

  useEffect(() => {
    loadData();
    // Refresh every 30 seconds
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      const [statusData, tasksData] = await Promise.all([
        apiClient.getSystemStatus(),
        apiClient.getSystemTasks(),
      ]);
      setStatus(statusData);
      setTasks(tasksData);
    } catch (error) {
      console.error('Failed to load system status:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRunTask = async (taskName: string) => {
    setRunningTask(taskName);
    try {
      await apiClient.runSystemTask(taskName);
      // Reload after a delay
      setTimeout(loadData, 2000);
    } catch (error) {
      console.error('Failed to run task:', error);
    } finally {
      setRunningTask(null);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
      </div>
    );
  }

  if (!status) {
    return (
      <div className="text-center py-12">
        <Server className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
        <h2 className="text-xl font-medium text-neutral-300 mb-2">Failed to Load Status</h2>
        <button onClick={loadData} className="text-sky-400 hover:text-sky-300">
          Try Again
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">System Status</h1>
          <p className="text-neutral-400 mt-1">Monitor system health, tasks, and resource usage</p>
        </div>
        <button
          onClick={loadData}
          className="flex items-center gap-2 px-4 py-2 text-neutral-400 hover:text-neutral-200 transition-colors"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Version */}
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="p-2 bg-sky-500/20 rounded-lg">
              <Server className="w-5 h-5 text-sky-400" />
            </div>
            <span className="text-sm text-neutral-400">Version</span>
          </div>
          <p className="text-2xl font-bold text-neutral-100">{status.version}</p>
          <p className="text-sm text-neutral-500 mt-1">{status.os}/{status.arch}</p>
        </div>

        {/* Uptime */}
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="p-2 bg-green-500/20 rounded-lg">
              <Clock className="w-5 h-5 text-green-400" />
            </div>
            <span className="text-sm text-neutral-400">Uptime</span>
          </div>
          <p className="text-2xl font-bold text-neutral-100">{status.uptime}</p>
          <p className="text-sm text-neutral-500 mt-1">
            Started: {new Date(status.startTime).toLocaleString()}
          </p>
        </div>

        {/* Database */}
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="p-2 bg-purple-500/20 rounded-lg">
              <Database className="w-5 h-5 text-purple-400" />
            </div>
            <span className="text-sm text-neutral-400">Database</span>
          </div>
          <p className="text-2xl font-bold text-neutral-100">{formatBytes(status.database.size)}</p>
          <div className="flex items-center gap-2 mt-1">
            <CheckCircle2 className="w-4 h-4 text-green-400" />
            <span className="text-sm text-green-400">{status.database.status}</span>
          </div>
        </div>

        {/* Active Connections */}
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="p-2 bg-amber-500/20 rounded-lg">
              <Wifi className="w-5 h-5 text-amber-400" />
            </div>
            <span className="text-sm text-neutral-400">Connections</span>
          </div>
          <p className="text-2xl font-bold text-neutral-100">{status.clients.webSockets}</p>
          <p className="text-sm text-neutral-500 mt-1">
            {status.clients.indexers} indexers, {status.clients.downloadClients} clients
          </p>
        </div>
      </div>

      {/* Library Stats */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-neutral-100 mb-4 flex items-center gap-2">
          <Activity className="w-5 h-5 text-sky-400" />
          Library Statistics
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <div>
            <p className="text-3xl font-bold text-neutral-100">{status.library.totalBooks}</p>
            <p className="text-sm text-neutral-500">Total Books</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-sky-400">{status.library.monitoredBooks}</p>
            <p className="text-sm text-neutral-500">Monitored</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-neutral-100">{status.library.totalAuthors}</p>
            <p className="text-sm text-neutral-500">Authors</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-neutral-100">{status.library.totalSeries}</p>
            <p className="text-sm text-neutral-500">Series</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-green-400">{status.library.totalMediaFiles}</p>
            <p className="text-sm text-neutral-500">Media Files</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-neutral-100">{formatBytes(status.library.totalSize)}</p>
            <p className="text-sm text-neutral-500">Total Size</p>
          </div>
        </div>
      </div>

      {/* Disk Status */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-neutral-100 mb-4 flex items-center gap-2">
          <HardDrive className="w-5 h-5 text-purple-400" />
          Storage Paths
        </h2>
        <div className="space-y-3">
          {[
            { name: 'Books', path: status.disk.booksPath },
            { name: 'Audiobooks', path: status.disk.audiobooksPath },
            { name: 'Downloads', path: status.disk.downloadsPath },
          ].map(({ name, path }) => (
            <div key={name} className="flex items-center justify-between p-3 bg-neutral-900/50 rounded-lg">
              <div className="flex items-center gap-3">
                <FileArchive className="w-5 h-5 text-neutral-400" />
                <div>
                  <p className="font-medium text-neutral-200">{name}</p>
                  <p className="text-sm text-neutral-500 font-mono">{path.path}</p>
                </div>
              </div>
              <div className="flex items-center gap-4">
                {path.usedBytes !== undefined && (
                  <span className="text-sm text-neutral-400">{formatBytes(path.usedBytes)}</span>
                )}
                <div className="flex items-center gap-2">
                  {path.exists ? (
                    <CheckCircle2 className="w-4 h-4 text-green-400" />
                  ) : (
                    <XCircle className="w-4 h-4 text-red-400" />
                  )}
                  {path.writable ? (
                    <span className="text-xs text-green-400">Writable</span>
                  ) : (
                    <span className="text-xs text-red-400">Read-only</span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Scheduled Tasks */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-neutral-100 mb-4 flex items-center gap-2">
          <Calendar className="w-5 h-5 text-amber-400" />
          Scheduled Tasks
        </h2>
        <div className="space-y-2">
          {tasks.map((task) => (
            <div
              key={task.name}
              className="flex items-center justify-between p-3 bg-neutral-900/50 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-lg ${
                  task.running ? 'bg-sky-500/20' : 
                  task.enabled ? 'bg-green-500/20' : 'bg-neutral-700'
                }`}>
                  {task.running ? (
                    <Loader2 className="w-4 h-4 text-sky-400 animate-spin" />
                  ) : (
                    <Clock className={`w-4 h-4 ${task.enabled ? 'text-green-400' : 'text-neutral-500'}`} />
                  )}
                </div>
                <div>
                  <p className="font-medium text-neutral-200">{task.name}</p>
                  <p className="text-sm text-neutral-500">
                    Interval: {task.interval} • Last: {task.lastStatus}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-4">
                <span className={`text-xs px-2 py-1 rounded ${
                  task.enabled ? 'bg-green-500/20 text-green-400' : 'bg-neutral-700 text-neutral-500'
                }`}>
                  {task.enabled ? 'Enabled' : 'Disabled'}
                </span>
                <button
                  onClick={() => handleRunTask(task.name)}
                  disabled={runningTask === task.name || task.running}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-sky-600 text-white text-sm rounded-lg hover:bg-sky-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {runningTask === task.name ? (
                    <Loader2 className="w-3 h-3 animate-spin" />
                  ) : (
                    <Play className="w-3 h-3" />
                  )}
                  Run Now
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Runtime Info */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-neutral-100 mb-4 flex items-center gap-2">
          <Cpu className="w-5 h-5 text-green-400" />
          Runtime Information
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <p className="text-neutral-500">Go Version</p>
            <p className="font-mono text-neutral-200">{status.goVersion}</p>
          </div>
          <div>
            <p className="text-neutral-500">Operating System</p>
            <p className="font-mono text-neutral-200">{status.os}</p>
          </div>
          <div>
            <p className="text-neutral-500">Architecture</p>
            <p className="font-mono text-neutral-200">{status.arch}</p>
          </div>
          <div>
            <p className="text-neutral-500">Database Type</p>
            <p className="font-mono text-neutral-200">{status.database.type}</p>
          </div>
        </div>
      </div>
    </div>
  );
}

