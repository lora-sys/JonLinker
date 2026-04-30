'use client';

import { useState } from 'react';
import { Bell, Shield, Trash2, User } from 'lucide-react';
import { useAuthStore } from '@/stores/auth';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import { LoadingSkeleton, ErrorState } from '@/components/ui';
import type { User as UserType } from '@/types';

function AccountSettingsSkeleton() {
  return (
    <Card className="p-6 bg-white/80 backdrop-blur-xl border border-white/20">
      <h2 className="text-lg font-semibold text-slate-900 mb-6 flex items-center gap-2">
        <User className="w-5 h-5 text-blue-600" />
        Account Settings
      </h2>
      <div className="space-y-6">
        <LoadingSkeleton count={4} variant="list" />
      </div>
    </Card>
  );
}

function AccountSettings({ user, onRefresh }: { user: UserType | null; onRefresh: () => void }) {
  return (
    <Card className="p-6 bg-white/80 backdrop-blur-xl border border-white/20">
      <h2 className="text-lg font-semibold text-slate-900 mb-6 flex items-center gap-2">
        <User className="w-5 h-5 text-blue-600" />
        Account Settings
      </h2>

      <div className="space-y-6">
        <div className="flex items-center justify-between py-3 border-b border-slate-100">
          <div>
            <label className="text-sm font-medium text-slate-500">Email</label>
            <p className="text-slate-900">{user?.email || 'Not signed in'}</p>
          </div>
          <Button variant="ghost" size="sm" className="cursor-pointer">
            Change
          </Button>
        </div>

        <div className="flex items-center justify-between py-3 border-b border-slate-100">
          <div>
            <label className="text-sm font-medium text-slate-500">Account Type</label>
            <p className="text-slate-900 capitalize">{user?.role || 'Unknown'}</p>
          </div>
        </div>

        <div className="flex items-center justify-between py-3 border-b border-slate-100">
          <div>
            <label className="text-sm font-medium text-slate-500">Organization</label>
            <p className="text-slate-900">{user?.organization_id || 'No organization'}</p>
          </div>
          <Button variant="ghost" size="sm" className="cursor-pointer">
            Manage
          </Button>
        </div>

        <div className="flex items-center justify-between py-3">
          <div>
            <label className="text-sm font-medium text-slate-500">Member Since</label>
            <p className="text-slate-900">
              {user?.created_at
                ? new Date(user.created_at).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })
                : 'Unknown'}
            </p>
          </div>
        </div>
      </div>
    </Card>
  );
}

function NotificationSettings() {
  return (
    <Card className="p-6 bg-white/80 backdrop-blur-xl border border-white/20">
      <h2 className="text-lg font-semibold text-slate-900 mb-6 flex items-center gap-2">
        <Bell className="w-5 h-5 text-blue-600" />
        Notifications
      </h2>

      <div className="space-y-4">
        {[
          { label: 'New match notifications', enabled: true },
          { label: 'Interview reminders', enabled: true },
          { label: 'Offer updates', enabled: false },
          { label: 'Marketing emails', enabled: false },
        ].map((setting) => (
          <div key={setting.label} className="flex items-center justify-between py-2">
            <span className="text-slate-700">{setting.label}</span>
            <button
              className={`relative w-11 h-6 rounded-full transition-colors duration-200 ${
                setting.enabled ? 'bg-blue-600' : 'bg-slate-200'
              }`}
            >
              <span
                className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200 ${
                  setting.enabled ? 'translate-x-5' : ''
                }`}
              />
            </button>
          </div>
        ))}
      </div>
    </Card>
  );
}

function DangerZone() {
  return (
    <Card className="p-6 bg-white/80 backdrop-blur-xl border border-red-200">
      <h2 className="text-lg font-semibold text-red-600 mb-4 flex items-center gap-2">
        <Shield className="w-5 h-5" />
        Danger Zone
      </h2>
      <p className="text-sm text-slate-600 mb-4">
        Once you delete your account, there is no going back. Please be certain.
      </p>
      <Button variant="ghost" className="border border-red-200 text-red-600 hover:bg-red-50 cursor-pointer flex items-center gap-2">
        <Trash2 className="w-4 h-4" />
        Delete Account
      </Button>
    </Card>
  );
}

export function SettingsContent({ initialUser }: { initialUser: UserType | null }) {
  const { user: storeUser } = useAuthStore();
  const [user] = useState<UserType | null>(initialUser || storeUser);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3 mb-8">
          Settings
        </h1>

        {user ? (
          <div className="space-y-6">
            <AccountSettings user={user} onRefresh={() => {}} />
            <NotificationSettings />
            <DangerZone />
          </div>
        ) : (
          <AccountSettingsSkeleton />
        )}
      </div>
    </div>
  );
}