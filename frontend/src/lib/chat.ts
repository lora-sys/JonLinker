import { useState, useRef, useCallback } from "react";

export function useSSEChat(sessionId: string | null, onComplete?: () => void) {
  const [messages, setMessages] = useState<{ role: string; content: string }[]>([]);
  const [streaming, setStreaming] = useState(false);
  const abortRef = useRef<AbortController | null>(null);
  const notifiedRef = useRef(false);

  const send = useCallback(async (text: string) => {
    if (!sessionId || streaming) return;
    setStreaming(true);
    setMessages((prev) => [...prev, { role: "user", content: text }]);

    const ac = new AbortController();
    abortRef.current = ac;

    try {
      const res = await fetch("/api/chat/resume", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId, message: text }),
        signal: ac.signal,
      });

      const reader = res.body?.getReader();
      if (!reader) return;
      const decoder = new TextDecoder();
      let buffer = "";

      setMessages((prev) => [...prev, { role: "assistant", content: "" }]);

      while (true) {
        const { done: rd, value } = await reader.read();
        if (rd) break;
        const chunk = value ? decoder.decode(value) : "";
        buffer += chunk;

        const parts = buffer.split("\n\n");
        buffer = parts.pop() || "";

        for (const part of parts) {
          if (!part.startsWith("data: ")) continue;
          const data = part.slice(6).trim();
          if (data === "[DONE]") continue;
          const token = data.replace(/\\n/g, "\n");

          if (!notifiedRef.current && token.includes('"complete"')) {
            try {
              const state = JSON.parse(token);
              if (state.complete) {
                notifiedRef.current = true;
                onComplete?.();
              }
            } catch { /* ignore parse errors on partial tokens */ }
          }

          setMessages((prev) => {
            const next = [...prev];
            const last = next[next.length - 1];
            if (last && last.role === "assistant") {
              let display = last.content + token;
              try {
                const parsed = JSON.parse(display);
                if (parsed.complete !== undefined) {
                  display = parsed.complete ? "✅ 个人资料已完善，可以开始搜索职位了！" : parsed.message || display;
                }
              } catch { /* not JSON, display as-is */ }
              next[next.length - 1] = { ...last, content: display };
            }
            return next;
          });
        }
      }
    } catch {
    } finally {
      setStreaming(false);
      abortRef.current = null;
    }
  }, [sessionId, streaming, onComplete]);

  const reset = useCallback(() => {
    setMessages([]);
    notifiedRef.current = false;
  }, []);

  return { messages, streaming, send, reset };
}

export function getTextFromParts(parts: { type: string; text?: string }[]): string {
  return parts.filter((p) => p.type === "text").map((p) => p.text ?? "").join("");
}
