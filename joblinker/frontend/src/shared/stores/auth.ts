import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import type { User } from '@/shared/types'

import { apiClient } from '@/lib/api_client'

function parseTenantId(token: string): string | undefined {
  try {
    const payload = token.split('.')[1]
    if (!payload)
      return undefined
    const decoded = JSON.parse(atob(payload))
    return decoded.tenant_id || undefined
  }
  catch {
    return undefined
  }
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  _hasRehydrated: boolean
  setAuth: (user: User, token: string) => void
  clearAuth: () => void
  updateUser: (user: Partial<User>) => void
  setRehydrated: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    set => ({
      user: null,
      token: null,
      isAuthenticated: false,
      _hasRehydrated: false,

      setAuth: (user, token) => {
        apiClient.setToken(token)
        apiClient.setUserId(user.id, undefined, parseTenantId(token))
        set({ user, token, isAuthenticated: true })
      },

      clearAuth: () => {
        apiClient.setToken(null)
        set({ user: null, token: null, isAuthenticated: false })
      },

      updateUser: userData =>
        set(state => ({
          user: state.user ? { ...state.user, ...userData } : null,
        })),

      setRehydrated: () => set({ _hasRehydrated: true }),
    }),
    {
      name: 'joblinker-auth',
      partialize: state => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.token) {
          apiClient.setToken(state.token)
          if (state.user?.id) {
            apiClient.setUserId(state.user.id)
          }
        }
      },
    },
  ),
)
