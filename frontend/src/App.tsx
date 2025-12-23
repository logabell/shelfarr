import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppLayout } from '@/components/layout/AppLayout'
import { LibraryPage } from '@/pages/LibraryPage'
import { SeriesPage } from '@/pages/SeriesPage'
import { AuthorsPage } from '@/pages/AuthorsPage'
import { SearchPage } from '@/pages/SearchPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { ManualImportPage } from '@/pages/ManualImportPage'
import { IndexersSettingsPage } from '@/pages/IndexersSettingsPage'
import { QualityProfilesSettingsPage } from '@/pages/QualityProfilesSettingsPage'
import { MediaManagementSettingsPage } from '@/pages/MediaManagementSettingsPage'
import { BookDetailPage } from '@/pages/BookDetailPage'
import ActivityPage from '@/pages/ActivityPage'
import WantedPage from '@/pages/WantedPage'
import AuthorDetailPage from '@/pages/AuthorDetailPage'
import SeriesDetailPage from '@/pages/SeriesDetailPage'
import HardcoverAuthorPage from '@/pages/HardcoverAuthorPage'
import HardcoverSeriesPage from '@/pages/HardcoverSeriesPage'
import HardcoverBookPage from '@/pages/HardcoverBookPage'
import SystemStatusPage from '@/pages/SystemStatusPage'
import DownloadClientsSettingsPage from '@/pages/DownloadClientsSettingsPage'
import NotificationsSettingsPage from '@/pages/NotificationsSettingsPage'
import ListsSettingsPage from '@/pages/ListsSettingsPage'
import LibrarySearchSettingsPage from '@/pages/LibrarySearchSettingsPage'
import GeneralSettingsPage from '@/pages/GeneralSettingsPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<AppLayout />}>
            <Route index element={<LibraryPage />} />
            <Route path="series" element={<SeriesPage />} />
            <Route path="authors" element={<AuthorsPage />} />
            <Route path="search" element={<SearchPage />} />
            <Route path="add" element={<Navigate to="/search" replace />} />
            <Route path="activity" element={<ActivityPage />} />
            <Route path="wanted" element={<WantedPage />} />
            <Route path="import" element={<ManualImportPage />} />
            <Route path="settings" element={<SettingsPage />} />
            <Route path="settings/general" element={<GeneralSettingsPage />} />
            <Route path="settings/library-search" element={<LibrarySearchSettingsPage />} />
            <Route path="settings/indexers" element={<IndexersSettingsPage />} />
            <Route path="settings/download-clients" element={<DownloadClientsSettingsPage />} />
            <Route path="settings/notifications" element={<NotificationsSettingsPage />} />
            <Route path="settings/lists" element={<ListsSettingsPage />} />
            <Route path="settings/profiles" element={<QualityProfilesSettingsPage />} />
            <Route path="settings/media" element={<MediaManagementSettingsPage />} />
            <Route path="settings/*" element={<ComingSoon title="Settings" />} />
            <Route path="books/:id" element={<BookDetailPage />} />
            <Route path="authors/:id" element={<AuthorDetailPage />} />
            <Route path="series/:id" element={<SeriesDetailPage />} />
            <Route path="hardcover/author/:id" element={<HardcoverAuthorPage />} />
            <Route path="hardcover/series/:id" element={<HardcoverSeriesPage />} />
            <Route path="hardcover/book/:id" element={<HardcoverBookPage />} />
            <Route path="system/status" element={<SystemStatusPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

// Placeholder component for routes not yet implemented
function ComingSoon({ title }: { title: string }) {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold">{title}</h1>
        <p className="text-muted-foreground mt-2">Coming soon...</p>
      </div>
    </div>
  )
}

export default App
