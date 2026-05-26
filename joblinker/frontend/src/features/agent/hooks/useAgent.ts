'use client'

import { useCallback } from 'react'

import type { CreateAgentRequest } from '@/shared/types/api'

import { getAuthToken } from '@/lib/api-utils'
import { useAgentStore } from '@/features/agent/stores/agent'

export function useAgent() {
  const {
    agents,
    currentAgent,
    isLoading,
    error,
    setAgents,
    setCurrentAgent,
    addAgent,
    updateAgent,
    removeAgent,
    setLoading,
    setError,
  } = useAgentStore()

  const getHeaders = () => {
    const token = getAuthToken()
    const headers: Record<string, string> = { 'Content-Type': 'application/json' }
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    return headers
  }

  const fetchAgents = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch('/api/agents', { headers: getHeaders() })
      if (!response.ok)
        throw new Error('Failed to fetch agents')
      const data = await response.json()
      setAgents(data.data || [])
    }
    catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
    finally {
      setLoading(false)
    }
  }, [setAgents, setLoading, setError])

  const createAgent = useCallback(
    async (agentData: CreateAgentRequest) => {
      setLoading(true)
      setError(null)
      try {
        const response = await fetch('/api/agents', {
          method: 'POST',
          headers: getHeaders(),
          body: JSON.stringify(agentData),
        })
        if (!response.ok)
          throw new Error('Failed to create agent')
        const data = await response.json()
        addAgent(data)
        return data
      }
      catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error')
        throw err
      }
      finally {
        setLoading(false)
      }
    },
    [addAgent, setLoading, setError],
  )

  const updateAgentById = useCallback(
    async (id: string, updates: Partial<CreateAgentRequest>) => {
      setLoading(true)
      setError(null)
      try {
        const response = await fetch(`/api/agents/${id}`, {
          method: 'PATCH',
          headers: getHeaders(),
          body: JSON.stringify(updates),
        })
        if (!response.ok)
          throw new Error('Failed to update agent')
        const data = await response.json()
        updateAgent(id, data)
        return data
      }
      catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error')
        throw err
      }
      finally {
        setLoading(false)
      }
    },
    [updateAgent, setLoading, setError],
  )

  const deleteAgent = useCallback(
    async (id: string) => {
      setLoading(true)
      setError(null)
      try {
        const response = await fetch(`/api/agents/${id}`, {
          method: 'DELETE',
          headers: getHeaders(),
        })
        if (!response.ok)
          throw new Error('Failed to delete agent')
        removeAgent(id)
      }
      catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error')
        throw err
      }
      finally {
        setLoading(false)
      }
    },
    [removeAgent, setLoading, setError],
  )

  return {
    agents,
    currentAgent,
    isLoading,
    error,
    fetchAgents,
    createAgent,
    updateAgent: updateAgentById,
    deleteAgent,
    setCurrentAgent,
  }
}
