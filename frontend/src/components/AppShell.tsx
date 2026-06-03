"use client";

import { ErrorBoundary } from "@/components/ErrorBoundary";
import { UnifiedChatIsland } from "@/components/UnifiedChatIsland";

export function AppShell() {
  return (
    <div className="flex flex-col min-h-screen bg-background">
      <header className="bg-card border-b px-6 py-3">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-lg font-bold text-foreground">JobLinker</h1>
            <p className="text-xs text-muted-foreground">AI 招聘助手</p>
          </div>
        </div>
      </header>
      <main className="flex-1 max-w-4xl mx-auto w-full p-4 min-h-0">
        <ErrorBoundary>
          <UnifiedChatIsland />
        </ErrorBoundary>
      </main>
    </div>
  );
}
