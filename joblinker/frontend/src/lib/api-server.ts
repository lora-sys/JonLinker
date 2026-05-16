import { getServerToken } from '@/lib/auth-utils'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface FetchServerResult<T> {
  data: T | null
  error: string | null
}

export async function fetchServer<T>(endpoint: string): Promise<T | null> {
  const token = await getServerToken()
  if (!token)
    return null

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok)
      return null
    return response.json()
  }
  catch {
    return null
  }
}

export async function fetchServerWithResult<T>(
  endpoint: string,
): Promise<FetchServerResult<T>> {
  const token = await getServerToken()
  if (!token) {
    return { data: null, error: 'Not authenticated' }
  }

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      return { data: null, error: `Failed to fetch: ${response.status}` }
    }

    const data = await response.json()
    return { data, error: null }
  }
  catch {
    return { data: null, error: 'Failed to connect to backend' }
  }
}
