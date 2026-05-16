'use client'

import { useAuthStore } from '@/stores/auth'

export type UserRole = 'seeker' | 'recruiter' | 'admin'

export function useRole(): UserRole {
  const user = useAuthStore(state => state.user)
  return (user?.role as UserRole) || 'seeker'
}

export function useIsAuthenticated(): boolean {
  return useAuthStore(state => state.isAuthenticated)
}
