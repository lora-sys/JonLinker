'use client';

import { useMatchStore } from '@/stores/match';
import Card from '@/components/ui/Card';
import ScoreBar from '@/components/ui/ScoreBar';
import Button from '@/components/ui/Button';

export default function MatchesPage() {
  const { matches } = useMatchStore();

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Matches</h1>
        {matches.length === 0 ? (
          <Card className="p-10 text-center bg-white/80 backdrop-blur-xl border border-white/20">
            <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">No matches yet</h2>
            <p className="text-slate-500">Create an agent and add jobs or resumes to start matching</p>
          </Card>
        ) : (
          <div className="space-y-4">
            {matches.map((match) => (
              <Card key={match.id} hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 group">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <p className="font-semibold text-slate-900">Match #{match.id.slice(0, 8)}</p>
                    <p className="text-sm text-slate-500 mt-1">
                      Seeker: {match.seeker_agent_id?.slice(0, 8) || 'N/A'} • Job: {match.job_id?.slice(0, 8) || 'N/A'}
                    </p>
                  </div>
                  <span className={`px-3 py-1 text-xs font-semibold rounded-full ${
                    match.status === 'mutual_interest' ? 'bg-green-100 text-green-700' :
                    match.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-slate-200 text-slate-600'
                  }`}>
                    {match.status.replace('_', ' ')}
                  </span>
                </div>

                {/* Match Score Bar */}
                <div className="mb-4">
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-slate-600">Match Score</span>
                    <span className="font-medium text-slate-900">{(match.score * 100).toFixed(1)}%</span>
                  </div>
                  <ScoreBar score={match.score * 100} size="md" />
                </div>

                {/* Action Buttons - Show on Hover */}
                <div className="flex gap-3 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                  <Button variant="primary" size="sm" className="flex-1">
                    Confirm
                  </Button>
                  <Button variant="ghost" size="sm" className="flex-1">
                    Decline
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}