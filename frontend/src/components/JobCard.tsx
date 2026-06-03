"use client";

import type { RankedJob } from "@/lib/types";

export function JobCard({
  job, sessionId, profileReady, applying, onApply,
}: {
  job: RankedJob;
  sessionId: string | null;
  profileReady: boolean;
  applying: string | null;
  onApply: (job: RankedJob) => void;
}) {
  return (
    <div className="bg-card rounded-xl border p-4 hover:shadow-sm transition-shadow">
      <div className="flex items-start justify-between mb-2">
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-foreground truncate">{job.title}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {job.company}{job.location && ` · ${job.location}`}{job.salary && ` · ${job.salary}`}
          </p>
        </div>
        <div className="flex-shrink-0 w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center ml-2">
          <span className="text-sm font-bold text-blue-600">{job.match_score}</span>
        </div>
      </div>
      <p className="text-xs text-muted-foreground mb-2">{job.summary}</p>
      {job.highlights?.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3">
            {job.highlights.map((h) => (
            <span key={h} className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded-full">{h}</span>
          ))}
        </div>
      )}
      <div className="flex gap-2">
        {job.url && (
          <a href={job.url} target="_blank" rel="noopener noreferrer"
             className="text-xs text-blue-600 hover:underline py-1">查看详情 →</a>
        )}
        <button
          className="ml-auto px-3 py-1 text-xs font-medium bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity"
          onClick={() => onApply(job)}
          disabled={applying === job.url || !sessionId || !profileReady}
          title={!sessionId ? "请先上传简历" : !profileReady ? "请先完善个人资料" : "生成申请"}
        >
          {applying === job.url ? "生成中..." : !sessionId ? "需先上传简历" : !profileReady ? "需先完成资料" : "生成申请"}
        </button>
      </div>
    </div>
  );
}
