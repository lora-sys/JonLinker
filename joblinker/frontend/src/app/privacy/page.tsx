'use client'

import { Download, Trash2 } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useState } from 'react'

import { Modal } from '@/components/ui/Modal'
import { apiClient } from '@/lib/api_client'
import { useAuthStore } from '@/stores/auth'

export default function PrivacyPage() {
  const router = useRouter()
  const clearAuth = useAuthStore(s => s.clearAuth)
  const [exporting, setExporting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [confirmExport, setConfirmExport] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleExport = async () => {
    setExporting(true)
    setConfirmExport(false)
    try {
      const data = await apiClient.post<Record<string, unknown>>('/api/privacy/export')
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `joblinker-export-${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
    }
    catch (err) {
      console.error('Export failed:', err)
    }
    finally {
      setExporting(false)
    }
  }

  const handleDelete = async () => {
    setDeleting(true)
    setConfirmDelete(false)
    try {
      await apiClient.delete('/api/privacy/account')
      clearAuth()
      router.push('/login')
    }
    catch (err) {
      console.error('Delete account failed:', err)
      setDeleting(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Privacy Settings</h1>
        <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6 space-y-6">
          <div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">Data Export</h2>
            <p className="text-slate-500 mb-4">Download all your data in a portable format</p>
            <button
              onClick={() => setConfirmExport(true)}
              disabled={exporting}
              className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-medium shadow-lg shadow-blue-600/25 transition-all duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            >
              <Download className="w-4 h-4" />
              {exporting ? 'Exporting...' : 'Export My Data'}
            </button>
          </div>
          <div className="border-t border-slate-200 pt-6">
            <h2 className="text-lg font-semibold text-slate-900 mb-2">Delete Account</h2>
            <p className="text-slate-500 mb-4">Permanently delete your account and all associated data</p>
            <button
              onClick={() => setConfirmDelete(true)}
              disabled={deleting}
              className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-red-600 to-rose-500 text-white rounded-xl hover:from-red-700 hover:to-rose-600 font-medium shadow-lg shadow-red-600/25 transition-all duration-200 focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50"
            >
              <Trash2 className="w-4 h-4" />
              {deleting ? 'Deleting...' : 'Delete My Account'}
            </button>
          </div>
        </div>
      </div>

      <Modal isOpen={confirmExport} onClose={() => setConfirmExport(false)} title="Export Data" size="sm">
        <p className="text-slate-600 mb-6">Your data will be downloaded as a JSON file. This includes your profile, matches, conversations, and offers.</p>
        <div className="flex gap-3 justify-end">
          <button onClick={() => setConfirmExport(false)} className="px-4 py-2 text-sm font-medium text-slate-700 bg-slate-100 rounded-lg hover:bg-slate-200">Cancel</button>
          <button onClick={handleExport} className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700">Export</button>
        </div>
      </Modal>

      <Modal isOpen={confirmDelete} onClose={() => setConfirmDelete(false)} title="Delete Account" size="sm">
        <p className="text-slate-600 mb-2">
          This action is
          {' '}
          <span className="font-semibold text-red-600">permanent</span>
          {' '}
          and cannot be undone.
        </p>
        <p className="text-slate-500 mb-6">All your data will be removed including matches, conversations, and offers.</p>
        <div className="flex gap-3 justify-end">
          <button onClick={() => setConfirmDelete(false)} className="px-4 py-2 text-sm font-medium text-slate-700 bg-slate-100 rounded-lg hover:bg-slate-200">Cancel</button>
          <button onClick={handleDelete} className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700">Delete Permanently</button>
        </div>
      </Modal>
    </div>
  )
}
