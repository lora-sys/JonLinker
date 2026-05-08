'use client';

import { useEffect, useState, useCallback } from 'react';
import { FileText, AlertTriangle, DollarSign, TrendingUp, Scale, Building2, ChevronDown, Check } from 'lucide-react';
import { apiClient } from '@/lib/api_client';
import { motion, AnimatePresence } from 'framer-motion';
import Card from '@/components/ui/Card';
import Badge from '@/components/ui/Badge';
import Button from '@/components/ui/Button';
import CompensationCard from '@/components/ui/CompensationCard';
import NegotiateSlider from '@/components/ui/NegotiateSlider';
import CountdownDisplay from '@/components/ui/Countdown';
import { LoadingSkeleton, ErrorState, EmptyState, RevealSection } from '@/components/ui';
import type { Offer, Compensation } from '@/types';

type OfferWithParsed = Omit<Offer, 'compensation'> & {
  compensation: Compensation;
};

function OfferCardInner({ offer, index, onUpdate }: { offer: OfferWithParsed; index: number; onUpdate: () => void }) {
  const [negotiateValue, setNegotiateValue] = useState(offer.compensation.base_salary);
  const [showNegotiate, setShowNegotiate] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);

  const statusColors: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    negotiating: 'bg-blue-100 text-blue-700 border-blue-200',
    accepted: 'bg-green-100 text-green-700 border-green-200',
    declined: 'bg-slate-100 text-slate-600 border-slate-200',
    expired: 'bg-red-100 text-red-700 border-red-200',
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: offer.compensation.currency || 'USD',
      maximumFractionDigits: 0,
    }).format(value);
  };

  const handleAccept = async () => {
    try {
      setIsUpdating(true);
      await apiClient.patch(`/api/offers/${offer.id}`, { status: 'accepted' });
      onUpdate();
    } catch (err) {
      console.error('Failed to accept offer:', err);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDecline = async () => {
    try {
      setIsUpdating(true);
      await apiClient.patch(`/api/offers/${offer.id}`, { status: 'declined' });
      onUpdate();
    } catch (err) {
      console.error('Failed to decline offer:', err);
    } finally {
      setIsUpdating(false);
    }
  };

  const totalComp = offer.compensation.base_salary +
    (offer.compensation.bonus?.amount || 0) +
    ((offer.compensation.equity?.shares || 0) * 100);

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.15 }}
      className="relative"
    >
      <Card hover className="overflow-hidden bg-white/80 backdrop-blur-xl border border-white/20 hover:shadow-xl transition-all duration-300">
        <div
          className="p-6 cursor-pointer"
          onClick={() => setExpanded(!expanded)}
        >
          <div className="flex items-start justify-between mb-4">
            <div>
              <h3 className="font-bold text-slate-900 text-xl">{offer.match?.job?.structured?.title || 'Offer'}</h3>
              <p className="text-slate-600 mt-0.5">{offer.match?.job?.structured?.title || 'Position'}</p>
            </div>
            <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${statusColors[offer.status] || 'bg-slate-100 text-slate-600 border-slate-200'}`}>
              {offer.status}
            </span>
          </div>

          <div className="flex items-center gap-6">
            <div>
              <p className="text-sm text-slate-500">Base Salary</p>
              <p className="text-lg font-semibold text-slate-900 flex items-center gap-1">
                <DollarSign className="w-4 h-4 text-green-600" />
                {formatCurrency(offer.compensation.base_salary)}
              </p>
            </div>
            <div className="w-px h-10 bg-slate-200" />
            <div>
              <p className="text-sm text-slate-500">Total Comp</p>
              <p className="text-lg font-semibold text-green-600 flex items-center gap-1">
                <TrendingUp className="w-4 h-4" />
                {formatCurrency(totalComp)}
              </p>
            </div>
            <div className="flex-1 flex justify-end">
              <ChevronDown className={`w-6 h-6 text-slate-400 transition-transform duration-300 ${expanded ? 'rotate-180' : ''}`} />
            </div>
          </div>
        </div>

        <AnimatePresence>
          {expanded && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.3 }}
              className="border-t border-slate-100"
            >
              <div className="p-6 space-y-5">
                {offer.status === 'pending' && offer.expires_at && (
                  <div className="flex items-center gap-3 p-4 bg-gradient-to-r from-amber-50 to-orange-50 rounded-xl border border-amber-100">
                    <div className="w-10 h-10 rounded-lg bg-amber-100 flex items-center justify-center">
                      <AlertTriangle className="w-5 h-5 text-amber-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-amber-800">Offer expires soon</p>
                      <div className="text-sm text-amber-600 flex items-center gap-1">
                        Expires in <CountdownDisplay targetDate={new Date(offer.expires_at)} showLabels />
                      </div>
                    </div>
                  </div>
                )}

                <div>
                  <h4 className="text-sm font-semibold text-slate-900 mb-3 flex items-center gap-2">
                    <FileText className="w-4 h-4" />
                    Compensation Breakdown
                  </h4>
                  <CompensationCard
                    breakdown={{
                      base: offer.compensation.base_salary,
                      bonus: offer.compensation.bonus?.amount,
                      equity: offer.compensation.equity?.shares,
                      benefits: offer.compensation.benefits?.length,
                    }}
                  />
                </div>

                <AnimatePresence>
                  {showNegotiate && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="p-4 bg-blue-50 rounded-xl"
                    >
                      <div className="flex items-center justify-between mb-3">
                        <h4 className="text-sm font-semibold text-slate-900">Negotiate Base Salary</h4>
                        <span className="text-lg font-bold text-blue-600">{formatCurrency(negotiateValue)}</span>
                      </div>
                      <NegotiateSlider
                        min={offer.compensation.base_salary * 0.8}
                        max={offer.compensation.base_salary * 1.3}
                        step={5000}
                        value={negotiateValue}
                        onChange={setNegotiateValue}
                      />
                      <div className="flex justify-between text-xs text-slate-500 mt-2">
                        <span>{(offer.compensation.base_salary * 0.8 / 1000).toFixed(0)}k</span>
                        <span>{(offer.compensation.base_salary * 1.3 / 1000).toFixed(0)}k</span>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>

                {(offer.status === 'pending' || offer.status === 'negotiating') && (
                  <div className="flex gap-3 pt-2">
                    <Button
                      variant="primary"
                      className="flex-1 bg-gradient-to-r from-green-600 to-emerald-500 hover:from-green-700 hover:to-emerald-600"
                      onClick={handleAccept}
                      disabled={isUpdating}
                    >
                      <Check className="w-4 h-4 mr-2" />
                      Accept Offer
                    </Button>
                    <Button
                      variant="secondary"
                      className="flex-1"
                      onClick={() => setShowNegotiate(!showNegotiate)}
                    >
                      {showNegotiate ? 'Cancel' : 'Negotiate'}
                    </Button>
                    <Button
                      variant="ghost"
                      className="flex-1 text-red-500 hover:text-red-600 hover:bg-red-50"
                      onClick={handleDecline}
                      disabled={isUpdating}
                    >
                      Decline
                    </Button>
                  </div>
                )}
                {offer.status === 'expired' && (
                  <div className="flex items-center justify-center p-4 bg-red-50 rounded-xl border border-red-100">
                    <p className="text-red-600 font-medium">This offer has expired</p>
                  </div>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </Card>
    </motion.div>
  );
}

export function OffersContent({ initialOffers }: { initialOffers: OfferWithParsed[] }) {
  const [offers, setOffers] = useState<OfferWithParsed[]>(initialOffers);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchOffers = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await apiClient.get<{offers: OfferWithParsed[]}>('/api/offers');
      const parsedOffers = (data?.offers || []).map(offer => ({
        ...offer,
        compensation: typeof offer.compensation === 'string'
          ? JSON.parse(offer.compensation)
          : offer.compensation,
      }));
      setOffers(parsedOffers);
    } catch (err) {
      setError(null);
      setOffers([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const activeOffers = offers.filter(o => o.status === 'pending' || o.status === 'negotiating').length;

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <RevealSection>
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-4xl font-bold text-slate-900 flex items-center gap-3">
                <FileText className="w-9 h-9 text-blue-600" />
                Offers
              </h1>
              <p className="text-slate-600 mt-1">Review and respond to your offers</p>
            </div>
            <div className="flex items-center gap-3">
              <div className="px-4 py-2 bg-green-50 rounded-xl border border-green-100">
                <span className="text-sm text-green-600 font-medium">{activeOffers} active</span>
              </div>
            </div>
          </div>
        </RevealSection>

        {isLoading ? (
          <LoadingSkeleton count={3} variant="card" />
        ) : error ? (
          <ErrorState message={error} onRetry={fetchOffers} />
        ) : offers.length === 0 ? (
          <EmptyState
            icon={<FileText className="w-10 h-10 text-blue-600" />}
            title="No offers yet"
            description="Offers from recruiters will appear here when you match with opportunities"
          />
        ) : (
          <div className="space-y-6">
            {offers.map((offer, index) => (
              <OfferCardInner key={offer.id} offer={offer} index={index} onUpdate={fetchOffers} />
            ))}
          </div>
        )}

        <RevealSection delay={3}>
          <div className="mt-16">
            <h2 className="text-xl font-semibold text-slate-900 mb-4 flex items-center gap-2">
              <Scale className="w-5 h-5 text-blue-600" />
              Tips for Evaluating Offers
            </h2>
            <div className="grid md:grid-cols-3 gap-4">
              {[
                { icon: DollarSign, title: 'Consider Total Comp', desc: 'Look beyond base salary — bonus, equity, and benefits add significant value' },
                { icon: TrendingUp, title: 'Negotiate Wisely', desc: 'Most offers have room for negotiation. Research market rates first' },
                { icon: Building2, title: 'Culture Matters', desc: 'Compensation is important, but company culture impacts long-term satisfaction' },
              ].map((tip, i) => (
                <motion.div
                  key={tip.title}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.3 + i * 0.1 }}
                  className="p-4 bg-white/60 backdrop-blur rounded-xl border border-white/50"
                >
                  <tip.icon className="w-6 h-6 text-blue-600 mb-2" />
                  <h3 className="font-semibold text-slate-900 mb-1">{tip.title}</h3>
                  <p className="text-sm text-slate-600">{tip.desc}</p>
                </motion.div>
              ))}
            </div>
          </div>
        </RevealSection>
      </div>
    </div>
  );
}