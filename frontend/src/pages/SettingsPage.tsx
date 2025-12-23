import { Topbar } from '@/components/layout/Topbar'
import { Settings, Database, Download, Bell, Users, Palette, BookSearch, SlidersHorizontal } from 'lucide-react'
import { Link } from 'react-router-dom'

const settingsSections = [
  {
    title: 'General',
    description: 'Basic application settings',
    icon: Settings,
    href: '/settings/general',
  },
  {
    title: 'Library Search Providers',
    description: 'Configure book metadata sources (Hardcover.app)',
    icon: BookSearch,
    href: '/settings/library-search',
  },
  {
    title: 'Media Management',
    description: 'Configure paths, naming, and import settings',
    icon: Database,
    href: '/settings/media',
  },
  {
    title: 'Quality Profiles',
    description: 'Configure format preferences for ebooks and audiobooks',
    icon: SlidersHorizontal,
    href: '/settings/profiles',
  },
  {
    title: 'Indexers',
    description: 'Configure search providers (MAM, Torznab, Anna)',
    icon: Download,
    href: '/settings/indexers',
  },
  {
    title: 'Download Clients',
    description: 'Configure torrent and usenet clients',
    icon: Download,
    href: '/settings/download-clients',
  },
  {
    title: 'Notifications',
    description: 'Configure webhooks and notifications',
    icon: Bell,
    href: '/settings/notifications',
  },
  {
    title: 'Users',
    description: 'Manage users and permissions',
    icon: Users,
    href: '/settings/users',
  },
  {
    title: 'UI',
    description: 'Customize appearance and theme',
    icon: Palette,
    href: '/settings/ui',
  },
]

export function SettingsPage() {
  return (
    <div className="flex flex-col h-full">
      <Topbar title="Settings" />
      
      <div className="flex-1 overflow-auto p-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-5xl">
          {settingsSections.map((section) => (
            <Link
              key={section.title}
              to={section.href}
              className="flex items-start gap-4 p-4 rounded-lg bg-card border border-border hover:border-primary transition-colors"
            >
              <div className="rounded-md bg-primary/10 p-2">
                <section.icon className="h-5 w-5 text-primary" />
              </div>
              <div>
                <h3 className="font-medium">{section.title}</h3>
                <p className="text-sm text-muted-foreground mt-1">
                  {section.description}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}

