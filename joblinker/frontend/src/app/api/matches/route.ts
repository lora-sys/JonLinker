import { NextResponse } from 'next/server';
import { getAuthHeader } from '@/lib/api-utils';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function GET() {
  try {
    const authHeader = getAuthHeader();
    const response = await fetch(`${API_BASE}/api/matches`, {
      headers: { ...authHeader },
    });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 });
  }
}