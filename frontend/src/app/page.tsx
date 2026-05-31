"use client";

import { useState, useRef, useEffect, useCallback } from "react";

// ---- Types ----
type Job = {
  title: string; company: string; location: string; salary: string;
  url: string; description: string; tags: string[]; source: string;
};
type RankedJob = Job & { match_score: number; summary: string; highlights: string[] };
type Application = { job_title: string; company: string; cover_letter: string; resume_md: string; highlights: string[]; generated_at: string };

// ---- Custom SSE Chat Hook ----
function useSSEChat(sessionId: string | null) {
  const [messages, setMessages] = useState<{ role: string; content: string }[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [done, setDone] = useState(false);
  const abortRef = useRef<AbortController | null>(null);

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
          if (data === "[DONE]") {
            setDone(true);
            continue;
          }
          const token = data.replace(/\\n/g, "\n");
          setMessages((prev) => {
            const next = [...prev];
            const last = next[next.length - 1];
            if (last && last.role === "assistant") {
              next[next.length - 1] = { ...last, content: last.content + token };
            }
            return next;
          });
        }
      }
    } catch {
      // aborted or error
    } finally {
      setStreaming(false);
      abortRef.current = null;
    }
  }, [sessionId, streaming]);

  return { messages, streaming, done, send };
}

// ---- Resume Chat Island ----
function ResumeChatIsland({ onSessionReady }: { onSessionReady: (sid: string) => void }) {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [input, setInput] = useState("");
  const { messages, streaming, send } = useSSEChat(sessionId);
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: "smooth" }); }, [messages]);

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const fd = new FormData();
      fd.set("file", file);
      const res = await fetch("/api/resume/upload", { method: "POST", body: fd });
      const data = await res.json();
      if (data.session_id) {
        setSessionId(data.session_id);
        onSessionReady(data.session_id);
      }
    } finally {
      setUploading(false);
    }
  }

  async function handleSend() {
    if (!input.trim() || streaming) return;
    const text = input;
    setInput("");
    await send(text);
  }

  return (
    <div className="flex flex-col h-full bg-white rounded-xl border border-gray-200 overflow-hidden">
      <div className="px-4 py-3 border-b border-gray-100">
        <h2 className="text-sm font-semibold text-gray-900">简历助手</h2>
        <p className="text-xs text-gray-500 mt-0.5">上传简历并聊天完善资料</p>
      </div>

      {!sessionId ? (
        <div className="flex-1 flex flex-col items-center justify-center gap-4 p-6">
          <label className="cursor-pointer inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors">
            {uploading ? (
              <><span className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full" /> 解析中...</>
            ) : (
              "上传 PDF 简历"
            )}
            <input type="file" accept=".pdf" className="hidden" onChange={handleUpload} disabled={uploading} />
          </label>
          <p className="text-xs text-gray-400">支持 PDF 格式</p>
        </div>
      ) : (
        <>
          <div className="flex-1 overflow-y-auto p-4 space-y-3">
            {messages.length === 0 && (
              <p className="text-sm text-gray-400 text-center py-8">
                简历已解析完成，开始聊天完善资料...
              </p>
            )}
            {messages.map((m, i) => (
              <div key={i} className={`flex ${m.role === "user" ? "justify-end" : "justify-start"}`}>
                <div className={`max-w-[85%] rounded-xl px-3.5 py-2 text-sm leading-relaxed ${
                  m.role === "user"
                    ? "bg-blue-600 text-white"
                    : "bg-gray-100 text-gray-800"
                }`}>
                  {m.content}
                </div>
              </div>
            ))}
            {streaming && (
              <div className="flex justify-start">
                <div className="bg-gray-100 rounded-xl px-3.5 py-2">
                  <span className="inline-block w-1.5 h-4 bg-gray-400 animate-pulse" />
                </div>
              </div>
            )}
            <div ref={chatEndRef} />
          </div>

          <div className="border-t border-gray-100 p-3 flex gap-2">
            <input
              className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="输入信息..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSend()}
              disabled={streaming}
            />
            <button
              className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              onClick={handleSend}
              disabled={streaming || !input.trim()}
            >
              发送
            </button>
          </div>
        </>
      )}
    </div>
  );
}

// ---- Application Postcard ----
function ApplicationCard({ app }: { app: Application }) {
  const [tab, setTab] = useState<"letter" | "resume">("letter");
  return (
    <div className="mt-4 bg-white rounded-xl border border-blue-200 overflow-hidden shadow-sm">
      <div className="bg-gradient-to-r from-blue-50 to-indigo-50 px-5 py-4 border-b border-blue-100">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-gray-900">申请已生成</h3>
            <p className="text-sm text-gray-600 mt-0.5">{app.job_title} · {app.company}</p>
          </div>
          <div className="flex gap-1.5">
            {app.highlights.slice(0, 3).map((h, i) => (
              <span key={i} className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">{h}</span>
            ))}
          </div>
        </div>
      </div>

      <div className="flex border-b border-gray-100">
        <button
          className={`flex-1 py-2.5 text-sm font-medium transition-colors ${tab === "letter" ? "text-blue-600 border-b-2 border-blue-600" : "text-gray-500 hover:text-gray-700"}`}
          onClick={() => setTab("letter")}
        >求职信</button>
        <button
          className={`flex-1 py-2.5 text-sm font-medium transition-colors ${tab === "resume" ? "text-blue-600 border-b-2 border-blue-600" : "text-gray-500 hover:text-gray-700"}`}
          onClick={() => setTab("resume")}
        >定制简历</button>
      </div>

      <div className="p-5 max-h-80 overflow-y-auto">
        <div className="text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
          {tab === "letter" ? app.cover_letter : app.resume_md}
        </div>
      </div>

      <div className="px-5 py-3 bg-gray-50 border-t border-gray-100 text-xs text-gray-400">
        生成时间：{new Date(app.generated_at).toLocaleString("zh-CN")}
      </div>
    </div>
  );
}

// ---- Search Island ----
function SearchIsland({ sessionId }: { sessionId: string | null }) {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [applying, setApplying] = useState<string | null>(null);
  const [result, setResult] = useState<{ jobs: RankedJob[] } | null>(null);
  const [applications, setApplications] = useState<Record<string, Application>>({});
  const [error, setError] = useState("");

  async function handleSearch() {
    if (!query.trim()) return;
    setLoading(true); setError(""); setResult(null);
    try {
      const res = await fetch("/api/search", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: query.trim() }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || `search failed (${res.status})`);
      }
      setResult(await res.json());
    } catch (e) {
      setError(e instanceof Error ? e.message : "search failed");
    } finally { setLoading(false); }
  }

  async function handleApply(jobUrl: string) {
    if (!sessionId) { setError("请先在左侧上传简历"); return; }
    setApplying(jobUrl);
    try {
      const res = await fetch("/api/apply", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ job_url: jobUrl, session_id: sessionId }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || "apply failed");
      }
      const app: Application = await res.json();
      setApplications((prev) => ({ ...prev, [jobUrl]: app }));
    } catch (e) {
      setError(e instanceof Error ? e.message : "apply failed");
    } finally { setApplying(null); }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Search bar */}
      <div className="flex gap-2 mb-4">
        <input
          className="flex-1 px-4 py-2.5 rounded-lg border border-gray-300 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="描述你想要的职位，如：找北京的前端岗位，3-5年经验，薪资25K以上"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          disabled={loading}
        />
        <button
          className="px-5 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
          onClick={handleSearch}
          disabled={loading || !query.trim()}
        >
          {loading ? "搜索中..." : "搜索"}
        </button>
      </div>

      {error && (
        <div className="p-3 mb-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">{error}</div>
      )}

      {/* Results */}
      <div className="flex-1 overflow-y-auto space-y-3">
        {loading && (
          <div className="flex items-center gap-2 text-sm text-gray-500 py-8 justify-center">
            <div className="animate-spin h-4 w-4 border-2 border-blue-600 border-t-transparent rounded-full" />
            AI 正在分析需求并搜索职位...
          </div>
        )}

        {result && result.jobs?.length === 0 && (
          <div className="p-8 text-center text-sm text-gray-400 bg-white rounded-lg border">未找到匹配职位</div>
        )}

        {result?.jobs?.map((job, i) => (
          <div key={i} className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-sm transition-shadow">
            <div className="flex items-start justify-between mb-2">
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-gray-900 truncate">{job.title}</h3>
                <p className="text-xs text-gray-500 mt-0.5">
                  {job.company}{job.location && ` · ${job.location}`}{job.salary && ` · ${job.salary}`}
                </p>
              </div>
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center ml-2">
                <span className="text-sm font-bold text-blue-600">{job.match_score}</span>
              </div>
            </div>

            <p className="text-xs text-gray-600 mb-2">{job.summary}</p>

            {job.highlights?.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-3">
                {job.highlights.map((h, j) => (
                  <span key={j} className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded-full">{h}</span>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              {job.url && (
                <a href={job.url} target="_blank" rel="noopener noreferrer"
                   className="text-xs text-blue-600 hover:underline py-1">查看详情 →</a>
              )}
              <button
                className="ml-auto px-3 py-1 text-xs font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
                onClick={() => handleApply(job.url)}
                disabled={applying === job.url}
              >
                {applying === job.url ? "生成中..." : "生成申请"}
              </button>
            </div>

            {applications[job.url] && <ApplicationCard app={applications[job.url]} />}
          </div>
        ))}
      </div>
    </div>
  );
}

// ---- Main Page ----
export default function Home() {
  const [sessionId, setSessionId] = useState<string | null>(null);

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-3">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-gray-900">JobLinker</h1>
            <p className="text-xs text-gray-500">AI 招聘助手</p>
          </div>
          {sessionId && (
            <span className="text-xs text-green-600 bg-green-50 px-2 py-1 rounded-full">简历已就绪</span>
          )}
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto w-full p-4 gap-4 grid grid-cols-1 lg:grid-cols-5">
        <div className="lg:col-span-2 min-h-[400px]">
          <ResumeChatIsland onSessionReady={setSessionId} />
        </div>
        <div className="lg:col-span-3 min-h-[400px]">
          <SearchIsland sessionId={sessionId} />
        </div>
      </main>
    </div>
  );
}
