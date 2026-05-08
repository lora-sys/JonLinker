import { streamText } from 'ai';
import { createOpenAI } from '@ai-sdk/openai';
import { getServerToken } from '@/lib/auth-utils';

const provider = createOpenAI({
  baseURL: process.env.AI_BASE_URL,
  apiKey: process.env.AI_API_KEY,
});

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function POST(req: Request) {
  const { messages, matchId } = await req.json();
  const token = await getServerToken();

  if (matchId && token) {
    const lastMessage = messages[messages.length - 1];
    if (lastMessage?.role === 'user') {
      fetch(`${API_BASE}/api/messages/${matchId}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content: lastMessage.content }),
      }).catch(console.error);
    }
  }

  const result = streamText({
    model: provider(process.env.AI_MODEL || 'gpt-4o-mini'),
    messages,
    system: 'You are a helpful recruitment assistant. Respond concisely and professionally.',
  });

  return result.toTextStreamResponse();
}
