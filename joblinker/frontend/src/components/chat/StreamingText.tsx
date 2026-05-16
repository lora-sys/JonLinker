'use client'

import { useEffect, useRef, useState } from 'react'

interface StreamingTextProps {
  content: string
  speed?: number // characters per second
  onComplete?: () => void
}

export function StreamingText({ content, speed = 30, onComplete }: StreamingTextProps) {
  const [displayed, setDisplayed] = useState('')
  const indexRef = useRef(0)
  const contentRef = useRef(content)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const completedRef = useRef(false)

  useEffect(() => {
    contentRef.current = content
  }, [content])

  useEffect(() => {
    indexRef.current = 0
    setDisplayed('')
    completedRef.current = false

    const intervalMs = Math.max(1000 / speed, 16)

    intervalRef.current = setInterval(() => {
      if (indexRef.current < contentRef.current.length) {
        const nextIndex = Math.min(
          indexRef.current + Math.max(1, Math.floor(speed * (intervalMs / 1000))),
          contentRef.current.length,
        )
        indexRef.current = nextIndex
        setDisplayed(contentRef.current.slice(0, nextIndex))
      }
      else {
        if (!completedRef.current) {
          completedRef.current = true
          if (intervalRef.current) {
            clearInterval(intervalRef.current)
            intervalRef.current = null
          }
          onComplete?.()
        }
      }
    }, intervalMs)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }
  }, [content, speed, onComplete])

  return (
    <span aria-live="polite" aria-atomic="false">
      <span className="whitespace-pre-wrap break-words">{displayed}</span>
      {!completedRef.current && (
        <span className="inline-block w-0.5 h-4 bg-current ml-0.5 animate-pulse" aria-hidden="true" />
      )}
    </span>
  )
}
