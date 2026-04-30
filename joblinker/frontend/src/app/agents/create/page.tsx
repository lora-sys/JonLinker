'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Sparkles, FileText } from 'lucide-react';

export default function AgentCreatePage() {
  const router = useRouter();
  const [agentType, setAgentType] = useState<'seeker' | 'recruiter'>('seeker');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      const response = await fetch('/api/agents', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type: agentType }),
      });
      if (!response.ok) throw new Error('Failed to create agent');
      const data = await response.json();
      if (data?.id) {
        router.push('/agents');
      } else {
        throw new Error('Failed to create agent');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create agent');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Create Agent</h1>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-3">
              Agent Type
            </label>
            <div className="grid grid-cols-2 gap-4">
              <button
                type="button"
                onClick={() => setAgentType('seeker')}
                className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                  agentType === 'seeker'
                    ? 'border-blue-600 bg-blue-50 text-blue-700'
                    : 'border-slate-200 text-slate-600 hover:border-blue-300 hover:bg-blue-50'
                }`}
              >
                <div className="font-medium">Job Seeker</div>
                <div className="text-sm opacity-75">Looking for opportunities</div>
              </button>
              <button
                type="button"
                onClick={() => setAgentType('recruiter')}
                className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                  agentType === 'recruiter'
                    ? 'border-blue-600 bg-blue-50 text-blue-700'
                    : 'border-slate-200 text-slate-600 hover:border-blue-300 hover:bg-blue-50'
                }`}
              >
                <div className="font-medium">Recruiter</div>
                <div className="text-sm opacity-75">Hiring talent</div>
              </button>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-sky-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg shadow-blue-600/25 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            {isLoading ? 'Creating...' : 'Create Agent'}
          </button>
        </form>

        {/* AI Resume Generation */}
        {agentType === 'seeker' && (
          <div className="mt-6 bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200 rounded-xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-slate-900 flex items-center gap-2">
                  <Sparkles className="w-5 h-5 text-amber-500" />
                  Don&apos;t have a resume?
                </h3>
                <p className="text-slate-600 text-sm mt-1">
                  Let our AI generate a professional resume for you in seconds
                </p>
              </div>
              <Link
                href="/resume/generate"
                className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-lg font-medium hover:from-amber-600 hover:to-orange-600 transition-all duration-200 shadow-lg shadow-amber-500/25"
              >
                <FileText className="w-4 h-4" />
                Generate with AI
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}