import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`
}

export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    downloaded: 'status-downloaded',
    unmonitored: 'status-unmonitored',
    downloading: 'status-downloading',
    missing: 'status-missing',
    unreleased: 'status-unreleased',
  }
  return colors[status] || 'status-black'
}

export function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    downloaded: 'Downloaded',
    unmonitored: 'Unmonitored',
    downloading: 'Downloading',
    missing: 'Missing',
    unreleased: 'Unreleased',
  }
  return labels[status] || 'Unknown'
}

