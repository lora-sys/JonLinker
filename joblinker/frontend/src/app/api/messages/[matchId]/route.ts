import { NextResponse } from 'next/server'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

function getTokenFromRequest(request: Request): string | null {
  // First try Authorization header
  const authHeader = request.headers.get('Authorization')
  if (authHeader?.startsWith('Bearer ')) {
    return authHeader.substring(7)
  }

  // Fall back to cookie parsing
  const cookieHeader = request.headers.get('cookie') || ''
  const cookieMatch = cookieHeader.match(/joblinker-auth=([^;]+)/)

  if (cookieMatch?.[1]) {
    try {
      const cookieData = JSON.parse(decodeURIComponent(cookieMatch[1]))
      // Handle Zustand persist format: {"state": {"token": "...", "user": {...}}}
      if (cookieData.state?.token)
        return cookieData.state.token
      // Handle simple format: {"token": "...", "userId": "..."}
      if (cookieData.token)
        return cookieData.token
    }
    catch (e) {
      console.error('Failed to parse cookie:', e)
    }
  }

  return null
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ matchId: string }> },
) {
  try {
    const { matchId } = await params
    const token = getTokenFromRequest(request)

    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const response = await fetch(`${API_BASE}/api/messages/${matchId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  }
  catch {
    return NextResponse.json({ error: 'Failed to fetch messages' }, { status: 500 })
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ matchId: string }> },
) {
  try {
    const { matchId } = await params
    const body = await request.json()
    const token = getTokenFromRequest(request)

    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const response = await fetch(`${API_BASE}/api/messages/${matchId}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  }
  catch {
    return NextResponse.json({ error: 'Failed to send message' }, { status: 500 })
  }
}
