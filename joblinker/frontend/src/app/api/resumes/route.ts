import { NextResponse } from 'next/server';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Proxy to backend AI endpoints - /api/resumes/generate -> /ai/resumes/generate
export async function POST(request: Request) {
  try {
    const body = await request.json();
    const headers = { 'Content-Type': 'application/json' };
    // The frontend calls /api/resumes/generate but backend has /ai/resumes/generate
    const response = await fetch(`${API_BASE}/ai/resumes/generate`, {
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
