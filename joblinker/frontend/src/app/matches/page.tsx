'use client';

import { useMatchStore } from '@/stores/match';

export default function MatchesPage() {
  const { matches } = useMatchStore();

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Matches</h1>
      {matches.length === 0 ? (
        <div className="bg-white rounded-xl p-8 shadow-sm border border-gray-200 text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </div>
          <h2 className="text-lg font-medium text-gray-900 mb-2">No matches yet</h2>
          <p className="text-gray-500">Create an agent and add jobs or resumes to start matching</p>
        </div>
      ) : (
        <div className="space-y-4">
          {matches.map((match) => (
            <div key={match.id} className="bg-white rounded-xl p-4 shadow-sm border border-gray-200">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-900">Match #{match.id.slice(0, 8)}</p>
                  <p className="text-sm text-gray-500">Score: {(match.score * 100).toFixed(1)}%</p>
                </div>
                <span className="px-3 py-1 text-sm font-medium rounded-full bg-blue-100 text-blue-700">
                  {match.status.replace('_', ' ')}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
