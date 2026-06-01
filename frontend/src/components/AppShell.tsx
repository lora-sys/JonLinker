"use client";

import { useState } from "react";
import { ResumeChatIsland } from "@/components/ResumeChatIsland";
import { SearchIsland } from "@/components/SearchIsland";

export function AppShell() {
  const [sessionId, setSessionId] = useState<string | null>(null);

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <header className="bg-card border-b px-6 py-3">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-foreground">JobLinker</h1>
            <p className="text-xs text-muted-foreground">AI 招聘助手</p>
          </div>
          {sessionId && (
            <span className="text-xs text-green-600 bg-green-50 px-2 py-1 rounded-full">简历已就绪</span>
          )}
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto w-full p-4 gap-4 grid grid-cols-1 lg:grid-cols-5 lg:grid-rows-1 min-h-0">
        <div className="lg:col-span-2 flex flex-col min-h-0">
          <ResumeChatIsland onSessionReady={setSessionId} />
        </div>
        <div className="lg:col-span-3 flex flex-col min-h-0">
          <SearchIsland sessionId={sessionId} />
        </div>
      </main>
    </div>
  );
}
