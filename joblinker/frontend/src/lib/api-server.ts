import { cookies } from 'next/headers';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function getAuthToken(): Promise<string | null> {
  const cookieStore = await cookies();
  const authCookie = cookieStore.get('joblinker-auth');
  if (!authCookie?.value) return null;
  try {
    const parsed = JSON.parse(authCookie.value);
    return parsed.token || null;
  } catch {
    return null;
  }
}

export async function fetchServer<T>(endpoint: string): Promise<T | null> {
  const token = await getAuthToken();
  if (!token) return null;

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    });

    if (!response.ok) return null;
    return response.json();
  } catch {
    return null;
  }
}
