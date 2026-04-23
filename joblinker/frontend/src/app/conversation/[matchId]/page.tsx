'use client';

import { use } from 'react';
import Link from 'next/link';
import { ChatWindow } from '@/components/chat/ChatWindow';

interface PageProps {
  params: Promise<{ matchId: string }>;
}

export default function ConversationPage({ params }: PageProps) {
  const { matchId } = use(params);

  const handleMilestoneConfirm = async () => {
    // Confirm the match transition to negotiating
    const token = localStorage.getItem('joblinker-auth');
    if (!token) return;
    const parsed = JSON.parse(token);
    const authToken = parsed.state?.token || parsed.token;

    try {
      await fetch(`/api/matches/${matchId}/confirm`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });
      alert('Milestone confirmed!');
    } catch (err) {
      console.error('Failed to confirm milestone:', err);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <Link href="/matches" className="text-gray-600 hover:text-gray-900">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </Link>
          <h1 className="text-xl font-bold text-gray-900">Conversation</h1>
        </div>
      </header>

      <main className="max-w-4xl mx-auto p-4">
        <div className="h-[calc(100vh-8rem)]">
          <ChatWindow matchId={matchId} onConfirmMilestone={handleMilestoneConfirm} />
        </div>
      </main>
    </div>
  );
}