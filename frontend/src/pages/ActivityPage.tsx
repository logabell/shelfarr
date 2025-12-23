import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  Download, 
  CheckCircle2, 
  XCircle, 
  Clock, 
  Loader2, 
  AlertCircle,
  FileDown,
  BookOpen,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { apiClient, ActivityEvent } from '../api/client';

const ITEMS_PER_PAGE = 25;

export default function ActivityPage() {
  const [activities, setActivities] = useState<ActivityEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    loadActivities();
  }, [page]);

  const loadActivities = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getActivityHistory(page, ITEMS_PER_PAGE);
      setActivities(data.activities);
      setTotal(data.total);
    } catch (error) {
      console.error('Failed to load activities:', error);
    } finally {
      setLoading(false);
    }
  };

  const getEventIcon = (type: string, status: string) => {
    if (status === 'error') return <XCircle className="w-5 h-5 text-red-400" />;
    if (status === 'warning') return <Clock className="w-5 h-5 text-amber-400" />;
    
    switch (type) {
      case 'download':
        return <Download className="w-5 h-5 text-sky-400" />;
      case 'import':
        return <FileDown className="w-5 h-5 text-green-400" />;
      case 'book':
        return <BookOpen className="w-5 h-5 text-purple-400" />;
      default:
        return <AlertCircle className="w-5 h-5 text-neutral-400" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return (
          <span className="flex items-center gap-1 text-xs text-green-400 bg-green-400/10 px-2 py-0.5 rounded">
            <CheckCircle2 className="w-3 h-3" /> Success
          </span>
        );
      case 'error':
        return (
          <span className="flex items-center gap-1 text-xs text-red-400 bg-red-400/10 px-2 py-0.5 rounded">
            <XCircle className="w-3 h-3" /> Failed
          </span>
        );
      case 'warning':
        return (
          <span className="flex items-center gap-1 text-xs text-amber-400 bg-amber-400/10 px-2 py-0.5 rounded">
            <Clock className="w-3 h-3" /> In Progress
          </span>
        );
      default:
        return null;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return date.toLocaleDateString();
  };

  const totalPages = Math.ceil(total / ITEMS_PER_PAGE);

  if (loading && activities.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-neutral-100">Activity</h1>
        <p className="text-neutral-400 mt-1">Recent downloads, imports, and system events</p>
      </div>

      {/* Activity List */}
      {activities.length === 0 ? (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
          <Clock className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
          <h3 className="text-lg font-medium text-neutral-300 mb-2">No Activity Yet</h3>
          <p className="text-neutral-500">Downloads and imports will appear here</p>
        </div>
      ) : (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-neutral-700 text-left">
                <th className="px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider">Event</th>
                <th className="px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider">Details</th>
                <th className="px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider text-right">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-700/50">
              {activities.map((activity) => (
                <tr key={`${activity.type}-${activity.id}`} className="hover:bg-neutral-700/20">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      {getEventIcon(activity.type, activity.status)}
                      <span className="text-neutral-200 font-medium">{activity.title}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="text-sm text-neutral-400">
                      {activity.message}
                      {activity.bookId && activity.bookTitle && (
                        <Link 
                          to={`/books/${activity.bookId}`}
                          className="ml-2 text-sky-400 hover:text-sky-300 transition-colors"
                        >
                          ({activity.bookTitle})
                        </Link>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    {getStatusBadge(activity.status)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <span className="text-sm text-neutral-500">
                      {formatTimestamp(activity.timestamp)}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t border-neutral-700">
              <div className="text-sm text-neutral-500">
                Showing {(page - 1) * ITEMS_PER_PAGE + 1} to {Math.min(page * ITEMS_PER_PAGE, total)} of {total}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="p-1 text-neutral-400 hover:text-neutral-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <span className="text-sm text-neutral-400">
                  Page {page} of {totalPages}
                </span>
                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages}
                  className="p-1 text-neutral-400 hover:text-neutral-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

