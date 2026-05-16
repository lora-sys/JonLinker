'use client'

import { Bell, Cpu, Shield, Trash2, User } from 'lucide-react'
import { useState } from 'react'

import type { User as UserType } from '@/types'

import { LoadingSkeleton } from '@/components/ui'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import { useAuthStore } from '@/stores/auth'

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
  )
}

function AccountSettings({ user, onRefresh: _onRefresh }: { user: UserType | null, onRefresh: () => void }) {
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
  )
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
        ].map(setting => (
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
  )
}

function AISettings() {
  const [aiConfig, setAiConfig] = useState({
    api_key: '',
    base_url: 'https://api.openai.com/v1',
    model: 'gpt-4o-mini',
    temperature: '0.7',
    max_tokens: '4000',
  })
  const [saved, setSaved] = useState(false)

  const handleSave = async () => {
    // Save to localStorage for now (backend would persist this)
    localStorage.setItem('ai_config', JSON.stringify(aiConfig))
    setSaved(true)
    setTimeout(setSaved, 2000, false)
  }

  return (
    <Card className="p-6 bg-white/80 backdrop-blur-xl border border-white/20">
      <h2 className="text-lg font-semibold text-slate-900 mb-6 flex items-center gap-2">
        <Cpu className="w-5 h-5 text-purple-600" />
        AI Configuration
      </h2>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">API Key</label>
          <input
            type="password"
            value={aiConfig.api_key}
            onChange={e => setAiConfig({ ...aiConfig, api_key: e.target.value })}
            placeholder="sk-..."
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Base URL</label>
          <input
            type="text"
            value={aiConfig.base_url}
            onChange={e => setAiConfig({ ...aiConfig, base_url: e.target.value })}
            placeholder="https://api.openai.com/v1"
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Model</label>
            <select
              value={aiConfig.model}
              onChange={e => setAiConfig({ ...aiConfig, model: e.target.value })}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
            >
              <option value="gpt-4o">GPT-4o</option>
              <option value="gpt-4o-mini">GPT-4o Mini</option>
              <option value="gpt-4-turbo">GPT-4 Turbo</option>
              <option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Temperature</label>
            <input
              type="number"
              step="0.1"
              min="0"
              max="2"
              value={aiConfig.temperature}
              onChange={e => setAiConfig({ ...aiConfig, temperature: e.target.value })}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Max Tokens</label>
          <input
            type="number"
            min="100"
            max="128000"
            value={aiConfig.max_tokens}
            onChange={e => setAiConfig({ ...aiConfig, max_tokens: e.target.value })}
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
          />
        </div>

        <div className="flex items-center gap-3 pt-2">
          <Button onClick={handleSave} className="cursor-pointer">
            {saved ? 'Saved!' : 'Save AI Config'}
          </Button>
          {saved && <span className="text-sm text-green-600">Configuration saved successfully</span>}
        </div>
      </div>
    </Card>
  )
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
  )
}

export function SettingsContent({ initialUser }: { initialUser: UserType | null }) {
  const storeUser = useAuthStore(state => state.user)
  const user = initialUser || storeUser

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3 mb-8">
          Settings
        </h1>

        {user
          ? (
              <div className="space-y-6">
                <AccountSettings user={user} onRefresh={() => {}} />
                <AISettings />
                <NotificationSettings />
                <DangerZone />
              </div>
            )
          : (
              <AccountSettingsSkeleton />
            )}
      </div>
    </div>
  )
}
