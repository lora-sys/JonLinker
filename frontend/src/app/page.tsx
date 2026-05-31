"use client";

import { useState } from "react";

type Job = {
  title: string;
  company: string;
  location: string;
  salary: string;
  url: string;
  description: string;
  tags: string[];
  source: string;
};

type RankedJob = Job & {
  match_score: number;
  summary: string;
  highlights: string[];
};

type SearchResponse = {
  jobs: RankedJob[];
};

export default function Home() {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SearchResponse | null>(null);
  const [error, setError] = useState("");

  async function handleSearch() {
    if (!query.trim()) return;
    setLoading(true);
    setError("");
    setResult(null);
    try {
      const res = await fetch("/api/search", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: query.trim() }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || `search failed (${res.status})`);
      }
      const data = await res.json();
      setResult(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "search failed");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") handleSearch();
  }

  return (
    <div className="flex flex-col items-center min-h-screen bg-gray-50 p-4">
      <header className="w-full max-w-3xl pt-12 pb-8 text-center">
        <h1 className="text-3xl font-bold text-gray-900">JobLinker</h1>
        <p className="text-gray-500 mt-2">AI 招聘助手 · 智能职位搜索</p>
      </header>

      <div className="w-full max-w-2xl flex gap-2 mb-8">
        <input
          className="flex-1 px-4 py-3 rounded-lg border border-gray-300 text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="描述你想要的职位，如：找北京的前端岗位，3-5年经验，薪资25K以上"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={loading}
        />
        <button
          className="px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          onClick={handleSearch}
          disabled={loading || !query.trim()}
        >
          {loading ? "搜索中..." : "搜索"}
        </button>
      </div>

      {error && (
        <div className="w-full max-w-2xl p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 mb-4">
          {error}
        </div>
      )}

      {loading && (
        <div className="flex items-center gap-3 text-gray-500 py-8">
          <div className="animate-spin h-5 w-5 border-2 border-blue-600 border-t-transparent rounded-full" />
          <span>AI 正在分析你的需求并搜索职位...</span>
        </div>
      )}

      {result && result.jobs && (
        <div className="w-full max-w-2xl space-y-4">
          <p className="text-sm text-gray-500">
            找到 {result.jobs.length} 个匹配职位
          </p>
          {result.jobs.map((job, i) => (
            <JobCard key={i} job={job} />
          ))}
        </div>
      )}

      {result && (!result.jobs || result.jobs.length === 0) && (
        <div className="w-full max-w-2xl p-8 text-center text-gray-400 bg-white rounded-lg border">
          未找到匹配职位，请尝试其他关键词
        </div>
      )}
    </div>
  );
}

function JobCard({ job }: { job: RankedJob }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-5 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-2">
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-semibold text-gray-900 truncate">
            {job.title}
          </h2>
          <p className="text-sm text-gray-600 mt-0.5">
            {job.company}
            {job.location && <span> · {job.location}</span>}
            {job.salary && <span> · {job.salary}</span>}
          </p>
        </div>
        <div className="flex-shrink-0 ml-3">
          <div className="w-12 h-12 rounded-full bg-blue-50 flex items-center justify-center">
            <span className="text-lg font-bold text-blue-600">
              {job.match_score}
            </span>
          </div>
        </div>
      </div>

      <p className="text-sm text-gray-700 mb-2">{job.summary}</p>

      {job.highlights && job.highlights.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {job.highlights.map((h, j) => (
            <span
              key={j}
              className="inline-block px-2 py-0.5 text-xs font-medium bg-green-50 text-green-700 rounded-full"
            >
              {h}
            </span>
          ))}
        </div>
      )}

      {job.url && (
        <a
          href={job.url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-block mt-3 text-sm text-blue-600 hover:underline"
        >
          查看详情 →
        </a>
      )}
    </div>
  );
}
