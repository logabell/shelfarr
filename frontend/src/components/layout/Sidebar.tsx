import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { 
  BookOpen, 
  Users, 
  Layers, 
  Search, 
  Settings, 
  Download,
  Import,
  Activity,
  Bell,
  Library
} from 'lucide-react'

const navigation = [
  { name: 'Library', href: '/', icon: Library },
  { name: 'Series', href: '/series', icon: Layers },
  { name: 'Authors', href: '/authors', icon: Users },
  { name: 'Search & Add', href: '/search', icon: Search },
  { name: 'Activity', href: '/activity', icon: Activity },
  { name: 'Wanted', href: '/wanted', icon: Download },
  { name: 'Manual Import', href: '/import', icon: Import },
]

const settingsNav = [
  { name: 'Settings', href: '/settings', icon: Settings },
  { name: 'Notifications', href: '/settings/notifications', icon: Bell },
]

export function Sidebar() {
  const location = useLocation()

  return (
    <div className="flex h-full w-64 flex-col bg-card border-r border-border">
      {/* Logo */}
      <div className="flex h-16 items-center px-6 border-b border-border">
        <BookOpen className="h-8 w-8 text-primary" />
        <span className="ml-3 text-xl font-bold tracking-tight">Shelfarr</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'group flex items-center rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )}
            >
              <item.icon
                className={cn(
                  'mr-3 h-5 w-5 flex-shrink-0',
                  isActive ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-accent-foreground'
                )}
              />
              {item.name}
            </Link>
          )
        })}

        <div className="my-4 border-t border-border" />

        {settingsNav.map((item) => {
          const isActive = location.pathname === item.href
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'group flex items-center rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )}
            >
              <item.icon
                className={cn(
                  'mr-3 h-5 w-5 flex-shrink-0',
                  isActive ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-accent-foreground'
                )}
              />
              {item.name}
            </Link>
          )
        })}
      </nav>

      {/* Version */}
      <div className="border-t border-border p-4">
        <p className="text-xs text-muted-foreground">
          Shelfarr v1.0.0
        </p>
      </div>
    </div>
  )
}

