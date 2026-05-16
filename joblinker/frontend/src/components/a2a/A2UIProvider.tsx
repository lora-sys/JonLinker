'use client'

import React, { createContext, useCallback, useContext, useRef, useState } from 'react'

import type { A2UIComponent, A2UIMessage, SurfaceState } from '@/hooks/useSSE'

interface A2UIContextValue {
  surfaces: Map<string, SurfaceState>
  addSurface: (sessionId: string, rootId: string) => void
  updateSurface: (sessionId: string, components: A2UIComponent[]) => void
  updateDataModel: (key: string, delta: string) => void
  getSurface: (sessionId: string) => SurfaceState | undefined
}

const A2UIContext = createContext<A2UIContextValue | null>(null)

export function useA2UI() {
  const ctx = useContext(A2UIContext)
  if (!ctx)
    throw new Error('useA2UI must be inside A2UIProvider')
  return ctx
}

export function A2UIProvider({ children }: { children: React.ReactNode }) {
  const [surfaces, setSurfaces] = useState<Map<string, SurfaceState>>(new Map())
  const dataModelsRef = useRef<Map<string, string>>(new Map())

  const addSurface = useCallback((sessionId: string, rootId: string) => {
    setSurfaces((prev) => {
      const next = new Map(prev)
      next.set(sessionId, {
        rootId,
        components: new Map(),
        dataModels: new Map(),
        sessionId,
      })
      return next
    })
  }, [])

  const updateSurface = useCallback((sessionId: string, components: A2UIComponent[]) => {
    setSurfaces((prev) => {
      const next = new Map(prev)
      const existing = next.get(sessionId)
      if (!existing)
        return prev
      const comps = new Map(existing.components)
      for (const comp of components) {
        comps.set(comp.id, comp)
      }
      next.set(sessionId, { ...existing, components: comps })
      return next
    })
  }, [])

  const updateDataModel = useCallback((key: string, delta: string) => {
    const dms = dataModelsRef.current
    const existing = dms.get(key) || ''
    dms.set(key, existing + delta)
    setSurfaces((prev) => {
      const next = new Map(prev)
      for (const [sid, surface] of next) {
        const dms2 = new Map(surface.dataModels)
        dms2.set(key, (dms2.get(key) || '') + delta)
        next.set(sid, { ...surface, dataModels: dms2 })
      }
      return next
    })
  }, [])

  const getSurface = useCallback((sessionId: string) => {
    return surfaces.get(sessionId)
  }, [surfaces])

  return (
    <A2UIContext.Provider value={{ surfaces, addSurface, updateSurface, updateDataModel, getSurface }}>
      {children}
    </A2UIContext.Provider>
  )
}

export function processA2UIMessage(msg: A2UIMessage, ctx: A2UIContextValue) {
  if (msg.beginRendering) {
    ctx.addSurface(msg.beginRendering.sessionId, msg.beginRendering.rootId)
  }
  if (msg.surfaceUpdate) {
    const sessionId = Array.from(ctx.surfaces.keys()).pop() || 'default'
    ctx.updateSurface(sessionId, msg.surfaceUpdate.components)
  }
  if (msg.dataModelUpdate) {
    ctx.updateDataModel(msg.dataModelUpdate.key, msg.dataModelUpdate.delta)
  }
}
