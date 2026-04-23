'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAgent } from '@/hooks/useAgent';

export default function AgentCreatePage() {
  const router = useRouter();
  const { createAgent, isLoading, error } = useAgent();
  const [agentType, setAgentType] = useState<'seeker' | 'recruiter'>('seeker');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createAgent({ type: agentType });
      router.push('/dashboard');
    } catch {
      // Error is handled by the hook
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Create Agent</h1>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Agent Type
          </label>
          <div className="grid grid-cols-2 gap-4">
            <button
              type="button"
              onClick={() => setAgentType('seeker')}
              className={`p-4 rounded-lg border-2 transition-colors ${
                agentType === 'seeker'
                  ? 'border-blue-600 bg-blue-50 text-blue-700'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300'
              }`}
            >
              <div className="font-medium">Job Seeker</div>
              <div className="text-sm opacity-75">Looking for opportunities</div>
            </button>
            <button
              type="button"
              onClick={() => setAgentType('recruiter')}
              className={`p-4 rounded-lg border-2 transition-colors ${
                agentType === 'recruiter'
                  ? 'border-blue-600 bg-blue-50 text-blue-700'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300'
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
          className="w-full py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isLoading ? 'Creating...' : 'Create Agent'}
        </button>
      </form>
    </div>
  );
}
