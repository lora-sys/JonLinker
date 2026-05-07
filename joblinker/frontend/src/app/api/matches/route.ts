import { NextResponse } from 'next/server';
import { getAuthHeaderFromCookie } from '@/lib/api-cookies';


function getAuthHeaderFromRequest(request: Request): Record<string, string> {
  const authHeader = request.headers.get('Authorization');
  if (authHeader?.startsWith('Bearer ')) {
    return { Authorization: authHeader };
  }
  return {};
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function GET(request: Request) {
  try {
    const headerAuth = getAuthHeaderFromRequest(request);
    const cookieAuth = await getAuthHeaderFromCookie();
    const authHeader = Object.keys(headerAuth).length > 0 ? headerAuth : cookieAuth;
    const response = await fetch(`${API_BASE}/api/matches`, {
      headers: { ...authHeader },
    });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 });
  }
}

export async function POST(request: Request) {
  try {
    const headerAuth = getAuthHeaderFromRequest(request);
    const cookieAuth = await getAuthHeaderFromCookie();
    const authHeader = Object.keys(headerAuth).length > 0 ? headerAuth : cookieAuth;
    const body = await request.json();
    const response = await fetch(`${API_BASE}/api/matches`, {
      method: 'POST',
      headers: { ...authHeader, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    return NextResponse.json({ error: 'Failed to connect to backend' }, { status: 500 });
  }
}
