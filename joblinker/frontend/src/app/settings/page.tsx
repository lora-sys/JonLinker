'use client';

import { useAuthStore } from '@/stores/auth';

export default function SettingsPage() {
  const { user } = useAuthStore();

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Settings</h1>
        <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
          <h2 className="text-lg font-semibold text-slate-900 mb-4">Account</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-500 mb-1">Email</label>
              <p className="text-slate-900">{user?.email || 'Not signed in'}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-500 mb-1">Role</label>
              <p className="text-slate-900 capitalize">{user?.role || 'Unknown'}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}