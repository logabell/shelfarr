import { useState, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Library, 
  Book as BookIcon, 
  CheckCircle2, 
  ArrowLeft,
  Plus,
  Loader2,
  ListPlus
} from 'lucide-react';
import { getSeriesDetail, addHardcoverBook, deleteBook } from '@/api/client';
import { Button } from '@/components/ui/button';
import { CatalogBookCard } from '@/components/library/CatalogBookCard';
import { AddSeriesModal } from '@/components/series/AddSeriesModal';
import { 
  BookSortFilter, 
  sortBooks, 
  filterBooks, 
  getDefaultSortFilterState,
  type SortFilterState 
} from '@/components/library/BookSortFilter';

export default function SeriesDetailPage() {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [addingBooks, setAddingBooks] = useState<Set<string>>(new Set());
  const [deletingBooks, setDeletingBooks] = useState<Set<number>>(new Set());
  const [isAddSeriesModalOpen, setIsAddSeriesModalOpen] = useState(false);
  const [sortFilterState, setSortFilterState] = useState<SortFilterState>(
    getDefaultSortFilterState(true)
  );

  const { data: series, isLoading, error } = useQuery({
    queryKey: ['series', id],
    queryFn: () => getSeriesDetail(parseInt(id!)),
    enabled: !!id,
  });

  const addBookMutation = useMutation({
    mutationFn: (hardcoverId: string) => addHardcoverBook(hardcoverId, { monitored: true }),
    onSuccess: (_, hardcoverId) => {
      setAddingBooks(prev => {
        const next = new Set(prev);
        next.delete(hardcoverId);
        return next;
      });
      queryClient.invalidateQueries({ queryKey: ['series', id] });
    },
    onError: (error: Error, hardcoverId) => {
      setAddingBooks(prev => {
        const next = new Set(prev);
        next.delete(hardcoverId);
        return next;
      });
      console.error('Failed to add book:', error.message);
    },
  });

  const deleteBookMutation = useMutation({
    mutationFn: (bookId: number) => deleteBook(bookId),
    onSuccess: (_, bookId) => {
      setDeletingBooks(prev => {
        const next = new Set(prev);
        next.delete(bookId);
        return next;
      });
      queryClient.invalidateQueries({ queryKey: ['series', id] });
      queryClient.invalidateQueries({ queryKey: ['library'] });
    },
    onError: (error: Error, bookId) => {
      setDeletingBooks(prev => {
        const next = new Set(prev);
        next.delete(bookId);
        return next;
      });
      console.error('Failed to delete book:', error.message);
    },
  });

  const handleAddBook = (hardcoverId: string) => {
    setAddingBooks(prev => new Set(prev).add(hardcoverId));
    addBookMutation.mutate(hardcoverId);
  };

  const handleDeleteBook = (bookId: number) => {
    setDeletingBooks(prev => new Set(prev).add(bookId));
    deleteBookMutation.mutate(bookId);
  };

  const handleAddAllMissing = () => {
    if (!series?.books) return;
    const missingBooks = series.books.filter(b => !b.inLibrary && b.hardcoverId);
    missingBooks.forEach(book => {
      if (book.hardcoverId) {
        handleAddBook(book.hardcoverId);
      }
    });
  };

  // Must compute derived data before early returns to satisfy React hooks rules
  const books = useMemo(() => series?.books || [], [series?.books]);
  const totalBooks = series?.totalBooks || books.length;
  const inLibraryCount = series?.inLibrary || books.filter(b => b.inLibrary).length;
  const downloadedCount = series?.downloadedCount || 0;
  const missingFromLibrary = books.filter(b => !b.inLibrary).length;
  const progressPercent = totalBooks > 0 ? (downloadedCount / totalBooks) * 100 : 0;

  // Apply sorting and filtering
  const processedBooks = useMemo(() => {
    // Map books to have consistent field names for sorting/filtering
    const mappedBooks = books.map(b => ({
      ...b,
      seriesIndex: b.index,
      rating: b.rating || b.book?.rating || 0,
      releaseYear: b.releaseYear || (b.book?.releaseDate ? new Date(b.book.releaseDate).getFullYear() : undefined),
      libraryBook: b.book,
    }));
    
    const filtered = filterBooks(mappedBooks, sortFilterState.filterStatus, sortFilterState.hideCompilations);
    return sortBooks(filtered, sortFilterState.sortField, sortFilterState.sortOrder);
  }, [books, sortFilterState]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-sky-500" />
      </div>
    );
  }

  if (error || !series) {
    return (
      <div className="text-center py-12">
        <Library className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
        <h2 className="text-xl font-medium text-neutral-300 mb-2">Series Not Found</h2>
        <Link to="/series" className="text-sky-400 hover:text-sky-300">
          Back to Series
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      {/* Back Link */}
      <Link
        to="/series"
        className="inline-flex items-center gap-2 text-neutral-400 hover:text-neutral-200 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Series
      </Link>

      {/* Series Header */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-6">
        <div className="flex items-start gap-6">
          {/* Series Art (collage of book covers) */}
          <div className="flex-shrink-0 w-32 h-40 bg-gradient-to-br from-amber-900 to-amber-700 rounded-xl flex items-center justify-center relative overflow-hidden">
            {books.slice(0, 4).some(b => b.coverUrl || b.book?.coverUrl) ? (
              <div className="grid grid-cols-2 gap-0.5 w-full h-full">
                {books.slice(0, 4).map((book, idx) => {
                  const coverUrl = book.coverUrl || book.book?.coverUrl;
                  return coverUrl ? (
                    <img
                      key={idx}
                      src={coverUrl}
                      alt=""
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div key={idx} className="bg-neutral-700" />
                  );
                })}
              </div>
            ) : (
              <Library className="w-16 h-16 text-amber-400" />
            )}
          </div>

          {/* Series Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between">
              <h1 className="text-2xl font-bold text-neutral-100">{series.name}</h1>
              {missingFromLibrary > 0 && (
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    onClick={() => setIsAddSeriesModalOpen(true)}
                    disabled={addingBooks.size > 0}
                    size="sm"
                  >
                    <ListPlus className="w-4 h-4 mr-2" />
                    Add Selected
                  </Button>
                  <Button
                    onClick={handleAddAllMissing}
                    disabled={addingBooks.size > 0}
                    size="sm"
                  >
                    {addingBooks.size > 0 ? (
                      <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    ) : (
                      <Plus className="w-4 h-4 mr-2" />
                    )}
                    Add All Missing ({missingFromLibrary})
                  </Button>
                </div>
              )}
            </div>
            
            {/* Stats */}
            <div className="flex items-center gap-6 mt-4">
              <div className="flex items-center gap-2 text-neutral-400">
                <BookIcon className="w-4 h-4" />
                <span>{totalBooks} Books</span>
              </div>
              <div className="flex items-center gap-2 text-sky-400">
                <Library className="w-4 h-4" />
                <span>{inLibraryCount} In Library</span>
              </div>
              <div className="flex items-center gap-2 text-green-400">
                <CheckCircle2 className="w-4 h-4" />
                <span>{downloadedCount} Downloaded</span>
              </div>
            </div>

            {/* Progress Bar */}
            <div className="mt-4">
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-neutral-400">Collection Progress</span>
                <span className="text-neutral-300">
                  {Math.round(progressPercent)}%
                </span>
              </div>
              <div className="h-2 bg-neutral-700 rounded-full overflow-hidden">
                <div 
                  className="h-full bg-gradient-to-r from-green-600 to-green-400 rounded-full transition-all"
                  style={{ width: `${progressPercent}%` }}
                />
              </div>
            </div>

            {/* Description */}
            {series.description && (
              <p className="text-neutral-400 mt-4 line-clamp-3">{series.description}</p>
            )}
          </div>
        </div>
      </div>

      {/* Books Section */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-neutral-100">Books in Series</h2>
        </div>
        
        {/* Sort/Filter Toolbar */}
        <BookSortFilter
          state={sortFilterState}
          onChange={setSortFilterState}
          showSeriesIndex={true}
          totalCount={books.length}
          filteredCount={processedBooks.length}
          className="mb-4"
        />
        
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {processedBooks.map((bookEntry, index) => (
            <CatalogBookCard
              key={bookEntry.hardcoverId || `book-${index}`}
              hardcoverId={bookEntry.hardcoverId || ''}
              title={bookEntry.title}
              coverUrl={bookEntry.coverUrl || bookEntry.book?.coverUrl}
              rating={bookEntry.rating || bookEntry.book?.rating}
              releaseYear={bookEntry.releaseYear}
              compilation={bookEntry.compilation}
              seriesIndex={bookEntry.index}
              authorName={bookEntry.authorName}
              inLibrary={bookEntry.inLibrary}
              libraryBook={bookEntry.book}
              onAdd={handleAddBook}
              isAdding={bookEntry.hardcoverId ? addingBooks.has(bookEntry.hardcoverId) : false}
              onDelete={handleDeleteBook}
              isDeleting={bookEntry.book?.id ? deletingBooks.has(bookEntry.book.id) : false}
            />
          ))}
        </div>
        
        {processedBooks.length === 0 && books.length > 0 && (
          <div className="text-center py-12 border border-dashed border-neutral-700 rounded-lg">
            <BookIcon className="w-12 h-12 mx-auto text-neutral-600 mb-4" />
            <p className="text-neutral-400">No books match your filter</p>
            <button 
              onClick={() => setSortFilterState(getDefaultSortFilterState(true))}
              className="text-sky-400 hover:text-sky-300 text-sm mt-2"
            >
              Clear filters
            </button>
          </div>
        )}
      </div>

      {/* Legend */}
      <div className="flex items-center gap-6 text-sm text-neutral-400 border-t border-neutral-700 pt-4">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-green-500" />
          <span>Downloaded</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-red-500" />
          <span>Missing File</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-sky-500" />
          <span>Downloading</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-purple-500" />
          <span>Unreleased</span>
        </div>
      </div>

      <AddSeriesModal
        isOpen={isAddSeriesModalOpen}
        onClose={() => setIsAddSeriesModalOpen(false)}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['series', id] });
        }}
        seriesId={parseInt(id!)}
        seriesName={series.name}
        books={books}
      />
    </div>
  );
}
