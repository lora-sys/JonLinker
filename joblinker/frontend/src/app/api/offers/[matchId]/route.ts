import { NextResponse } from 'next/server'

import { getAuthHeaderFromCookie } from '@/lib/api-cookies'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

function getAuthHeaderFromRequest(request: Request): Record<string, string> {
  const authHeader = request.headers.get('Authorization')
  if (authHeader?.startsWith('Bearer ')) {
    return { Authorization: authHeader }
  }
  return {}
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ matchId: string }> },
) {
  try {
    const { matchId } = await params
    const headerAuth = getAuthHeaderFromRequest(request)
    const cookieAuth = await getAuthHeaderFromCookie()
    const authHeader = Object.keys(headerAuth).length > 0 ? headerAuth : cookieAuth
    const response = await fetch(`${API_BASE}/api/offers/${matchId}`, {
      headers: { ...authHeader },
    })
    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  }
  catch {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 })
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ matchId: string }> },
) {
  try {
    const { matchId } = await params
    const body = await request.json()
    const headerAuth = getAuthHeaderFromRequest(request)
    const cookieAuth = await getAuthHeaderFromCookie()
    const authHeader = Object.keys(headerAuth).length > 0 ? headerAuth : cookieAuth
    const headers = { 'Content-Type': 'application/json', ...authHeader }
    const response = await fetch(`${API_BASE}/api/offers/${matchId}/${body.action || ''}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    })
    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  }
  catch {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 })
  }
}
