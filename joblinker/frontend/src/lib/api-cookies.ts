import { cookies } from 'next/headers'

export async function getAuthTokenFromCookie(): Promise<string | null> {
  const cookieStore = await cookies()
  const authCookie = cookieStore.get('joblinker-auth')
  if (!authCookie?.value)
    return null
  try {
    // Cookie values are URL-encoded by browsers, need to decode first
    const decoded = decodeURIComponent(authCookie.value)
    const parsed = JSON.parse(decoded)
    return parsed.token || null
  }
  catch {
    return null
  }
}

export async function getAuthHeaderFromCookie(): Promise<Record<string, string>> {
  const token = await getAuthTokenFromCookie()
  if (token)
    return { Authorization: `Bearer ${token}` }
  return {}
}
