'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth';
import WSDebugPanel from '@/components/ws-debug-panel';
import { BackgroundBeams } from '@/components/ui/background-beams';
import { Activity } from 'lucide-react';

export default function WSTestPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace('/login');
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-zinc-950 relative overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 bg-gradient-to-br from-cyan-950/20 via-zinc-950 to-zinc-950 pointer-events-none" />
      <BackgroundBeams />

      {/* Header */}
      <div className="relative z-10 pt-8 px-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-10 h-10 rounded-lg bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center">
            <Activity className="h-5 w-5 text-cyan-400" />
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">
            WebSocket Test
          </h1>
        </div>
        <p className="text-neutral-400 text-sm">
          Test bidirectional WebSocket communication with the server
        </p>
      </div>

      {/* Debug Panel */}
      <div className="relative z-10 p-8 h-[calc(100vh-200px)]">
        <div className="h-full max-w-3xl mx-auto bg-zinc-900/50 backdrop-blur-sm rounded-xl border border-zinc-800 overflow-hidden shadow-[0_0_60px_-15px_rgba(6,182,212,0.15)]">
          <WSDebugPanel />
        </div>
      </div>
    </div>
  );
}