import { NextResponse } from 'next/server';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function getAuthHeader(request: Request): Record<string, string> {
  const authHeader = request.headers.get('Authorization');
  if (authHeader) {
    return { 'Authorization': authHeader };
  }
  return {};
}

export async function GET(request: Request) {
  try {
    const headers = getAuthHeader(request);
    const response = await fetch(`${API_BASE}/api/agents`, { headers });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 });
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const headers = { 'Content-Type': 'application/json', ...getAuthHeader(request) };
    const response = await fetch(`${API_BASE}/api/agents`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 });
  }
}