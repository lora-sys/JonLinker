import type { Match } from '@/types';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

interface MatchCardProps {
  match: Match;
  onClick?: () => void;
}

export function MatchCard({ match, onClick }: MatchCardProps) {
  const scorePercent = Math.round((match.score || 0) * 100);

  const getScoreColor = (score: number) => {
    if (score >= 0.8) return 'bg-emerald-100 text-emerald-700';
    if (score >= 0.6) return 'bg-yellow-100 text-yellow-700';
    return 'bg-red-100 text-red-700';
  };

  const statusLabels: Record<string, string> = {
    active: 'Active',
    pending: 'Pending',
    accepted: 'Accepted',
    declined: 'Declined',
    expired: 'Expired',
  };

  return (
    <div
      className="bg-white rounded-xl border border-slate-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${getScoreColor(match.score || 0)}`}>
              {scorePercent}% Match
            </span>
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200">
              {statusLabels[match.status] || match.status}
            </span>
          </div>

          {match.job && (
            <p className="text-sm text-slate-600 mt-2 truncate">
              {match.job.structured?.title || 'Unknown Position'}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center justify-between mt-3">
        <span className="text-xs text-slate-400">
          Matched {match.created_at ? new Date(match.created_at).toLocaleDateString() : 'recently'}
        </span>
        <Link
          href={`/conversation/${match.id}`}
          className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700 font-medium"
        >
          Chat <ArrowRight className="w-3 h-3" />
        </Link>
      </div>
    </div>
  );
}

export default MatchCard;