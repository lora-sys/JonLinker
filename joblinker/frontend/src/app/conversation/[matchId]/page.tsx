'use client';

import { useEffect, useState } from 'react';
import { use } from 'react';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { ChatWindow } from '@/components/chat/ChatWindow';
import { InterviewCard } from '@/components/interview/InterviewCard';
import { OfferCard } from '@/components/offer/OfferCard';
import { apiClient } from '@/lib/api_client';
import { getAuthToken } from '@/lib/api-utils';
import type { Match, Interview, Offer } from '@/types';

interface PageProps {
  params: Promise<{ matchId: string }>;
}

export default function ConversationPage({ params }: PageProps) {
  const { matchId } = use(params);
  const [match, setMatch] = useState<Match | null>(null);
  const [interview, setInterview] = useState<Interview | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const token = getAuthToken();

        // Fetch match details
        const matchHeaders: HeadersInit = token ? { Authorization: `Bearer ${token}` } : {};
        const matchRes = await fetch(`/api/matches/${matchId}`, { headers: matchHeaders });
        if (matchRes.ok) {
          const matchData = await matchRes.json();
          setMatch(matchData);

          // If match status is interview_scheduled, fetch interview
          if (matchData.status === 'interview_scheduled') {
            try {
              const interviewRes = await fetch(`/api/interviews/${matchId}`, { headers: matchHeaders });
              if (interviewRes.ok) {
                const interviewData = await interviewRes.json();
                setInterview(interviewData);
              }
            } catch {
              // No interview yet
            }
          }

          // If match status is offer_sent, fetch offer
          if (matchData.status === 'offer_sent' || matchData.status === 'offered') {
            try {
              const offerRes = await fetch(`/api/offers/${matchId}`, { headers: matchHeaders });
              if (offerRes.ok) {
                const offerData = await offerRes.json();
                setOffer(offerData);
              }
            } catch {
              // No offer yet
            }
          }
        }
      } catch (err) {
        console.error('Failed to fetch conversation data:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [matchId]);

  const handleInterviewConfirm = async (id: string) => {
    try {
      const token = getAuthToken();
      await fetch(`/api/interviews/${matchId}/confirm`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      // Refresh interview data
      const interviewRes = await fetch(`/api/interviews/${matchId}`);
      if (interviewRes.ok) {
        setInterview(await interviewRes.json());
      }
    } catch (err) {
      console.error('Failed to confirm interview:', err);
    }
  };

  const handleInterviewCancel = async (id: string) => {
    try {
      const token = getAuthToken();
      await fetch(`/api/interviews/${matchId}/cancel`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      // Refresh interview data
      const interviewRes = await fetch(`/api/interviews/${matchId}`);
      if (interviewRes.ok) {
        setInterview(await interviewRes.json());
      }
    } catch (err) {
      console.error('Failed to cancel interview:', err);
    }
  };

  const handleOfferAccept = async (id: string) => {
    try {
      const token = getAuthToken();
      await fetch(`/api/offers/${matchId}/accept`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      // Refresh offer data
      const offerRes = await fetch(`/api/offers/${matchId}`);
      if (offerRes.ok) {
        setOffer(await offerRes.json());
      }
    } catch (err) {
      console.error('Failed to accept offer:', err);
    }
  };

  const handleOfferDecline = async (id: string) => {
    try {
      const token = getAuthToken();
      await fetch(`/api/offers/${matchId}/decline`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      // Refresh offer data
      const offerRes = await fetch(`/api/offers/${matchId}`);
      if (offerRes.ok) {
        setOffer(await offerRes.json());
      }
    } catch (err) {
      console.error('Failed to decline offer:', err);
    }
  };

  const handleMilestoneConfirm = async () => {
    const token = getAuthToken();
    if (!token) return;

    try {
      await fetch(`/api/matches/${matchId}/confirm`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      alert('Milestone confirmed!');
    } catch (err) {
      console.error('Failed to confirm milestone:', err);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <header className="bg-white/80 backdrop-blur-xl border-b border-white/20 shadow-sm rounded-xl mb-6">
          <div className="px-6 py-4 flex items-center gap-4">
            <Link href="/matches" className="text-slate-600 hover:text-slate-900 cursor-pointer">
              <ArrowLeft className="w-6 h-6" />
            </Link>
            <h1 className="text-xl font-bold text-slate-900">Conversation</h1>
          </div>
        </header>

        <main className="px-4 sm:px-6 lg:px-8">
          {/* Interview Card */}
          {interview && (
            <div className="mb-6">
              <InterviewCard
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                interview={interview as any}
                onConfirm={handleInterviewConfirm}
                onCancel={handleInterviewCancel}
              />
            </div>
          )}

          {/* Offer Card */}
          {offer && (
            <div className="mb-6">
              <OfferCard
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                offer={offer as any}
                onRespond={handleOfferAccept}
              />
            </div>
          )}

          <div className="h-[calc(100vh-8rem)]">
            <ChatWindow matchId={matchId} onConfirmMilestone={handleMilestoneConfirm} />
          </div>
        </main>
      </div>
    </div>
  );
}