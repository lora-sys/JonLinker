'use client';

import { Card } from '@/components/ui/Card';
import type { Offer as DomainOffer } from '@/types';

type OfferCardProps = {
  offer: DomainOffer;
  onRespond?: (id: string, response: 'accept' | 'decline' | 'negotiate') => void;
};

export function OfferCard({ offer, onRespond }: OfferCardProps) {
  const formatCurrency = (amount: number, currency = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      });
    } catch {
      return dateStr;
    }
  };

  const statusColors: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    accepted: 'bg-green-100 text-green-800',
    declined: 'bg-red-100 text-red-800',
    negotiating: 'bg-blue-100 text-blue-800',
    withdrawn: 'bg-gray-100 text-gray-800',
  };

  const comp = offer.compensation;

  return (
    <Card variant="outlined" padding="lg" className="hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">
            {offer.match?.job?.structured?.title || 'Offer'}
          </h3>
          <p className="text-sm text-gray-500 mt-1">
            Received {formatDate(offer.created_at)}
          </p>
        </div>
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[offer.status] || 'bg-gray-100 text-gray-800'}`}>
          {offer.status}
        </span>
      </div>

      <div className="space-y-3 mb-4">
        <div className="flex justify-between items-center py-2 border-b border-gray-100">
          <span className="text-sm text-gray-600">Base Salary</span>
          <span className="font-semibold text-gray-900">
            {formatCurrency(comp.base_salary, comp.currency)}
          </span>
        </div>

        {comp.bonus && comp.bonus.amount > 0 && (
          <div className="flex justify-between items-center py-2 border-b border-gray-100">
            <span className="text-sm text-gray-600">{comp.bonus.description || 'Bonus'}</span>
            <span className="font-semibold text-gray-900">
              {formatCurrency(comp.bonus.amount, comp.currency)}
            </span>
          </div>
        )}

        {comp.equity && comp.equity.shares > 0 && (
          <div className="flex justify-between items-center py-2 border-b border-gray-100">
            <span className="text-sm text-gray-600">Equity ({comp.equity.vesting_period})</span>
            <span className="font-semibold text-gray-900">
              {comp.equity.shares.toLocaleString()} shares
            </span>
          </div>
        )}
      </div>

      <div className="mb-4">
        <p className="text-sm text-gray-600">
          <span className="font-medium">Start Date:</span> {formatDate(offer.start_date)}
        </p>
      </div>

      {comp.benefits && comp.benefits.length > 0 && (
        <div className="mb-4">
          <p className="text-sm font-medium text-gray-700 mb-2">Benefits</p>
          <div className="flex flex-wrap gap-2">
            {comp.benefits.map((benefit, idx) => (
              <span key={idx} className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700">
                {benefit}
              </span>
            ))}
          </div>
        </div>
      )}

      {offer.status === 'pending' && onRespond && (
        <div className="flex items-center gap-3 pt-4 border-t border-gray-100">
          <button
            onClick={() => onRespond(offer.id, 'accept')}
            className="flex-1 px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 transition-colors"
          >
            Accept
          </button>
          <button
            onClick={() => onRespond(offer.id, 'negotiate')}
            className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
          >
            Negotiate
          </button>
          <button
            onClick={() => onRespond(offer.id, 'decline')}
            className="flex-1 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors"
          >
            Decline
          </button>
        </div>
      )}
    </Card>
  );
}
