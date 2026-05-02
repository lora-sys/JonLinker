export const AI_CONFIG = {
  api: '/api/chat',
  model: 'gpt-4o-mini',
  streaming: {
    enabled: true,
    throttleMs: 16, // ~60fps
    chunkSize: 1,
  },
  maxTokens: 4096,
  temperature: 0.7,
} as const;

export function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('joblinker-auth');
  const authToken = token
    ? (() => {
        try {
          const parsed = JSON.parse(token);
          return parsed.state?.token || parsed.token;
        } catch {
          return token;
        }
      })()
    : '';
  return {
    Authorization: `Bearer ${authToken}`,
  };
}
