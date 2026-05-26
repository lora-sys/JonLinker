import { create } from 'zustand'

import type { Match } from '@/shared/types'

interface MatchState {
  matches: Match[]
  currentMatch: Match | null
  isLoading: boolean
  error: string | null
  setMatches: (matches: Match[]) => void
  setCurrentMatch: (match: Match | null) => void
  addMatch: (match: Match) => void
  updateMatch: (id: string, updates: Partial<Match>) => void
  setLoading: (isLoading: boolean) => void
  setError: (error: string | null) => void
}

export const useMatchStore = create<MatchState>(set => ({
  matches: [],
  currentMatch: null,
  isLoading: false,
  error: null,

  setMatches: matches => set({ matches }),

  setCurrentMatch: match => set({ currentMatch: match }),

  addMatch: match =>
    set(state => ({ matches: [...state.matches, match] })),

  updateMatch: (id, updates) =>
    set(state => ({
      matches: state.matches.map(m =>
        m.id === id ? { ...m, ...updates } : m,
      ),
      currentMatch:
        state.currentMatch?.id === id
          ? { ...state.currentMatch, ...updates }
          : state.currentMatch,
    })),

  setLoading: isLoading => set({ isLoading }),

  setError: error => set({ error }),
}))
