export function getAuthToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  try {
    const stored = localStorage.getItem('joblinker-auth');
    if (stored) {
      const parsed = JSON.parse(stored);
      // Zustand persist wraps state in 'state' key
      const token = parsed.state?.token || parsed.token;
      return token || null;
    }
  } catch {
    // ignore
  }
  return null;
}

export function getAuthHeader(): Record<string, string> {
  const token = getAuthToken();
  if (token) {
    return { 'Authorization': `Bearer ${token}` };
  }
  return {};
}