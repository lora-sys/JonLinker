import type { NextRequest } from 'next/server'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { message, sessionId } = body

    const backendUrl = new URL(`${API_BASE}/api/a2a/chat`)
    backendUrl.searchParams.set('session_id', sessionId || crypto.randomUUID())

    const response = await fetch(backendUrl.toString(), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ message }),
    })

    if (!response.ok) {
      return new Response(JSON.stringify({ error: 'Backend request failed' }), {
        status: response.status,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Proxy SSE stream
    const reader = response.body?.getReader()
    if (!reader) {
      return new Response(JSON.stringify({ error: 'No response body' }), { status: 502 })
    }

    const stream = new ReadableStream({
      async start(controller) {
        const decoder = new TextDecoder()
        try {
          while (true) {
            const { done, value } = await reader.read()
            if (done)
              break
            const text = decoder.decode(value, { stream: true })
            controller.enqueue(new TextEncoder().encode(text))
          }
        }
        catch {
          // Stream ended
        }
        finally {
          controller.close()
        }
      },
    })

    return new Response(stream, {
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive',
      },
    })
  }
  catch {
    return new Response(JSON.stringify({ error: 'Failed to connect to backend' }), { status: 500 })
  }
}
