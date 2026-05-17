'use client'

import { ArrowLeftRight, Briefcase, Calendar, FileText, LayoutDashboard, LogOut, MessageCircle, Settings, Upload, Users } from 'lucide-react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'

import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

export function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()
  const { sidebarOpen, toggleSidebar } = useUIStore()
  const { isAuthenticated, clearAuth } = useAuthStore()

  const handleLogout = async () => {
    try {
      await fetch('/api/auth/logout', { method: 'POST' })
    }
    finally {
      clearAuth()
      router.push('/login')
    }
  }

  if (!isAuthenticated)
    return null

  const navItems = [
    { label: 'Dashboard', href: '/dashboard', icon: <LayoutDashboard className="w-5 h-5" /> },
    { label: 'Agents', href: '/agents', icon: <Users className="w-5 h-5" /> },
    { label: 'Jobs', href: '/jobs', icon: <Briefcase className="w-5 h-5" /> },
    { label: 'Offers', href: '/offers', icon: <FileText className="w-5 h-5" /> },
    { label: 'Interviews', href: '/interviews', icon: <Calendar className="w-5 h-5" /> },
    { label: 'Matches', href: '/matches', icon: <ArrowLeftRight className="w-5 h-5" /> },
    { label: 'Messages', href: '/messages', icon: <MessageCircle className="w-5 h-5" /> },
    { label: 'Resume', href: '/resume', icon: <Upload className="w-5 h-5" /> },
    { label: 'Settings', href: '/settings', icon: <Settings className="w-5 h-5" /> },
    { label: 'Privacy', href: '/privacy', icon: <FileText className="w-5 h-5" /> },
    { label: 'Logout', href: '#', icon: <LogOut className="w-5 h-5" />, onClick: handleLogout },
  ]

  return (
    <>
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={toggleSidebar}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          fixed top-16 left-0 h-[calc(100vh-4rem)] bg-white border-r border-gray-200
          transition-all duration-300 z-50
          lg:relative lg:top-0 lg:h-full lg:bg-transparent lg:border-r-0 lg:border-none
          ${sidebarOpen ? 'w-64' : 'w-0 lg:w-16 lg:relative'}
          overflow-hidden
        `}
      >
        <nav className="p-3 space-y-1">
          {navItems.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)
            const isLogout = item.label === 'Logout'

            if (isLogout) {
              return (
                <button
                  key={item.href}
                  onClick={() => item.onClick?.()}
                  className={`
                    w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200
                    text-gray-600 hover:bg-gray-50 hover:text-gray-900
                    ${!sidebarOpen && 'lg:justify-center'}
                  `}
                >
                  {item.icon}
                  <span className={`font-medium ${!sidebarOpen && 'lg:hidden'}`}>{item.label}</span>
                </button>
              )
            }

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => {
                  if (window.innerWidth < 1024)
                    toggleSidebar()
                }}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200
                  ${isActive
                ? 'bg-blue-50 text-blue-600'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              }
                  ${!sidebarOpen && 'lg:justify-center'}
                `}
              >
                {item.icon}
                <span className={`font-medium ${!sidebarOpen && 'lg:hidden'}`}>{item.label}</span>
              </Link>
            )
          })}
        </nav>
      </aside>
    </>
  )
}
