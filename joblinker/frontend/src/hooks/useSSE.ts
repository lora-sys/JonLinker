'use client'

import { useCallback, useEffect, useRef, useState } from 'react'

export interface A2UIMessage {
  beginRendering?: { sessionId: string, rootId: string }
  surfaceUpdate?: { components: A2UIComponent[] }
  dataModelUpdate?: { key: string, delta: string }
  deleteSurface?: { sessionID: string }
  interruptRequest?: { interruptId: string, question: string, options?: string[] }
}

export interface A2UIComponent {
  id: string
  type: string
  props: Record<string, unknown>
  children?: string[]
  dataKey?: string
}

export interface SurfaceState {
  rootId: string
  components: Map<string, A2UIComponent>
  dataModels: Map<string, string>
  sessionId: string
}

interface UseSSEOptions {
  url: string
  onMessage?: (msg: A2UIMessage) => void
  onError?: (err: Error) => void
  onComplete?: () => void
}

export function useSSE({ url, onMessage, onError, onComplete }: UseSSEOptions) {
  const [isConnected, setIsConnected] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  const start = useCallback(() => {
    if (abortRef.current)
      abortRef.current.abort()
    const controller = new AbortController()
    abortRef.current = controller

    const connect = async () => {
      try {
        const response = await fetch(url, {
          signal: controller.signal,
          headers: { Accept: 'text/event-stream' },
        })

        if (!response.ok) {
          throw new Error(`SSE connection failed: ${response.status}`)
        }

        setIsConnected(true)
        const reader = response.body?.getReader()
        if (!reader)
          throw new Error('No response body')

        const decoder = new TextDecoder()
        let buffer = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done)
            break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''

          for (const line of lines) {
            const trimmed = line.trim()
            if (trimmed.startsWith('data:')) {
              try {
                const jsonStr = trimmed.slice(5).trimStart()
                const msg: A2UIMessage = JSON.parse(jsonStr)
                onMessage?.(msg)
              }
              catch (e) {
                console.error('SSE parse error:', e)
              }
            }
          }
        }
      }
      catch (err) {
        if ((err as Error).name === 'AbortError')
          return
        onError?.(err as Error)
      }
      finally {
        setIsConnected(false)
        onComplete?.()
      }
    }

    connect()
  }, [url, onMessage, onError, onComplete])

  const stop = useCallback(() => {
    abortRef.current?.abort()
    setIsConnected(false)
  }, [])

  useEffect(() => {
    return () => stop()
  }, [stop])

  return { start, stop, isConnected }
}
