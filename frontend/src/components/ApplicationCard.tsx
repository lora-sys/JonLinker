"use client";

import { useState } from "react";
import type { Application } from "@/lib/types";
import { MessageResponse } from "@/components/ai-elements/message";

export function ApplicationCard({ app }: { app: Application }) {
  const [tab, setTab] = useState<"letter" | "resume">("letter");
  return (
    <div className="mt-4 bg-card rounded-xl border border-blue-200 overflow-hidden shadow-sm">
      <div className="bg-gradient-to-r from-blue-50 to-indigo-50 px-5 py-4 border-b border-blue-100">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-foreground">申请已生成</h3>
            <p className="text-sm text-muted-foreground mt-0.5">{app.job_title} · {app.company}</p>
          </div>
          <div className="flex gap-1.5">
            {app.highlights.slice(0, 3).map((h, i) => (
              <span key={i} className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">{h}</span>
            ))}
          </div>
        </div>
      </div>

      <div className="flex border-b border-border">
        <button
          className={`flex-1 py-2.5 text-sm font-medium transition-colors ${tab === "letter" ? "text-blue-600 border-b-2 border-blue-600" : "text-muted-foreground hover:text-foreground"}`}
          onClick={() => setTab("letter")}
        >求职信</button>
        <button
          className={`flex-1 py-2.5 text-sm font-medium transition-colors ${tab === "resume" ? "text-blue-600 border-b-2 border-blue-600" : "text-muted-foreground hover:text-foreground"}`}
          onClick={() => setTab("resume")}
        >定制简历</button>
      </div>

      <div className="p-5 max-h-80 overflow-y-auto">
        <MessageResponse>
          {tab === "letter" ? app.cover_letter : app.resume_md}
        </MessageResponse>
      </div>

      <div className="px-5 py-3 bg-muted/30 border-t border-border text-xs text-muted-foreground">
        生成时间：{new Date(app.generated_at).toLocaleString("zh-CN")}
      </div>
    </div>
  );
}
