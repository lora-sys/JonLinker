'use client';

import Card from '@/components/ui/Card';
import Badge from '@/components/ui/Badge';
import Button from '@/components/ui/Button';
import CompensationCard from '@/components/ui/CompensationCard';
import NegotiateSlider from '@/components/ui/NegotiateSlider';
import Countdown from '@/components/ui/Countdown';
import { useState } from 'react';

interface Offer {
  id: string;
  company: string;
  role: string;
  expiresAt: Date;
  compensation: {
    base: number;
    bonus?: number;
    equity?: number;
  };
}

const MOCK_OFFERS: Offer[] = [
  {
    id: '1',
    company: 'TechCorp',
    role: 'Senior React Developer',
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    compensation: {
      base: 140000,
      bonus: 15000,
      equity: 20000,
    },
  },
];

export default function OffersPage() {
  const [negotiateValue, setNegotiateValue] = useState(140000);
  const [showNegotiate, setShowNegotiate] = useState<string | null>(null);

  return (
    <div className="max-w-4xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold text-slate-900 mb-6" style={{ fontFamily: 'Plus Jakarta Sans, sans-serif' }}>Offers</h1>

      {MOCK_OFFERS.length === 0 ? (
        <Card className="p-10 text-center">
          <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h2 className="text-lg font-semibold text-slate-900 mb-2">No offers yet</h2>
          <p className="text-slate-500">Offers from recruiters will appear here</p>
        </Card>
      ) : (
        <div className="space-y-6">
          {MOCK_OFFERS.map((offer) => (
            <Card key={offer.id} className="p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="font-semibold text-slate-900">{offer.role}</h3>
                  <p className="text-sm text-slate-600">{offer.company}</p>
                </div>
                <Badge status="negotiating">Active</Badge>
              </div>

              {/* Expiration Countdown */}
              <div className="flex items-center gap-2 text-sm text-slate-500 mb-4 p-3 bg-slate-50 rounded-lg">
                <svg className="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <span>Expires in</span>
                <Countdown targetDate={offer.expiresAt} showLabels />
              </div>

              {/* Compensation Breakdown */}
              <CompensationCard
                breakdown={offer.compensation}
                className="mb-4"
              />

              {/* Negotiate Slider */}
              {showNegotiate === offer.id && (
                <div className="mb-4 p-4 bg-blue-50 rounded-lg">
                  <NegotiateSlider
                    min={offer.compensation.base * 0.8}
                    max={offer.compensation.base * 1.3}
                    step={5000}
                    value={negotiateValue}
                    onChange={setNegotiateValue}
                  />
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex gap-3">
                <Button variant="primary" className="flex-1">
                  Accept
                </Button>
                <Button
                  variant="secondary"
                  className="flex-1"
                  onClick={() => setShowNegotiate(showNegotiate === offer.id ? null : offer.id)}
                >
                  {showNegotiate === offer.id ? 'Cancel' : 'Negotiate'}
                </Button>
                <Button variant="ghost" className="flex-1">
                  Decline
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
