import { create } from 'zustand'

import type { Agent } from '@/types'

interface AgentState {
  agents: Agent[]
  currentAgent: Agent | null
  isLoading: boolean
  error: string | null
  setAgents: (agents: Agent[]) => void
  setCurrentAgent: (agent: Agent | null) => void
  addAgent: (agent: Agent) => void
  updateAgent: (id: string, updates: Partial<Agent>) => void
  removeAgent: (id: string) => void
  setLoading: (isLoading: boolean) => void
  setError: (error: string | null) => void
}

export const useAgentStore = create<AgentState>(set => ({
  agents: [],
  currentAgent: null,
  isLoading: false,
  error: null,

  setAgents: agents => set({ agents }),

  setCurrentAgent: agent => set({ currentAgent: agent }),

  addAgent: agent =>
    set(state => ({ agents: [...state.agents, agent] })),

  updateAgent: (id, updates) =>
    set(state => ({
      agents: state.agents.map(a =>
        a.id === id ? { ...a, ...updates } : a,
      ),
      currentAgent:
        state.currentAgent?.id === id
          ? { ...state.currentAgent, ...updates }
          : state.currentAgent,
    })),

  removeAgent: id =>
    set(state => ({
      agents: state.agents.filter(a => a.id !== id),
      currentAgent:
        state.currentAgent?.id === id ? null : state.currentAgent,
    })),

  setLoading: isLoading => set({ isLoading }),

  setError: error => set({ error }),
}))
