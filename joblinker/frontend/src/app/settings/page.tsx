'use client';

import { useAuthStore } from '@/stores/auth';

export default function SettingsPage() {
  const { user } = useAuthStore();

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Settings</h1>
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Account</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-500 mb-1">Email</label>
            <p className="text-gray-900">{user?.email || 'Not signed in'}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-500 mb-1">Role</label>
            <p className="text-gray-900 capitalize">{user?.role || 'Unknown'}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
