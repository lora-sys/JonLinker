import type { Offer } from '@/types';
import { DollarSign } from 'lucide-react';

interface OfferCardProps {
  offer: Offer;
  onAccept?: () => void;
  onDecline?: () => void;
}

export function OfferCard({ offer, onAccept, onDecline }: OfferCardProps) {
  const formatCurrency = (value: number, currency = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const statusColors: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    negotiating: 'bg-blue-100 text-blue-700 border-blue-200',
    accepted: 'bg-green-100 text-green-700 border-green-200',
    declined: 'bg-slate-100 text-slate-600 border-slate-200',
    expired: 'bg-red-100 text-red-700 border-red-200',
  };

  const statusLabels: Record<string, string> = {
    pending: 'Pending',
    negotiating: 'Negotiating',
    accepted: 'Accepted',
    declined: 'Declined',
    expired: 'Expired',
  };

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium border ${statusColors[offer.status] || 'bg-slate-100 text-slate-600 border-slate-200'}`}>
              {statusLabels[offer.status] || offer.status}
            </span>
          </div>

          <div className="space-y-1">
            <p className="text-2xl font-bold text-slate-900">
              {formatCurrency(offer.compensation.base_salary, offer.compensation.currency)}
            </p>
            <p className="text-sm text-slate-500">
              Base Salary • Start: {offer.start_date || 'TBD'}
            </p>
          </div>
        </div>
      </div>

      {offer.compensation.bonus && (
        <div className="mt-3 flex gap-4">
          {offer.compensation.bonus.amount > 0 && (
            <div className="text-sm">
              <span className="text-slate-500">Bonus:</span>{' '}
              <span className="font-medium text-slate-700">
                {formatCurrency(offer.compensation.bonus.amount, offer.compensation.currency)}
              </span>
            </div>
          )}
          {offer.compensation.equity && (
            <div className="text-sm">
              <span className="text-slate-500">Equity:</span>{' '}
              <span className="font-medium text-slate-700">
                {offer.compensation.equity.shares} shares ({offer.compensation.equity.vesting_period})
              </span>
            </div>
          )}
        </div>
      )}

      <div className="flex gap-2 mt-4">
        {offer.status === 'pending' && (
          <>
            {onAccept && (
              <button
                onClick={onAccept}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium"
              >
                Accept
              </button>
            )}
            {onDecline && (
              <button
                onClick={onDecline}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 transition-colors text-sm font-medium"
              >
                Decline
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
}

export default OfferCard;