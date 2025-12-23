import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  Search, 
  BookX, 
  Loader2, 
  Filter,
  ChevronLeft,
  ChevronRight,
  ArrowUpCircle,
  Download,
  Book,
  Headphones
} from 'lucide-react';
import { apiClient, WantedBook } from '../api/client';

type TabType = 'missing' | 'cutoff';

const ITEMS_PER_PAGE = 25;

export default function WantedPage() {
  const [activeTab, setActiveTab] = useState<TabType>('missing');
  const [books, setBooks] = useState<WantedBook[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [mediaTypeFilter, setMediaTypeFilter] = useState<'all' | 'ebook' | 'audiobook'>('all');
  const [searchingBook, setSearchingBook] = useState<number | null>(null);

  useEffect(() => {
    loadBooks();
  }, [activeTab, page, mediaTypeFilter]);

  const loadBooks = async () => {
    setLoading(true);
    try {
      let data;
      if (activeTab === 'missing') {
        data = await apiClient.getWantedMissing(
          page, 
          ITEMS_PER_PAGE,
          mediaTypeFilter === 'all' ? undefined : mediaTypeFilter
        );
      } else {
        data = await apiClient.getWantedCutoff(page, ITEMS_PER_PAGE);
      }
      setBooks(data.books);
      setTotal(data.total);
    } catch (error) {
      console.error('Failed to load wanted books:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAutomaticSearch = async (bookId: number, mediaType?: string) => {
    setSearchingBook(bookId);
    try {
      await apiClient.automaticSearch(bookId, mediaType);
      // Reload to reflect changes
      loadBooks();
    } catch (error) {
      console.error('Failed to start automatic search:', error);
    } finally {
      setSearchingBook(null);
    }
  };

  const totalPages = Math.ceil(total / ITEMS_PER_PAGE);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">Wanted</h1>
          <p className="text-neutral-400 mt-1">Books missing media files or needing quality upgrades</p>
        </div>
        <button
          onClick={() => {
            // Search all visible books
            books.forEach((book) => handleAutomaticSearch(book.id));
          }}
          className="flex items-center gap-2 px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-500 transition-colors"
        >
          <Search className="w-4 h-4" />
          Search All
        </button>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-4 border-b border-neutral-700">
        <button
          onClick={() => { setActiveTab('missing'); setPage(1); }}
          className={`flex items-center gap-2 px-4 py-3 border-b-2 transition-colors ${
            activeTab === 'missing'
              ? 'border-sky-500 text-sky-400'
              : 'border-transparent text-neutral-400 hover:text-neutral-200'
          }`}
        >
          <BookX className="w-4 h-4" />
          Missing
        </button>
        <button
          onClick={() => { setActiveTab('cutoff'); setPage(1); }}
          className={`flex items-center gap-2 px-4 py-3 border-b-2 transition-colors ${
            activeTab === 'cutoff'
              ? 'border-sky-500 text-sky-400'
              : 'border-transparent text-neutral-400 hover:text-neutral-200'
          }`}
        >
          <ArrowUpCircle className="w-4 h-4" />
          Cutoff Unmet
        </button>
      </div>

      {/* Filters */}
      {activeTab === 'missing' && (
        <div className="flex items-center gap-4">
          <Filter className="w-4 h-4 text-neutral-400" />
          <select
            value={mediaTypeFilter}
            onChange={(e) => { setMediaTypeFilter(e.target.value as typeof mediaTypeFilter); setPage(1); }}
            className="px-3 py-1.5 bg-neutral-800 border border-neutral-700 rounded-lg text-sm text-neutral-200 focus:outline-none focus:border-sky-500"
          >
            <option value="all">All Media Types</option>
            <option value="ebook">Ebooks Only</option>
            <option value="audiobook">Audiobooks Only</option>
          </select>
        </div>
      )}

      {/* Loading */}
      {loading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
        </div>
      ) : books.length === 0 ? (
        <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
          <BookX className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
          <h3 className="text-lg font-medium text-neutral-300 mb-2">
            {activeTab === 'missing' ? 'No Missing Books' : 'No Cutoff Unmet'}
          </h3>
          <p className="text-neutral-500">
            {activeTab === 'missing'
              ? 'All monitored books have media files'
              : 'All books meet their quality profile cutoff'}
          </p>
        </div>
      ) : (
        <>
          {/* Books List */}
          <div className="grid gap-3">
            {books.map((book) => (
              <div
                key={book.id}
                className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-4 hover:border-neutral-600 transition-colors"
              >
                <div className="flex items-center gap-4">
                  {/* Cover */}
                  <Link to={`/books/${book.id}`} className="flex-shrink-0">
                    {book.coverUrl ? (
                      <img
                        src={book.coverUrl}
                        alt={book.title}
                        className="w-16 h-24 object-cover rounded-lg"
                      />
                    ) : (
                      <div className="w-16 h-24 bg-neutral-700 rounded-lg flex items-center justify-center">
                        <BookX className="w-8 h-8 text-neutral-500" />
                      </div>
                    )}
                  </Link>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <Link
                      to={`/books/${book.id}`}
                      className="text-lg font-medium text-neutral-100 hover:text-sky-400 transition-colors line-clamp-1"
                    >
                      {book.title}
                    </Link>
                    <p className="text-sm text-neutral-400">{book.authorName}</p>
                    {book.seriesName && (
                      <p className="text-xs text-neutral-500">
                        {book.seriesName} #{book.seriesIndex}
                      </p>
                    )}
                    
                    {/* Media Status */}
                    <div className="flex items-center gap-3 mt-2">
                      <span className={`flex items-center gap-1 text-xs ${
                        book.hasEbook ? 'text-green-400' : 'text-red-400'
                      }`}>
                        <Book className="w-3 h-3" />
                        {book.hasEbook ? 'Has Ebook' : 'Missing Ebook'}
                      </span>
                      <span className={`flex items-center gap-1 text-xs ${
                        book.hasAudiobook ? 'text-green-400' : 'text-red-400'
                      }`}>
                        <Headphones className="w-3 h-3" />
                        {book.hasAudiobook ? 'Has Audiobook' : 'Missing Audiobook'}
                      </span>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 flex-shrink-0">
                    {!book.hasEbook && (
                      <button
                        onClick={() => handleAutomaticSearch(book.id, 'ebook')}
                        disabled={searchingBook === book.id}
                        className="flex items-center gap-1.5 px-3 py-1.5 bg-neutral-700 text-neutral-200 text-sm rounded-lg hover:bg-neutral-600 disabled:opacity-50 transition-colors"
                      >
                        {searchingBook === book.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Book className="w-4 h-4" />
                        )}
                        Search Ebook
                      </button>
                    )}
                    {!book.hasAudiobook && (
                      <button
                        onClick={() => handleAutomaticSearch(book.id, 'audiobook')}
                        disabled={searchingBook === book.id}
                        className="flex items-center gap-1.5 px-3 py-1.5 bg-neutral-700 text-neutral-200 text-sm rounded-lg hover:bg-neutral-600 disabled:opacity-50 transition-colors"
                      >
                        {searchingBook === book.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Headphones className="w-4 h-4" />
                        )}
                        Search Audiobook
                      </button>
                    )}
                    <Link
                      to={`/books/${book.id}`}
                      className="flex items-center gap-1.5 px-3 py-1.5 bg-sky-600 text-white text-sm rounded-lg hover:bg-sky-500 transition-colors"
                    >
                      <Download className="w-4 h-4" />
                      Manual
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <div className="text-sm text-neutral-500">
                Showing {(page - 1) * ITEMS_PER_PAGE + 1} to {Math.min(page * ITEMS_PER_PAGE, total)} of {total} books
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
        </>
      )}
    </div>
  );
}

