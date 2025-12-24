import { useState, useMemo } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  User, 
  Book as BookIcon, 
  Eye, 
  EyeOff, 
  Loader2, 
  ArrowLeft,
  Plus,
  CheckCircle2,
  Library,
  Trash2
} from 'lucide-react';
import { getAuthor, updateAuthor, addHardcoverBook, addOpenLibraryBook, deleteBook, deleteAuthor } from '@/api/client';
import { Button } from '@/components/ui/button';
import { DeleteConfirmDialog } from '@/components/ui/delete-confirm-dialog';
import { CatalogBookCard } from '@/components/library/CatalogBookCard';
import { 
  BookSortFilter, 
  sortBooks, 
  filterBooks, 
  getDefaultSortFilterState,
  type SortFilterState 
} from '@/components/library/BookSortFilter';

export default function AuthorDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [addingBooks, setAddingBooks] = useState<Set<string>>(new Set());
  const [deletingBooks, setDeletingBooks] = useState<Set<number>>(new Set());
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [sortFilterState, setSortFilterState] = useState<SortFilterState>(
    getDefaultSortFilterState(false)
  );

  const { data: author, isLoading, error } = useQuery({
    queryKey: ['author', id],
    queryFn: () => getAuthor(parseInt(id!)),
    enabled: !!id,
  });

  const updateAuthorMutation = useMutation({
    mutationFn: ({ authorId, monitored }: { authorId: number; monitored: boolean }) => 
      updateAuthor(authorId, { monitored }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['author', id] });
      queryClient.invalidateQueries({ queryKey: ['authors'] });
    },
  });

  const deleteAuthorMutation = useMutation({
    mutationFn: ({ authorId, deleteFiles }: { authorId: number; deleteFiles: boolean }) => 
      deleteAuthor(authorId, deleteFiles),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['authors'] });
      queryClient.invalidateQueries({ queryKey: ['library'] });
      navigate('/authors');
    },
    onError: (error: Error) => {
      console.error('Failed to delete author:', error.message);
    }
  });

  const addHardcoverBookMutation = useMutation({
    mutationFn: (hardcoverId: string) => addHardcoverBook(hardcoverId, { monitored: true }),
    onSuccess: (_, hardcoverId) => {
      setAddingBooks(prev => {
        const next = new Set(prev);
        next.delete(hardcoverId);
        return next;
      });
      queryClient.invalidateQueries({ queryKey: ['author', id] });
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

  const addOpenLibraryBookMutation = useMutation({
    mutationFn: (openLibraryWorkId: string) => addOpenLibraryBook(openLibraryWorkId, true),
    onSuccess: (_, workId) => {
      setAddingBooks(prev => {
        const next = new Set(prev);
        next.delete(workId);
        return next;
      });
      queryClient.invalidateQueries({ queryKey: ['author', id] });
      queryClient.invalidateQueries({ queryKey: ['library'] });
    },
    onError: (error: Error, workId) => {
      setAddingBooks(prev => {
        const next = new Set(prev);
        next.delete(workId);
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
      queryClient.invalidateQueries({ queryKey: ['author', id] });
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

  const handleAddBook = (bookId: string, source: 'hardcover' | 'openlibrary') => {
    setAddingBooks(prev => new Set(prev).add(bookId));
    if (source === 'hardcover') {
      addHardcoverBookMutation.mutate(bookId);
    } else {
      addOpenLibraryBookMutation.mutate(bookId);
    }
  };

  const handleDeleteBook = (bookId: number) => {
    setDeletingBooks(prev => new Set(prev).add(bookId));
    deleteBookMutation.mutate(bookId);
  };

  const handleDeleteAuthor = (deleteFiles: boolean) => {
    if (author) {
      deleteAuthorMutation.mutate({ authorId: author.id, deleteFiles });
    }
  };

  const handleAddAllMissing = () => {
    if (!author?.books) return;
    const missingBooks = author.books.filter(b => !b.inLibrary && (b.hardcoverId || b.openLibraryWorkId));
    missingBooks.forEach(book => {
      if (book.openLibraryWorkId) {
        handleAddBook(book.openLibraryWorkId, 'openlibrary');
      } else if (book.hardcoverId) {
        handleAddBook(book.hardcoverId, 'hardcover');
      }
    });
  };

  const handleToggleMonitored = () => {
    if (!author) return;
    updateAuthorMutation.mutate({ authorId: author.id, monitored: !author.monitored });
  };

  // Must compute derived data before early returns to satisfy React hooks rules
  const books = useMemo(() => author?.books || [], [author?.books]);
  const totalBooks = author?.totalBooks || books.length;
  const inLibraryCount = author?.inLibrary || books.filter(b => b.inLibrary).length;
  const downloadedCount = author?.downloadedCount || 0;
  const missingFromLibrary = books.filter(b => !b.inLibrary).length;
  const progressPercent = totalBooks > 0 ? (downloadedCount / totalBooks) * 100 : 0;

  // Group books by series
  const seriesGroups = useMemo(() => {
    const seriesMap = new Map<string, { 
      id: string
      name: string
      books: typeof books 
    }>();

    books.forEach(book => {
      if (book.seriesId && book.seriesName) {
        const existing = seriesMap.get(book.seriesId);
        if (existing) {
          existing.books.push(book);
        } else {
          seriesMap.set(book.seriesId, {
            id: book.seriesId,
            name: book.seriesName,
            books: [book],
          });
        }
      }
    });

    // Sort books within each series by series index
    seriesMap.forEach(series => {
      series.books.sort((a, b) => (a.seriesIndex || 0) - (b.seriesIndex || 0));
    });

    return Array.from(seriesMap.values())
      .filter(series => series.books.length > 1)
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [books]);

  // Apply sorting and filtering
  const processedBooks = useMemo(() => {
    // Map books to have consistent field names for sorting/filtering
    const mappedBooks = books.map(b => ({
      ...b,
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

  if (error || !author) {
    return (
      <div className="text-center py-12">
        <User className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
        <h2 className="text-xl font-medium text-neutral-300 mb-2">Author Not Found</h2>
        <Link to="/authors" className="text-sky-400 hover:text-sky-300">
          Back to Authors
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <DeleteConfirmDialog 
        isOpen={isDeleteDialogOpen}
        onClose={() => setIsDeleteDialogOpen(false)}
        onConfirm={handleDeleteAuthor}
        title={`Delete ${author.name}?`}
        description="This will remove the author and all their books from your library."
        isDeleting={deleteAuthorMutation.isPending}
      />

      {/* Back Link */}
      <Link
        to="/authors"
        className="inline-flex items-center gap-2 text-neutral-400 hover:text-neutral-200 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Authors
      </Link>

      {/* Author Header */}
      <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl overflow-hidden">
        <div className="flex items-start gap-6 p-6">
          {/* Author Image */}
          <div className="flex-shrink-0">
            {author.imageUrl ? (
              <img
                src={author.imageUrl}
                alt={author.name}
                className="w-32 h-32 object-cover rounded-xl"
              />
            ) : (
              <div className="w-32 h-32 bg-gradient-to-br from-sky-900 to-sky-700 rounded-xl flex items-center justify-center">
                <User className="w-16 h-16 text-sky-400" />
              </div>
            )}
          </div>

          {/* Author Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-bold text-neutral-100">{author.name}</h1>
                {author.sortName && author.sortName !== author.name && (
                  <p className="text-neutral-500 text-sm mt-1">Sort: {author.sortName}</p>
                )}
              </div>

              {/* Actions */}
              <div className="flex items-center gap-2">
                {/* Monitor Toggle */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleMonitored}
                  disabled={updateAuthorMutation.isPending}
                  className={author.monitored ? 'bg-sky-600/20 border-sky-600 text-sky-400' : ''}
                >
                  {updateAuthorMutation.isPending ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : author.monitored ? (
                    <Eye className="w-4 h-4 mr-2" />
                  ) : (
                    <EyeOff className="w-4 h-4 mr-2" />
                  )}
                  {author.monitored ? 'Monitored' : 'Not Monitored'}
                </Button>

                {/* Add All Missing */}
                {missingFromLibrary > 0 && (
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
                )}

                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => setIsDeleteDialogOpen(true)}
                  disabled={deleteAuthorMutation.isPending}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </Button>
              </div>
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

            {/* Biography */}
            {author.biography && (
              <p className="text-neutral-400 mt-4 line-clamp-3">{author.biography}</p>
            )}
          </div>
        </div>
      </div>

      {/* Series Section */}
      {seriesGroups.length > 0 && (
        <div>
          <h2 className="text-xl font-semibold text-neutral-100 mb-4">Series ({seriesGroups.length})</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {seriesGroups.map(series => {
              const inLibraryInSeries = series.books.filter(b => b.inLibrary).length;
              const downloadedInSeries = series.books.filter(b => b.book?.status === 'downloaded').length;
              const librarySeriesId =
                series.books.find(b => b.inLibrary && b.book?.series?.id != null)?.book?.series?.id ?? null;
              const target =
                librarySeriesId != null ? `/series/${librarySeriesId}` : `/hardcover/series/${series.id}`;
              
              return (
                <Link
                  key={series.id}
                  to={target}
                  className="flex items-center gap-4 p-4 rounded-lg bg-neutral-800/50 border border-neutral-700 hover:border-sky-500/50 transition-colors"
                >
                  <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-amber-900/50 to-amber-700/50 flex items-center justify-center shrink-0">
                    <Library className="w-6 h-6 text-amber-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-medium text-neutral-200 truncate">{series.name}</h3>
                    <div className="flex items-center gap-3 text-xs mt-1">
                      <span className="text-neutral-400">{series.books.length} books</span>
                      {inLibraryInSeries > 0 && (
                        <span className="text-sky-400">{inLibraryInSeries} in library</span>
                      )}
                      {downloadedInSeries > 0 && (
                        <span className="text-green-400">{downloadedInSeries} downloaded</span>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </div>
      )}

      {/* Books Section */}
      <div>
        <h2 className="text-xl font-semibold text-neutral-100 mb-4">Books</h2>

        {books.length === 0 ? (
          <div className="bg-neutral-800/50 border border-neutral-700 rounded-xl p-12 text-center">
            <BookIcon className="w-16 h-16 mx-auto text-neutral-600 mb-4" />
            <h3 className="text-lg font-medium text-neutral-300 mb-2">No Books Found</h3>
            <p className="text-neutral-500">
              No books found for this author in Open Library.
            </p>
          </div>
        ) : (
          <>
            {/* Sort/Filter Toolbar */}
            <BookSortFilter
              state={sortFilterState}
              onChange={setSortFilterState}
              showSeriesIndex={false}
              totalCount={books.length}
              filteredCount={processedBooks.length}
              className="mb-4"
            />
            
            {processedBooks.length > 0 ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
                {processedBooks.map((bookEntry, index) => (
                  <CatalogBookCard
                    key={bookEntry.hardcoverId || bookEntry.openLibraryWorkId || `book-${index}`}
                    hardcoverId={bookEntry.hardcoverId}
                    openLibraryWorkId={bookEntry.openLibraryWorkId}
                    title={bookEntry.title}
                    coverUrl={bookEntry.coverUrl || bookEntry.book?.coverUrl}
                    rating={bookEntry.rating || bookEntry.book?.rating}
                    releaseYear={bookEntry.releaseYear}
                    compilation={bookEntry.compilation}
                    seriesName={bookEntry.seriesName}
                    seriesIndex={bookEntry.seriesIndex}
                    authorName={bookEntry.authorName || author.name}
                    inLibrary={bookEntry.inLibrary}
                    libraryBook={bookEntry.book}
                    onAdd={handleAddBook}
                    onDelete={handleDeleteBook}
                    isDeleting={bookEntry.book?.id ? deletingBooks.has(bookEntry.book.id) : false}
                    isAdding={addingBooks.has(bookEntry.hardcoverId || bookEntry.openLibraryWorkId || '')}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-12 border border-dashed border-neutral-700 rounded-lg">
                <BookIcon className="w-12 h-12 mx-auto text-neutral-600 mb-4" />
                <p className="text-neutral-400">No books match your filter</p>
                <button 
                  onClick={() => setSortFilterState(getDefaultSortFilterState(false))}
                  className="text-sky-400 hover:text-sky-300 text-sm mt-2"
                >
                  Clear filters
                </button>
              </div>
            )}
          </>
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
    </div>
  );
}
