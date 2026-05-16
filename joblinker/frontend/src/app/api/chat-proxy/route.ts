import { NextResponse } from 'next/server'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

function getTokenFromRequest(request: Request): string | null {
  const authHeader = request.headers.get('Authorization')
  if (authHeader?.startsWith('Bearer ')) {
    return authHeader.substring(7)
  }
  const cookieHeader = request.headers.get('cookie') || ''
  const cookieMatch = cookieHeader.match(/joblinker-auth=([^;]+)/)
  if (cookieMatch?.[1]) {
    try {
      const cookieData = JSON.parse(decodeURIComponent(cookieMatch[1]))
      if (cookieData.state?.token)
        return cookieData.state.token
      if (cookieData.token)
        return cookieData.token
    }
    catch {}
  }
  return null
}

function parseXmlContent(contentXml: string): string {
  if (!contentXml)
    return ''
  try {
    const paramsMatch = contentXml.match(/<parameters>([^<]+)<\/parameters>/)
    if (paramsMatch?.[1]) {
      try {
        const params = JSON.parse(paramsMatch[1])
        if (params.message)
          return params.message
        if (params.title)
          return `${params.title} - ${params.location || ''}`
        return paramsMatch[1]
      }
      catch {
        return paramsMatch[1]
      }
    }
    const textMatch = contentXml.match(/<text>([^<]*)<\/text>/)
    if (textMatch?.[1])
      return textMatch[1]
    return contentXml
  }
  catch {
    return contentXml
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { messages, id: matchId } = body as { messages: Array<{ content: string }>, id: string }

    if (!matchId || !messages?.length) {
      return NextResponse.json({ error: 'matchId and messages required' }, { status: 400 })
    }

    const token = getTokenFromRequest(request)
    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const lastMsg = messages[messages.length - 1]
    if (!lastMsg?.content) {
      return NextResponse.json({ error: 'last message has no content' }, { status: 400 })
    }
    const xml = `<message><payload><intent>INQUIRY</intent><parameters>{"message":"${lastMsg.content.replace(/"/g, '\\"')}"}</parameters></payload></message>`

    // POST user message to Go backend (fire-and-forget, triggers A2A)
    await fetch(`${API_BASE}/api/messages/${matchId}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ content_xml: xml, intent_type: 'INQUIRY' }),
    })

    // Poll for new assistant messages (max 30s)
    const knownCount = messages.length
    const startTime = Date.now()
    const maxWait = 30000
    const pollInterval = 1000

    const encoder = new TextEncoder()
    const stream = new ReadableStream({
      async start(controller) {
        const poll = async () => {
          if (Date.now() - startTime > maxWait) {
            controller.enqueue(encoder.encode('data: [DONE]\n\n'))
            controller.close()
            return
          }
          try {
            const resp = await fetch(`${API_BASE}/api/conversation/${matchId}`, {
              headers: { Authorization: `Bearer ${token}` },
            })
            if (resp.ok) {
              const data: Array<Record<string, unknown>> = await resp.json()
              const newMsgs = data.slice(knownCount)
              for (const m of newMsgs) {
                const content = parseXmlContent(String(m.content_xml || ''))
                const id = String(m.id || crypto.randomUUID())
                const payload = JSON.stringify({
                  id,
                  role: 'assistant',
                  content,
                  sender_agent_id: m.sender_agent_id,
                  intent_type: m.intent_type,
                  created_at: m.created_at,
                })
                controller.enqueue(encoder.encode(`data: ${payload}\n\n`))
              }
              if (newMsgs.length > 0) {
                controller.enqueue(encoder.encode('data: [DONE]\n\n'))
                controller.close()
                return
              }
            }
          }
          catch {}
          setTimeout(poll, pollInterval)
        }
        poll()
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
    return NextResponse.json({ error: 'Failed to proxy chat' }, { status: 500 })
  }
}
