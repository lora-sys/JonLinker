'use client'

import type { ChatMessage, ConversationThread } from '@/types/ai'

const STORAGE_KEY = 'joblinker-ai-threads'
const MAX_STORED_MESSAGES = 1000

export function saveThread(thread: ConversationThread): void {
  try {
    const threads = getAllThreads()
    const existingIndex = threads.findIndex(t => t.id === thread.id)

    const trimmedThread = {
      ...thread,
      messages: thread.messages.slice(-MAX_STORED_MESSAGES),
    }

    if (existingIndex >= 0) {
      threads[existingIndex] = trimmedThread
    }
    else {
      threads.push(trimmedThread)
    }

    localStorage.setItem(STORAGE_KEY, JSON.stringify(threads))
  }
  catch {
    // Storage full or unavailable
  }
}

export function loadThread(threadId: string): ConversationThread | null {
  try {
    const threads = getAllThreads()
    const thread = threads.find(t => t.id === threadId)
    if (!thread)
      return null

    return {
      ...thread,
      createdAt: new Date(thread.createdAt),
      updatedAt: new Date(thread.updatedAt),
      messages: thread.messages.map(m => ({
        ...m,
        createdAt: new Date(m.createdAt),
      })),
    }
  }
  catch {
    return null
  }
}

export function getAllThreads(): ConversationThread[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw)
      return []
    return JSON.parse(raw)
  }
  catch {
    return []
  }
}

export function deleteThread(threadId: string): void {
  try {
    const threads = getAllThreads().filter(t => t.id !== threadId)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(threads))
  }
  catch {
    // Ignore
  }
}

export function exportThreadMessages(threadId: string): ChatMessage[] {
  const thread = loadThread(threadId)
  return thread?.messages ?? []
}
