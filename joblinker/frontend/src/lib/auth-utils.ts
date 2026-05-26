import { cookies } from 'next/headers'

import { useAuthStore } from '@/shared/stores/auth'

const COOKIE_NAME = 'joblinker-auth'

interface AuthCookieValue {
  token: string
  userId: string
}

function parseAuthCookie(raw: string): AuthCookieValue | null {
  try {
    const decoded = decodeURIComponent(raw)
    const parsed = JSON.parse(decoded)
    if (parsed.token)
      return parsed
    return null
  }
  catch {
    return null
  }
}

export async function getServerToken(): Promise<string | null> {
  const cookieStore = await cookies()
  const authCookie = cookieStore.get(COOKIE_NAME)
  if (!authCookie?.value)
    return null
  const parsed = parseAuthCookie(authCookie.value)
  return parsed?.token ?? null
}

export async function getServerAuthHeader(): Promise<Record<string, string>> {
  const token = await getServerToken()
  if (token)
    return { Authorization: `Bearer ${token}` }
  return {}
}

export function getClientToken(): string | null {
  if (typeof window === 'undefined')
    return null
  const state = useAuthStore.getState()
  if (state.token)
    return state.token
  return null
}

export function getClientAuthHeader(): Record<string, string> {
  const token = getClientToken()
  if (token)
    return { Authorization: `Bearer ${token}` }
  return {}
}
